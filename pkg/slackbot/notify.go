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

	prn, details, err := getPullRequestNumber(activity)
	if err != nil {
		log.Logger().Warnf("failed to get PullRequest number for activity %s", activity.Name)
	}

	if !o.matchesPipeline(activity, cfg, prn) {
		return false, nil, nil, nil
	}
	if !cfg.Branch.Matches(details.BranchName) {
		log.Logger().Infof("Ignoring %s because it has a different branch: %s\n", activity.Name, details.BranchName)
		return false, nil, nil, nil
	}
	if !cfg.Context.Matches(details.Context) {
		log.Logger().Infof("Ignoring %s because it has a different context: %s\n", activity.Name, details.Context)
		return false, nil, nil, nil
	}

	if prn <= 0 {
		return true, nil, nil, nil
	}
	var pr *scm.PullRequest
	var resolver *users.GitUserResolver
	pr, resolver, err = o.getPullRequest(context.TODO(), activity, prn)
	if err != nil {
		return false, nil, nil, errors.WithStack(err)
	}

	if pr == nil {
		log.Logger().Warnf("no Pull Request found for %s PR %d", activity.Name, prn)
		pr = &scm.PullRequest{}
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

func (o *SlackBotOptions) matchesPipeline(activity *jenkinsv1.PipelineActivity, cfg *v1alpha1.SlackNotify, prn int) bool {
	switch cfg.Pipeline {
	case v1alpha1.PipelineKindAll, v1alpha1.PipelineKindNone:
		return true
	case v1alpha1.PipelineKindRelease:
		return prn <= 0
	case v1alpha1.PipelineKindPullRequest:
		return prn > 0
	default:
		log.Logger().Infof("unknown pipeline kind %s for activity %s", string(cfg.Pipeline), activity.Name)
		return false
	}
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
