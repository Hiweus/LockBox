package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hiweus/LockBox/credentials"
	"github.com/hiweus/LockBox/encryption"
	"github.com/hiweus/LockBox/externalSync"

	"github.com/manifoldco/promptui"
)

func autocompleteCredential(allCredentials []credentials.Credential) *credentials.Credential {
	search := func(input string, index int) bool {
		c := allCredentials[index]
		lowerInput := strings.ToLower(input)
		lowerKey := strings.ToLower(c.Key)
		lowerType := strings.ToLower(c.Type)
		return strings.Contains(lowerKey, lowerInput) || strings.Contains(lowerType, lowerInput)
	}

	autocompletePrompt := promptui.Select{
		Label:             "Select credential",
		Items:             allCredentials,
		Searcher:          search,
		StartInSearchMode: true,
		Size:              5,
	}

	index, _, err := autocompletePrompt.Run()
	if err != nil {
		return nil
	}

	return &allCredentials[index]
}

func getInput(label string, secret bool, minimumCharacters bool) (string, error) {

	validate := func(input string) error {
		if minimumCharacters && len(input) < 8 {
			return fmt.Errorf("must be at least 8 characters")
		}
		return nil
	}

	var prompt promptui.Prompt
	if secret {
		prompt = promptui.Prompt{
			Label:       label,
			HideEntered: secret,
			Mask:        '*',
			Validate:    validate,
		}
	} else {
		prompt = promptui.Prompt{
			Label:    label,
			Validate: validate,
		}
	}

	result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}

func getOptions(options []string) (string, error) {
	prompt := promptui.Select{
		Label: "Type",
		Items: options,
	}

	_, result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}

func performLogin(vault *encryption.Vault) {
	if len(os.Args) < 2 {
		return
	}

	username, err := getInput("Username", false, false)
	if err != nil {
		log.Fatal("Canceled, exiting")
	}

	token, err := getInput("Token", true, false)
	if err != nil {
		log.Fatal("Canceled, exiting")
	}

	syncTool := externalSync.New(username, token, vault)
	syncTool.Persist()
	syncTool.Push()

}

func main() {
	credentialMasterKey, err := getInput("Master key", true, false)
	if err != nil {
		log.Fatal("Canceled, exiting")
	}
	vault := encryption.New(credentialMasterKey)

	performLogin(vault)

	syncAgent := externalSync.Load(vault)
	syncAgent.Push()
	syncAgent.Pull()

	go func() {
		for {
			time.Sleep(5 * time.Second)
			syncAgent.Push()
		}
	}()

	fmt.Println("Welcome to LockBox")

	credentialRepository := credentials.New(vault)

	for {
		selectedAction, err := getOptions([]string{
			"Find credential",
			"Add credential",
		})

		if err != nil {
			log.Fatal("Canceled, exiting")
		}

		if selectedAction == "Find credential" {
			credentialSelected := autocompleteCredential(credentialRepository.Fetch())
			if credentialSelected == nil {
				log.Fatal("Credential not found")
			}

			fmt.Println(credentialSelected.GetContent())
		} else if selectedAction == "Add credential" {
			credentialType, err := getOptions([]string{
				"totp",
				"password",
			})
			if err != nil {
				continue
			}

			credentialKey, err := getInput("Key", false, false)
			if err != nil {
				continue
			}

			credentialValue, err := getInput("Content", true, false)
			if err != nil {
				continue
			}

			newCredential := credentials.Credential{
				Type:      credentialType,
				Key:       fmt.Sprintf("%s:%s", credentialKey, credentialType),
				Content:   credentialValue,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				Synced:    false,
			}

			err = credentialRepository.Save(newCredential)
			if err != nil {
				log.Fatal("Error saving credential", err)
			} else {
				fmt.Println("Credential saved")
				go syncAgent.Push()
			}
		} else {
			log.Fatal("Invalid action")
		}
	}
}
