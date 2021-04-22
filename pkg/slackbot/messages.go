package slackbot

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jenkins-x-plugins/jx-gitops/pkg/apis/gitops/v1alpha1"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/jenkins-x/jx-helpers/v3/pkg/scmhelpers"
	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jenkins-x-plugins/jx-gitops/pkg/sourceconfigs"
	"k8s.io/apimachinery/pkg/types"

	"github.com/jenkins-x-plugins/jx-changelog/pkg/users"

	"github.com/pkg/errors"

	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/slack-go/slack"
)

func (o *Options) PipelineMessage(activity *jenkinsv1.PipelineActivity) error {
	if activity.Name == "" {
		return fmt.Errorf("PipelineActivity name cannot be empty")
	}

	cfg := o.getSlackConfigForPipeline(activity)
	if cfg == nil || cfg.Channel == "" {
		log.Logger().Infof("no slack configuration for %s", activity.Name)
		return nil
	}
	channel := channelName(cfg.Channel)
	enabled, pullRequest, resolver, err := o.NotifyPipeline(activity, cfg)
	if err != nil {
		return errors.Wrapf(errors.WithStack(err), "failed to verify if message should be sent")
	}
	if !enabled {
		return nil
	}

	options, createIfMissing, err := o.createPipelineMessage(activity, pullRequest)
	if err != nil {
		return err
	}
	if cfg.Channel != "" {
		err = o.postMessage(channel, false, pipelineMessageType, activity, nil, options, createIfMissing)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error posting cfg for %s to channel %s", activity.Name,
				channel))
		}
		log.Logger().Infof("Channel message sent to %s\n", cfg.Channel)
	}
	if cfg.DirectMessage.ToBool() {
		if pullRequest != nil {
			id, err := o.resolveGitUserToSlackUser(&pullRequest.Author, resolver)
			if err != nil {
				return errors.Wrapf(err, "Cannot resolve Slack ID for Git user %s", pullRequest.Author.Name)
			}
			if id != "" {
				err = o.postMessage(id, true, pipelineMessageType, activity, nil, options, createIfMissing)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("error sending direct pipeline for %s to %s", activity.Name,
						id))
				}
				log.Logger().Infof("Direct message sent to %s\n", pullRequest.Author.Name)
			}
		}
	}
	return nil
}

func (o *Options) getSlackConfigForPipeline(activity *jenkinsv1.PipelineActivity) *v1alpha1.SlackNotify {
	ps := &activity.Spec
	gitServer := ""
	owner := ps.GitOwner
	repoName := ps.GitRepository

	repoConfig := sourceconfigs.GetOrCreateRepositoryFor(o.SourceConfigs, gitServer, owner, repoName)
	return repoConfig.Slack
}

func (o *Options) ReviewRequestMessage(activity *jenkinsv1.PipelineActivity) error {
	if activity.Name == "" {
		return fmt.Errorf("PipelineActivity name cannot be empty")
	}

	prn, _, err := getPullRequestNumber(activity)
	if err != nil {
		return errors.Wrapf(err, "getting pull request number %s", activity.Name)
	}
	if prn <= 0 {
		return nil
	}
	ctx := context.TODO()
	cfg := o.getSlackConfigForPipeline(activity)
	if cfg == nil || cfg.Channel == "" {
		return nil
	}

	enabled, pullRequest, resolver, err := o.NotifyPipeline(activity, cfg)
	if err != nil {
		return errors.WithStack(err)
	}
	if !enabled {
		return nil
	}

	log.Logger().Infof("Preparing review request message for %s\n", activity.Name)
	oldestActivity, latestActivity, all, err := o.findPipelineActivities(ctx, activity)
	if err != nil {
		return err
	}
	buildNumber, err := strconv.Atoi(CreatePipelineDetails(activity).Build)
	if err != nil {
		return err
	}
	latestBuildNumber := -1
	if latestActivity != nil {
		// TODO Some activities could be missing the labels that identify them properly,
		// in that case just display what we have?
		latestBuildNumber, err = strconv.Atoi(CreatePipelineDetails(latestActivity).Build)
	}
	if oldestActivity == nil {
		// TODO Some activities could be missing the labels that identify them so what do we do?
		// We at least try to not error
		oldestActivity = activity
	}
	if buildNumber >= latestBuildNumber {
		attachments, reviewers, buildStatus, err := o.createReviewersMessage(activity, cfg.NotifyReviewers.ToBool(),
			pullRequest, resolver)
		if err != nil {
			return err
		}
		createIfMissing := true
		if buildStatus == defaultStatuses.Merged || buildStatus == defaultStatuses.Closed {
			createIfMissing = false
		}
		if attachments != nil {
			options := []slack.MsgOption{
				slack.MsgOptionAttachments(attachments...),
			}
			if cfg.Channel != "" {
				channel := channelName(cfg.Channel)
				err := o.postMessage(channel, false, pullRequestReviewMessageType, oldestActivity,
					all, options, createIfMissing)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("error posting PR review request for %s to channel %s",
						activity.Name,
						channel))
				}
			}
			if cfg.DirectMessage.ToBool() && cfg.NotifyReviewers.ToBool() {
				for _, user := range reviewers {
					if user != nil {
						err = o.postMessage(user.ID, true, pullRequestReviewMessageType, oldestActivity,
							all, options, createIfMissing)
						if err != nil {
							return errors.Wrap(err, fmt.Sprintf("error sending direct PR review request for %s to %s",
								activity.Name,
								user.ID))
						}
					}
				}

			}
		}
	} else {
		log.Logger().Infof("Skipping %v as it is older than latest build number %d\n", activity.Name,
			latestBuildNumber)
	}
	return nil
}

