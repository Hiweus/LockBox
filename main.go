package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/hiweus/LockBox/credentials"

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

func main() {

	credentialMasterKey, err := getInput("Master key", true, false)
	if err != nil {
		log.Fatal("Canceled, exiting")
	}
	credentialRepository := credentials.New(credentialMasterKey)

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
				Type:    credentialType,
				Key:     fmt.Sprintf("%s:%s", credentialKey, credentialType),
				Content: credentialValue,
			}

			err = credentialRepository.Save(newCredential)
			if err != nil {
				log.Fatal("Error saving credential", err)
			} else {
				fmt.Println("Credential saved")
			}
		} else {
			log.Fatal("Invalid action")
		}
	}
}
