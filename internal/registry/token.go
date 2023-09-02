package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/akakream/MultiPlatform2IPFS/internal/fs"
)

func getCachedOrNewToken(imageName string, imageTag string) (string, error) {
	cacheHit, token, err := fs.GetCachedToken(imageName, imageTag)
	if err != nil {
		return "", err
	}

	if !cacheHit {
		token, err = getToken(imageName)
		if err != nil {
			return "", err
		}
		err = fs.FillCache(imageName, token)
		if err != nil {
			return "", err
		}
		fmt.Println(
			"New token for the repositoy " + imageName + " is fetched and stored.",
		)
	} else {
		fmt.Println("Cached token for the repositoy " + imageName + " is being used.")
	}

	return token, nil
}

func getToken(imageName string) (string, error) {
	url := "https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/" + imageName + ":pull"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var jsonResult TokenResponse
	if err := json.Unmarshal(body, &jsonResult); err != nil { // Parse []byte to the go struct pointer
		return "", err
	}
	return jsonResult.Token, nil
}
