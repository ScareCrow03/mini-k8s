## httputils和message的使用方式

### httputils

封装了两个函数Post和Get

使用方法是

```
Post(url string, requestBody map[string]interface{})
```

第一个参数url是请求的url字符串

第二个参数是以map[string]{interface}的形式传入请求体，这表示索引为string，值可以为任何类型的映射，用于表示json数据。

当需要传递请求体时，建议用如下方式生成：

```
requestBody := make(map[string]interface{})
requestBody["filename"] = filename //按照键-值方式填写，可以嵌套迭代
requestBody["ispause"] = true
httputils.Post("localhost:8080/", requestbody) //发送请求
```

在这样的约定下，接收方只需要准备一个map[string]interface{}结构对请求体进行绑定，就可以拿到里面的数据

```
var requestBody map[string]interface{}
c.BindJSON(&requestBody)
filepath := requestBody["filename"].(string) //获取时要定义好其类型
```

### message

封装了两个消息队列函数Consume和Publish

+ Publish使用方法

  ```
  Publish(name string, msg []byte)
  ```

  第一个参数是期望在哪一个消息队列中发送消息

  第二个参数是发送消息的具体内容，生成方法如下所示：

  ```
  requestBody := make(map[string]interface{})
  requestBody["filename"] = filename //按照键-值方式填写，可以嵌套迭代
  requestBody["ispause"] = true
  msg, _ := json.Marshal(requestBody)
  message.Publish("testname", json.Marshal(msg)) //进行Marshal即可生成[]byte类型
  ```

+ Consume使用方法

  ```
  Consume(name string, callback func(map[string]interface{}))
  ```

​		使用consume方式，可以无限监听名字为name的消息队列，当产生消息时，则可以使用回调函数callback进行处理，内部已经完成了从消息格式到map的转化，所以参数可以从map中根据名字获取。