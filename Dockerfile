FROM arm32v7/golang

COPY . /go/src/goser
WORKDIR /go/src/goser

RUN go get -d -v ./...
RUN go build -o main *.go

CMD ["./main"]
