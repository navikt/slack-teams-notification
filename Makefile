all: slack-teams-notification test check fmt

slack-teams-notification:
	go build -o bin/slack-teams-notification cmd/slack-teams-notification/main.go

test:
	go test ./...

check:
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest -test ./...

fmt:
	go run mvdan.cc/gofumpt@latest -w ./

static:
	CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/slack-teams-notification cmd/slack-teams-notification/main.go