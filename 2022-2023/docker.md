# docker

# docker

# **docker简介**

## **docker背景**

一款产品从开发到上线，从操作系统，到运行环境，再到应用配置。作为开发+运维之间的协作我们需要关心很多东西，这也是很多互联网公司都不得不面对的问题，特别是各种版本的迭代之后，不同版本环境的兼容，对运维人员都是考验 **Docker**之所以发展如此迅速，也是因为它对此给出了一个标准化的解决方案。 环境配置如此麻烦，换一台机器，就要重来一次，费力费时。很多人想到，能不能从根本上解决问题，软件可以带环境安装? 也就是说，安装的时候，把原始环境一模一样地复制过来。开发人员利用Docker可以消除协作编码时"在我的机器上可正常工作"的问题。

## **docker理念**

Docker是基于Go语言实现的云开源项目。Docker的主要目标是"**Build, Ship\[ and Run Any App,Anywhere**"，也就是通过对应用组件的封装、分发、部署、运行等生命期的管理，使用户的APP (可以是一个WEB应用或数据库应用等等)及其运行环境能够做到"**一次封装，到处运行**"。 Linux容器技术的出现就解决了这样一个问题，而Docker就是在它的基础上发展过来的。将应用运行在Docker容器上面，而Docker容器在任何操作系统上都是一致的，这就实现了跨平台、跨服务器。**只需要一次配置好环境，换到别的机子上就可以一键部署好，大大简化了操作** 解决了运行环境和配置问题的软件容器，方便做持续集成并有助于整体发布的容器虚拟化技术

## **docker作用**

#### **之前的虚拟机技术**

虚拟机\*\*(virtual machine)\*\*就是带环境安装的一种解决方案。 它可以在一种操作系统里面运行另一种作系统，比如在**Windows系统里面运行Linux系统**。 应用程序对此毫无感知，因为虚拟机看上去跟真实系统一模一样； 而对于底层系统来说，虚拟机就是一个普通文件，不需要了就删掉，对其他部分毫无影响。 这类虚拟机完美的运行了另一套系统，能够使应用程序，操作系统和硬件三者之间的逻辑不变。  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=ZDY1MGFhNjcxNDhlOTYyMDIzMzJlNjA4ZDY1ZTJkMTNfYmMwMmYwY2RlMWYxM2U2YzMwZmM4YjQxOGMyMWQ0OGZfSUQ6NzIzMDM2NjY2MDI0MTQ3MzUzOF8xNzgxNzk1MDkzOjE3ODE3OTg2OTNfVjM) 虚拟机的缺点: 1、资源占用多 2、冗余步骤多 3、启动慢

#### **容器虚拟化技术**

由于前面虛拟机存在这些缺点，**Linux** 发展出了另一种虚拟化技术: **Linux 容器**(Linux Containers,缩为LXC)。 **Linux容器不是模拟一个完整的操作系统**，而是对进程进行隔离。有了容器，就可以将软件运行所的所有资源打包到一个隔离的容器中。容器与虚拟机不同，不需要捆绑一整套操作系统，只需要软件工作所需的库资源和设置。系统因此而变得高效轻量并保证部署在任何环境中的软件都能始终如一地运行。.  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=OTJlODQ2MDgzMjk1Y2U0MTRkMTkxZTg5NzE3ZTY5MTdfYzU5ZDY0ZGEwN2QwNTU1NWE0Zjg5ZDI4MWQ0ZTI3ZDRfSUQ6NzIzMDM2NjY1ODUzNDk5ODAyMF8xNzgxNzk1MDkzOjE3ODE3OTg2OTNfVjM) 比较了**Docker**和传统虚拟化方式的不同之处: 1、传统虚拟机技术是虚拟出一套硬件后，在其上运行一个完整操作系统，在该系统上再运行所需应用进程; 2、而容器内的应用进程直接运行于宿主的内核，容器内没有自己的内核，**而且也没有进行硬件虚拟**。因此容器要比传统虚拟机为轻便。 3、每个容器之间互相隔离，每个容器有自己的文件系统，容器之间进程不会相互影响，能区分计算资源。

#### **下载与仓库**

##### **1、官网**

docker官网： <https://www.docker.com/>

##### **2、仓库**

Docker Hub官网：<https://hub.docker.com/>

# **docker常用命令**

## **镜像**

在Docker中，镜像是一个包含应用程序及相关依赖库的文件，在Docker容器启动的过程中，它以只读的方式被用于创建容器运行的基础环境。如果把容器理解为应用程序运行的虚拟环境，那么镜像就可以被看作这个环境的持久化副本。通过镜像，我们可以很容易地保存虚拟环境的运行状态，并可以很方便地镜像迁移及反复构造相同的运行环境。

### **镜像管理命令**

#### **获取镜像**

docker pull 与git pull相似 使用docker pull 命令就可以从网上的镜像库下载所需要的镜像。 下面以mysql镜像为例 使用docker pull mysql 就可以将该镜像拉去到本地了。 \[root@VM-20-8-centos docker-test\]# docker pull mysql\nUsing default tag: latest\nlatest: Pulling from library/mysql\n72a69066d2fe: Pull complete\n93619dbc5b36: Pull complete\n99da31dd6142: Pull complete\n626033c43d70: Pull complete\n37d5d7efb64e: Pull complete\nac563158d721: Pull complete\nd2ba16033dad: Pull complete\n688ba7d5c01a: Pull complete\n00e060b6d11d: Pull complete\n1c04857f594f: Pull complete\n4d7cfa90e6ea: Pull complete\ne0431212d27d: Pull complete\nDigest: sha256:e9027fe4d91c0153429607251656806cc784e914937271037f7738bd5b8e7709\nStatus: Downloaded newer image for mysql:latest\ndocker.io/library/mysql:latest 可以看到 **using default tag:latest** 这一行，镜像名还包括了标签名，如果pull的时候不指定tag则默认是最新版本的镜像 同样我们也加上标签，拉取指定版本mysql \[root@VM-20-8-centos docker-test\]# docker pull mysql:5\n5: Pulling from library/mysql\n72a69066d2fe: Already exists\n93619dbc5b36: Already exists\n99da31dd6142: Already exists\n626033c43d70: Already exists\n37d5d7efb64e: Already exists\nac563158d721: Already exists\nd2ba16033dad: Already exists\n0ceb82207cd7: Pull complete\n37f2405cae96: Pull complete\ne2482e017e53: Pull complete\n70deed891d42: Pull complete\nDigest: sha256:f2ad209efe9c67104167fc609cca6973c8422939491c9345270175a300419f94\nStatus: Downloaded newer image for mysql:5\ndocker.io/library/mysql:5

