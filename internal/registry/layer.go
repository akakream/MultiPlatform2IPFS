package registry

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

func downloadLayer(
	repoName string,
	digest string,
	token string,
	destination string,
	wg *sync.WaitGroup,
) error {
	defer wg.Done()

	url := registryEndpoint + repoName + "/blobs/" + digest

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Non-OK HTTP status:", resp.StatusCode)
		return ErrNonOKhttpStatus
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = os.WriteFile(destination, body, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
