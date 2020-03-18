package slackbot

var (
	// contains mappings of known pipeline steps to user friendly slack descriptions
	slackMessageMapping = map[string]string{
		"meta pipeline": "Generating pipeline",
	}
)

// returns a user friendly slack description for a pipeline step if one exists
func getUserFriendlyMapping(stepName string) string {
	if slackMessageMapping[stepName] != "" {
		return slackMessageMapping[stepName]
	}
	return stepName
}
