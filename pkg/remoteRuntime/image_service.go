package remoteRuntime

// 这是一个项目底层文件
// 这个文件用于操作docker SDK提供的client服务，向docker守护进程发请求，简单提供操作镜像的服务。原生k8s底层直接使用的是gRPC来发送符合CRI规范的请求，这比较复杂，为了简化开发过程考虑使用docker SDK、来屏蔽掉具体的发请求细节。
import (
	"context"
	"mini-k8s/pkg/logger"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// struct from k8s and modified
// remoteImageService is a gRPC implementation of internalapi.ImageManagerService.
// 这个类不会在外部单独创建，而是随着remoteRuntimeService一起创建，然后被封装在里面，外部可以通过remoteRuntimeService的ImgSvc字段访问它的方法
type remoteImageService struct {
	timeout     time.Duration
	imageClient *client.Client
}

type ImageServiceInterface interface { // 这个服务可以向上提供哪些方法，仅做声明用（视为一个本文件的摘要）；我们把它封装成长得很像CRI规范的接口，但docker SDK的底层调用并不是gRPC
	PullImage(imageName string, alwaysPull bool) error
	GetImage(imageName string) (image.Summary, error)
	ListImages() ([]image.Summary, error)
	Close()
}

// 建立一个新的远程镜像服务，请持有好这个指针
func newRemoteImageService(timeout time.Duration) *remoteImageService {
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

	return &remoteImageService{
		timeout:     timeout,
		imageClient: cli,
	}
}

// 简单释放资源；请注意，Close方法只能在某个位置调用一次，请不要重复释放！
func (r *remoteImageService) Close() {
	r.imageClient.Close()
	r.imageClient = nil
}

// 从docker hub上拉取镜像的逻辑；alwaysPull表示是否强制拉取，即使本地已经有了
func (r *remoteImageService) PullImage(imageName string, alwaysPull bool) error {
	logger.KInfo("Start to Pull Image: %v, alwaysPull: %v, please wait...", imageName, alwaysPull)
	reader, err := r.imageClient.ImagePull(context.Background(), imageName, image.PullOptions{All: alwaysPull})
	if err != nil {
		logger.KError("Pull image failed in PullImage!")
		return err
	}

	defer reader.Close()
	return nil
}

// 获取某个指定key的镜像；逻辑是先查看本地镜像列表，然后再找到对应的镜像是否存在
func (r *remoteImageService) GetImage(imageName string) (image.Summary, error) {
	images, err := r.ListImages()
	if err != nil {
		logger.KError("List images failed in GetImage!")
		return image.Summary{}, err
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageName {
				return image, nil
			}
		}
	}

	// 这是找不到镜像的正常情况，而不是出错！
	return image.Summary{}, nil
}

// 列出所有的镜像，每个Summary是一个结构体，包含一些可能需要的信息
func (r *remoteImageService) ListImages() ([]image.Summary, error) {
	images, err := r.imageClient.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		logger.KError("List images failed in ListImages!")
	}
	return images, err
}

func (r *remoteImageService) RemoveImage(imageName string) error {
	_, err := r.imageClient.ImageRemove(context.Background(), imageName, image.RemoveOptions{})
	if err != nil {
		logger.KError("Remove image failed in RemoveImage!")
	}
	return err
}
