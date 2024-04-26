package img

// 这是一个项目底层文件
// 这个文件用于操作docker SDK提供的client服务，向docker守护进程发请求，简单提供操作镜像的服务。原生k8s底层直接使用的是gRPC来发送符合CRI规范的请求，这比较复杂，为了简化开发过程考虑使用docker SDK、来屏蔽掉具体的发请求细节。
import (
	"context"
	"io"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/protocol"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// 描述本文件实现的一些方法，用接口的形式
type ImageServiceInterface interface {
	PullImage(imageName string, alwaysPull protocol.ImagePullPolicyType) error
	GetImage(imageName string) (image.Summary, error)
	ListImages() ([]image.Summary, error)
	RemoveImageById(imageId string) error
	RemoveImageByName(imageName string) error
	RemoveImageByPrefixName(imageName string) error
	Close()
}

// struct from k8s and modified
// RemoteImageService is a gRPC implementation of internalapi.ImageManagerService.
// 这个类不会在外部单独创建，而是随着remoteRuntimeService一起创建，然后被封装在里面，外部可以通过remoteRuntimeService的ImgSvc字段访问它的方法
type RemoteImageService struct {
	Timeout     time.Duration
	ImageClient *client.Client
}

// 建立一个新的远程镜像服务，请持有好这个指针
func NewRemoteImageService(timeout time.Duration) *RemoteImageService {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		logger.KError("Create docker client failed in NewRemoteImageService!")
		panic(err)
	}

	// 检查与Docker守护进程的连通性
	_, err = cli.Ping(context.Background())
	if err != nil {
		logger.KError("Failed to connect to Docker daemon!")
		panic(err)
	}

	// 获取Docker守护进程的详细信息
	info, err := cli.Info(context.Background())
	if err != nil {
		logger.KError("Failed to get Docker daemon info!")
		panic(err)
	}

	logger.KInfo("Connected to Docker daemon: %s", info.ID)

	return &RemoteImageService{
		Timeout:     timeout,
		ImageClient: cli,
	}
}

// 简单释放资源；请注意，Close方法只能在某个位置调用一次，请不要重复释放！
func (r *RemoteImageService) Close() {
	r.ImageClient.Close()
	r.ImageClient = nil
}

// 从docker hub上拉取镜像的逻辑；
//
//	经实验，这个方法的确会在本地已经有镜像的情况下，再去到远端拉取一次
func (r *RemoteImageService) PullImage(imageName string, alwaysPull protocol.ImagePullPolicyType) error {

	switch alwaysPull {
	case protocol.AlwaysPull:
		logger.KInfo("Always pull image: %s", imageName)
	case protocol.PullIfNotPresent:
		summary, err := r.GetImage(imageName)
		if err != nil {
			logger.KError("Get image failed in PullImage!")
			return err
		}
		if summary.ID != "" {
			logger.KInfo("Image %s already exists, no need to pull!", imageName)
			return nil
		}
		logger.KInfo("Pull image if not exist: %s", imageName)

	case protocol.NeverPull:
		_, err := r.GetImage(imageName)
		if err == nil {
			logger.KInfo("Image %s already exists, no need to pull!", imageName)
		} else {
			logger.KError("Image %s not exists, but you set NeverPull, so no image will be pulled!", imageName)
		}
		return nil
	}

	logger.KInfo("Start to Pull Image: %s", imageName)
	reader, err := r.ImageClient.ImagePull(context.Background(), imageName, image.PullOptions{})
	if err != nil {
		logger.KError("Pull image failed in PullImage!")
		return err
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)
	return nil
}

// 获取某个指定Name的镜像；逻辑是先查看本地镜像列表，然后再找到对应的镜像是否存在
// 同一时刻在docker本地，一个完整的镜像名（包括仓库名和标签，例如"redis:latest"）在是唯一的。需要注意的是，一个镜像可以有多个标签（标记着版本），因此可能会有多个完整镜像名指向同一个镜像。例如以下两个完整镜像名，"redis:latest"和"redis:3.2"可能指向同一个镜像。
func (r *RemoteImageService) GetImage(imageName string) (image.Summary, error) {
	images, err := r.ListImages()
	if err != nil {
		logger.KError("List images failed in GetImage!")
		return image.Summary{}, err
	}

	for _, image := range images {
		for _, tag := range image.RepoTags { // 检查当前镜像的所有镜像名。上述已经说到一个镜像名至多对应一个Image，所以这里只要找到一个就可以返回了
			if tag == imageName {
				return image, nil
			}
		}
	}

	// 这是找不到镜像的正常情况，而不是出错！
	return image.Summary{}, nil
}

// 列出所有的镜像，每个Summary是一个结构体，包含一些可能需要的信息
func (r *RemoteImageService) ListImages() ([]image.Summary, error) {
	images, err := r.ImageClient.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		logger.KError("List images failed in ListImages!")
	}
	return images, err
}

func (r *RemoteImageService) RemoveImageById(imageId string) error {
	logger.KInfo("Start to Remove ImageId: %v, please wait...", imageId)
	_, err := r.ImageClient.ImageRemove(context.Background(), imageId, image.RemoveOptions{})
	if err != nil {
		logger.KWarning("Failed to remove image %s, maybe no need to remove! Reason: %s", imageId, err)
	} else {
		logger.KInfo("Successfully removed image %s", imageId)
	}
	return nil
}

func (r *RemoteImageService) RemoveImageByName(imageName string) error {
	logger.KInfo("Start to Remove Image With Name: %v, please wait...", imageName)
	summary, err := r.GetImage(imageName)
	if err != nil {
		logger.KError("Get image failed in RemoveImageByName!")
		return err
	}

	if summary.ID == "" {
		logger.KInfo("No image found with name: %v, no need to remove!", imageName)
	}

	return r.RemoveImageById(summary.ID)
}

// 删除所有以它为前缀的镜像名的镜像；比如指定一个redis，那么可以删除redis:latest、redis:3.2等等
func (r *RemoteImageService) RemoveImageByPrefixName(imageName string) error {
	logger.KInfo("Start to Remove Image With Name: %v, please wait...", imageName)
	images, err := r.ImageClient.ImageList(context.Background(), image.ListOptions{}) // 镜像查询是不需要指定All字段的，如果指定了会返回中间层镜像，这是不需要的
	if err != nil {
		logger.KError("List images failed in RemoveImageByName! Reason: %s", err)
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if strings.HasPrefix(tag, imageName) {
				_, err := r.ImageClient.ImageRemove(context.Background(), tag, image.RemoveOptions{Force: true})
				if err != nil {
					logger.KError("Failed to remove image %s: %s\n", tag, err)
				} else {
					logger.KInfo("Successfully removed image %s\n", tag)
				}
			}
		}
	}
	return err
}
