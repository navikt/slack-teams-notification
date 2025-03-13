module github.com/nais/slack-teams-notification

go 1.24

tool (
	golang.org/x/tools/cmd/deadcode
	golang.org/x/vuln/cmd/govulncheck
	honnef.co/go/tools/cmd/staticcheck
	mvdan.cc/gofumpt
)

require (
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/sethvargo/go-envconfig v1.1.1
	github.com/sirupsen/logrus v1.9.3
	github.com/slack-go/slack v0.16.0
)

require (
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/exp/typeparams v0.0.0-20231108232855-2478ac86f678 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/telemetry v0.0.0-20240522233618-39ace7a40ae7 // indirect
	golang.org/x/tools v0.31.0 // indirect
	golang.org/x/vuln v1.1.4 // indirect
	honnef.co/go/tools v0.6.1 // indirect
	mvdan.cc/gofumpt v0.7.0 // indirect
)
