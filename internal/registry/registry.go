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


func GetToken(repoName string) string{
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

func HeadManifestType(repoName string, token string) error {
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
    req.Header.Set("Authorization", "Bearer " + token)

    resp, err := client.Do(req)
    if err != nil {
       log.Fatalln(err)
    }

    if resp.StatusCode != http.StatusOK {
        fmt.Println("Non-OK HTTP status:", resp.StatusCode)
        return ErrManifestIsNotFat
    }

    if resp.Header.Get("content-type") != "application/vnd.docker.distribution.manifest.list.v2+json" {
        return ErrManifestIsNotFat
    }

    fmt.Println("The manifest is a fat one.")

    return nil
}