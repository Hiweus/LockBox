package credentials

import (
	"fmt"
	"time"

	"github.com/hiweus/LockBox/totp"
)

type Credential struct {
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Synced    bool      `json:"synced"`
}

func (c Credential) GetContent() string {
	if c.Type == "totp" {
		return fmt.Sprint(totp.GenerateTOTP(c.Content, 6, 30, time.Now().Unix()))
	}

	return c.Content
}

func (c Credential) String() string {
	return c.Key
}
