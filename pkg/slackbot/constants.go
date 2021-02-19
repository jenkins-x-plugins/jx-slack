package slackbot

const (
	SlackAnnotationPrefix        = "bot.slack.apps.jenkins-x.io"
	pullRequestReviewMessageType = "pr"
	pipelineMessageType          = "pipeline"
)

var knownPipelineStageTypes = []string{"setup", "setVersion", "preBuild", "build", "postBuild", "promote", "pipeline"}

var defaultStatuses = Statuses{
	Merged: &Status{
		Emoji: ":purple_heart:",
		Text:  "merged",
	},
	Closed: &Status{
		Emoji: ":closed_book:",
		Text:  "closed and not merged",
	},
	Aborted: &Status{
		Emoji: ":red_circle:",
		Text:  "build aborted",
	},
	Errored: &Status{
		Emoji: ":red_circle:",
		Text:  "build errored",
	},
	Failed: &Status{
		Emoji: ":red_circle:",
		Text:  "build failed",
	},
	Approved: &Status{
		Emoji: ":+1:",
		Text:  "approved",
	},
	NotApproved: &Status{
		Emoji: ":wave:",
		Text:  "not approved",
	},
	NeedsOkToTest: &Status{
		Emoji: ":wave:",
		Text:  "needs /ok-to-test",
	},
	Hold: &Status{
		Emoji: ":octagonal_sign:",
		Text:  "hold",
	},
	Pending: &Status{
		Emoji: ":question:",
		Text:  "build pending",
	},
	Running: &Status{
		Emoji: ":white_circle:",
		Text:  "build running",
	},
	Succeeded: &Status{
		Emoji: ":white_check_mark:",
		Text:  "build succeeded",
	},
	LGTM: &Status{
		Emoji: ":+1:",
		Text:  "lgtm",
	},
	Unknown: &Status{
		Emoji: ":grey_question:",
		Text:  "",
	},
}