func (o *Options) isLgtmRepo(_ *jenkinsv1.PipelineActivity) (bool, error) {
	/*
		options := prow.Options{
			KubeClient: o.KubeClient,
			NS:         o.Namespace,
		}
		cfg, _, err := options.GetProwConfig()
		if err != nil {
			return false, errors.Wrapf(err, "getting prow config")
		}
		pipeDetails := CreatePipelineDetails(activity)
		for _, query := range cfg.Keeper.Queries {
			if query.ForRepo(pipeDetails.GitOwner, pipeDetails.GitRepository) {
				if util.Contains(query.Labels, "lgtm") {
					return true, nil
				}
			}
		}
	*/
	return false, nil
}

func (o *Options) findPipelineActivities(ctx context.Context, activity *jenkinsv1.PipelineActivity) (oldest *jenkinsv1.PipelineActivity, latest *jenkinsv1.PipelineActivity, all []jenkinsv1.PipelineActivity, err error) {
	// This is the trigger activity. Working out the correct slack message is a bit tricky,
	// as we have a 1:n mapping between PRs and PipelineActivities (which store the message info).
	// The algorithm in use just picks the earliest pipeline activity as determined by build number
	prn, _, err := getPullRequestNumber(activity)
	if err != nil {
		return nil, nil, nil, err
	}

	pipelineDetails := CreatePipelineDetails(activity)

	acts, err := o.getPipelineActivities(ctx, pipelineDetails.GitOwner, pipelineDetails.GitRepository, prn)

	if err != nil {
		return nil, nil, nil, err
	}
	if len(acts.Items) > 0 {
		sort.Sort(byBuildNumber(acts.Items))
		return &acts.Items[0], &acts.Items[len(acts.Items)-1], acts.Items, nil
	} else {
		log.Logger().Warnf("No pipeline activities exist for %s/%s/pr-%d", pipelineDetails.GitOwner, pipelineDetails.GitRepository, prn)
	}
	return nil, nil, nil, nil
}

func getStatus(overrideStatus *Status, defaultStatus *Status) *Status {
	if overrideStatus == nil {
		return defaultStatus
	}
	return overrideStatus
}

