module github.com/jenkins-x-labs/slack

require (
	github.com/briandowns/spinner v1.9.0 // indirect
	github.com/c2h5oh/datasize v0.0.0-20200112174442-28bbd4740fee // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/jenkins-x/go-scm v1.5.143
	github.com/jenkins-x/jx-logging v0.0.10
	github.com/jenkins-x/jx/v2 v2.1.84
	github.com/jenkins-x/lighthouse v0.0.650
	github.com/knative/build v0.7.0 // indirect
	github.com/knative/pkg v0.0.0-20190624141606-d82505e6c5b4 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.7.0
	github.com/sirupsen/logrus v1.6.0
	github.com/slack-go/slack v0.6.3
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.6.0
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975 // indirect
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/metrics v0.17.3 // indirect
	k8s.io/test-infra v0.0.0-20190131093439-a22cef183a8f
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451 // indirect
)

replace k8s.io/api => k8s.io/api v0.16.5

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20190819143841-305e1cef1ab1

replace k8s.io/apimachinery => k8s.io/apimachinery v0.16.5

replace k8s.io/client-go => k8s.io/client-go v0.16.5

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190819143637-0dbe462fe92d

replace github.com/sirupsen/logrus => github.com/jtnord/logrus v1.4.2-0.20190423161236-606ffcaf8f5d

replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v23.2.0+incompatible

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.1+incompatible

replace github.com/banzaicloud/bank-vaults => github.com/banzaicloud/bank-vaults v0.0.0-20191212164220-b327d7f2b681

replace github.com/banzaicloud/bank-vaults/pkg/sdk => github.com/banzaicloud/bank-vaults/pkg/sdk v0.0.0-20191212164220-b327d7f2b681

replace k8s.io/test-infra => github.com/jenkins-x/test-infra v0.0.0-20200611142252-211a92405c22

replace gomodules.xyz/jsonpatch/v2 => gomodules.xyz/jsonpatch/v2 v2.0.1

replace github.com/heptio/sonobuoy => github.com/heptio/sonobuoy v0.16.0

replace github.com/jenkins-x/lighthouse => github.com/abayer/lighthouse v0.0.0-20200624191510-286888fb3279

go 1.13
