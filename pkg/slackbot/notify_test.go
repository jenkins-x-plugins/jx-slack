package slackbot_test

import (
	"github.com/jenkins-x-plugins/slack/pkg/slackbot"
	"github.com/jenkins-x-plugins/slack/pkg/testpipelines"
	"github.com/jenkins-x/go-scm/scm"
	fakescm "github.com/jenkins-x/go-scm/scm/driver/fake"
	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	fakejx "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned/fake"
	"github.com/jenkins-x/jx-gitops/pkg/apis/gitops/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			expected: []string{"myorg-myrepo-main-1", "myorg-myrepo-main-2", "myorg-myrepo-pr-1234-mycontext-1", "myorg-myrepo-pr-4567-mycontext-1"},
		},
		{
			name: "matchesReleases",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailureOrFirstSuccess,
				Pipeline: v1alpha1.PipelineKindRelease,
			},
			expected: []string{"myorg-myrepo-main-1"},
		},
		{
			name: "release-with-branch",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailure,
				Pipeline: v1alpha1.PipelineKindRelease,
				Branch: &v1alpha1.Pattern{
					Name: "main",
				},
			},
			expected: []string{"myorg-myrepo-main-1"},
		},
		{
			name: "release-with-branch-include",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailure,
				Pipeline: v1alpha1.PipelineKindRelease,
				Branch: &v1alpha1.Pattern{
					Includes: []string{"main"},
				},
			},
			expected: []string{"myorg-myrepo-main-1"},
		},
		{
			name: "release-with-branch-exclude-not-exist",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailure,
				Pipeline: v1alpha1.PipelineKindRelease,
				Branch: &v1alpha1.Pattern{
					Excludes: []string{"does-not-exist"},
				},
			},
			expected: []string{"myorg-myrepo-main-1"},
		},
		{
			name: "release-with-branch-exclude",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailure,
				Pipeline: v1alpha1.PipelineKindRelease,
				Branch: &v1alpha1.Pattern{
					Excludes: []string{"main"},
				},
			},
		},
		{
			name: "release-with-branch-missing-include",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailure,
				Pipeline: v1alpha1.PipelineKindRelease,
				Branch: &v1alpha1.Pattern{
					Includes: []string{"PR-"},
				},
			},
		},
		{
			name: "pr-with-context-name",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailure,
				Pipeline: v1alpha1.PipelineKindPullRequest,
				Context: &v1alpha1.Pattern{
					Name: "mycontext",
				},
			},
			expected: []string{"myorg-myrepo-pr-1234-mycontext-1"},
		},
		{
			name: "pr-with-context-include",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailure,
				Pipeline: v1alpha1.PipelineKindPullRequest,
				Context: &v1alpha1.Pattern{
					Includes: []string{"mycontext"},
				},
			},
			expected: []string{"myorg-myrepo-pr-1234-mycontext-1"},
		},
		{
			name: "pr-with-context-exclude",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailure,
				Pipeline: v1alpha1.PipelineKindPullRequest,
				Context: &v1alpha1.Pattern{
					Excludes: []string{"mycontext"},
				},
			},
		},
		{
			name: "pr-with-label-name",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailure,
				Pipeline: v1alpha1.PipelineKindPullRequest,
				PullRequestLabel: &v1alpha1.Pattern{
					Name:     "dependencies",
					Includes: nil,
					Excludes: nil,
				},
			},
			expected: []string{"myorg-myrepo-pr-1234-mycontext-1"},
		},
		{
			name: "pr-with-label-includes",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailure,
				Pipeline: v1alpha1.PipelineKindPullRequest,
				PullRequestLabel: &v1alpha1.Pattern{
					Includes: []string{"dependencies"},
					Excludes: []string{"does-not-exist"},
				},
			},
			expected: []string{"myorg-myrepo-pr-1234-mycontext-1"},
		},
		{
			name: "pr-with-label-excludes",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailure,
				Pipeline: v1alpha1.PipelineKindPullRequest,
				PullRequestLabel: &v1alpha1.Pattern{
					Excludes: []string{"dependencies"},
				},
			},
		},
		{
			name: "pr-with-label-includes-no-match",
			cfg: v1alpha1.SlackNotify{
				Kind:     v1alpha1.NotifyKindFailure,
				Pipeline: v1alpha1.PipelineKindPullRequest,
				PullRequestLabel: &v1alpha1.Pattern{
					Includes: []string{"does-not-exist"},
				},
			},
		},
	}

	ns := "jx"
	owner := "myorg"
	repo := "myrepo"

	activities := []*jenkinsv1.PipelineActivity{
		testpipelines.CreateTestPipelineActivity(ns, owner, repo, "main", "", "1", jenkinsv1.ActivityStatusTypeFailed),
		testpipelines.CreateTestPipelineActivity(ns, owner, repo, "main", "", "2", jenkinsv1.ActivityStatusTypeRunning),
		testpipelines.CreateTestPipelineActivity(ns, owner, repo, "PR-1234", "mycontext", "1", jenkinsv1.ActivityStatusTypeFailed),
		testpipelines.CreateTestPipelineActivity(ns, owner, repo, "PR-4567", "mycontext", "1", jenkinsv1.ActivityStatusTypeSucceeded),
	}

	var jxObjects []runtime.Object
	for _, pa := range activities {
		jxObjects = append(jxObjects, pa)
	}
	jxClient := fakejx.NewSimpleClientset(jxObjects...)
	scmClient, scmData := fakescm.NewDefault()

	for _, prNumber := range []int{1234, 4567} {
		var labels []*scm.Label
		if prNumber == 1234 {
			labels = append(labels, &scm.Label{
				Name: "mylabel",
			}, &scm.Label{
				Name: "dependencies",
			})
		}
		scmData.PullRequests[prNumber] = &scm.PullRequest{
			Number: prNumber,
			Title:  "my awesome pull request",
			Body:   "some text",
			Source: "my-branch",
			Labels: labels,
		}
	}

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
