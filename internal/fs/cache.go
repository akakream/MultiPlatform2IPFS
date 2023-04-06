package fs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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
		token, err := getTokenFromCache(repoName)
		if err != nil {
			// Deal with this
			log.Fatalln("DEAL WITH THIS")
		}
		if isTokenValid(repoName, token) {
			return true, token, nil
		}
	} else {
		if err := createTokenFile(); err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
	}

	return false, "", nil
}

func FillCache(repoName string, token string) error {
	if !pathExists(tokenPath) {
		if err := createTokenFile(); err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
	}

	tokens := readTokensFromCache()
	repoExists, repoIndex := repositoryIsInCache(repoName, &tokens)
	if repoExists {
		tokens.Tokens[repoIndex].Token = token // renew the token
	} else {
		addTokenToCache(repoName, token, &tokens)
	}
	writeTokensToCache(tokens)
	return nil
}

func addTokenToCache(repoName string, token string, tokens *TokenStore) {
	// create the repository in the file and save token
	newToken := TokenItem{Repository: repoName, Token: token}
	updatedTokens := append(tokens.Tokens, newToken)
	tokens.Tokens = updatedTokens
}

func repositoryIsInCache(repoName string, tokens *TokenStore) (bool, int) {
	if len(tokens.Tokens) > 0 {
		for index, repo := range tokens.Tokens {
			if repoName == repo.Repository {
				return true, index
			}
		}
	}
	return false, -1
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

func getTokenFromCache(repoName string) (string, error) {
	tokens := readTokensFromCache()
	repoExists, repoIndex := repositoryIsInCache(repoName, &tokens)
	if repoExists {
		return tokens.Tokens[repoIndex].Token, nil
	}
	return "", errors.New("token could not be fetched from the cache")
}

func readTokensFromCache() TokenStore {
	// read File
	file, err := os.ReadFile(tokenPath)
	if err != nil {
		log.Fatalln("The file " + tokenPath + " cannot be read.")
		os.Exit(1)
	}
	var tokens TokenStore
	if err := json.Unmarshal(file, &tokens); err != nil {
		log.Fatalln("Could not read the tokens from " + tokenPath)
		os.Exit(1)
	}

	return tokens
}

func writeTokensToCache(tokens TokenStore) {
	tokenStoreMarshalled, err := json.Marshal(tokens)
	if err != nil {
		log.Fatalln("The token could not updated.")
	}
	if err := os.WriteFile(tokenPath, tokenStoreMarshalled, 0644); err != nil {
		log.Fatalln("Could not write to cache/tokens.json")
	}

}

func isTokenValid(repoName string, token string) bool {
	acceptList := [7]string{"application/vnd.docker.distribution.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.docker.distribution.manifest.list.v2+json",
		"application/vnd.docker.container.image.v1+json",
		"application/vnd.docker.image.rootfs.diff.tar.gzip",
		"application/vnd.docker.image.rootfs.foreign.diff.tar.gzip",
		"application/vnd.docker.plugin.v1+json"}

	url := "https://registry-1.docker.io/v2/library/" + repoName + "/manifests/latest"

	client := &http.Client{}
	req, _ := http.NewRequest("HEAD", url, nil)
	req.Header.Set("Accept", strings.Join(acceptList[:], ", "))
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Non-OK HTTP status:", resp.StatusCode)
		return false
	}
	return true
}
