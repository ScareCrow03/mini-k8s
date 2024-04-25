package etcd

// 这个文件实现了一个简单的etcd客户端包装EtcdStore，提供了创建、释放、Get、Put、Del、Watch、GetWithPrefix、WatchWithPrefix、DelWithPrefix的方法
import (
	"context"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/logger"

	etcd_v3 "go.etcd.io/etcd/client/v3"
)

// CompareAndSwap比较版本、实现乐观并发控制的事不放在这里做，它是应用逻辑，如果需要、放在apiServer里做

const ( // 用于WatchReply的opType字段，区分发生了什么变化
	CREATE_OP = 0
	UPDATE_OP = 1
	DELETE_OP = 2
)

type EtcdStore struct {
	client *etcd_v3.Client // 私有字段，外部应该调用被包装后的成员方法
}

type GetReply struct {
	Key   string
	Value []byte
	// 待补充更多需要的字段
}

type WatchReply struct {
	Key             string // 在通知时也许要知道是哪个key发生了变化
	Value           []byte
	OpType          int   // 0: create, 1: update, 2: delete
	ResourceVersion int64 // 这个字段在etcd内部就维护好为ModRevision；它后续会被映射到应用层的metadata.resourceVersion
	CreateVersion   int64 // 用于记录创建时的版本号，是etcd内部维护的CreateRevision
	// 待补充更多需要的字段
}

// 一个工厂方法，接收Etcd服务器的IP:port信息（因为是etcd集群，允许接收多个），返回一个EtcdStore对象
func NewEtcdStore(endpoints []string) (*EtcdStore, error) {
	logger.KInfo("init one etcd client, endpoints: %v", endpoints)
	cfg := etcd_v3.Config{
		Endpoints:   endpoints,            // etcd集群每个端点的IP:port
		DialTimeout: constant.EtcdTimeout, // 5s连接超时，则创建client失败
		// 也许需要其他配置字段，例如Username、Password等
	}

	client, err := etcd_v3.New(cfg)
	if err != nil {
		logger.KError("etcd create lient error: %v", err)
		return nil, err
	}

	return &EtcdStore{client: client}, nil
}

// 释放资源；它必须是对象方法，因为可能调用多次NewEtcdStore
func (st *EtcdStore) Close() error {
	logger.KInfo("close one etcd client, endpoints: %v", st.client.Endpoints())
	return st.client.Close()
}

// EtcdStore的成员方法Get、Put、Del、Watch、GetWithPrefix、WatchWithPrefix、DelWithPrefix
func (st *EtcdStore) Get(key string) (GetReply, error) {
	res, err := st.client.Get(context.Background(), key)
	if err != nil { // 出错
		logger.KError("etcd Get error: %v", err)
		return GetReply{}, err
	}

	if len(res.Kvs) == 0 { // 不存在；这样Value是一个空的[]byte，OK
		return GetReply{}, nil
	}

	reply := GetReply{ // 这里res的Kvs是一个数组，但我们调用了GET，它至多有一个元素
		Key:   string(res.Kvs[0].Key),
		Value: res.Kvs[0].Value,
	}
	return reply, nil
}

func (st *EtcdStore) Put(key string, value []byte) error {
	_, err := st.client.Put(context.Background(), key, string(value))
	if err != nil {
		logger.KError("etcd Put error: %v", err)
	}
	return err
}

func (st *EtcdStore) Del(key string) error {
	_, err := st.client.Delete(context.Background(), key)
	if err != nil {
		logger.KError("etcd Del error: %v", err)
	}
	return err
}

func (st *EtcdStore) Watch(key string) (context.CancelFunc, <-chan WatchReply) {
	// 建立一个可供取消的上下文，在外部调用这个cancel后，watcher会停止监听，也一并关闭watchChan对外通道
	ctx, cancel := context.WithCancel(context.Background())
	watchChan := make(chan WatchReply)   // 建立一个通道，用于传递watch的结果
	watcher := st.client.Watch(ctx, key) // 通过etcd的watch方法、建立一个watcher；

	go func() { // 开启一个goroutine、是轻量级线程，用于持续监听结果（开启后它就在后台运行）
		for res := range watcher { // 循环从watcher中取出WatchResponse
			for _, ev := range res.Events { // 查看每一个变化事件；每当value发生变化，就会新增一个事件（为什么要有循环？因为可能监听多个key，那么一个revison中有多个value发生变化，它们被包含在一个WatchResponse里）
				var opType int // 判断事件类型
				switch ev.Type {
				case etcd_v3.EventTypePut:
					if ev.IsCreate() {
						opType = CREATE_OP
					} else {
						opType = UPDATE_OP
					}
				case etcd_v3.EventTypeDelete:
					opType = DELETE_OP
				}
				// 将结果包装好，传递给watchChan
				watchChan <- WatchReply{
					Key:             string(ev.Kv.Key),
					Value:           ev.Kv.Value,
					OpType:          opType,
					ResourceVersion: ev.Kv.ModRevision,
					CreateVersion:   ev.Kv.CreateRevision,
				}
			}
		}
		close(watchChan) // 当watcher关闭时，也关闭对外传递结果的watchChan通道
	}()

	return cancel, watchChan
}

// 前缀操作
func (st *EtcdStore) GetWithPrefix(prefix string) ([]GetReply, error) {
	resp, err := st.client.Get(context.Background(), prefix, etcd_v3.WithPrefix())
	if err != nil {
		logger.KError("etcd GetPrefix error: %v", err)
		return nil, err
	}

	var getReplies []GetReply
	for _, kv := range resp.Kvs { // 遍历每一个结果，做包装
		getReplies = append(getReplies, GetReply{Key: string(kv.Key), Value: kv.Value})
	}

	return getReplies, nil
}

func (st *EtcdStore) WatchWithPrefix(prefix string) (context.CancelFunc, <-chan WatchReply) { // 结构与上述的Watch方法类似
	ctx, cancel := context.WithCancel(context.Background())
	watchChan := make(chan WatchReply)
	watcher := st.client.Watch(ctx, prefix, etcd_v3.WithPrefix())

	go func() {
		for res := range watcher {
			for _, ev := range res.Events {
				var opType int // 判断事件类型
				switch ev.Type {
				case etcd_v3.EventTypePut:
					if ev.IsCreate() {
						opType = CREATE_OP
					} else {
						opType = UPDATE_OP
					}
				case etcd_v3.EventTypeDelete:
					opType = DELETE_OP
				}
				// 将结果包装好，传递给watchChan
				watchChan <- WatchReply{
					Key:             string(ev.Kv.Key),
					Value:           ev.Kv.Value,
					OpType:          opType,
					ResourceVersion: ev.Kv.ModRevision,
					CreateVersion:   ev.Kv.CreateRevision,
				}
			}
		}
		close(watchChan) // 当watcher关闭时，也关闭对外传递结果的watchChan通道
	}()

	return cancel, watchChan
}

func (st *EtcdStore) DelWithPrefix(prefix string) error { // 传递""则等效于DelAll
	_, err := st.client.Delete(context.Background(), prefix, etcd_v3.WithPrefix())
	if err != nil {
		logger.KError("etcd DelPrefix error: %v", err)
	}
	return err
}
