module github.com/jenkins-x-plugins/jx-slack

require (
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/jenkins-x-plugins/jx-changelog v0.0.42
	github.com/jenkins-x-plugins/jx-gitops v0.2.63
	github.com/jenkins-x/go-scm v1.6.18
	github.com/jenkins-x/jx-api/v4 v4.0.28
	github.com/jenkins-x/jx-helpers/v3 v3.0.104
	github.com/jenkins-x/jx-logging/v3 v3.0.3
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.15.0
	github.com/sethvargo/go-envconfig v0.3.2
	github.com/slack-go/slack v0.8.1
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.6.1
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
)

replace (
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
)

go 1.15
