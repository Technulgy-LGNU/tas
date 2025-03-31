package notifications

import (
	"encoding/json"
	discordwebhook "github.com/bensch777/discord-webhook-golang"
	"log"
	"tas/src/config"
)

func SendDiscordEmbed(embeds discordwebhook.Embed, config *config.CFG) {
	hook := discordwebhook.Hook{
		Username:    "TAS",
		Avatar_url:  "https://raw.githubusercontent.com/Technulgy-LGNU/technulgy-website/refs/heads/dev/public/favicon.ico",
		Embeds:      []discordwebhook.Embed{embeds},
		Attachments: nil,
	}

	payload, err := json.Marshal(&hook)
	if err != nil {
		log.Printf("Error marshalling webhook payload: %v\n", err)
		return
	}

	err = discordwebhook.ExecuteWebhook(config.DiscordWebhook, payload)
	if err != nil {
		log.Printf("Error sending webhook: %v\n", err)
		return
	}
}
