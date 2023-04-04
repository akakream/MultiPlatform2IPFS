package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	// ErrManifestIsNotFat is error for when the repository is not multi-platform
	ErrManifestIsNotFat = errors.New("the repository is not multi-platform")
	// ErrNonOKhttpStatus is error for when the http status is not OK. 
	ErrNonOKhttpStatus = errors.New("the http status is not OK")
)

type TokenResponse struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}

type FatManifest struct {
	Manifests []struct {
		Digest    string `json:"digest"`
		MediaType string `json:"mediaType"`
		Platform  struct {
			Architecture string `json:"architecture"`
			Os           string `json:"os"`
		} `json:"platform,omitempty"`
		Size      int `json:"size"`
	} `json:"manifests"`
	MediaType     string `json:"mediaType"`
	SchemaVersion int    `json:"schemaVersion"`
}

	
type Manifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"layers"`
}


func CopyImage(repoName string) {
	token := getToken(repoName)
	fatManifest, fatManifestErr := getFatManifest(repoName, token)
    if fatManifestErr != nil {
        fmt.Println(fatManifestErr)
    }

    for _, manifestValue := range fatManifest.Manifests {
        fmt.Println(manifestValue.Digest)
        manifest, manifestErr := getManifest(repoName, manifestValue.Digest, token)
        if manifestErr != nil {
            fmt.Println(manifestErr)
        }
        fmt.Println(manifest)
    }
}

func getToken(repoName string) string{
    url := "https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/" + repoName + ":pull"

    client := &http.Client{}
    req, _ := http.NewRequest("GET", url, nil)

    resp, err := client.Do(req)
    if err != nil {
       log.Fatalln(err)
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatalln(err)
    }

    var jsonResult TokenResponse
    if err := json.Unmarshal(body, &jsonResult); err != nil {  // Parse []byte to the go struct pointer
        fmt.Println("Can not unmarshal JSON")
    }

    fmt.Println(jsonResult.Token)
    return jsonResult.Token
}

func getFatManifest(repoName string, token string) (FatManifest, error) {
    acceptList := [7]string{"application/vnd.docker.distribution.manifest.v1+json", 
                            "application/vnd.docker.distribution.manifest.v2+json", 
                            "application/vnd.docker.distribution.manifest.list.v2+json",  
                            "application/vnd.docker.container.image.v1+json", 
                            "application/vnd.docker.image.rootfs.diff.tar.gzip", 
                            "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip", 
                            "application/vnd.docker.plugin.v1+json"}
    
    url := "https://registry-1.docker.io/v2/library/" + repoName + "/manifests/latest"

    client := &http.Client{}
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Accept", strings.Join(acceptList[:], ", "))
    req.Header.Set("Authorization", "Bearer " + token)

    resp, err := client.Do(req)
    if err != nil {
       log.Fatalln(err)
    }

    if resp.StatusCode != http.StatusOK {
        fmt.Println("Non-OK HTTP status:", resp.StatusCode)
        return FatManifest{}, ErrNonOKhttpStatus
    }

    if resp.Header.Get("content-type") != "application/vnd.docker.distribution.manifest.list.v2+json" {
        return FatManifest{}, ErrManifestIsNotFat
    }
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatalln(err)
    }

    var fatManifest FatManifest
    if err := json.Unmarshal(body, &fatManifest); err != nil {  // Parse []byte to the go struct pointer
        fmt.Println("Can not unmarshal JSON")
    }
    fmt.Println(fatManifest)

    return fatManifest, nil
}

func getManifest(repoName string, digest string, token string) (Manifest, error) {
    acceptList := [7]string{"application/vnd.docker.distribution.manifest.v1+json", 
                            "application/vnd.docker.distribution.manifest.v2+json", 
                            "application/vnd.docker.distribution.manifest.list.v2+json",  
                            "application/vnd.docker.container.image.v1+json", 
                            "application/vnd.docker.image.rootfs.diff.tar.gzip", 
                            "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip", 
                            "application/vnd.docker.plugin.v1+json"}
    
    url := "https://registry-1.docker.io/v2/library/" + repoName + "/manifests/" + digest

    client := &http.Client{}
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Accept", strings.Join(acceptList[:], ", "))
    req.Header.Set("Authorization", "Bearer " + token)

    resp, err := client.Do(req)
    if err != nil {
       log.Fatalln(err)
    }

    if resp.StatusCode != http.StatusOK {
        fmt.Println("Non-OK HTTP status:", resp.StatusCode)
        return Manifest{}, ErrNonOKhttpStatus
    }
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatalln(err)
    }

    var manifest Manifest
    if err := json.Unmarshal(body, &manifest); err != nil {  // Parse []byte to the go struct pointer
        fmt.Println("Can not unmarshal JSON")
    }

    return manifest, nil
}