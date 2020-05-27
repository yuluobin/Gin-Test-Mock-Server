FROM golang:latest

COPY . $GOPATH/src/github.com/brown-csci1380-s20/puddlestorenew-puddlestorenew-cwang147-byu18-mxu57
WORKDIR $GOPATH/src/github.com/brown-csci1380-s20/puddlestorenew-puddlestorenew-cwang147-byu18-mxu57/mockServer


#ADD . $GOPATH/src/github.com/brown-csci1380-s20/puddlestorenew-puddlestorenew-cwang147-byu18-mxu57/mockServer/
#WORKDIR $GOPATH/src/github.com/brown-csci1380-s20/puddlestorenew-puddlestorenew-cwang147-byu18-mxu57/mockServer
RUN go get github.com/gin-gonic/gin
RUN go build .
EXPOSE 3000/tcp
ENTRYPOINT ["./mockServer"]