// createReviewersMessage will return a slackapp message notifying reviewers of a PR, or nil if the activity is not a PR
func (o *Options) createReviewersMessage(activity *jenkinsv1.PipelineActivity, notifyReviewers bool, pr *scm.PullRequest, resolver *users.GitUserResolver) ([]slack.Attachment, []*slack.User, *Status, error) {
	author, err := resolver.Resolve(&pr.Author)
	if err != nil {
		return nil, nil, nil, errors.WithStack(err)
	}
	if author == nil || pr == nil {
		return nil, nil, nil, nil
	}
	var attachments []slack.Attachment
	var actions []slack.AttachmentAction
	var fallback []string
	status := pipelineStatus(activity)

	authorName, err := o.mentionOrLinkUser(author)
	if err != nil {
		return nil, nil, nil, err
	}

	mentions := make([]string, 0)
	reviewers := make([]*slack.User, 0)
	if notifyReviewers {

		// Match requested requested reviewers to slack users (if possible)
		for i := range pr.Reviewers {
			r := &pr.Reviewers[i]
			u, err := resolver.Resolve(r)
			if err != nil {
				return nil, nil, nil, errors.Wrapf(err, "resolving %s user %s as Jenkins X user",
					resolver.GitProviderKey(), r.Login)
			}
			if u != nil {
				mention, err := o.mentionOrLinkUser(u)
				if err != nil {
					return nil, nil, nil, errors.Wrapf(err,
						"generating mention or link for user record %s with email %s", u.Name, u.Email)
				}
				mentions = append(mentions, mention)
			}
		}
	}

	// The default state is not approved
	reviewStatus := getStatus(o.Statuses.NotApproved, defaultStatuses.NotApproved)

	// A bit of a hacky way to do this,
	// but until we get a better CRD based interface to the prow this will work
	lgtmRepo, err := o.isLgtmRepo(activity)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "checking if repo for %s is configured for lgtm", activity.Name)
	}
	if lgtmRepo {
		if containsOneOf(pr.Labels, "lgtm") {
			reviewStatus = getStatus(o.Statuses.LGTM, defaultStatuses.LGTM)
		}
	} else {
		if containsOneOf(pr.Labels, "approved") {
			reviewStatus = getStatus(o.Statuses.Approved, defaultStatuses.Approved)
		}
	}
	if containsOneOf(pr.Labels, "do-not-merge/hold") {
		reviewStatus = getStatus(o.Statuses.Hold, defaultStatuses.Hold)
	}
	if containsOneOf(pr.Labels, "needs-ok-to-test") {
		reviewStatus = getStatus(o.Statuses.NeedsOkToTest, defaultStatuses.NeedsOkToTest)
	}

	// The default build state is unknown
	buildStatus := getStatus(o.Statuses.Unknown, defaultStatuses.Unknown)
	if pr.Merged {
		buildStatus = getStatus(o.Statuses.Merged, defaultStatuses.Merged)
	} else if pr.Closed {
		buildStatus = getStatus(o.Statuses.Closed, defaultStatuses.Closed)
	} else {
		switch activity.Spec.Status {
		case jenkinsv1.ActivityStatusTypePending:
			buildStatus = getStatus(o.Statuses.Pending, defaultStatuses.Pending)
		case jenkinsv1.ActivityStatusTypeRunning:
			buildStatus = getStatus(o.Statuses.Running, defaultStatuses.Running)
		case jenkinsv1.ActivityStatusTypeSucceeded:
			buildStatus = getStatus(o.Statuses.Succeeded, defaultStatuses.Succeeded)
		case jenkinsv1.ActivityStatusTypeFailed:
			buildStatus = getStatus(o.Statuses.Failed, defaultStatuses.Failed)
		case jenkinsv1.ActivityStatusTypeError:
			buildStatus = getStatus(o.Statuses.Errored, defaultStatuses.Errored)
		case jenkinsv1.ActivityStatusTypeAborted:
			buildStatus = getStatus(o.Statuses.Aborted, defaultStatuses.Aborted)
		}
	}

	mentionsString := strings.Join(mentions, " ")
	pleaseText := "please"
	if len(mentions) == 0 {
		pleaseText = "Please"
	}
	messageText := fmt.Sprintf("%s %s review %s created on %s by %s",
		mentionsString,
		pleaseText,
		link(fmt.Sprintf("Pull Request %s (%s)", pullRequestName(pr.Link), pr.Title), pr.Link),
		repositoryName(activity),
		authorName)
	attachment := slack.Attachment{
		CallbackID: "preview:" + activity.Name,
		Color:      attachmentColor(status),
		Text:       messageText,

		Fallback: strings.Join(fallback, ", "),
		Actions:  actions,
		Fields: []slack.AttachmentField{
			{
				Value: fmt.Sprintf("%s %s", reviewStatus.Emoji, reviewStatus.Text),
				Short: true,
			},
			{
				Value: fmt.Sprintf("%s %s", buildStatus.Emoji, buildStatus.Text),
				Short: true,
			},
		},
	}
	updatedEpochTime := getLastUpdatedTime(pr, activity)
	if updatedEpochTime > 0 {
		attachment.Ts = json.Number(strconv.FormatInt(updatedEpochTime, 10))
	}

	attachments = append(attachments, attachment)

	return attachments, reviewers, buildStatus, nil
}

