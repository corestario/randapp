FROM dkglib_testnet:latest

RUN mkdir /root/tmp
ENV GO111MODULE=off
ENV PATH /go/bin:$PATH
ENV GOPATH /go
ENV RAPATH /go/src/github.com/corestario/randapp
RUN mkdir -p /go/src/github.com/corestario/randapp

COPY . $RAPATH

WORKDIR $RAPATH

RUN go install $RAPATH/cmd/rd
RUN go install $RAPATH/cmd/rcli

WORKDIR $RAPATH/scripts

EXPOSE 26656