#### **查看镜像**

使用docker images 命令即可查看本地镜像 \[root@VM-20-8-centos docker-test\]# docker images\nREPOSITORY  TAG    IMAGE ID    CREATED     SIZE\ngolang    1.17    276895edf967  4 months ago  941MB\ngolang    latest   276895edf967  4 months ago  941MB\nmysql     5     c20987f18b13  4 months ago  448MB\nmysql     latest   3218b38490ce  4 months ago  516MB 每个镜像都有一个独立的imageID,标识每个镜像

#### **删除镜像**

**docker rmi IMAGEID 或 REPOSITORY:TAG** 后面跟上镜像ID **或者**镜像名:标签 指定镜像删除 \[root@VM-20-8-centos docker-test\]# docker rmi mysql:latest\nUntagged: mysql:latest\nUntagged: mysql@sha256:e9027fe4d91c0153429607251656806cc784e914937271037f7738bd5b8e7709\nDeleted: sha256:3218b38490cec8d31976a40b92e09d61377359eab878db49f025e5d464367f3b\nDeleted: sha256:aa81ca46575069829fe1b3c654d9e8feb43b4373932159fe2cad1ac13524a2f5\nDeleted: sha256:0558823b9fbe967ea6d7174999be3cc9250b3423036370dc1a6888168cbd224d\nDeleted: sha256:a46013db1d31231a0e1bac7eeda5ad4786dea0b1773927b45f92ea352a6d7ff9\nDeleted: sha256:af161a47bb22852e9e3caf39f1dcd590b64bb8fae54315f9c2e7dc35b025e4e3\nDeleted: sha256:feff1495e6982a7e91edc59b96ea74fd80e03674d92c7ec8a502b417268822ff\n\[root@VM-20-8-centos docker-test\]# docker images\nREPOSITORY  TAG    IMAGE ID    CREATED     SIZE\nmysql     5     c20987f18b13  4 months ago  448MB docker的镜像是一个多层结构，镜像的每一层都是在原有层的基础上进行改动的。镜像的分层机制与Git的版本控制原理类似，每层镜像都可以被视为一个提交，并且拥有独立的ID，最顶层的ID就被视为镜像ID。

## **容器**

在Docker中，容器是基于镜像运行的轻量级环境，是Docker封装和管理应用程序或微服务的"集装箱"。运行中的容器读取了镜像中基础的程序和依赖库的代码，并将修改保存在一个沙盒环境中，充分保障了应用程序运行的虚拟性和隔离性。

### **容器管理命令**

#### **创建容器**

**docker create IMAGEID 或 REPOSITORY:TAG** 现在将刚刚拉取的mysql镜像构造成容器,构造成功后返回**容器ID** \[root@VM-20-8-centos docker-test\]# docker create mysql:5\ne9a9b932e468477a7d225f1acae1f39a8ed0f64f0a5d396f016b39070def3e55

#### **查看容器**

**docker ps** \[root@VM-20-8-centos docker-test\]# docker ps\nCONTAINER ID  IMAGE   COMMAND  CREATED  STATUS   PORTS   NAMES\n\[root@VM-20-8-centos docker-test\]# 竟然空空如也，回顾一下ps 命令，ps命令是Process Status的缩写，用来列出当前运行的进程。很明显刚刚我们只是创建了容器，但容器并没有启动，docker ps 列出的是运行中的容器状态。想要查看刚刚创建的容器只要加上 **-a** （也可以写成 **-all** 顾名思义列出所有容器包括未启动的）参数就行了 \[root@VM-20-8-centos docker-test\]# docker ps -a\nCONTAINER ID  IMAGE   COMMAND          CREATED        STATUS   PORTS   NAMES\ne9a9b932e468  mysql:5  "docker-entrypoint.s…"  About a minute ago  Created       elastic_kirch 下面是查看结果解析

>  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=ZTAxN2I5MjNhZGI3ZGQwZWI2ZWY0NTY3ZDVhMGIyYTNfMmNlZDA2NzA5ZGVkMmRiMGVkZDI2M2Q3MmZmYTk5OGFfSUQ6NzIzMDM2NjY1MzIzMjU3ODU2M18xNzgxNzk1MDkzOjE3ODE3OTg2OTNfVjM) 可以看到为了方便显示，docker ps 命令显示结果都是截断的，(docker id和command 指令都显示不全)。需要完整查看 再加上 **--no-trunc** 参数。 \[root@VM-20-8-centos docker-test\]# docker ps -a --no-trunc\nCONTAINER ID                            IMAGE   COMMAND             CREATED     STATUS   PORTS   NAMES\ne9a9b932e468477a7d225f1acae1f39a8ed0f64f0a5d396f016b39070def3e55  mysql:5  "docker-entrypoint.sh mysqld"  2 minutes ago  Created       elastic_kirch

#### **运行容器**

