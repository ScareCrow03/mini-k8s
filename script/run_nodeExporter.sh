#!/bin/bash

# node_exporter可以在每个节点上运行，用于收集节点的监控数据
# 可以部署成一个系统服务，也可以直接容器启动
docker run -d --name minik8s-node-exporter -p 9100:9100 --net="host" --pid="host" -v "/:/host:ro,rslave" quay.io/prometheus/node-exporter:v1.8.0