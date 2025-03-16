package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hiweus/LockBox/credentials"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Missing required argument")
	}

	credentialKey := os.Args[1]

	credentialRepository := credentials.New("masterkey")
	expectedCredential, err := credentialRepository.Find(credentialKey)
	if err != nil {
		log.Fatal(err)
	}

	if expectedCredential == nil {
		log.Fatal("Credential not found")
	}

	fmt.Println(expectedCredential.GetContent())
}
