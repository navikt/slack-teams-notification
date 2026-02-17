ARG GO_VERSION="1.25"
FROM golang:${GO_VERSION} AS builder

WORKDIR /src
COPY go.* ./
RUN go mod download
COPY . ./
RUN go build -o ./notifier ./main.go

FROM gcr.io/distroless/base
COPY --from=builder /src/notifier /app/notifier
CMD ["/app/notifier"]
