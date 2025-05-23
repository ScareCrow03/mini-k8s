# Minik8s 开题报告

### 人员组成

|  姓名  |     学号     |            邮箱             |
| :----: | :----------: | :-------------------------: |
| 郑宇城 | 521010910025 |     iloly10@sjtu.edu.cn     |
| 李瑞涵 | 521021910288 | lrh200219791964@sjtu.edu.cn |
| 谢书洋 | 521021910260 |    ocean_xie@sjtu.edu.cn    |



### 自选功能

MicroService



### 时间安排

- 第一次迭代 2024/04/12 - 2024/05/05
  - 学习 go 语言
  - 学习 k8s 基本功能及设计
  - 实现 Pod 抽象
  - 实现 CNI 功能
  - 使用 CI/CD 进行部署
- 第二次迭代 2024/05/06 - 2024/05/18
  - 实现 Service 抽象
  - 实现 Pod ReplicaSet 抽象
  - 动态伸缩
  - DNS 与转发
  - 容错
  - 多机 minik8s
- 第三次迭代 2024/05/19 - 2024/05/31
  - 实现 MicroService
  - 实现三项个人作业



### 人员分工

- 郑宇城：实现 Pod 抽象，实现 CNI 功能，动态伸缩，多机 minik8s，实现 MicroService，持久化存储
- 李瑞涵：实现 Pod 抽象，实现 Service 抽象，DNS 与转发，使用 CI/CD 进行部署，实现 MicroService，⽀持GPU应⽤
- 谢书洋：实现 Pod 抽象，实现 Pod ReplicaSet 抽象，容错，实现 MicroService，⽇志与监控



### gitee 目录

https://gitee.com/dye-one-s-hair/mini-k8s



### 第一次迭代

#### 已实现功能

- 搭建 minik8s 基本组件的框架，包括 kubectl、api-server、scheduler
- 实现 pod 抽象，pod 内容器通信，创建 pod
- 实现组件之间的通信，使用 http 及消息队列

#### TODO

- pod 相关其他操作，pod 间通信，获取 pod 状态
- kubelet 启动时注册 node
- etcd 存储规范
- 完善 api 接口文档