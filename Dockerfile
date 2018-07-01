FROM arm32v7/golang

COPY . /go/src/goser
WORKDIR /go/src/goser

EXPOSE 8080

RUN go get -d -v ./...
RUN go build -o main *.go

ENV GK_TOKEN="d42a152bff711f187479d8613ccb47925d82b21a"
ENV GK_SERVER="http://10.1.1.1:8080"
ENV SERIAL_DEVICE="/dev/ttyS0"
ENV SERIAL_DEVICE_BAUD=9600

CMD ["./main"]
