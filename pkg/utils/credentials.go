package utils

import (
	"errors"
	"os"
)

var (
	CredentialFileNotFoundError = errors.New("credentials file not found")
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func FindCredentials(filename string) ([]byte, error) {
	if filename == "" {
		return nil, CredentialFileNotFoundError
	}

	var path string
	var found bool

	var paths = []string{".", "..", "./credentials"}

	for _, p := range paths {
		path = p + "/" + filename
		if fileExists(path) {
			found = true
			break
		}
	}

	if !found {
		return nil, CredentialFileNotFoundError
	}

	return os.ReadFile(path)
}
