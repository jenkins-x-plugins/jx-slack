package slackbot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/pluginhelp"
	"k8s.io/test-infra/prow/pluginhelp/externalplugins"

	"github.com/jenkins-x/jx/v2/pkg/log"
	"k8s.io/test-infra/prow/github"
)

func (s *SlackBots) ProwExternalPluginServer() error {
	isLighthouse := true
	_, err := s.KubeClient.AppsV1().Deployments(s.Namespace).Get("lighthouse-keeper", metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			isLighthouse = false
		} else {
			return err
		}
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if isLighthouse {
			err := s.handleLighthouseEvent(r)
			if err != nil {
				log.Logger().WithError(err).Error("Error parsing Lighthouse event")
			}
		} else {
			webhookToken, err := s.getWebHookToken()
			eventType, eventGUID, payload, ok, errCode := github.ValidateWebhook(w, r, webhookToken)
			if !ok {
				if errCode == 200 {
					// then it's a health check
				} else if webhookToken == nil {
					if err == nil {
						log.Logger().Errorf("Unable to load HMAC token as not specified\n")
					} else {
						log.Logger().Errorf("Unable to load HMAC token as %v\n", err)
					}
				} else {
					log.Logger().Errorf("Error validating WebHook, error code is %d\n", errCode)
				}
				return
			}
			fmt.Fprint(w, "Event received. Have a nice day.")
			if err := s.handleProwEvent(eventType, eventGUID, payload); err != nil {
				log.Logger().Error("Error parsing event.")
			}
		}
	})
	externalplugins.ServeExternalPluginHelp(http.DefaultServeMux, logrus.StandardLogger().WithField("plugin",
		"slackbot"),
		helpProvider)
	return http.ListenAndServe("0.0.0.0:"+strconv.Itoa(s.Port), nil)
}

func (s *SlackBots) handleProwPullRequest(pr github.PullRequestEvent) error {
	if pr.Action == github.PullRequestActionReviewRequested || pr.Action == github.
		PullRequestActionReviewRequestRemoved {
		return s.processPR(pr.Repo.Owner.Login, pr.Repo.Name, pr.Number)
	}
	return nil
}

func (s *SlackBots) handleLighthousePullRequest(pr *scm.PullRequestHook) error {
	if pr.Action == scm.ActionReviewRequested || pr.Action == scm.ActionReviewRequestRemoved {
		return s.processPR(pr.Repo.Namespace, pr.Repo.Name, pr.PullRequest.Number)
	}
	return nil
}

func (s *SlackBots) processPR(owner, repo string, number int) error {
	// This is the trigger. Working out the correct slack message is a bit tricky,
	// as we have a 1:n mapping between PRs and PipelineActivities (which store the message info).
	// The algorithm in use just picks the earliest pipeline activity as determined by build number
	acts, err := s.getPipelineActivities(owner, repo, number)

	if err != nil {
		return err
	}
	if len(acts.Items) > 0 {
		sort.Sort(byBuildNumber(acts.Items))
		act := acts.Items[0]
		// now we can just run the bots for the activity
		for _, bot := range s.Items {
			err := bot.ReviewRequestMessage(&act)
			if err != nil {
				return err
			}
		}
	} else {
		log.Logger().Warnf("No pipeline activities exist for %s/%s/pr-%d", owner, repo, number)
	}

	return nil
}

func (s *SlackBots) handleProwEvent(eventType, eventGUID string, payload []byte) error {
	switch eventType {
	case "pull_request":
		var pr github.PullRequestEvent
		if err := json.Unmarshal(payload, &pr); err != nil {
			return err
		}
		go func() {
			if err := s.handleProwPullRequest(pr); err != nil {
				log.Logger().Infof("Refreshing slack message failed because %v\n", err)
			}
		}()
	default:
		logrus.Debugf("skipping event of type %q", eventType)
	}
	return nil
}

func (s *SlackBots) handleLighthouseEvent(r *http.Request) error {
	if r.Method != http.MethodPost {
		// liveness probe etc
		log.Logger().WithField("method", r.Method).Debug("invalid http method so returning 200")
		return nil
	}
	kind := os.Getenv("GIT_KIND")
	if kind == "" {
		kind = "github"
	}
	serverURL := os.Getenv("GIT_SERVER")

	client, err := factory.NewClient(kind, serverURL, "")
	if err != nil {
		log.Logger().Errorf("Could not create client to parse webhook: %s", err)
		return err
	}
	webhook, err := client.Webhooks.Parse(r, func(webhook scm.Webhook) (string, error) {
		tokenBytes, err := s.getWebHookToken()
		return string(tokenBytes), err
	})
	if err != nil {
		log.Logger().Warnf("failed to parse webhook: %s", err.Error())
		return err
	}
	if webhook == nil {
		log.Logger().Error("no webhook was parsed")
		return nil
	}
	prHook, ok := webhook.(*scm.PullRequestHook)
	if ok {
		go func() {
			if err := s.handleLighthousePullRequest(prHook); err != nil {
				log.Logger().Infof("Refreshing slack message failed because %v\n", err)
			}
		}()
	} else {
		log.Logger().Debugf("skipping event of type %q", webhook.Kind())
	}
	return nil
}

func helpProvider(enabledRepos []string) (*pluginhelp.PluginHelp, error) {
	pluginHelp := &pluginhelp.PluginHelp{
		Description: `The slackbot plugin for Jenkins X Labs is used for communicating between PRs and Slack. 
It will notify any reviewers on slack when a PR changes state`,
	}
	return pluginHelp, nil
}

func (s *SlackBots) getWebHookToken() ([]byte, error) {
	if s.HmacSecretName == "" || s.HmacSecretName == "REPLACE_ME" {
		// Not configured
		return nil, nil
	}
	secret, err := s.KubeClient.CoreV1().Secrets(s.Namespace).Get(s.HmacSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return secret.Data["hmac"], nil
}
