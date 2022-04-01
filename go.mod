module github.com/jenkins-x-plugins/jx-slack

require (
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/jenkins-x-plugins/jx-changelog v0.0.42
	github.com/jenkins-x-plugins/jx-gitops v0.7.5
	github.com/jenkins-x/go-scm v1.11.4
	github.com/jenkins-x/jx-api/v4 v4.3.4
	github.com/jenkins-x/jx-helpers/v3 v3.2.4
	github.com/jenkins-x/jx-logging/v3 v3.0.6
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a
	github.com/pkg/errors v0.9.1
	github.com/sethvargo/go-envconfig v0.3.5
	github.com/slack-go/slack v0.10.2
	github.com/spf13/cobra v1.4.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/oauth2 v0.0.0-20220223155221-ee480838109b // indirect
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
)

replace (
	github.com/nats-io/jwt/v2 => github.com/nats-io/jwt/v2 v2.0.1
	github.com/nats-io/nats-server/v2 => github.com/nats-io/nats-server/v2 v2.7.2
	// Fixing dependabot alerts
	github.com/tidwall/gjson => github.com/tidwall/gjson v1.9.3
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
)

go 1.16
