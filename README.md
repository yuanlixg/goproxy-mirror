
# GOPROXY mirror

给[https://github.com/goproxyio/goproxy](github.com/goproxyio/goproxy)添加从镜像站点获取的功能


## 原英文 README 见下面


## 此程序只适用于一次一人使用，文件写入竟态不安全

无论是 [goproxyio](https://github.com/goproxyio/goproxy) 还是 [Athens](https://github.com/gomods/athens) 都是利用 go get -d 的功能。
Athens每次都设定一个单独环境、一个临时目录运行go get -d，但它只利用了 go(go.exe) 本身，是否竟态安全不清楚。
goproxyio 完全依靠 go get -d 来决定是否竟态安全，随着 pkg/mod/cache/vcs 下的模块增多，大概率的工作都让http文件服务处理了，
竟态概率越来越小，因golang卓越的文件服务性能让它越来越快。


此程序只是作为 [https://goproxy.io](https://goproxy.io) 不能访问时的应急措施，没能力做到竟态安全。
只有 mirror.go 是新加的，main.go做了少量修改，版权属于 @goproxyio 。


使用的原则是能不用GOPROXY，就尽量不用。临时执行下行处理好依赖，就正常操作。

	GOPROXY=http://localhost:8081 go mod tidy


## 警告

无论是本程序还是原 goproxyio，都要安装到一个单独的 GOPATH 下，不能和 golang 的 GOPATH 相同，否则会引入额外的竟态。


假设 golang 的 GOPATH 是 $HOME/go

	＃ 执行一次
	mkdir -p "$HOME"/goproxy
	cp -p ./goproxy "$HOME"/goproxy/

	# 每次运行，清理 go get -d 引入的indirect依赖
	cd "$HOME"/goproxy
	rm -f go.mod go.sum
	echo "module m">go.mod
	GOPATH="$HOME/goproxy" ./goproxy


## 添加新的mirror

参见main.go文件 func main() 中注释掉的例子。一类站点加一行就行。


## Build 技巧

alpine和debian同为linux/amd64，但默认编译结果并不通用，一个是musl C库，一个是glibc C库。
macOS下编译的库路径和版本也有可能不相同。按交叉编译处理最通用。

非本机使用(unix同类型，适用于go1.11.x、go1.10.x)

	CGO_ENABLED=0 MACOSX_DEPLOYMENT_TARGET=10.10 go build -a

给Windows用

	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -a


## Docker 技巧

	docker run -it goproxyio/goproxy

用 -v 参数，利用本地命名持久化卷保存proxy module数据:

	docker run -it -v go-repo-volume:/go/pkg/mod goproxyio/goproxy

原英文指令保存数据不全，容器会越来越大，每次启动要重新下载部分VCS。
/go/pkg/mod/cache/download是部分结果，/go/pkg/mod/cache/vcs是原始数据。
可能还有 /go/pkg/mod/github.com /go/pkg/mod/golang.org ...


原英文指令用本机目录，Windows下运行可靠性差的不是一个数量级，unix下安全性差一些。

从持久化卷获取数据(备份):

	docker run -i -v go-repo-volume:/go/pkg/mod goproxyio/goproxy tar -C /go/pkg/mod -cf - cache | gzip > pkg-mod-cache.tar.gz


#


********


# GOPROXY [![CircleCI](https://circleci.com/gh/goproxyio/goproxy.svg?style=svg)](https://circleci.com/gh/goproxyio/goproxy)

A global proxy for go modules. see: [https://goproxy.io](https://goproxy.io)

## Build

    go build

## Started

    ./goproxy -listen=0.0.0.0:80

## Docker

    docker run -it goproxyio/goproxy

Use the -v flag to persisting the proxy module data (change ___go_repo___ to your own dir):

    docker run -it -v go_repo:/go/pkg/mod/cache/download goproxyio/goproxy

## Docker Compose

    docker-compose up


