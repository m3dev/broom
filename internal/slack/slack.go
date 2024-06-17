package slack

import (
	"fmt"

	"github.com/slack-go/slack"
)

const (
	slackUserName = "Broom"
	slackIconURL  = "https://raw.githubusercontent.com/Kitsuya0828/broom-assets/main/logo_dark.png"
)

type ContainerUpdate struct {
	Name            string `json:"name"`
	BeforeMemory    string `json:"before_memory"`
	AfterMemory     string `json:"after_memory"`
	MaxLimitReached bool   `json:"max_limit_reached"`
}

type UpdateResult struct {
	CronJobNamespace string            `json:"cronjob_namespace"`
	CronJobName      string            `json:"cronjob_name"`
	ContainerUpdates []ContainerUpdate `json:"container_updates"`
	RestartedJobName string            `json:"restarted_job_name"`
}

// SendMessage sends a message to the Slack via webhook
func SendMessage(res *UpdateResult, webhookURL string, webhookChannel string) error {
	if len(res.ContainerUpdates) == 0 {
		return nil
	}

	var memoryChanges string
	for _, u := range res.ContainerUpdates {
		var maxLimitReachedMessage string
		if u.MaxLimitReached {
			maxLimitReachedMessage = ":rotating_light: MAX LIMIT REACHED"
		}
		memoryChanges += fmt.Sprintf(
			"\t:sparkles: *%s (%s → %s %s)*\n",
			u.Name, u.BeforeMemory, u.AfterMemory, maxLimitReachedMessage,
		)
	}

	restartedJob := res.RestartedJobName
	if restartedJob == "" {
		restartedJob = "None"
	}

	attatchment := slack.Attachment{
		Text: fmt.Sprintf(
			"Namespace: *%s*\nCronJob name: *%s*\nContainer memory changes:\n%sRestarted Job name: *%s*\n",
			res.CronJobNamespace,
			res.CronJobName,
			memoryChanges,
			restartedJob,
		),
	}

	msg := slack.WebhookMessage{
		Username:    slackUserName,
		IconURL:     slackIconURL,
		Channel:     webhookChannel,
		Text:        ":broom: updated OOM CronJob",
		Attachments: []slack.Attachment{attatchment},
	}
	err := slack.PostWebhook(webhookURL, &msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
