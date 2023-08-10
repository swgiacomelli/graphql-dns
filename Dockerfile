FROM golang:1.19-alpine as builder

WORKDIR /go/src/github.com/swgiacomelli/graphql-dns
COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dnsd -ldflags '-s -w -extldflags "-static"' .

FROM scratch
COPY --from=builder /go/src/github.com/swgiacomelli/graphql-dns/dnsd /app/

ENTRYPOINT ["/app/dnsd"]