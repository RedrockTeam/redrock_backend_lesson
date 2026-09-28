# Nginx kubernetes

# Nginx kubernetes

Nginx 一款高效的HTTP和反向代理服务器

## 安装nginx

docker run一个来学习把

```Bash
bash
```

docker run --name nginx -p 8080:80 nginx

```Bash
开启了一个在本地端口8080的nginx服务 通过docker exec 进入容器内部对文件进行查看 先看看版本
root@b3ea444ddda8:/# nginx -v
nginx version: nginx/1.23.3
```

可以看出来 版本为1.23.3

## 配置文件

对于nginx的学习主要是研究他的配置文件

```Thrift
root@b3ea444ddda8:/etc/nginx# cat nginx.conf
user  nginx;
worker_processes  auto;
error_log  /var/log/nginx/error.log notice;
pid        /var/run/nginx.pid;
events {
    worker_connections  1024;
}
http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';
    access_log  /var/log/nginx/access.log  main;
    sendfile        on;
    #tcp_nopush     on;
    keepalive_timeout  65;
    #gzip  on;
    include /etc/nginx/conf.d/*.conf;
}
```

他的组成主要有 main event http server 和location 组成

### main

```Thrift
user  nginx;
worker_processes  auto;
error_log  /var/log/nginx/error.log notice;
pid        /var/run/nginx.pid
```

这一部分 为全局配置` user`指定了用户组 ` worker_processes` 就是`worker`数目 默认是和`cpu`核心数一样的数目的`worker` 同时有一个`master`负载调配`work`，当将其改成指定的值的时候。可以实现资源的限制 。`err_log`指定了错误的输出文件夹，`pid` 指定了进程id的存储位置，

### event

```Thrift
events {
    worker_connections  1024;
}
```

`evnet `指定了工作模式以及最大连接，这里没有工作模式 其中工作模式分为 `select poll kqueue epoll rtsig and dev/poll`

### Http

```Thrift
http {
    include       /etc/nginx/mime.types; #支持的mime类型
    default_type  application/octet-stream;
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';
                      ## 日志格式
    access_log  /var/log/nginx/access.log  main; ## 访问日志位置
    sendfile        on; ## 
    #tcp_nopush     on;
    keepalive_timeout  65;  ## 最大保持连接时间
    #gzip  on; ## 是否开启gzip压缩
    include /etc/nginx/conf.d/*.conf; ## 包含的其他的配置文件
}
```

`include ` 指定了其他配置文件的位置和`mime`类型文件夹，比如 `default`的配置文件位置` include /etc/nginx/conf.d/default.conf;`就是默认打开地址的欢迎界面 `default_type` 为默认的类型为二进制文件流 `log_format` 访问日志输出格式         `docker `里面写到了标准输出我就不去复制了 ` sendfile` 开启高效文件传输 `keepalive_timeout` 客户端连接保持活动的时间 `gzip` 文件压缩用于提高传输效率

### Server

这里打开`/etc/nginx/conf.d/default.conf`的默认欢迎界面

```Thrift
root@b3ea444ddda8:/etc/nginx/conf.d# cat default.conf
server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;
    #access_log  /var/log/nginx/host.access.log  main;
    location / {
        root   /usr/share/nginx/html;
        index  index.html index.htm;
    }
    #error_page  404              /404.html;
    # redirect server error pages to the static page /50x.html
    #
    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }
    # proxy the PHP scripts to Apache listening on 127.0.0.1:80
    #
    #location ~ \.php$ {
    #    proxy_pass   http://127.0.0.1;
    #}
    # pass the PHP scripts to FastCGI server listening on 127.0.0.1:9000
    #
    #location ~ \.php$ {
    #    root           html;
    #    fastcgi_pass   127.0.0.1:9000;
    #    fastcgi_index  index.php;
    #    fastcgi_param  SCRIPT_FILENAME  /scripts$fastcgi_script_name;
    #    include        fastcgi_params;
    #}
    # deny access to .htaccess files, if Apache's document root
    # concurs with nginx's one
    #
    #location ~ /\.ht {
    #    deny  all;
    #}
}
```

`server`定义一个虚拟主机 `listen`指定监听的地址 `server_name`指定域名 既当不同的域名以同一端口访问时的辨别 `err_page`当访问出错时展现的页面 `root`指定静态文件的地址 `index`当以`http://localhost/`访问时 默认的后缀

### 配置文件检查

`nginx -t `检查配置文件正确性

```Thrift
root@b3ea444ddda8:/etc/nginx/conf.d# nginx -t
nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
nginx: configuration file /etc/nginx/nginx.conf test is successful
```

### 负载均衡

