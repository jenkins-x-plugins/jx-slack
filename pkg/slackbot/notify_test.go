package slackbot_test

import (
	"github.com/jenkins-x-plugins/slack/pkg/slackbot"
	fakescm "github.com/jenkins-x/go-scm/scm/driver/fake"
	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	fakejx "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned/fake"
	"github.com/jenkins-x/jx-gitops/pkg/apis/gitops/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"testing"
)

func TestNotifyPipeline(t *testing.T) {
	testCases := []struct {
		name     string
		cfg      v1alpha1.SlackNotify
		expected []string
	}{
		{
			name: "all",
			cfg: v1alpha1.SlackNotify{
				Kind: v1alpha1.NotifyKindAlways,
			},
			expected: []string{"release-running-1", "pr-failed-1"},
		},
		{
			name: "matchesReleases",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailureOrFirstSuccess,
				Pipeline: v1alpha1.PipelineKindRelease,
			},
		},
	}

	ns := "jx"
	activities := []*jenkinsv1.PipelineActivity{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "release-running-1",
				Namespace: ns,
			},
			Spec: jenkinsv1.PipelineActivitySpec{
				Build:     "1",
				Status:    jenkinsv1.ActivityStatusTypeRunning,
				GitBranch: "main",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pr-failed-1",
				Namespace: ns,
			},
			Spec: jenkinsv1.PipelineActivitySpec{
				Build:     "1",
				Status:    jenkinsv1.ActivityStatusTypeFailed,
				GitBranch: "PR-1234",
			},
		},
	}

	var jxObjects []runtime.Object
	for _, pa := range activities {
		jxObjects = append(jxObjects, pa)
	}
	jxClient := fakejx.NewSimpleClientset(jxObjects...)
	scmClient, _ := fakescm.NewDefault()

	for _, tc := range testCases {
		name := tc.name
		cfg := &tc.cfg
		t.Logf("running case %s\n", name)

		o := &slackbot.SlackBotOptions{
			KubeClient: fake.NewSimpleClientset(),
			JXClient:   jxClient,
			ScmClient:  scmClient,
		}

		var matched []string
		for _, pa := range activities {
			flag, _, _, err := o.NotifyPipeline(pa, cfg)
			require.NoError(t, err, "failed to run NotifyPipeline for activity %s test: %s", pa.Name, name)
			if flag {
				matched = append(matched, pa.Name)
			}
		}

		assert.Equal(t, tc.expected, matched, "matched pipeline names for test: %s", name)
	}
}
