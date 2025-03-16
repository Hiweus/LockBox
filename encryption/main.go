package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"golang.org/x/crypto/argon2"
)

// Encrypt encrypts a file in memory with a password using Argon2 and AES-GCM.
// It concatenates the salt to the output ciphertext and returns the result in base64 encoding.
func Encrypt(data []byte, password string) ([]byte, error) {
	// Generate a random salt
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return []byte(""), err
	}

	// Derive key from password using Argon2
	key := argon2.Key([]byte(password), salt, 1, 64*1024, 4, 32)

	// Create AES-GCM cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte(""), err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return []byte(""), err
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return []byte(""), err
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	// Concatenate the salt to the ciphertext (salt + ciphertext)
	result := append(salt, ciphertext...)

	// Encode the result in base64
	encoded := base64.StdEncoding.EncodeToString(result)
	return []byte(encoded), nil
}

// Decrypt decrypts a file in memory with a password using Argon2 and AES-GCM.
// It extracts the salt from the ciphertext and uses it for key derivation.
func Decrypt(encoded []byte, password string) ([]byte, error) {
	// Decode the base64-encoded ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return nil, err
	}

	// Extract the salt from the beginning of the ciphertext
	if len(ciphertext) < 16 {
		return nil, io.ErrShortBuffer
	}

	salt := ciphertext[:16]
	ciphertext = ciphertext[16:]

	// Derive key from password using Argon2
	key := argon2.Key([]byte(password), salt, 1, 64*1024, 4, 32)

	// Create AES-GCM cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// The nonce is the first part of the ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, io.ErrShortBuffer
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