```Thrift
upstream mysvr { 
    server 192.168.10.121:3333;
    server 192.168.10.122:3333;
}
server {
    ....
    location  ~*^.+$ {         
        proxy_pass  http://mysvr;  #请求转向mysvr 定义的服务器列表         
    }
}
//负载均衡策略
//热备
//轮询
//加权轮询
//ip_hash
//
```

# kubernetes

kubernetes 也简称k8s 因为他中间有8个字母 k8s 是一个开源的容器编排引擎，用来对容器化应用进行自动化部署、 扩缩和管理。,由Google开发。 对于应用部署，经历了三个阶段  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=ZmJhODRmNzExMDVjODQzYzEzYTI5NDQzNTMxMjIyMTRfNjdlZmRlMzVjN2Y3NWNkMzVjYTQxY2NlMTdmNzBjYjBfSUQ6NzI1NTYwNzM1MjMwNTY2NDAwNF8xNzgxNzk1MDkwOjE3ODE3OTg2OTBfVjM) 物理机部署：多个应用直接部署到物理机上，共同分配物理机资源，无法做到应用隔离，发生bug会影响到其他应用 虚拟机部署：在一个物理机上通过虚拟化技术，隔离出来独立的虚拟机 常见的有KVM虚拟化等,可以限制虚拟机资源 容器部署：对于容器化来说，其类似于VM但是具备更松的隔离策略，基于进程级，共享操作系统，更加轻量化，能够更好的分配资源

## docker 与k8s

Docker 是用于构建、分发、运行（Build, Ship and Run）容器的平台和工具 而 K8s 实际上是一个使用 Docker 容器等 进行编排的系统，主要围绕 pods 进行工作

## 特性

* **服务发现和负载均衡** Kubernetes 可以使用 DNS 名称或自己的 IP 地址来暴露容器。 可以实现负载均衡。
* **存储编排** Kubernetes 允许你自动挂载你选择的存储系统，例如本地存储、公共云（NFS，Ceph，GFS）提供商等。
* **自动部署和回滚** 可以使用 Kubernetes 更新应用，对于出现问题自动回滚，平滑升级
* **自动完成装箱计算**
* **自我修复** Kubernetes 将重新启动失败的容器、替换容器、杀死不响应用户定义的运行状况检查的容器
* **密钥与配置管理** Kubernetes 允许你存储和管理敏感信息，例如密码、OAuth 令牌和 SSH 密钥。

## 架构

 ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=NzBlZDFmZjg4NDJlZjhiN2RhNWFlMzI2N2Q2Njk5MjJfOWQ5MTUzMzA3ZTUyYWEwODhhYmE1MDAzZjc4ZjYwZmNfSUQ6NzI1NTYwNjkyODkwMTI1OTI2N18xNzgxNzk1MDkwOjE3ODE3OTg2OTBfVjM) **kubenetes**由两种节点构成

### master

提供集群的管理控制中心。

#### kube-apiserver

`kube-apiserver`用于暴露Kubernetes API。任何的资源请求/调用操作都是通过kube-apiserver提供的接口进行。

#### etcd

`etcd`是Kubernetes提供默认的存储系统，保存所有集群数据，使用时需要为etcd数据提供备份计划。

#### kube-controller-manager

`kube-controller-manager`运行管理控制器，负责管理控制器进程，通过监视控制器保障集群工作状态

#### kube-scheduler

`kube-scheduler` 负责进行调度，为新建的pod选取一个合适的node进行调度

### node

#### kubelet

`kubelet` 监视node节点上的资源和服务状态并汇报给apiserver；跟容器引擎交互实现容器的生命周期管理

#### kube-proxy

* 是实现负载均衡的重要组件，它将service的请求转发到我们后端的pod上
* 为集群IP提供整个集群的DNS，让他们能在集群内都能被访问到

#### Container Runtime

容器运行环境是负责运行容器的软件。比如`docker`

### 核心概念

#### Pod  docker run

Pod是Kubernetes创建或部署的最小/最简单的基本单位， 一个Pod 代表集群上正在运行的一个进程，可以把Pod理解成豌豆荚，而同一Pod内的每个容器是一颗颗豌豆

#### 工作资源

控制器 用来控制Pod 对于控制器来说，主要有

* **Deployment** 无状态应用部署。Deployment 的作用是管理和控制Pod和Replicaset， 管控它们运行在用户期望的状态中 Replicaset: 确保预期的Pod副本数量
* **Daemonset** 确保所有节点运行同一类Pod，保证每个节点上都有一个此类Pod运行，通常用于实现系统级后台任务，比如资源监控
* **Statefulset** 有状态应用部署,比如需要一个稳定的网络标识符，稳定持久的存储，有序优雅平滑的部署，扩缩和更新
* **Job** 一次性执行任务，类似Linux中的job
* **Cronjob** 周期性任务，像Linux的Crontab一样。

