FROM golang as builder

ENV GO111MODULE=on

COPY ./ /src/

WORKDIR /src/samples/idobot-gopher

RUN go clean -cache

RUN go clean -modcache

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app

FROM scratch

COPY --from=builder /app /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app"]