docker run IMAGEID 或 REPOSITORY:TAG docker run 让容器在创建完成后运行起来 Docker容器有以下两种运行态。 **前台交互式**:容器运行在前台，容器运行时直接连接到了程序中运行的程序上，当我们通过命令退出和关闭连接的程序时，就意味着容器停止了运行。在这种场景下，我们通常会通过附加的参数打开容器的伪终端和输入流，这样我们就可以和容器中的程序实现交互了。 虽然容器在前台运行，但我们只是连接到了容器中的应用程序，并没有形成与应用程序完整的交互，要与容器形成完整的交互，还需要使用-t和-i两个参数。在docker run命令中携带\*\*-t或--tty**参数，可以让Docker为这个容器分配一个伪终端，为实现交互提供基础。而带入**-i**或**-- interactive\*\*则打开了交互模式，这时候输入流会一直保持，以便于我们向容器中输入命令或数据。 一般容器启动后，docker会给这个容器随机命名，如果我们想要自己命名容器的话需要加上 **--name** 参数 后面跟的是容器名 \[root@VM-20-8-centos docker-test\]# docker run -it --name mymysql mysql:5\n2022-05-13 06:39:59+00:00 \[Note\] \[Entrypoint\]: Entrypoint script for MySQL Server 5.7.36-1debian10 started.\n2022-05-13 06:39:59+00:00 \[Note\] \[Entrypoint\]: Switching to dedicated user 'mysql'\n2022-05-13 06:39:59+00:00 \[Note\] \[Entrypoint\]: Entrypoint script for MySQL Server 5.7.36-1debian10 started.\n2022-05-13 06:39:59+00:00 \[ERROR\] \[Entrypoint\]: Database is uninitialized and password option is not specified\nYou need to specify one of the following:\n\\- MYSQL_ROOT_PASSWORD\n\\- MYSQL_ALLOW_EMPTY_PASSWORD\n\\- MYSQL_RANDOM_ROOT_PASSWORD 第一次运行mysql容器时，需要指定密码，这里我们添加\*\*-e\*\*参数指定环境遍量就行了 MYSQL_ROOT_PASSWORD 代表root密码 \[root@VM-20-8-centos docker-test\]# docker run -it -e MYSQL_ROOT_PASSWORD="123456" --name mymysql mysql 由于携带 -it 参数，使用了交互模式，控制台上会打印一堆log  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=OTA1MjM4OTdkZjczZmY3ZDg0YWJhOTk3MDlhYTgzNzlfYzc5ZWY0NmIzYzUyYTY2MjQ0NmYyMDA1Y2U0N2Y1NmVfSUQ6NzIzMDM2NjY1MTk5NTMyNDQxN18xNzgxNzk1MDkzOjE3ODE3OTg2OTNfVjM) 到最后显示mysqld 准备好连接时，代表mysql服务器已经成功启动。但此时，mysql服务器进程一直占据终端，不得不重新开启一个终端。 像这种情况，我们可以直接让其在后台运行。 **后台守护式**:容器运行在后台，运行的过程不会占用到当前输入指令的终端，也不会连接到容器内的应用程序上。因为我们无法连接到容器中的程序，所以处于这种运行态的容器必须通过exec指令再次进入容器。 后台运行只需要加上\*\*-d\*\* 参数。 \[root@VM-20-8-centos \~\]# docker run -d -e MYSQL_ROOT_PASSWORD="123456" --name mymysql mysql\n05e81fe9d6a45bfa64c473577aa1fdd8a4e9070912be432930f4cd10fc3cdbdd\n\[root@VM-20-8-centos \~\]# docker ps\nCONTAINER ID  IMAGE   COMMAND          CREATED     STATUS     PORTS         NAMES\n05e81fe9d6a4  mysql   "docker-entrypoint.s…"  2 seconds ago  Up 1 second  3306/tcp, 33060/tcp  mymysql 可以看到后台启动了mysql 并且返回 容器ID。

#### **连接到容器**

**docker attach** docker attach命令可以attach到一个已经运行的容器的[stdin](https://so.csdn.net/so/search?q=stdin&spm=1001.2101.3001.7020)，然后进行命令执行的动作。但是需要注意的是，使用docker attach命令可以衔接到容器中运行的主进程上，并且衔接后就依附到了主进程之中，当我们退出程序时，也会导致容器一起停止。 docker attach 容器名 或者 容器ID **docker exec** 那么有没有其他的办法来实现在容器中进行操作呢? Docker提供了docker exec命令来达成我们的需求。 docker exec命令可以让我们在容器中执行一条全新的指令，并开启新的进程。使用docker exec命令的方法很简单，给出容器名称或容器ID并提供需要执行的命令即可。进入容器后再用之前设置的密码成功连接数据库  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=NzBlZmY1NTY2OTlhZGFmZWRkYzBmNWZjNTQxYWUzMWZfNjRmNmEzZGQxZTg4MDU2NTlkNWZmNThiYzUwMDExY2VfSUQ6NzIzMDM2NjY1NDAyNTIzNjUwOF8xNzgxNzk1MDkzOjE3ODE3OTg2OTNfVjM) mymysql 后 bash 命令，表示进入容器后自动执行bash命令。

## **数据卷**

容器内的文件环境，虽然能够让程序随意操作其中的文件，但所有的读写都是在沙盒环境中进行的。当容器停止运行并被删除时，这个临时记录着文件修改的层就会被一同丢弃。 Docker提出了\*\*数据卷(Data Volume）\*\*的概念。简而言之，数据卷就是一个挂载在容器内文件系统中的文件或目录。在容器中，数据卷和其他的文件或目录看起来别无二致，但是因为数据卷是从外界挂载在容器中的，所以它可以脱离容器的生命周期而独立存在。正是由于数据卷的生命周期并不等同于容器的生命周期，在容器退出乃至删除之后，数据卷仍不会受到影响，会依然存在于Docker之中。所以，如果通过数据卷保存文件，那么这些文件就不会因为容器的终结而消失。

### **指定路径挂载**

参数 -v src:dst，src是宿主机的目录，dst 是容器内的目录。  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=MTZjYjdhZDBlODJhYjAyYWQ4YTQ1OTk1Y2YwOWFmNzNfY2EyMDdhZWFmYzhiYTQyMDkwNGY3NjQzYTI4NTQyZmFfSUQ6NzIzMDM2NjY1OTIxNzgzNDAxMl8xNzgxNzk1MDkzOjE3ODE3OTg2OTNfVjM) 先在 root目录下创建一个volume 目录，下面再创建file文件，随便写点内容进去。后面我们将使用volume目录作为宿主机挂载目录，挂载在docker的\*\*/volume\*\*目录下  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=ZjlhNmY3OTZkMjFiODZmYmVjZjQ1ZGU5ZWIxN2EyNjJfODY1NTZiMDY0MTQyYWM0OTRmZDY2MTM2NTFiYzY3ODBfSUQ6NzIzMDM2NjY1MTUyMjE1NDQ5N18xNzgxNzk1MDkzOjE3ODE3OTg2OTNfVjM) 挂载成功后可以看到，容器内挂载目录下内容与宿主机一致。 现在尝试修改内容  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=OTEyMTc5YjA4MzY2NDI0NjJjNWY4MmJkZGQ3Nzg2ZmNfMDM4NDY5MGVlYzk2YzkzOWEyZGMyNmU1ZGU5NWUyMGVfSUQ6NzIzMDM2NjY1MjU3ODIzNDM3MV8xNzgxNzk1MDkzOjE3ODE3OTg2OTNfVjM) 然后在另一个终端中查看宿主机下挂载的文件信息  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=NjU4NmI2ZGI5ZjJlZDIyNWIzZjMyNGNjOWU1ZTZlOWRfZmI0ODJkZGQ4NmMzOWZiZjNiZmNjODNlNjlkMTkwN2FfSUQ6NzIzMDM2NjY1NTM2NzM2NDYxMl8xNzgxNzk1MDkzOjE3ODE3OTg2OTNfVjM)

