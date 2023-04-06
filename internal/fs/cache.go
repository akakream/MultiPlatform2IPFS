package fs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type TokenStore struct {
	Tokens []TokenItem
}

type TokenItem struct {
	Repository string
	Token      string
}

const tokenPath = "cache/tokens.json"

func GetCachedToken(repoName string) (bool, string, error) {
	if pathExists(tokenPath) {
		tokenExists, token := getTokenFromCache(repoName)
		if tokenExists {
			if tokenValid(token) {
				return true, token, nil
			}
		}
	} else {
		if err := createTokenFile(); err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
	}

	return false, "", nil
	// Check if cache/token.txt exists
	// 		if does not exist, return error
	//		else do a HEAD request
	//			if request unauthorized return error
	//			else return token
}

func FillCache(repoName string, token string) error {
	if !pathExists(tokenPath) {
		if err := createTokenFile(); err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
	}

	// read File
	file, err := os.ReadFile(tokenPath)
	if err != nil {
		log.Fatalln("The file " + tokenPath + " cannot be read.")
		os.Exit(1)
	}
	var tokens TokenStore
	json.Unmarshal(file, &tokens)
	// Check if the repository exists in the file.
	if len(tokens.Tokens) > 0 {
		for _, repo := range tokens.Tokens {
			if repoName == repo.Repository {
				repo.Token = token // renew the token
			} else {
				addTokenToCache(repoName, token, &tokens)
			}
		}
	} else {
		addTokenToCache(repoName, token, &tokens)
	}

	tokenStoreMarshalled, err := json.Marshal(tokens)
	if err != nil {
		log.Fatalln("The token could not updated.")
	}
	if err := os.WriteFile(tokenPath, tokenStoreMarshalled, 0644); err != nil {
		log.Fatalln("Could not write to cache/tokens.json")
	}

	return nil
}

func addTokenToCache(repoName string, token string, tokens *TokenStore) {
	// create the repository in the file and save token
	newToken := TokenItem{Repository: repoName, Token: token}
	updatedTokens := append(tokens.Tokens, newToken)
	tokens.Tokens = updatedTokens
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		fmt.Printf("File does not exist\n")
		return false
	}
	fmt.Printf("File exists\n")
	return true
}

func createTokenFile() error {
	_, err := os.Create(tokenPath)
	if err != nil {
		return err
	}
	return nil
}

func getTokenFromCache(repoName string) (bool, string) {
	return false, ""
}

func tokenValid(token string) bool {
	return false
}
