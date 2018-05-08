FROM golang:1.10.1
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLE=0 GOOS=linux go build -o echo .

FROM debian
COPY --from=0 /go/src/app/echo .
ENTRYPOINT ["/echo"]