func getLastUpdatedTime(pr *scm.PullRequest, activity *jenkinsv1.PipelineActivity) int64 {
	updatedEpochTime := int64(-1)
	if pr != nil {
		updatedEpochTime = pr.Updated.Unix()
	} else {
		updatedEpochTime = activity.CreationTimestamp.Unix()
	}

	// Check if there is a started or completion timestamp that is more recent
	if activity != nil && activity.Spec.StartedTimestamp != nil {
		if start := activity.Spec.StartedTimestamp.Unix(); start > updatedEpochTime {
			updatedEpochTime = start
		}
	}
	if activity != nil && activity.Spec.CompletedTimestamp != nil {
		if completed := activity.Spec.CompletedTimestamp.Unix(); completed > updatedEpochTime {
			updatedEpochTime = completed
		}
	}
	return updatedEpochTime
}

func containsOneOf(a []*scm.Label, x ...string) bool {
	for _, n := range a {
		for _, y := range x {
			if y == n.Name {
				return true
			}
		}
	}
	return false
}

func (o *Options) createPipelineMessage(activity *jenkinsv1.PipelineActivity, pr *scm.PullRequest) ([]slack.MsgOption, bool, error) {
	format := &o.MessageFormat
	spec := &activity.Spec
	status := pipelineStatus(activity)
	icon := pipelineIcon(status)
	pipelineName, err := pipelineName(activity)
	if err != nil {
		return nil, false, errors.Wrapf(err, "getting pipeline name for %s", activity.Name)
	}
	messageText := icon + pipelineName + " " + repositoryName(activity)
	if prn, _, err := getPullRequestNumber(activity); err != nil {
		return nil, false, err
	} else if prn > 0 {
		messageText = fmt.Sprintf("%s : PR %s", messageText, link(pullRequestName(pr.Link), pr.Link))
	}

	buildURL := spec.BuildURL
	if format.DashboardURL != "" {
		owner := spec.GitOwner
		repoName := spec.GitRepository
		branch := spec.GitBranch
		build := spec.Build
		if owner != "" && repoName != "" && branch != "" && build != "" {
			buildURL = stringhelpers.UrlJoin(format.DashboardURL, owner, repoName, branch, build)
		}
	}

	buildNumber := link("#"+spec.Build, buildURL)
	pipelineCtx := spec.Context
	if pipelineCtx == "" {
		pipelineCtx = "Build"
	}
	messageText = fmt.Sprintf("%s (%s %s)", messageText, pipelineCtx, buildNumber)

	// lets ignore old pipelines
	dayAgo := time.Now().Add((-24) * time.Hour).Unix()
	createIfMissing := true
	lastUpdatedTime := getLastUpdatedTime(nil, activity)
	if lastUpdatedTime < dayAgo {
		createIfMissing = false
	}

	var attachments []slack.Attachment
	var actions []slack.AttachmentAction
	versionPrefix := spec.Version
	if versionPrefix != "" {
		versionPrefix += " "
	}
	var fallback []string
	if format.ShowRepository && spec.GitURL != "" {
		fallback = append(fallback, "Repo: "+spec.GitURL)
		actions = append(actions, slack.AttachmentAction{
			Type: "button",
			Text: "Repository",
			URL:  spec.GitURL,
		})
	}
	if format.ShowBuildURL && buildURL != "" {
		fallback = append(fallback, "Build: "+buildURL)
		actions = append(actions, slack.AttachmentAction{
			Type: "button",
			Text: "Pipeline",
			URL:  buildURL,
		})
	}
	if format.ShowBuildLogs && spec.BuildLogsURL != "" {
		fallback = append(fallback, "Logs: "+spec.BuildLogsURL)
		actions = append(actions, slack.AttachmentAction{
			Type: "button",
			Text: "Build Logs",
			URL:  strings.Replace(spec.BuildLogsURL, "gs://", "https://storage.cloud.google.com/", -1),
		})
	}
	if format.ShowReleaseNotes && spec.ReleaseNotesURL != "" {
		fallback = append(fallback, "Release Notes: "+spec.BuildLogsURL)
		actions = append(actions, slack.AttachmentAction{
			Type: "button",
			Text: "Release Notes",
			URL:  spec.ReleaseNotesURL,
		})
	}
	attachment := slack.Attachment{
		CallbackID: "pipelineactivity:" + activity.Name,
		Color:      attachmentColor(status),
		Title:      messageText,
		Fallback:   strings.Join(fallback, ", "),
		Actions:    actions,
	}

	if lastUpdatedTime > 0 {
		attachment.Ts = json.Number(strconv.FormatInt(lastUpdatedTime, 10))
	}

	attachments = append(attachments, attachment)

	if format.ShowSteps {
		for _, step := range spec.Steps {
			stepAttachments := o.createAttachments(activity, &step)
			if len(stepAttachments) > 0 {
				attachments = append(attachments, stepAttachments...)
			}
		}
	}
	options := []slack.MsgOption{
		slack.MsgOptionAttachments(attachments...),
	}
	return options, createIfMissing, nil
}

