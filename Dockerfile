ARG GO_VERSION="1.24"
FROM golang:${GO_VERSION} AS builder

WORKDIR /src
COPY go.* ./
RUN go mod download
COPY . ./
RUN go build -o ./bin/slack-teams-notification ./cmd/slack-teams-notification

FROM cgr.dev/chainguard/static
COPY --from=builder /src/bin/slack-teams-notification /app/slack-teams-notification
CMD ["/app/slack-teams-notification"]