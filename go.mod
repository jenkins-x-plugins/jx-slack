module github.com/jenkins-x-plugins/jx-slack

require (
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/jenkins-x-plugins/jx-changelog v0.0.42
	github.com/jenkins-x-plugins/jx-gitops v0.2.91
	github.com/jenkins-x/go-scm v1.9.0
	github.com/jenkins-x/jx-api/v4 v4.0.33
	github.com/jenkins-x/jx-helpers/v3 v3.0.114
	github.com/jenkins-x/jx-logging/v3 v3.0.6
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.15.0
	github.com/sethvargo/go-envconfig v0.3.2
	github.com/slack-go/slack v0.8.1
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.7.0
	k8s.io/apimachinery v0.20.7
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
)

go 1.16
