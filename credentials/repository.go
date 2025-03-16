package credentials

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/hiweus/LockBox/encryption"
)

type CredentialRepository struct {
	masterKey   string
	credentials []Credential
}

func getCompleteFileName() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(homeDir, ".credentials-lb.json")
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func (c *CredentialRepository) loadCredentials() ([]Credential, error) {
	if c.credentials != nil {
		return c.credentials, nil
	}

	if !fileExists(getCompleteFileName()) {
		encryptedFile, err := encryption.Encrypt([]byte("[]"), c.masterKey)
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(getCompleteFileName(), encryptedFile, 0644)
		if err != nil {
			return nil, err
		}
	}

	credentialFileInput, err := os.ReadFile(getCompleteFileName())
	if err != nil {
		return nil, err
	}

	decryptedCredentials, err := encryption.Decrypt([]byte(credentialFileInput), c.masterKey)

	if err != nil {
		return nil, err
	}

	var credentials []Credential
	err = json.Unmarshal(decryptedCredentials, &credentials)
	if err != nil {
		return nil, err
	}

	c.credentials = credentials
	return credentials, nil
}

func (c *CredentialRepository) Save(credential Credential) error {
	credentials, err := c.loadCredentials()
	if err != nil {
		return err
	}

	credentials = append(credentials, credential)

	credentialsJson, err := json.Marshal(credentials)
	if err != nil {
		return err
	}

	encryptedCredentials, err := encryption.Encrypt(credentialsJson, c.masterKey)
	if err != nil {
		return err
	}

	err = os.WriteFile(getCompleteFileName(), encryptedCredentials, 0644)
	if err != nil {
		return err
	}

	c.credentials = credentials
	return nil
}

func (c *CredentialRepository) Find(key string) (*Credential, error) {
	credentials, err := c.loadCredentials()
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
	c := &CredentialRepository{
		masterKey: masterKey,
	}
	_, err := c.loadCredentials()
	if err != nil {
		panic(err)
	}

	return c
}
