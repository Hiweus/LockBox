package externalSync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/hiweus/LockBox/credentials"
	"github.com/hiweus/LockBox/encryption"
)

const CredentialsFilename = ".credentials-lb.json"
const SyncFilename = ".sync-lb"

type SyncConfiguration struct {
	Token    string `json:"token"`
	Username string `json:"username"`

	vault *encryption.Vault

	mutex *sync.Mutex
}

type GistFile struct {
	Filename string `json:"filename"`
	RawUrl   string `json:"raw_url"`
}

type Gist struct {
	ID          string              `json:"id"`
	Description string              `json:"description"`
	Files       map[string]GistFile `json:"files"`
}

func getCompleteFileName(filename string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(homeDir, filename)
}

func getCredentialsFilename() string {
	return getCompleteFileName(CredentialsFilename)
}

func getSyncFilename() string {
	return getCompleteFileName(SyncFilename)
}

func saveToFile(url, filepath string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

func (sc *SyncConfiguration) findLockBoxGist() (*Gist, error) {
	response, err := http.Get(fmt.Sprintf("https://api.github.com/users/%s/gists", sc.Username))
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var gists []Gist
	if err := json.Unmarshal(data, &gists); err != nil {
		return nil, err
	}

	for _, gist := range gists {
		if gist.Description == "lockbox" {
			return &gist, nil
		}
	}

	return nil, nil
}

func (sc *SyncConfiguration) loadCredentials() ([]credentials.Credential, error) {
	credentialsFilename := getCredentialsFilename()
	file, err := os.ReadFile(credentialsFilename)
	if err != nil {
		return nil, err
	}

	unencryptedFile, err := sc.vault.Decrypt(file)
	if err != nil {
		return nil, err
	}

	var parsedCredentials []credentials.Credential
	if err := json.Unmarshal(unencryptedFile, &parsedCredentials); err != nil {
		return nil, err
	}

	return parsedCredentials, nil
}

func (sc *SyncConfiguration) Push() error {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	syncConfigurationFilename := getSyncFilename()
	credentialsFilename := getCredentialsFilename()

	parsedCredentials, err := sc.loadCredentials()
	if err != nil {
		return err
	}

	shouldPush := false
	for i := range parsedCredentials {
		if !parsedCredentials[i].Synced {
			shouldPush = true
		}
		parsedCredentials[i].Synced = true
	}

	if !shouldPush {
		return nil
	}

	serializedCredentials, err := json.Marshal(parsedCredentials)
	if err != nil {
		return err
	}

	encryptedCredentials, err := sc.vault.Encrypt(serializedCredentials)
	if err != nil {
		return err
	}

	if len(sc.Token) == 0 {
		var storedSync *SyncConfiguration
		file, err := os.ReadFile(syncConfigurationFilename)
		if err != nil {
			return err
		}

		unencryptedFile, err := sc.vault.Decrypt(file)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(unencryptedFile, &storedSync); err != nil {
			return err
		}

		sc.Token = storedSync.Token
	}

	desiredGist, err := sc.findLockBoxGist()
	if err != nil {
		return err
	}

	if desiredGist == nil {
		return fmt.Errorf("no gist found, create one first on github panel")
	}

	syncFile, err := os.ReadFile(syncConfigurationFilename)
	if err != nil {
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest("PATCH", fmt.Sprintf("https://api.github.com/gists/%s", desiredGist.ID), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sc.Token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	var requestBody map[string]interface{} = map[string]interface{}{
		"files": map[string]interface{}{
			SyncFilename: map[string]interface{}{
				"content": string(syncFile),
			},
			CredentialsFilename: map[string]interface{}{
				"content": string(encryptedCredentials),
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(requestBody); err != nil {
		return err
	}

	req.Body = io.NopCloser(&buf)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("error on status code: %d", resp.StatusCode)
	}

	err = os.WriteFile(credentialsFilename, encryptedCredentials, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (sc *SyncConfiguration) Pull() error {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	syncConfigurationFilename := getSyncFilename()
	credentialsFilename := getCredentialsFilename()

	desiredGist, err := sc.findLockBoxGist()

	if err != nil {
		return err
	}

	if desiredGist == nil {
		return fmt.Errorf("no gist found")
	}

	rawSyncFileUrl := desiredGist.Files[".sync-lb"].RawUrl
	err = saveToFile(rawSyncFileUrl, syncConfigurationFilename)

	if err != nil {
		return err
	}

	rawCredentialsFileUrl := desiredGist.Files[".credentials-lb.json"].RawUrl
	err = saveToFile(rawCredentialsFileUrl, credentialsFilename)

	return err
}

func (sc *SyncConfiguration) Persist() {
	syncConfigurationFilename := getSyncFilename()
	credentialsFilename := getCredentialsFilename()

	serializedSync, err := json.Marshal(sc)
	if err != nil {
		panic(err)
	}

	encryptedSync, err := sc.vault.Encrypt(serializedSync)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(syncConfigurationFilename, encryptedSync, 0644)
	if err != nil {
		panic(err)
	}

	encryptedDefaultFile, err := sc.vault.Encrypt([]byte("[]"))
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(
		credentialsFilename,
		encryptedDefaultFile,
		0644,
	)

	if err != nil {
		panic(err)
	}
}

func New(username, token string, vault *encryption.Vault) *SyncConfiguration {
	return &SyncConfiguration{
		Username: username,
		Token:    token,
		vault:    vault,
		mutex:    &sync.Mutex{},
	}
}

func Load(vault *encryption.Vault) *SyncConfiguration {
	syncConfigurationFilename := getSyncFilename()
	file, err := os.ReadFile(syncConfigurationFilename)
	if err != nil {
		panic(err)
	}

	unencryptedFile, err := vault.Decrypt(file)
	if err != nil {
		panic(err)
	}

	var sc *SyncConfiguration
	if err := json.Unmarshal(unencryptedFile, &sc); err != nil {
		panic(err)
	}

	sc.vault = vault
	sc.mutex = &sync.Mutex{}

	return sc
}
