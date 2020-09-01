package slackbot

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	jxv1 "github.com/jenkins-x/jx/v2/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx/v2/pkg/cmd/opts"
	"github.com/jenkins-x/jx/v2/pkg/cmd/testhelpers"
	"github.com/jenkins-x/jx/v2/pkg/gits"
	helm_test "github.com/jenkins-x/jx/v2/pkg/helm/mocks"
	resources_test "github.com/jenkins-x/jx/v2/pkg/kube/resources/mocks"
	"github.com/jenkins-x/lighthouse/pkg/util"
	slackappapi "github.com/jenkins-x/slack/pkg/apis/slack/v1alpha1"
	"github.com/jenkins-x/slack/pkg/client/clientset/versioned/fake"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slacktest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testOrgName  = "test-org"
	testRepoName = "test-repo"
	secretName   = "some-secret"
	testChannel  = "some-channel"
	testNs       = "jx"
)

func TestLighthouseDelivery(t *testing.T) {
	testcases := []struct {
		name            string
		payload         string
		headers         map[string]string
		expectedMessage string
	}{{
		name: "something",
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			server := createServer()
			go server.Start()

			botConfig := createSlackBot()
			clients := createTestClients(server, botConfig, []*jxv1.PipelineActivity{})

			bots := SlackBots{
				GlobalClients:  clients,
				HmacSecretName: secretName,
			}
			handler := bots.ExternalPluginServer()

			s := httptest.NewServer(handler)
			defer s.Close()
		})
	}

}

func createServer(customizers ...func(slacktest.Customize)) *slacktest.Server {
	s := slacktest.NewTestServer()
	for _, c := range customizers {
		c(s)
	}
	return s
}

func createSlackBot() *slackappapi.SlackBot {
	return &slackappapi.SlackBot{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test_slack_bot",
		},
		Spec: slackappapi.SlackBotSpec{
			TokenReference: jxv1.ResourceReference{
				Kind: "Secret",
				Name: secretName,
			},
			Namespace: testNs,
			PullRequests: []slackappapi.SlackBotMode{{
				Channel: testChannel,
				Orgs: []slackappapi.Org{{
					Name:  testOrgName,
					Repos: []string{testRepoName},
				}},
			}},
			Pipelines: []slackappapi.SlackBotMode{{
				Channel: testChannel,
				Orgs: []slackappapi.Org{{
					Name:  testOrgName,
					Repos: []string{testRepoName},
				}},
			}},
		},
	}
}

func createTestClients(svr *slacktest.Server, bot *slackappapi.SlackBot, activities []*jxv1.PipelineActivity) *GlobalClients {
	commonOpts := &opts.CommonOptions{}

	fakeRepo, _ := gits.NewFakeRepository(testOrgName, testRepoName, nil, nil)
	fakeGitProvider := gits.NewFakeProvider(fakeRepo)
	gitter := gits.NewGitCLI()
	mockHelmer := helm_test.NewMockHelmer()
	installerMock := resources_test.NewMockInstaller()

	var jxObjects []runtime.Object
	for _, a := range activities {
		jxObjects = append(jxObjects, a)
	}
	var kubeObjects []runtime.Object
	kubeObjects = append(kubeObjects, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: testNs,
		},
		Data: map[string][]byte{
			"hmac": []byte("abcd1234"),
		},
	})
	testhelpers.ConfigureTestOptionsWithResources(commonOpts,
		kubeObjects,
		jxObjects,
		gitter,
		fakeGitProvider,
		mockHelmer,
		installerMock)

	kubeClient, _ := commonOpts.KubeClient()
	jxClient, _, _ := commonOpts.JXClient()
	return &GlobalClients{
		SlackAppClient: fake.NewSimpleClientset(bot),
		Namespace:      testNs,
		KubeClient:     kubeClient,
		JXClient:       jxClient,
		slackClientHelper: &listenerFakeSlackClient{
			serverURL: svr.GetAPIURL(),
		},
		CommonOptions: commonOpts,
	}
}

type listenerFakeSlackClient struct {
	*slack.Client
	serverURL string
}

func (f *listenerFakeSlackClient) getSlackClient(token string, options ...slack.Option) *slack.Client {
	return slack.New(token, slack.OptionAPIURL(f.serverURL))
}

// SendHook sends an event of type eventType to the provided address.
func SendHook(address, eventType string, headers map[string]string, payload, hmacToken []byte) error {
	req, err := http.NewRequest(http.MethodPost, address, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	mac := hmac.New(sha256.New, hmacToken)
	_, err = mac.Write(payload)
	if err != nil {
		return err
	}
	sum := mac.Sum(nil)
	signature := "sha256=" + hex.EncodeToString(sum)
	req.Header.Set(util.LighthouseSignatureHeader, signature)

	req.Header.Set("content-type", "application/json")

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("response from hook has status %d and body %s", resp.StatusCode, string(bytes.TrimSpace(rb)))
	}
	return nil
}
