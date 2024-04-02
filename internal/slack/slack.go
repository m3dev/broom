package slack

import (
	"fmt"

	"github.com/slack-go/slack"
)

const (
	slackUserName  = "Broom"
	slackIconEmoji = ":broom:"
)

type ContainerUpdate struct {
	Name         string `json:"name"`
	BeforeMemory string `json:"before_memory"`
	AfterMemory  string `json:"after_memory"`
}

type UpdateResult struct {
	CronJobNamespace string            `json:"cronjob_namespace"`
	CronJobName      string            `json:"cronjob_name"`
	ContainerUpdates []ContainerUpdate `json:"container_updates"`
	RestartedJobName string            `json:"restarted_job_name"`
}

func SendMessage(res UpdateResult, webhookURL string, webhookChannel string) error {
	if len(res.ContainerUpdates) == 0 {
		return nil
	}

	var memoryChanges string
	for _, u := range res.ContainerUpdates {
		memoryChanges += fmt.Sprintf("\t:sparkles: *%s (%s → %s)*\n", u.Name, u.BeforeMemory, u.AfterMemory)
	}

	restartedJob := res.RestartedJobName
	if restartedJob == "" {
		restartedJob = "None"
	}

	attatchment := slack.Attachment{
		Text: fmt.Sprintf("Namespace: *%s*\nName: *%s*\nContainer memory changes:\n%sRestarted Job: *%s*\n",
			res.CronJobNamespace,
			res.CronJobName,
			memoryChanges,
			restartedJob,
		),
	}

	msg := slack.WebhookMessage{
		Username:    slackUserName,
		IconEmoji:   slackIconEmoji,
		Channel:     webhookChannel,
		Text:        ":broom: CronJob jobTemplate updated",
		Attachments: []slack.Attachment{attatchment},
	}
	err := slack.PostWebhook(webhookURL, &msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
