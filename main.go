package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hiweus/LockBox/totp"
)

const credentialFile = "credentials.json"

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

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Missing required argument")
	}

	credentialKey := os.Args[1]

	if !fileExists(filepath.Join(".", credentialFile)) {
		fmt.Println("First time running, creating credentials file")
		os.WriteFile(filepath.Join(".", credentialFile), []byte("[]"), 0644)
	}

	credentialFileInput, err := os.ReadFile(filepath.Join(".", credentialFile))
	if err != nil {
		log.Fatal(err)
	}

	var credentials []Credential
	err = json.Unmarshal(credentialFileInput, &credentials)
	if err != nil {
		log.Fatal(err)
	}

	var expectedCredential *Credential
	for _, credential := range credentials {
		if credential.Key == credentialKey {
			expectedCredential = &credential
			break
		}
	}

	if expectedCredential == nil {
		log.Fatal("Credential not found")
	}

	fmt.Println(expectedCredential.GetContent())
}
