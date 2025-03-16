package credentials

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type CredentialRepository struct {
	masterKey   string
	credentials []Credential
}

const credentialFile = "credentials.json"

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func loadCredentials() ([]Credential, error) {
	if !fileExists(credentialFile) {
		err := os.WriteFile(credentialFile, []byte("[]"), 0644)
		if err != nil {
			return nil, err
		}
	}

	credentialFileInput, err := os.ReadFile(filepath.Join(".", credentialFile))
	if err != nil {
		return nil, err
	}

	var credentials []Credential
	err = json.Unmarshal(credentialFileInput, &credentials)
	if err != nil {
		return nil, err
	}

	return credentials, nil
}

func (c *CredentialRepository) Save(credential Credential) error {
	credentials, err := loadCredentials()
	if err != nil {
		return err
	}

	credentials = append(credentials, credential)

	credentialsJson, err := json.Marshal(credentials)
	if err != nil {
		return err
	}

	err = os.WriteFile(credentialFile, credentialsJson, 0644)
	if err != nil {
		return err
	}

	c.credentials = credentials
	return nil
}

func (c *CredentialRepository) Find(key string) (*Credential, error) {
	credentials, err := loadCredentials()
	if err != nil {
		return nil, err
	}

	for _, credential := range credentials {
		if credential.Key == key {
			return &credential, nil
		}
	}

	return nil, nil
}

func (c *CredentialRepository) Fetch() []Credential {
	return c.credentials
}

func New(masterKey string) *CredentialRepository {
	credentials, err := loadCredentials()
	if err != nil {
		panic(err)
	}
	return &CredentialRepository{
		masterKey:   masterKey,
		credentials: credentials,
	}
}
