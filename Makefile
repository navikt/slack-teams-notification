.PHONY: all
all: fmt test check build

.PHONY: fmt
fmt:
	go tool mvdan.cc/gofumpt -w ./

.PHONY: test
test:
	go test --race -v ./...

.PHONY: check
check: staticcheck vulncheck deadcode

.PHONY: staticcheck
staticcheck:
	go tool honnef.co/go/tools/cmd/staticcheck ./...

.PHONY: vulncheck
vulncheck:
	go tool golang.org/x/vuln/cmd/govulncheck ./...

.PHONY: deadcode
deadcode:
	go tool golang.org/x/tools/cmd/deadcode -test ./...

.PHONY: build
build:
	go build -o ./bin/slack-teams-notification ./cmd/slack-teams-notification