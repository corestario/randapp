FROM dkglib_testnet:latest

RUN mkdir /root/tmp
ENV GO111MODULE=off
ENV PATH /go/bin:$PATH
ENV GOPATH /go
ENV RAPATH /go/src/github.com/corestario/randapp
RUN mkdir -p /go/src/github.com/corestario/randapp

COPY . $RAPATH

WORKDIR $RAPATH

RUN go install $RAPATH/cmd/randappd
RUN go install $RAPATH/cmd/randappcli

WORKDIR $RAPATH/scripts

EXPOSE 26656
