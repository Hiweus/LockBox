package credentials

import (
	"fmt"
	"time"

	"github.com/hiweus/LockBox/totp"
)

type Credential struct {
	Content string `json:"content"`
	Type    string `json:"type"`
	Key     string `json:"key"`
}

func (c Credential) GetContent() string {
	if c.Type == "totp" {
		return fmt.Sprintf("%d", totp.GenerateTOTP(c.Content, 6, 30, time.Now().Unix()))
	}

	return c.Content
}
