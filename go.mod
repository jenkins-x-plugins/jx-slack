module github.com/jenkins-x-labs/slack

require (
	github.com/briandowns/spinner v1.9.0 // indirect
	github.com/c2h5oh/datasize v0.0.0-20200112174442-28bbd4740fee // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/jenkins-x/jx v1.3.1119
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.4.1
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/slack-go/slack v0.6.3
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver v0.0.0-20190718185103-d1ef975d28ce
	k8s.io/apimachinery v0.17.4
	k8s.io/apiserver v0.17.4 // indirect
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.17.3
	k8s.io/gengo v0.0.0-20190822140433-26a664648505
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	k8s.io/metrics v0.17.3 // indirect
	k8s.io/test-infra v0.0.0-20190131093439-a22cef183a8f
)

replace k8s.io/api => k8s.io/api v0.0.0-20190528110122-9ad12a4af326

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190528110200-4f3abb12cae2

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20181128195641-3954d62a524d

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221084156-01f179d85dbc

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190528110544-fa58353d80f3

replace github.com/sirupsen/logrus => github.com/jtnord/logrus v1.4.2-0.20190423161236-606ffcaf8f5d

replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v21.1.0+incompatible

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v10.14.0+incompatible

replace github.com/banzaicloud/bank-vaults => github.com/banzaicloud/bank-vaults v0.0.0-20190508130850-5673d28c46bd

replace github.com/heptio/sonobuoy => github.com/heptio/sonobuoy v0.16.0

replace github.com/jenkins-x/jx => github.com/jenkins-x/jx v0.0.0-20200305075422-16ecced629f9

go 1.13
