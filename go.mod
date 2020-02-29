module github.com/jenkins-x-labs/app-slack

require (
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/evanphx/json-patch v4.1.0+incompatible
	github.com/gogo/protobuf v1.1.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/jenkins-x/jx v1.3.830
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a
	github.com/nlopes/slack v0.0.0-20180721202243-347a74b1ea30
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.8.0
	github.com/prometheus/common v0.2.0
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5 // indirect
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.1 // indirect
	k8s.io/api v0.0.0-20190126160303-ccdd560a045f
	k8s.io/apiextensions-apiserver v0.0.0-20181128195303-1f84094d7e8e
	k8s.io/apimachinery v0.0.0-20190122181752-bebe27e40fb7
	k8s.io/client-go v9.0.0+incompatible
	k8s.io/code-generator v0.0.0-20181017053441-8c97d6ab64da
	k8s.io/gengo v0.0.0-20180718083919-906d99f89cd6
	k8s.io/kube-openapi v0.0.0-20180719232738-d8ea2fe547a4
	k8s.io/metrics v0.0.0-20190126173137-e22014de0362 // indirect
	k8s.io/test-infra v0.0.0-20190131093439-a22cef183a8f
)

replace k8s.io/api => k8s.io/api v0.0.0-20181128191700-6db15a15d2d3

replace k8s.io/client-go => k8s.io/client-go v2.0.0-alpha.0.0.20190115164855-701b91367003+incompatible

replace github.com/heptio/sonobuoy => github.com/jenkins-x/sonobuoy v0.11.7-0.20190131193045-dad27c12bf17

go 1.13
