FROM golang:1.17 as builder

WORKDIR /src

COPY . .

RUN go build -o /bin/conntest *.go

FROM golang:1.17 as app

COPY --from=builder /bin/conntest /bin/conntest

ENTRYPOINT [ "conntest" ]