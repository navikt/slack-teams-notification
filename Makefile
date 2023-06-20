all: slack-teams-notification test check fmt

slack-teams-notification:
	go build -o bin/slack-teams-notification cmd/slack-teams-notification/main.go

test:
	go test ./...

check:
	go run honnef.co/go/tools/cmd/staticcheck ./...
	go run golang.org/x/vuln/cmd/govulncheck -v -test ./...

fmt:
	go run mvdan.cc/gofumpt -w ./

static:
	CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/slack-teams-notification cmd/slack-teams-notification/main.go