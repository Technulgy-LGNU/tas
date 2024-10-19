package mail

import "tas/src/config"

type NewEmailData struct {
	Email     string
	Password  string
	Redirects []string
}

// NewEmail creates a new email in mailcow with the config provided
func NewEmail(data NewEmailData, cfg *config.CFG) {

}
