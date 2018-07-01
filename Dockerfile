FROM arm32v7/golang as builder

COPY . /go/src/goser
WORKDIR /go/src/goser

RUN go get -d -v ./...
RUN go build -o spackler *.go

FROM arm32v7/ubuntu

COPY --from=builder /go/src/goser/spackler /app/spackler/.
WORKDIR /app/spackler

EXPOSE 8080

ENV GK_TOKEN="d42a152bff711f187479d8613ccb47925d82b21a"
ENV GK_SERVER="http://10.1.1.1:8080"
ENV SERIAL_DEVICE="/dev/ttyS0"
ENV SERIAL_DEVICE_BAUD=9600

CMD ["./spackler"]
