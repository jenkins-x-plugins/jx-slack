package slackbot

var (
	// contains mappings of known pipeline steps to user friendly slack descriptions
	slackMessageMapping = map[string]string{
		"Meta Pipeline": "Generating Pipeline",
	}
)

// returns a user friendly slack description for a pipeline step if one exists
func getUserFriendlyMapping(stepName string) string {
	if slackMessageMapping[stepName] != "" {
		return slackMessageMapping[stepName]
	}
	return stepName
}