func (o *Options) getSlackUserID(gitUser *scm.User, resolver *users.GitUserResolver) (string, error) {
	if gitUser == nil {
		return "", fmt.Errorf("user cannot be nil")
	}
	user, err := resolver.Resolve(gitUser)
	if err != nil {
		return "", err
	}

	return o.SlackUserResolver.SlackUserLogin(user)
}

// getPullRequestNumber extracts the pull request number from the activity or returns 0 if it's not a pull request
func getPullRequestNumber(activity *jenkinsv1.PipelineActivity) (int, *PipelineDetails, error) {
	pipelineDetails := CreatePipelineDetails(activity)
	if strings.HasPrefix(strings.ToLower(pipelineDetails.BranchName), "pr-") {
		prn, err := strconv.Atoi(strings.TrimPrefix(strings.ToLower(pipelineDetails.BranchName), "pr-"))
		return prn, pipelineDetails, err
	}
	return 0, pipelineDetails, nil
}

func (o *Options) postMessage(channel string, directMessage bool, messageType string,
	activity *jenkinsv1.PipelineActivity, all []jenkinsv1.PipelineActivity, options []slack.MsgOption,
	createIfMissing bool) error {
	timestamp := o.FakeTimestamp
	messageRef := o.findMessageRefViaAnnotations(activity, channel, messageType)
	channelId := channel

	if messageRef == nil {
		// couldn't find the message ref on a Pipeline Activity so attempt to find the message ref in memory
		messageRef = o.Timestamps[channel][activity.Name]
	}
	if messageRef != nil {
		timestamp = messageRef.Timestamp
		channelId = messageRef.ChannelID
	}

	if o.Timestamps == nil {
		o.Timestamps = map[string]map[string]*MessageReference{}
	}
	if _, ok := o.Timestamps[channel]; !ok {
		o.Timestamps[channel] = make(map[string]*MessageReference, 0)
	}

	if directMessage {
		channel, _, _, err := o.SlackClient.OpenConversation(&slack.OpenConversationParameters{
			Users: []string{
				channel,
			},
		})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("(open converation channelId: %s)", channelId))
		}
		channelId = channel.ID
	}
	post := true
	if timestamp != "" {
		options = append(options, slack.MsgOptionUpdate(timestamp))
		log.Logger().Infof("Updating message for %s with timestamp %s\n", activity.Name, timestamp)
	} else {
		if createIfMissing {
			log.Logger().Infof("Creating new message for %s\n", activity.Name)
		} else {
			log.Logger().Infof("No existing message to update, ignoring, for %s\n", activity.Name)
			post = false
		}

	}
	if post {
		ctx := context.TODO()

		channelId, timestamp, _, err := o.SlackClient.SendMessage(channelId, options...)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("(post channelId: %s, timestamp: %s)", channelId, timestamp))
		}
		o.Timestamps[channel][activity.Name] = &MessageReference{
			ChannelID: channelId,
			Timestamp: timestamp,
		}
		key := annotationKey(channel, messageType)
		value := annotationValue(channelId, timestamp)
		if all == nil {
			if activity.Annotations[key] != value {
				err = o.annotatePipelineActivity(ctx, activity, key, value)
				if err != nil {
					return err
				}
			}
		} else {
			for _, a := range all {
				if a.Annotations[key] != value {
					err = o.annotatePipelineActivity(ctx, &a, key, value)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

//getPullRequest will return the PullRequestInfo for the activity, or nil if it's not a pull request
func (o *Options) getPullRequest(ctx context.Context, activity *jenkinsv1.PipelineActivity, prn int) (pr *scm.PullRequest, resolver *users.GitUserResolver, err error) {
	if activity.Spec.GitURL == "" {
		return nil, nil, fmt.Errorf("no GitURL on PipelineActivity %s", activity.Name)
	}
	gitInfo, err := giturl.ParseGitURL(activity.Spec.GitURL)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to parse git URL %s", activity.Spec.GitURL)
	}

	resolver = &users.GitUserResolver{
		GitProvider: o.ScmClient,
	}
	fullName := scm.Join(gitInfo.Organisation, gitInfo.Name)
	pr, _, err = o.ScmClient.PullRequests.Find(ctx, fullName, prn)
	if scmhelpers.IsScmNotFound(err) {
		return pr, resolver, nil
	}
	return pr, resolver, err
}

func (o *Options) findMessageRefViaAnnotations(activity *jenkinsv1.PipelineActivity,
	channel string, messageType string) *MessageReference {
	annotations := activity.Annotations
	if annotations != nil {
		key := annotationKey(channel, messageType)
		value := annotations[key]
		if value != "" {
			values := strings.SplitN(value, "/", 2)
			if len(values) > 1 {
				log.Logger().Infof("Found annotation %s: %s for %s\n", key, value, activity.Name)
				return &MessageReference{values[0], values[1]}
			}
		}
		log.Logger().Infof("Could not find annotation %s for %s\n", key, activity.Name)
	}
	return nil
}

func annotationKey(channel string, messageType string) string {
	return fmt.Sprintf("%s-%s/%s", SlackAnnotationPrefix, messageType, strings.TrimPrefix(channel, "#"))
}

func annotationValue(channelId string, timestamp string) string {
	return fmt.Sprintf("%s/%s", channelId, timestamp)
}

func (o *Options) createAttachments(activity *jenkinsv1.PipelineActivity,
	step *jenkinsv1.PipelineActivityStep) []slack.Attachment {
	stage := step.Stage
	promote := step.Promote
	if stage != nil {
		return o.createStageAttachments(activity, stage)
	} else if promote != nil {
		return o.createPromoteAttachments(activity, promote)
	}
	return []slack.Attachment{}

}

func (o *Options) createStageAttachments(activity *jenkinsv1.PipelineActivity,
	stage *jenkinsv1.StageActivityStep) []slack.Attachment {
	name := stage.Name
	if name == "" {
		name = "Stage"
	}
	version := activity.Spec.Version
	if name == "Release" {
		if version != "" {
			name = "release " + link(version, activity.Spec.ReleaseNotesURL)
		}
	}

	attachments := []slack.Attachment{
		o.createStepAttachment(stage.CoreActivityStep, name, "", ""),
	}
	if stage.CoreActivityStep.Name != "meta pipeline" {
		for _, step := range stage.Steps {
			// filter out tekton generated steps
			if isUserPipelineStep(step.Name) {
				attachments = append(attachments, o.createStepAttachment(step, "", "", ""))
			}
		}
	}

	return attachments
}

func isUserPipelineStep(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	ss := strings.Fields(name)
	firstWord := ss[0]

	if containsIgnoreCase(knownPipelineStageTypes, firstWord) {
		return true
	}
	return false
}

func (o *Options) createStepAttachment(step jenkinsv1.CoreActivityStep, name string, description string,
	iconUrl string) slack.Attachment {
	text := step.Description
	if description != "" {
		if text == "" {
			text = description
		} else {
			text += description
		}
	}
	textName := strings.Title(name)
	if textName == "" {
		textName = step.Name
	}

	textName = getUserFriendlyMapping(textName)

	stepStatus := step.Status
	textMessage := o.statusString(stepStatus) + " " + textName
	if text != "" {
		textMessage += " " + text
	}

	return slack.Attachment{
		Text:       textMessage,
		FooterIcon: iconUrl,
		MarkdownIn: []string{"fields"},
		Color:      attachmentColor(stepStatus),
	}
}

func (o *Options) createPromoteAttachments(activity *jenkinsv1.PipelineActivity, parent *jenkinsv1.PromoteActivityStep) []slack.Attachment {
	envName := strings.Title(parent.Environment)
	attachments := []slack.Attachment{
		o.createStepAttachment(parent.CoreActivityStep, "promote to *"+envName+"*", "", ""),
	}

	pullRequest := parent.PullRequest
	update := parent.Update
	if pullRequest != nil {
		iconUrl := pullRequestIcon(pullRequest)
		attachments = append(attachments, o.createStepAttachment(pullRequest.CoreActivityStep, "PR", describePromotePullRequest(activity, pullRequest), iconUrl))
	}
	if update != nil {
		attachments = append(attachments, o.createStepAttachment(update.CoreActivityStep, "update", describePromoteUpdate(update), ""))
	}
	appURL := parent.ApplicationURL
	if appURL != "" {
		attachments = append(attachments, o.createStepAttachment(update.CoreActivityStep, ":star: application now in "+link(envName, appURL), "", ""))
	}
	return attachments
}

func (o *Options) annotatePipelineActivity(ctx context.Context, activity *jenkinsv1.PipelineActivity, key string, value string) error {
	newActivity := activity.DeepCopy()
	if newActivity.Annotations == nil {
		newActivity.Annotations = make(map[string]string)
	}
	newActivity.Annotations[key] = value
	patch, err := CreatePatch(activity, newActivity)
	if err != nil {
		return errors.Wrapf(err, "creating patch to add annotation %s=%s to %s", key, value, activity.Name)
	}
	jsonPatch, err := json.Marshal(patch)
	if err != nil {
		return errors.Wrapf(err, "marshaling patch to add annotation %s=%s to %s", key, value, activity.Name)
	}
	_, err = o.JXClient.JenkinsV1().PipelineActivities(o.Namespace).Patch(ctx, activity.Name, types.JSONPatchType,
		jsonPatch, metav1.PatchOptions{})
	return err
}

func describePromotePullRequest(activity *jenkinsv1.PipelineActivity, promote *jenkinsv1.PromotePullRequestStep) string {
	description := ""
	if promote.PullRequestURL != "" {
		description += " " + link(pullRequestName(promote.PullRequestURL), promote.PullRequestURL)
	}
	if promote.MergeCommitSHA != "" {
		// lets not use a URL
		gitUrl := activity.Spec.GitURL
		description += " merged " + mergeShaText(gitUrl, promote.MergeCommitSHA)
	}
	return description
}

func pullRequestName(url string) string {
	idx := strings.LastIndex(url, "/")
	if idx > 0 {
		return url[idx+1:]
	}
	return url
}

func pipelineName(activity *jenkinsv1.PipelineActivity) (string, error) {
	name := activity.Spec.Pipeline
	if strings.HasSuffix(name, "/master") {
		return "Release Pipeline", nil
	}
	prn, _, err := getPullRequestNumber(activity)
	if err != nil {
		return "", errors.Wrapf(err, "getting pull request number from %s", activity.Name)
	}
	if prn > 0 {
		return "Pull Request Pipeline", nil
	}
	return "Pipeline", nil
}

func repositoryName(act *jenkinsv1.PipelineActivity) string {
	details := CreatePipelineDetails(act)
	gitURL := act.Spec.GitURL
	ownerURL := strings.TrimSuffix(gitURL, "/")
	idx := strings.LastIndex(ownerURL, "/")
	if idx > 0 {
		ownerURL = ownerURL[0 : idx+1]
	}
	return link(details.GitOwner, ownerURL) + "/" + link(details.GitRepository, gitURL)
}

type PipelineDetails struct {
	GitOwner      string
	GitRepository string
	BranchName    string
	Pipeline      string
	Build         string
	Context       string
}

// CreatePipelineDetails creates a PipelineDetails object populated from the activity
func CreatePipelineDetails(activity *jenkinsv1.PipelineActivity) *PipelineDetails {
	spec := &activity.Spec
	repoOwner := spec.GitOwner
	repoName := spec.GitRepository
	buildNumber := spec.Build
	branchName := spec.GitBranch
	pipelineCtx := spec.Context
	pipeline := spec.Pipeline
	if pipeline != "" {
		paths := strings.Split(pipeline, "/")
		if len(paths) > 2 {
			if repoOwner == "" {
				repoOwner = paths[0]
			}
			if repoName == "" {
				repoName = paths[1]
			}
			if branchName == "" {
				branchName = paths[2]
			}
		}
	}
	if branchName == "" {
		branchName = "master"
	}
	if pipeline == "" && (repoName != "" && repoOwner != "") {
		pipeline = repoOwner + "/" + repoName + "/" + branchName
	}
	return &PipelineDetails{
		GitOwner:      repoOwner,
		GitRepository: repoName,
		BranchName:    branchName,
		Pipeline:      pipeline,
		Build:         buildNumber,
		Context:       pipelineCtx,
	}
}

func (o *Options) mentionOrLinkUser(user *jenkinsv1.UserDetails) (string, error) {
	id, err := o.SlackUserResolver.SlackUserLogin(user)
	if err != nil {
		return "", err
	}
	if id != "" {
		return mentionUser(id), nil
	}
	if user.Name != "" && user.URL != "" {
		return link(user.Name, user.URL), nil
	}
	if user.Name != "" {
		return user.Name, nil
	}
	return "", nil
}

func channelName(channel string) string {
	if !strings.HasPrefix(channel, "#") {
		return fmt.Sprintf("#%s", channel)
	}
	return channel
}

func link(text string, url string) string {
	if url != "" {
		if text == "" {
			text = url
		}
		return "<" + url + "|" + text + ">"
	} else {
		return text
	}
}

func mergeShaText(gitURL, sha string) string {
	short := sha[0:7]
	cleanUrl := strings.TrimSuffix(gitURL, ".git")
	if cleanUrl != "" {
		cleanUrl = stringhelpers.UrlJoin(cleanUrl, "commit", sha)
	}
	return link(short, cleanUrl)
}

func describePromoteUpdate(promote *jenkinsv1.PromoteUpdateStep) string {
	description := ""
	for _, status := range promote.Statuses {
		url := status.URL
		state := status.Status

		if url != "" && state != "" {
			description += " " + link(pullRequestStatusString(state), url)
		}
	}
	return description
}

func pullRequestStatusString(text string) string {
	title := strings.Title(text)
	switch text {
	case "success":
		return title
	case "error", "failed":
		return title
	default:
		return title
	}
}

func (o *Options) resolveGitUserToSlackUser(user *scm.User, resolver *users.GitUserResolver) (string,
	error) {
	resolved, err := resolver.Resolve(user)
	if err != nil {
		return "", err
	}
	return o.SlackUserResolver.SlackUserLogin(resolved)
}

func (o *Options) statusString(statusType jenkinsv1.ActivityStatusType) string {
	switch statusType {
	case jenkinsv1.ActivityStatusTypeFailed:
		return getStatus(o.Statuses.Failed, defaultStatuses.Failed).Emoji
	case jenkinsv1.ActivityStatusTypeError:
		return getStatus(o.Statuses.Errored, defaultStatuses.Errored).Emoji
	case jenkinsv1.ActivityStatusTypeSucceeded:
		return getStatus(o.Statuses.Succeeded, defaultStatuses.Succeeded).Emoji
	case jenkinsv1.ActivityStatusTypeRunning:
		return getStatus(o.Statuses.Running, defaultStatuses.Running).Emoji
	}
	return ""
}

func attachmentColor(statusType jenkinsv1.ActivityStatusType) string {
	switch statusType {
	case jenkinsv1.ActivityStatusTypeFailed, jenkinsv1.ActivityStatusTypeError:
		return "danger"
	case jenkinsv1.ActivityStatusTypeSucceeded:
		return "good"
	case jenkinsv1.ActivityStatusTypeRunning:
		return "#3AA3E3"
	}
	return ""
}

func pullRequestIcon(step *jenkinsv1.PromotePullRequestStep) string {
	state := "open"
	switch step.Status {
	case jenkinsv1.ActivityStatusTypeFailed, jenkinsv1.ActivityStatusTypeError:
		state = "closed"
	case jenkinsv1.ActivityStatusTypeSucceeded:
		state = "merged"
	}
	return "https://images.atomist.com/rug/pull-request-" + state + ".png"
}

func pipelineStatus(activity *jenkinsv1.PipelineActivity) jenkinsv1.ActivityStatusType {
	statusType := activity.Spec.Status
	switch statusType {
	case jenkinsv1.ActivityStatusTypeFailed, jenkinsv1.ActivityStatusTypeError:
	case jenkinsv1.ActivityStatusTypeSucceeded:
		return statusType
	}
	// lets try find the last status
	for _, step := range activity.Spec.Steps {
		stage := step.Stage
		promote := step.Promote
		if stage != nil {
			statusType = stage.Status
		} else if promote != nil {
			statusType = promote.Status
		}
	}
	return statusType
}

func pipelineIcon(statusType jenkinsv1.ActivityStatusType) string {
	switch statusType {
	case jenkinsv1.ActivityStatusTypeFailed, jenkinsv1.ActivityStatusTypeError:
		return ""
	case jenkinsv1.ActivityStatusTypeSucceeded:
		return ""
	case jenkinsv1.ActivityStatusTypeRunning:
		return ""
	}
	return ""
}

func mentionUser(id string) string {
	return fmt.Sprintf("<@%s>", id)
}
