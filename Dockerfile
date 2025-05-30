FROM golang:alpine AS go-builder

RUN apk --no-cache add ca-certificates

WORKDIR /go/src/app

COPY go.mod go.sum /go/src/app/

RUN go mod download

COPY . /go/src/app/

RUN CGO_ENABLED=0 go build -a -trimpath -tags netgo -ldflags '-s -w -extldflags "-static"' -o ./ ./cmd/...

FROM scratch

LABEL maintainer="Flemming Andersen <flemming@flamingcode.com>"

EXPOSE 8080

WORKDIR /app

CMD ["/app/taleplade"]

COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=go-builder /go/src/app/taleplade /app
COPY --from=go-builder /go/src/app/public /app/public