#### **具名挂载**

\[root@VM-20-8-centos \~\]# docker run -d -e MYSQL_ROOT_PASSWORD=yes --name mymysql3 -v volume:/volume mysql:5\nc3873912615bb191e41b74003d7507edad6582615f974fae7a8807af4766b63c\n\[root@VM-20-8-centos \~\]# docker volume inspect volume  #查看数据卷信息\n\[\n{\n"CreatedAt": "2022-05-12T17:43:20+08:00",\n"Driver": "local",\n"Labels": null,\n"Mountpoint": "/var/lib/docker/volumes/volume/_data",\n"Name": "volume",\n"Options": null,\n"Scope": "local"\n}\n\] 如果宿主机挂载-v后的src不是目录或者文件的话，将会创建一个数据卷并以该参数命名。通过 docker volume inspect 查看到该数据卷实际挂载点为 **/var/lib/docker/volumes/volume/_data** \[root@VM-20-8-centos \~\]# docker volume ls  #查看数据卷docker\nDRIVER   VOLUME NAME\nlocal   volume\n\[root@VM-20-8-centos \~\]# docker volume rm volume  #删除数据卷 注意需要先删除与其挂载的所有容器\nvolume

## **构建镜像(Dockerfile)**

前面一直说的是将镜像pull到本地后的操作，那么镜像是怎么构建的呢？ Docker提供了一种通过配置文件创建镜像的方式――使用Dockerfile构建镜像。这种方式是将制作镜像的操作全部写入到一个文件中，而docker build命令可以读取这个文件中的所有操作，并根据这些配置创建相应的镜像。Dockerfile让创建镜像的过程变得更加独立和透明，也使整个过程可以轻松容易地往复执行，是构建镜像和进行容器迁移时一个非常优秀的辅助工具。

### **基础指令**

#### **FROM**

FROM指令就是用来指定我们所要构建的镜像是基于哪个镜像建立的。作为Dockerfile必不可少的最基础的指令，FROM指令必须作为第一条指令，也就是说，它应该出现在除注释以外的第一行里。不过，在一个Dockerfile中是允许出现多个FROM指令的，以每个FROM指令为界限，都会生成不同的镜像，但是我们还是推荐将生成不同镜像的命令拆分到不同的Dockerfile中。 FROM <image>:<tag>

#### **MAINTAINER**

MAINTAINER(维护者)指令的用处是提供镜像的作者信息，其使用格式为: MAINTAINER <name>

### **控制指令**

#### **RUN**

在构建镜像的过程中，我们需要在基础镜像中做很多操作，RUN指令就是用来给定这些需要被执行的操作的。由于在构建过程中进行各项操作是不可或缺的过程，可以说RUN指令是Dockerfile中最常用的指令。 Docker会在一个新的镜像层中执行我们给出的命令，并且在执行完成后提交镜像层，用作Dockerfile中下一个指令执行的基础。正是因为Docker使用了镜像分层设计，才使得镜像的提交变得非常"廉价"，即使每次执行RUN指令都创建一个新的镜像层，也不会对镜像的体积、性能等造成巨大的影响。 RUN command param1 param2 ...

#### **WORKDIR**

WORKDIR指令用于切换构建过程中的工作目录，如果我们在使用RUN指令、ADD指令、COPY指令，以及容器运行时才会执行的CMD指令、ENTRYPOINT指令中使用了相对目录，相对目录所基于的就是当前的工作目录。 WORKDIR dir

### **引入指令**

在很多场合下，我们希望将文件加入到即将构建的镜像中，引入指令就能够帮助我们实现这个目的。

#### **ADD**

在构建容器的过程中，可能需要将一些软件源码、配置文件、执行脚本等导入到镜像的构建过程，这时可以使用ADD指令将文件从外部传递到镜像内部。 ADD <src> <dst> ADD指令能够将我们在<src>里指定的**文件、目录**，乃至通过URL指定的远程文件复制到镜像的<dest>目标路径中。我们可以分别指定多个<src>文件或目录，也可以使用通配符指定多个文件或目录。 `ADD`遵守以下规则：

