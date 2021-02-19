package slackbot

import (
	"context"
	"github.com/jenkins-x-plugins/jx-changelog/pkg/users"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-gitops/pkg/apis/gitops/v1alpha1"
	"strings"

	"github.com/pkg/errors"

	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
)

// NotifyPipeline returns true if the given pipeline activity matches the configuration
func (o *SlackBotOptions) NotifyPipeline(activity *jenkinsv1.PipelineActivity, cfg *v1alpha1.SlackNotify) (bool, *scm.PullRequest, *users.GitUserResolver, error) {
	enabled := o.shouldSendPipelineMessage(activity, cfg)
	if !enabled {
		return false, nil, nil, nil
	}
	var pr *scm.PullRequest
	var err error
	var resolver *users.GitUserResolver
	pr, resolver, err = o.getPullRequest(context.TODO(), activity)
	if err != nil {
		return false, nil, nil, errors.WithStack(err)
	}
	if pr == nil {
		return false, nil, nil, nil
	}

	var labels []string
	for _, v := range pr.Labels {
		labels = append(labels, v.Name)
	}
	if !cfg.PullRequestLabel.MatchesLabels(labels) {
		log.Logger().Infof("Ignoring %s because it has labels %s\n", activity.Name, strings.Join(labels, ", "))
		return false, nil, nil, nil
	}
	return true, pr, resolver, nil
}
func (o *SlackBotOptions) shouldSendPipelineMessage(activity *jenkinsv1.PipelineActivity, cfg *v1alpha1.SlackNotify) bool {
	failed := activity.Spec.Status == jenkinsv1.ActivityStatusTypeError || activity.Spec.Status == jenkinsv1.ActivityStatusTypeFailed
	succeeded := activity.Spec.Status == jenkinsv1.ActivityStatusTypeSucceeded
	switch cfg.Kind {
	case v1alpha1.NotifyKindNone, v1alpha1.NotifyKindNever:
		return false
	case v1alpha1.NotifyKindAlways:
		return true
	case v1alpha1.NotifyKindFailure:
		return failed
	case v1alpha1.NotifyKindFailureOrFirstSuccess:
		if succeeded {
			// TODO lets find if the last status we logged was a fail...
			flag, err := o.previousPipelineFailed(activity)
			if err != nil {
				log.Logger().Warnf("failed to find if previous pipeline of %s was failed: %s", activity.Name, err.Error())
				return false
			}
			return flag
		}
		return failed
	case v1alpha1.NotifyKindSuccess:
		return succeeded
	default:
		log.Logger().Warnf("invalid notify kind %s", string(cfg.Kind))
		return false
	}
}
