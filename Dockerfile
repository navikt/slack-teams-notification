FROM golang:1.23-alpine AS builder

WORKDIR /src
COPY go.* ./
RUN go mod download
COPY . .

RUN go test -v ./...
RUN go run honnef.co/go/tools/cmd/staticcheck@latest ./...
RUN go run golang.org/x/vuln/cmd/govulncheck@latest ./...
RUN go run golang.org/x/tools/cmd/deadcode@latest -test ./...
RUN go build -o ./bin/slack-teams-notification ./cmd/slack-teams-notification

FROM cgr.dev/chainguard/static
COPY --from=builder /src/bin/slack-teams-notification /app/slack-teams-notification
CMD ["/app/slack-teams-notification"]