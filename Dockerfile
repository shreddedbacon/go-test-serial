FROM arm32v7/golang

COPY . /go/src/goser
WORKDIR /go/src/goser

RUN go get -d -v ./...
RUN go build -o main *.go

ENV GK_TOKEN="d42a152bff711f187479d8613ccb47925d82b21a"
ENV GK_SERVER="http://10.1.1.1:8080"

CMD ["./main"]
