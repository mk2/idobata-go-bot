FROM golang as builder

ENV GO111MODULE=on

WORKDIR $GOPATH/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app

FROM scratch

COPY --from=builder /app /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app"]