* 该`<src>`路径必须在构建的\*上下文中；\*你不能`ADD ../something /something`，因为 a 的第一步 `docker build`是将上下文目录（和子目录）发送到 docker 守护进程。
* 如果`<src>`是一个 URL 并且`<dest>`不以尾部斜杠结尾，则会从该 URL 下载一个文件并将其复制到`<dest>`.
* 如果`<src>`是一个 URL 并且`<dest>`确实以尾部斜杠结尾，那么文件名是从 URL 推断出来的，文件被下载到 `<dest>/<filename>`. 例如，`ADD http://example.com/foobar /`将创建文件`/foobar`. URL 必须有一个重要的路径，以便在这种情况下可以发现适当的文件名（`http://example.com` 将不起作用）。
* ADD指令还能够自动完成对压缩文件的解压。如果我们提供的<src>是Docker能够识别的压缩文件格式(gzip、bzip2、xz），则其中的内容会解压到<dest>中，而非直接进行复制。Docker识别文件是否是压缩文件，是根据文件内容中的特征码来实现的，如果一个文件并非压缩文件，而只是文件后缀为压缩文件的扩展名，那么Docker并不会尝试解压它。另外，自动解压文件只对本地文件有效，如果<src>给出的是网络路径，Docker不会对文件进行解压。
* 如果`<src>`是目录，则复制目录的全部内容，包括文件系统元数据。
* 如果`<src>`是任何其他类型的文件（不是目录），它将连同其元数据一起单独复制。在这种情况下，如果`<dest>`以尾部斜杠结尾`/`，它将被视为一个目录，其内容`<src>`将写入`<dest>/base(<src>)`.
* 如果`<src>`直接或由于使用通配符指定了多个资源，则`<dest>`必须是目录，并且必须以斜杠结尾`/`。
* # 起初有demo1.txt和demo2.txt两个文件\nADD demo\*.txt /data/ # 对\nADD demo\*.txt /data  # 错
* 如果`<dest>`不以尾部斜杠结尾，它将被视为常规文件（即不是目录），其内容`<src>`将写入`<dest>`.
* ADD demo1.txt /data # data就是文件，不是目录
* 如果`<dest>`不存在，则会创建它及其路径中所有缺失的目录。

> * 不复制目录本身，只复制其内容。
> * ADD /data /var/ # 将宿主机的data目录里的所有内容放到var目录下，但是data目录本身并没有放进去

#### **COPY**

在Dockerfile中还有一种引入文件的方式:使用COPY指令。与ADD指令非常相似。 COPY <src> <dst> COPY指令在源路径<src>、目标路径<dest>，<dest>绝对路径，或相对于WORKDIR。以及文件通配符的使用上，与ADD指令的规则几乎是一致的。 主要的区别就在于COPY指令不能识别网络地址，也不会自动对压缩文件进行解压。在不需要自动解压或者没有网络文件需求的时候，使用COPY指令是一个不错的选择。 `COPY`遵守以下规则：

* 该`<src>`路径必须在构建的\*上下文中；\*你不能`COPY ../something /something`，因为 a 的第一步 `docker build`是将上下文目录（和子目录）发送到 docker 守护进程。
* 如果`<src>`是目录，则复制目录的全部内容，包括文件系统元数据。
* 如果`<src>`是任何其他类型的文件，它将连同其元数据一起单独复制。在这种情况下，如果`<dest>`以尾部斜杠 结尾`/`，它将被视为一个目录，其内容`<src>`将写入`<dest>/base(<src>)`.
* 如果`<src>`直接或由于使用通配符指定了多个资源，则`<dest>`必须是目录，并且必须以斜杠结尾`/`。
* 如果`<dest>`不以尾部斜杠结尾，它将被视为常规文件，其内容`<src>`将写入`<dest>`.
* 如果`<dest>`不存在，则会创建它及其路径中所有缺失的目录。

> 不复制目录本身，只复制其内容。

### **执行指令**

#### **CMD**

Docker容器是为运行单独应用程序而设计的，当Docker容器启动时，实际上是对程序的启动，而**容器是否停止也以程序是否结束为标准**。而在Dockerfile中，就可以通过CMD指令来指定由镜像创建的容器中的主体程序。 CMD \["executable", "paraml","param2",...\] 或\nCMD comrnand paraml param2 ... 需要注意的是，因为容器中只会绑定一个应用程序，所以在Dockerfile中只存在一个CMD指令，如果我们给出了多个CMD指令，之后的指令会覆盖掉之前的指令。 **提示**:不要混淆了CMD指令和RUN指令，它们的区别是很大的。RUN指令是在镜像构建的过程中执行，并将执行结果提交到新的镜像层中;CMD指令在镜像构建的过程中不能执行，它只是配置镜像的默认入口程序。 #如果在docker run 命令时 指定了command 将会替换掉dockerfile中的CMD命令\ndocker run \[OPTIONS\] IMAGE \[COMMAND\] \[ARG...\]

#### **ENTRYPOINT**

镜像所指定的应用程序在容器中运行时，难免需要一些系统服务或其他程序的支持和配合，我们可以在CMD指令中启动这些服务，但这样做会让启动服务的命令与启动主程序的命令杂糅在一起，显得比较混乱。ENTRYPOINT指令就是专门用于主程序启动前的准备工作的。 当ENTRYPOINT指令被指定时，所有的CMD指令或通过docker run等方式指定的应用程序启动命令，不会在容器启动时被直接执行，而是把这些命令当作参数，拼接到ENTRYPOINT指令给出的命令之后，传入ENTRYPOINT指令给出的程序中。 ENTRYPOINT \["executable", "paraml","param2",...\] 或\nENTRYPOINT comrnand paraml param2 ...

### **Build**

docker build \[OPTIONS\] PATH | URL | - 使用docker build 命令就可以通过dockerfile构建出镜像了,可以使用 **-t** 参数 可以指定镜像名，**-f** 参数指定dockerfile 路径。 用下面的dockerfile测试一下上面的cmd和entrypoint。 \[root@VM-20-8-centos \~\]# cat Dockerfile\nFROM centos\nENTRYPOINT \["ls"\]\nCMD \["-l"\] \[root@VM-20-8-centos \~\]# docker build -t test:1.0 .\nSending build context to Docker daemon  38.4kB\nStep 1/3 : FROM centos\n---> 5d0da3dc9764\nStep 2/3 : ENTRYPOINT \["ls"\]\n---> Using cache\n---> 639fd3a7aa2d\nStep 3/3 : CMD \["-l"\]\n---> Running in 15a58e3c2afe\nRemoving intermediate container 15a58e3c2afe\n---> 8ca3f3439709\nSuccessfully built 8ca3f3439709\nSuccessfully tagged test:1.0 \[root@VM-20-8-centos \~\]# docker run -it test:1.0\ntotal 48\nlrwxrwxrwx  1 root root   7 Nov  3  2020 bin -> usr/bin\ndrwxr-xr-x  5 root root  360 May 13 02:23 dev\ndrwxr-xr-x  1 root root 4096 May 13 02:23 etc\ndrwxr-xr-x  2 root root 4096 Nov  3  2020 home\nlrwxrwxrwx  1 root root   7 Nov  3  2020 lib -> usr/lib\nlrwxrwxrwx  1 root root   9 Nov  3  2020 lib64 -> usr/lib64\ndrwx------  2 root root 4096 Sep 15  2021 lost+found\ndrwxr-xr-x  2 root root 4096 Nov  3  2020 media\ndrwxr-xr-x  2 root root 4096 Nov  3  2020 mnt\ndrwxr-xr-x  2 root root 4096 Nov  3  2020 opt\ndr-xr-xr-x 110 root root   0 May 13 02:23 proc\ndr-xr-x---  2 root root 4096 Sep 15  2021 root\ndrwxr-xr-x  11 root root 4096 Sep 15  2021 run\nlrwxrwxrwx  1 root root   8 Nov  3  2020 sbin -> usr/sbin\ndrwxr-xr-x  2 root root 4096 Nov  3  2020 srv\ndr-xr-xr-x  13 root root   0 May 13 02:10 sys\ndrwxrwxrwt  7 root root 4096 Sep 15  2021 tmp\ndrwxr-xr-x  12 root root 4096 Sep 15  2021 usr\ndrwxr-xr-x  20 root root 4096 Sep 15  2021 var \[root@VM-20-8-centos \~\]# docker ps -a\nCONTAINER ID  IMAGE    COMMAND  CREATED        STATUS              PORTS   NAMES\n4ede4b54ce3f  test:1.0  "ls -l"  About a minute ago  Exited (0) About a minute ago       beautiful_noether 上面Dockerfile 通过简单的例子可以看到使用 ENTRYPOINT 后，CMD的命令将会以拼接的形式与ENTRYPOINT一起执行，最后docker 启动程序为 ls -l。 docker run 的时候如果指定了启动程序，相会覆盖dockerfile中CMD命令。ls -a 表示显示目录下所有文件包括隐藏文件 \[root@VM-20-8-centos \~\]# docker run -it test:1.0 -a\n.  .dockerenv  dev  home  lib64    media  opt  root  sbin  sys  usr\n..  bin     etc  lib  lost+found  mnt   proc  run  srv   tmp  var\n\[root@VM-20-8-centos \~\]# docker ps -a\nCONTAINER ID  IMAGE    COMMAND  CREATED     STATUS           PORTS   NAMES\n5772d9a5b3ff  test:1.0  "ls -a"  4 seconds ago  Exited (0) 3 seconds ago       zealous_faraday\n4ede4b54ce3f  test:1.0  "ls -l"  4 minutes ago  Exited (0) 4 minutes ago       beautiful_noether 现在我们来尝试下构建go项目的容器 以下是代码和dockerfile 示例 \[root@VM-20-8-centos docker-test\]# cat main.go\npackage main import (\n"github.com/gin-gonic/gin"\n"net/http"\n) func main(){\napp:=gin.Default()\napp.GET("/", func(c \*gin.Context) {\nc.JSON(http.StatusOK,"hello word")\n})\napp.Run(":80")\n} \[root@VM-20-8-centos docker-test\]# cat Dockerfile\nFROM golang:1.17  AS builder ENV GO111MODULE=on \\  CGO_ENABLED=0 \\  GOOS=linux \\  GOARCH=amd64 ENV GOPROXY=https://goproxy.cn,direct # 移动到工作目录 build\nWORKDIR /build # 复制项目中的 go.mod 和 go.sum文件并下载依赖信息\nCOPY go.mod .\nCOPY go.sum .\nRUN go mod download # 将代码复制到容器中\nCOPY . . # 将代码编译为可执行文件到app\nRUN go build -o app . # 创建一个小镜像\n#FROM scratch\nFROM busybox # 从builder镜像中将拉取 build/app 到当前目录\nCOPY --from=builder /build/app / ENTRYPOINT \["/app"\] ENV 是配置指令，上述中用于配置go环境遍量。 scratch是一个空镜像，只能用于构建其他镜像，比如你要运行一个包含所有依赖的二进制文件，如Golang程序，可以直接使用scratch作为基础镜像。如果希望镜像里可以包含一些常用的Linux工具，busybox镜像是个不错选择，镜像本身只有1.16M，非常便于构建小镜像。 # 使用docker build 构建镜像\n\[root@VM-20-8-centos docker-test\]# docker build -t server .\nSending build context to Docker daemon  10.75kB\nStep 1/13 : FROM golang:1.17  AS builder\n1.17: Pulling from library/golang\nDigest: sha256:c72fa9afc50b3303e8044cf28fb358b48032a548e1825819420fd40155a131cb\nStatus: Downloaded newer image for golang:1.17\n---> 276895edf967\nStep 2/13 : ENV GO111MODULE=on   CGO_ENABLED=0   GOOS=linux   GOARCH=amd64\n---> Using cache\n---> 255745e3f6c7\nStep 3/13 : ENV GOPROXY=https://goproxy.cn,direct\n---> Using cache\n---> 1906f1d8c7c4\nStep 4/13 : WORKDIR /build\n---> Using cache\n---> 30d605cf3d7f\nStep 5/13 : RUN ls\n---> Running in a4f3176aff7f\nRemoving intermediate container a4f3176aff7f\n---> ba32e4d27570\nStep 6/13 : COPY go.mod .\n---> 83c440022445\nStep 7/13 : COPY go.sum .\n---> 9e0346ae13dc\nStep 8/13 : RUN go mod download\n---> Running in f2dd5f5390da\nRemoving intermediate container f2dd5f5390da\n---> 0ed8f1840596\nStep 9/13 : COPY . .\n---> 0398c5eb5bec\nStep 10/13 : RUN go build -o app .\n---> Running in 1c069c41f98b\nRemoving intermediate container 1c069c41f98b\n---> 1b20ebf19a9b\nStep 11/13 : FROM busybox\nlatest: Pulling from library/busybox\n5cc84ad355aa: Pull complete\nDigest: sha256:5acba83a746c7608ed544dc1533b87c737a0b0fb730301639a0179f9344b1678\nStatus: Downloaded newer image for busybox:latest\n---> beae173ccac6\nStep 12/13 : COPY --from=builder /build/app /\n---> 3f7a6c6138fe\nStep 13/13 : ENTRYPOINT \["/app"\]\n---> Running in dec2a9d7d118\nRemoving intermediate container dec2a9d7d118\n---> 58d2cda3f21d\nSuccessfully built 58d2cda3f21d\nSuccessfully tagged server:latest #构建成功后，在刚刚构建的镜像上启动容器\n\[root@VM-20-8-centos docker-test\]# docker run -it server\n\[GIN-debug\] \[WARNING\] Creating an Engine instance with the Logger and Recovery middleware already attached. \[GIN-debug\] \[WARNING\] Running in "debug" mode. Switch to "release" mode in production.\n\\- using env:   export GIN_MODE=release\n\\- using code:  gin.SetMode(gin.ReleaseMode) \[GIN-debug\] GET   /             --> main.main.func1 (3 handlers)\n\[GIN-debug\] \[WARNING\] You trusted all proxies, this is NOT safe. We recommend you to set a value.\nPlease check https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies for details.\n\[GIN-debug\] Listening and serving HTTP on :80 虽然容器启动成功，但因为容器是隔离的状态此时外部仍无法访问到容器。所以在启动的时候还需要配置**端口映射**。将容器的端口与宿主机端口关联起来，这样外部通过访问宿主机某个端口间接访问到容器了。 可以通过 -P 或 -p 参数来指定端口映射。 #docker -p src:dst  src为宿主机端口 dst为容器端口\n\[root@VM-20-8-centos \~\]# docker run -d -p 80:80 server\n0e4e4f831e0432d36ecb8a2194508e4404ef6bf503650dbc75dc2218363e84f1\n\[root@VM-20-8-centos \~\]# docker ps\nCONTAINER ID  IMAGE   COMMAND  CREATED     STATUS     PORTS                NAMES\n0e4e4f831e04  server   "/app"   3 seconds ago  Up 2 seconds  0.0.0.0:80->80/tcp, :::80->80/tcp  recursing_kare #docker -P 参数可以不指定映射关系\n#会随机的将宿主机空闲端口与容器端口绑定，前提是容器内部需要开放端口\n#我们可以在dockerfile中使用配置指令 EXPOSE port 指定开放哪些端口\n#像上述情况我们需要开放80端口 只需在之前的dockerfile中添加 EXPOSE 80 重新构建镜像server2即可\n\[root@VM-20-8-centos docker-test\]# docker run -d -P server2\n9084a097c6f9e13fa221f509c771668ced49e350e59e087e896ccb4bfadc5fc2\n\[root@VM-20-8-centos docker-test\]# docker ps\nCONTAINER ID  IMAGE   COMMAND  CREATED     STATUS     PORTS                   NAMES\n9084a097c6f9  server2  "/app"   3 seconds ago  Up 2 seconds  0.0.0.0:49154->80/tcp, :::49154->80/tcp  busy_chaplygin\n#将宿主机49154端口与容器80端口绑定

## **Link**

刚刚的端口映射其实是外部访问了宿主机，宿主机再去访问容器。那么现在我们要将刚刚的go 示例中连接上mysql容器，容器怎么访问容器呢？ 之前的项目里mysql连接配置的ip一般都是127.0.0.1 ，可是当我们启动go-server容器后，该容器相当于一个实例，如果继续用127.0.0.1:3306连接mysql的话，其实访问的是自己容器内部的端口。 package main import (\n"github.com/gin-gonic/gin"\n"gorm.io/driver/mysql"\n"gorm.io/gorm"\n"log"\n"net/http"\n) func main(){\nInitDB(GetDsn())\napp:=gin.Default()\napp.GET("/", func(c \*gin.Context) {\nc.JSON(http.StatusOK,"hello word")\n})\napp.Run(":80")\n} func InitDB(dsn string) {\n_, err := gorm.Open(mysql.Open(dsn))\nif err != nil {\nlog.Fatalln("open database success failed",err)\n}\nlog.Println("open database success")\n} func GetDsn() string {\nvar (\nhost   = "127.0.0.1"\nport   = "3306"\nusername = "root"\npassword = "123456"\ndbname  = "test"\n)\ndbs := username + ":" + password + "@(" + host + ":" + port + ")/" + dbname + "?charset=utf8mb4&parseTime=True&loc=Local"\nreturn dbs\n} \[root@VM-20-8-centos docker-test\]# docker build -t goserver .\n# ...构建过程省略 #先启动mysql容器，进去创建一个上述需要连接的test数据库(该创建过程省略)\n\[root@VM-20-8-centos docker-test\]# docker run -d -e MYSQL_ROOT_PASSWORD="123456" --name mymysql mysql\nbafa13668582d1a9b02782b910f1f5d427be43055b883021f6bf8362dcb7d170\n\[root@VM-20-8-centos docker-test\]# docker ps\nCONTAINER ID  IMAGE   COMMAND          CREATED      STATUS     PORTS         NAMES\nbafa13668582  mysql   "docker-entrypoint.s…"  10 seconds ago  Up 9 seconds  3306/tcp, 33060/tcp  mymysql #可以看到之前我们连接数据库的配置在此不能用\n\[root@VM-20-8-centos docker-test\]# docker run -it -p 80:80 goserver\n2022/05/13 05:54:01 /build/main.go:21\n\[error\] failed to initialize database, got error dial tcp 127.0.0.1:3306: connect: connection refused\n2022/05/13 05:54:01 open database success failed dial tcp 127.0.0.1:3306: connect: connection refused\n\[root@VM-20-8-centos docker-test\]# 针对这种情况，我们只需在启动go容器的时候 ，添加 **--link 参数 +mysql容器名+端口** ，之后我们将访问ip换成mysql容器名就能访问成功了。 # host ="mymydocker" //你启动的mysql容器名\n# 重新构建镜像 注意还要修改dockerfile 让容器中的3306端口暴露出来 EXPOSE 3306\n# \[root@VM-20-8-centos docker-test\]# docker run -it --link mymysql:3306 goserver\n2022/05/13 06:08:52 open database success //这里启动成功了\n\[GIN-debug\] \[WARNING\] Creating an Engine instance with the Logger and Recovery middleware already attached. \[GIN-debug\] \[WARNING\] Running in "debug" mode. Switch to "release" mode in production.\n\\- using env:   export GIN_MODE=release\n\\- using code:  gin.SetMode(gin.ReleaseMode) \[GIN-debug\] GET   /             --> main.main.func1 (3 handlers)\n\[GIN-debug\] \[WARNING\] You trusted all proxies, this is NOT safe. We recommend you to set a value.\nPlease check https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies for details.\n\[GIN-debug\] Listening and serving HTTP on :80

# **开发最佳实践**

### **如何让你的image变小**

启动容器或服务时，小image可以更快地通过网络拉取并更快地加载到内存中。有一些经验法则可以保持image较小：

* 从合适的基础镜像开始。例如，如果您需要 JDK，请考虑将您的映像基于官方`openjdk`映像，而不是从通用`ubuntu`映像开始并`openjdk`作为 Dockerfile 的一部分进行安装。
* [使用多阶段构建](https://docs.docker.com/build/building/multi-stage/)。例如，您可以使用`maven`image构建 Java 应用程序，然后重置为`tomcat`image并将 Java 工件复制到正确的位置以部署您的应用程序，所有这些都在同一个 Dockerfile 中。这意味着您的最终映像不包含构建引入的所有库和依赖项，而只包含运行它们所需的工件和环境。
  * `RUN`如果您需要使用不包含多阶段构建的 Docker 版本，请尝试通过最小化Dockerfile 中单独命令的数量来减少映像中的层数。`RUN`您可以通过将多个命令合并到一行并使用 shell 的机制将它们组合在一起来实现。考虑以下两个片段。第一个在image中创建两个层，而第二个只创建一个。
  * RUN apt-get -y update\nRUN apt-get install -y python
  * RUN apt-get -y update && apt-get install -y python
* 如果您有多个具有很多共同点的image，请考虑 使用共享组件创建您自己的[基础image，并以此为基础创建您独特的image。](https://docs.docker.com/build/building/base-images/)Docker 只需要加载一次公共层，它们会被缓存。这意味着您的衍生image可以更有效地使用 Docker 主机上的内存并加载得更快。

### **在哪里以及如何保存应用程序数据**

* **避免使用**[存储驱动程序](https://docs.docker.com/storage/storagedriver/select-storage-driver/)将应用程序数据存储在容器的可写层中 。这会增加容器的大小，并且从 I/O 的角度来看，**效率低于使用卷或绑定挂载**。
* 相反，使用[volumes](https://docs.docker.com/storage/volumes/)存储数据。

# **Dockerfile最佳实践**

### **排除 .dockerignore**

要排除与构建无关的文件，而不重构源存储库，请使用文件`.dockerignore`。此文件支持类似于`.gitignore`文件的排除模式。有关创建文件的信息，请参阅 [.dockerignore 文件](https://docs.docker.com/engine/reference/builder/#dockerignore-file)。

### **使用多阶段构建**

[多阶段构建](https://docs.docker.com/build/building/multi-stage/)允许您大幅减小最终image的大小，而无需努力减少中间层和文件的数量。 因为image是在构建过程的最后阶段构建的，所以您可以通过[利用构建缓存](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#leverage-build-cache)来最小化image层。 例如，如果您的构建包含多个层并且您希望确保构建缓存可重用，您可以将它们从更改频率较低的顺序排列到更改频率较高的顺序。以下列表是指令顺序的示例：


1. 安装构建应用程序所需的工具
2. 安装或更新库依赖项
3. 生成您的应用程序 Go 应用程序的 Dockerfile 可能如下所示： # syntax=docker/dockerfile:1\nFROM golang:1.16-alpine AS build # Install tools required for project\n# Run `docker build --no-cache .` to update dependencies\nRUN apk add --no-cache git\nRUN go get github.com/golang/dep/cmd/dep # List project dependencies with Gopkg.toml and Gopkg.lock\n# These layers are only re-built when Gopkg files are updated\nCOPY Gopkg.lock Gopkg.toml /go/src/project/\nWORKDIR /go/src/project/\n# Install library dependencies\nRUN dep ensure -vendor-only # Copy the entire project and build it\n# This layer is rebuilt when a file changes in the project directory\nCOPY . /go/src/project/\nRUN go build -o /bin/project # This results in a single layer image\nFROM scratch\nCOPY --from=build /bin/project /bin/project\nENTRYPOINT \["/bin/project"\]\nCMD \["--help"\]

### **不要安装不必要的包**

避免安装额外的或不必要的包，因为它们可能很好。例如，您不需要在数据库image中包含文本编辑器。 当您避免安装额外或不必要的包时，您的image将降低复杂性、减少依赖性、减小文件大小并缩短构建时间。

### **解耦应用程序**

每个容器应该只有一个程序。将应用程序解耦到多个容器中可以更轻松地水平扩展和重用容器。例如，一个 Web 应用程序堆栈可能由三个独立的容器组成，每个容器都有自己独特的image，以分离的方式管理 Web 应用程序、数据库和内存缓存。 将每个容器限制为一个进程是一个很好的经验法则，但这不是一个硬性规定。 使用你最好的判断来保持容器尽可能干净和模块化。如果容器相互依赖，您可以使用[Docker 容器网络](https://docs.docker.com/network/) 来确保这些容器可以通信。

### **尽量减少层数**

在旧版本的 Docker 中，尽量减少image中的层数以确保它们的性能非常重要。添加了以下功能以减少此限制：

* 只有指令`RUN`, `COPY`,`ADD`创建图层。其他指令创建临时中间image，并且不会增加构建的大小。
* 在可能的情况下，使用[多阶段构建](https://docs.docker.com/build/building/multi-stage/)，并且只将您需要的工件复制到最终image中。这允许您在中间构建阶段包含工具和调试信息，而无需增加最终image的大小。

### **对多行参数进行排序**

只要有可能，通过按字母数字顺序对多行参数进行排序来简化以后的更改。这有助于避免包重复并使列表更容易更新。这也使 PR 更容易阅读和审查。在反斜杠 ( `\`) 前添加一个空格也有帮助。 [这是来自buildpack-deps image](https://github.com/docker-library/buildpack-deps)的示例： RUN apt-get update && apt-get install -y \\  bzr \\  cvs \\  git \\  mercurial \\  subversion \\  && rm -rf /var/lib/apt/lists/\*

### **利用构建缓存**

构建映像时，Docker 会逐步执行 Dockerfile 中的指令，并按照指定的顺序执行每条指令。在检查每条指令时，Docker 会在其缓存中查找可以重用的现有image，而不是创建新的重复image。 如果您根本不想使用缓存，可以使用命令`--no-cache=true` 上的选项`docker build`。但是，如果您确实让 Docker 使用它的缓存，那么了解它什么时候可以，什么时候不能找到匹配的image是很重要的。Docker 遵循的基本规则概述如下：

* 从缓存中已有的父image开始，将下一条指令与从该基础image派生的所有子image进行比较，以查看其中一个是否是使用完全相同的指令构建的。如果不是，则缓存无效。
* 在大多数情况下，只需将 Dockerfile 中的指令与其中一个子image进行比较就足够了。但是，某些说明需要更多的检查和解释。
* 对于`ADD`和`COPY`指令，检查image中每个文件的内容并为每个文件计算校验和。这些校验和不考虑每个文件的上次修改时间和上次访问时间。在缓存查找期间，将校验和与现有image中的校验和进行比较。如果任何文件中发生了任何更改，例如内容和元数据，则缓存将失效。
* 除了`ADD`和`COPY`命令之外，缓存检查不会查看容器中的文件来确定缓存匹配。例如，在处理命令时，`RUN apt-get -y update`不会检查容器中更新的文件以确定是否存在缓存命中。在这种情况下，仅使用命令字符串本身来查找匹配项。 一旦缓存失效，所有后续的 Dockerfile 命令都会生成新image并且不会使用缓存。

# **作业**

* lv1：开发"登录与注册"http或RPC服务，并使用docker**分别对应用程序和数据库服务打包**，数据需要持久化到宿主机
* lv2：基于一个操作系统镜像，编写Dockerfile，构建自己的mysql或redis镜像，并且push到dockerhub中，提交镜像ID即可。