#### 服务

* Pod存在生命周期，它自己虽然有IP，但是却无法提供一个固定的访问接口给客户端，属于用完即抛的资源
* 因此需要使用Service来提供服务发现，也可以提供一个固定的访问接口，还能实现负载均衡（一个Service对应多个Pod）

#### Label

* 通过label，我们给资源对象设置一个标签，允许灵活地分类我们的资源
* 在需要使用这部分资源的时候，在label选择器中加上我们设置的标签，就能精准分配我们的资源

#### Namespace

* k8s中将资源用Namespace逻辑上地隔离开，让我们每个用户只需要关心自己名称空间下的资源
* 主要用于分组，同时各个分组还能共享整个集群的资源，我们可以在部署时指定部署到的名称空间 还有很多很多很多，诸如存储，配置，安全，策略，调度，集群管理 这里只说这几种基础的，在此基础上通过一个service暴露我们的服务

### 安装

这里通过kubeadm进行安装可以参照官网示例\[安装 kubeadm | Kubernetes\] [安装 kubeadm](https://kubernetes.io/zh-cn/docs/setup/production-environment/tools/kubeadm/install-kubeadm/) 在配置好前面的环境后，通过

```Thrift
kubeadm init --pod-network-cidr=192.168.0.0/16
```

初始化一个集群 pod的ip地址块为`192.168.0.0/16`

```Thrift
Your Kubernetes control-plane has initialized successfully!
To start using your cluster, you need to run the following as a regular user:
//启动集群
  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config
Alternatively, if you are the root user, you can run:
  export KUBECONFIG=/etc/kubernetes/admin.conf
You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  https://kubernetes.io/docs/concepts/cluster-administration/addons/
Then you can join any number of worker nodes by running the following on each as root:
//node节点加入命令
kubeadm join xxx.xxx.xxx.xxx:6443 --token 15kk8k.t7ls5hhmp2dkxoku \
        --discovery-token-ca-cert-hash sha256:cc517c2ae48ff4fe690c0e25e4033d48def37284d7ea1a6d5b4d13c7344b71cb 
```

现在已经可以查看集群状态了

```Thrift
root@sianao:~# kubectl get node
NAME     STATUS     ROLES           AGE   VERSION
sianao   NotReady   control-plane   98s   v1.27.3
```

为NotReady

```Thrift
container runtime network not ready: 
NetworkReady=false reason:NetworkPluginNotReady message:
Network plugin returns error: cni plugin not initialized
```

部署网络插件 calico

```Thrift
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.1/manifests/tigera-operator.yaml
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.1/manifests/custom-resources.yaml
```

```Thrift
root@sianao:~# kubectl get node
NAME     STATUS   ROLES           AGE     VERSION
sianao   Ready    control-plane   7m43s   v1.27.3
```

做的是单节点去除掉node的污点

```Thrift
kubectl taint nodes --all node-role.kubernetes.io/control-plane-
```

现在已经基本做好了

### kubectl

基础语法

```Thrift
kubectl [command] [TYPE] [NAME] [flags]
#command：指定对资源的操作
#TYPE：指定资源的类型
#NAME：资源名
#flag：指定一些参数
```

常用命令

```Thrift
kubectl apply 
kubectl create
kubectl delete 
kubectl get 
kubectl describe 
```

### deployment

* 在metadata块配置一些元数据，比如deployment名和标签
* 在spec块中配置副本数，标签选择器
* 主要在template块配置你所期望的pod模板，在应用deployment后，pod就会以template中所写的来部署
* container下可以有多个容器模板，这里是一个数组的格式，用-来标识多个数组元素

#### replicas

```Thrift
root@sianao:~# cat nginx.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
```

```Thrift
root@sianao:~# kubectl apply -f nginx.yaml 
deployment.apps/nginx created
```

```Thrift
root@sianao:~# kubectl get pod
NAME                     READY   STATUS    RESTARTS   AGE
nginx-6c449797b9-595g7   1/1     Running   0          13s
nginx-6c449797b9-6ksqg   1/1     Running   0          13s
nginx-6c449797b9-tjv9p   1/1     Running   0          13s
```

```Thrift
root@sianao:~# kubectl delete  -f nginx.yaml 
deployment.apps "nginx" deleted
```

#### StatefulSet

```Thrift
apiVersion: v1
kind: Service
metadata:
  name: mysql
spec:
  ports:
  - port: 3306
  selector:
    app: mysql
  clusterIP: None
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql
spec:
  selector:
    matchLabels:
      app: mysql
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: mysql
    spec:
      containers:
      - image: mysql:5.6
        name: mysql
        ports:
        - containerPort: 3306
          name: mysql
        volumeMounts:
        - name: mysql-persistent-storage
          mountPath: /var/lib/mysql
      volumes:
      - name: mysql-persistent-storage
        persistentVolumeClaim:
          claimName: mysql-pv-claim
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: mysql-pv-volume
  labels:
    type: local
spec:
  storageClassName: manual
  capacity:
    storage: 20Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/mnt/data"
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-pv-claim
spec:
  storageClassName: manual
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi
```

```Thrift
root@sianao:~# kubectl apply -f mysql.yaml 
service/mysql created
deployment.apps/mysql created
persistentvolume/mysql-pv-volume created
persistentvolumeclaim/mysql-pv-claim created        
```

对于暴露一个服务而言，我们可以采用 NodePort LoadBalancer Ingress 这里说使用的NodePort 对其他的感兴趣的可以自行了解

```Thrift
apiVersion: v1
kind: Service
metadata:
  name: nginx
spec:
  selector:
    pod: nginx-server     
  type: NodePort
  ports:
  - port: 80
    targetPort: 80
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  selector:
    matchLabels:
      pod: nginx-server
  template:
    metadata:
      labels:
        pod: nginx-server
    spec:
      containers:
      - image: nginx
        name: nginx
        ports:
        - containerPort: 80
```

```Thrift
root@sianao:~# kubectl get svc
NAME         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
kubernetes   ClusterIP   10.96.0.1        <none>        443/TCP          60m
mysql        NodePort    10.111.251.239   <none>        3306:30007/TCP   13m
nginx        NodePort    10.107.132.128   <none>        80:30195/TCP     3m20s
root@sianao:~# curl localhost:30195
<!DOCTYPE html>
<html>
<head>
# Welcome to nginx!
<style>
html { color-scheme: light dark; }
body { width: 35em; margin: 0 auto;
font-family: Tahoma, Verdana, Arial, sans-serif; }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>
<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>
<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```

## 作业

* Lv1使用nginx caddy trafik 等实现一个负载均衡（你也可以手搓）(必须做的)
* Lv2  可以使用 minkube  kind  或者 docker  部署一个 k8s
  * 并部署一个 service 通过NodePort暴露  (sevice deployment)
* Lv3 搭建一个Kubenetes集群 可以使用腾讯云按时计费的服务器  部署一个deployment

```YAML
root@master:~# kubectl get deployment nginx -o yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    deployment.kubernetes.io/revision: "1"
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"apps/v1","kind":"Deployment","metadata":{"annotations":{},"name":"nginx","namespace":"default"},"spec":{"selector":{"matchLabels":{"pod":"nginx-server"}},"template":{"metadata":{"labels":{"pod":"nginx-server"}},"spec":{"containers":[{"image":"nginx","name":"nginx","ports":[{"containerPort":80}]}]}}}}
  creationTimestamp: "2024-06-02T08:09:06Z"
  generation: 1
  name: nginx
  namespace: default
  resourceVersion: "14277"
  uid: facc3bf2-f670-4284-bb38-2166a7c5c0e4
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      pod: nginx-server
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        pod: nginx-server
    spec:
      containers:
      - image: nginx
        imagePullPolicy: Always
        name: nginx
        ports:
        - containerPort: 80
          protocol: TCP
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
status:
  availableReplicas: 1
  conditions:
  - lastTransitionTime: "2024-06-02T08:09:18Z"
    lastUpdateTime: "2024-06-02T08:09:18Z"
    message: Deployment has minimum availability.
    reason: MinimumReplicasAvailable
    status: "True"
    type: Available
  - lastTransitionTime: "2024-06-02T08:09:06Z"
    lastUpdateTime: "2024-06-02T08:09:18Z"
    message: ReplicaSet "nginx-6cc89f4b5f" has successfully progressed.
    reason: NewReplicaSetAvailable
    status: "True"
    type: Progressing
  observedGeneration: 1
  readyReplicas: 1
  replicas: 1
  updatedReplicas: 1
```

* Lv4 通过ingress暴露自己的服务  配置文件 能够通过 web访问的地址或者截图

## Refrence

https://k8s.easydoc.net/docs \[Kubernetes 文档 | Kubernetes\](https://kubernetes.io/zh-cn/docs/home/) \[kind – Quick Start (k8s.io)\](https://kind.sigs.k8s.io/docs/user/quick-start/#configuring-your-kind-cluster) \[Welcome! | minikube (k8s.io)\](https://minikube.sigs.k8s.io/docs/)