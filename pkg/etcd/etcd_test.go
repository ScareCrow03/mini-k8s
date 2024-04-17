package etcd

import (
	"mini-k8s/pkg/constant"
	"strconv"
	"testing"
	"time"
)

var st *EtcdStore = nil
var testStrs = []string{"test0", "test1", "test2", "test3"}

func TestMain(m *testing.M) {
	var err error // 特别注意，这里的err必须先声明，否则如果直接下面写:=来初始化两个返回值，会导致st新定义成了一个本函数的局部变量、作用域覆盖！
	st, err = NewEtcdStore(constant.EtcdIpPortInTestEnv)
	if err != nil {
		panic(err)
	}
	st.DelWithPrefix(constant.EtcdTestUriPrefix)
	m.Run()
	st.DelWithPrefix(constant.EtcdTestUriPrefix)
	st.Close()
}

func TestPut(t *testing.T) {
	// 报错逻辑已经在内部声明好
	for i, str := range testStrs {
		err := st.Put(constant.EtcdTestUriPrefix+"/test/hello"+strconv.Itoa(i), []byte(str))
		if err != nil {
			t.Errorf("Put error: %v", err)
		}
	}
}

func TestGet(t *testing.T) {
	for i, str := range testStrs {
		reply, err := st.Get(constant.EtcdTestUriPrefix + "/test/hello" + strconv.Itoa(i))
		if err != nil {
			t.Errorf("Get error: %v", err)
		} else {
			if string(reply.Value) != str {
				t.Errorf("Get error: %v", err)
			}
		}
	}
}

func TestGetPrefix(t *testing.T) {
	reply, err := st.GetWithPrefix(constant.EtcdTestUriPrefix + "/test")
	if err != nil {
		t.Errorf("GetWithPrefix error: %v", err)
	} else {
		if len(reply) != len(testStrs) {
			t.Errorf("GetWithPrefix error: %v", err)
		}
	}
}

func TestDel(t *testing.T) {
	// 只删一个；删的是key不是value！
	err := st.Del(constant.EtcdTestUriPrefix + "/test/hello0")
	// 检查是否删干净了
	if err != nil {
		t.Errorf("Del error: %v", err)
	}
	reply, err := st.Get(constant.EtcdTestUriPrefix + "/test/hello0")

	if err != nil {
		t.Errorf("Get error: %v", err)
	} else {
		if len(reply.Value) != 0 {

			t.Errorf("Del error: %v", err)
		}
	}
}

func TestPrefixDel(t *testing.T) {
	// 全部删掉
	err := st.DelWithPrefix(constant.EtcdTestUriPrefix + "/test")
	if err != nil {
		t.Errorf("DelWithPrefix error: %v", err)
	}
	reply, err := st.GetWithPrefix(constant.EtcdTestUriPrefix + "/test")
	if err != nil {
		t.Errorf("GetWithPrefix error: %v", err)
	} else {
		if len(reply) != 0 {
			t.Errorf("DelWithPrefix error: %v", err)
		}
	}
}

func TestWatch(t *testing.T) {
	key := constant.EtcdTestUriPrefix + "/foo"
	value := []byte("bar")

	// 先注册监听变化（没有key也可以注册）
	cancel, watchChan := st.Watch(key)
	defer cancel() // defer关键字表示延迟调用，当前函数执行完后LIFO执行被标记为defer的语句，常用于释放资源

	go func() { // 另外开一个操作etcd的线程，每隔1s新增、修改、删除这个数据
		time.Sleep(time.Second)

		// Create the key
		_ = st.Put(key, value)

		time.Sleep(time.Second)

		// Update the key
		_ = st.Put(key, []byte("baz"))

		time.Sleep(time.Second)

		// Delete the key
		_ = st.Del(key)

	}()

	// 主线程，轮询一段时间，计数是否能正确收到上述的事件
	opEventsCount := map[int]int{}
	for {
		select { // select语句会阻塞，直到其中的一个case成立；如果同时有多个case成立（例如写了多个恒成立条件），则随机选择一个执行；故这里的for死循环不会导致CPU占用过高
		case reply := <-watchChan: // 获取到一个事件
			{
				opEventsCount[reply.OpType]++
			}

		case <-time.After(5 * time.Second): // 在进入select体后，建立一个新的定时channel，在5s后发送一个事件；如果本select体中的case没有一个其他满足，则5s后该case会被执行；如果在5s内有其他case满足，则这个定时channel会被丢弃（因为它的生命周期就是select体，它们都处于一个for循环体内）
			{ // 这里就是说，如果5s内没有从watchChan中收到任何事件，这里就产生一个事件使case成立，退出监测检验（但是watchChan在最后才关闭，虽然没有继续获取它的返回值）
				goto END
			}
		}
	}
END:
	// 退出监测后，检验计数是否正确
	if opEventsCount[CREATE_OP] != 1 || opEventsCount[UPDATE_OP] != 1 || opEventsCount[DELETE_OP] != 1 {
		t.Errorf("Watch error: %v", opEventsCount)
	}
}
