FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /server ./cmd/server

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=builder /server /server

EXPOSE 50051

ENTRYPOINT ["/server"]
