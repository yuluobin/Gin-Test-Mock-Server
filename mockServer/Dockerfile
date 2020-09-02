FROM csighub.tencentyun.com/tlinux/tlinux-mini:2.2-stke
LABEL MAINTAINER="bookeryu@tencent.com"

RUN go get github.com/gin-gonic/gin

COPY . /go/src/github.com/yuluobin/Gin-Test-Mock-Server/mockServer
WORKDIR /go/src/github.com/yuluobin/Gin-Test-Mock-Server/mockServer
#ADD . $GOPATH/src/github.com/brown-csci1380-s20/puddlestorenew-puddlestorenew-cwang147-byu18-mxu57/mockServer/
#WORKDIR $GOPATH/src/github.com/brown-csci1380-s20/puddlestorenew-puddlestorenew-cwang147-byu18-mxu57/mockServer
#RUN go get ./...
RUN go build .
EXPOSE 8000
ENTRYPOINT ["./mockServer"]
