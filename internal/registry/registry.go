package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/akakream/MultiPlatform2IPFS/internal/fs"
)

var (
	// ErrManifestIsNotFat is error for when the repository is not multi-platform
	ErrManifestIsNotFat = errors.New("the repository is not multi-platform")
	// ErrNonOKhttpStatus is error for when the http status is not OK.
	ErrNonOKhttpStatus = errors.New("the http status is not OK")
)

func CopyImage(repoName string) {
	cacheHit, token, tokenCacheError := fs.GetCachedToken(repoName)
	if tokenCacheError != nil {
		os.Exit(1)
	}
	if !cacheHit {
		token = getToken(repoName)
		fs.FillCache(repoName, token)
		fmt.Println("New token for the repositoy " + repoName + " is fetched and stored.")
	} else {
		fmt.Println("Cached token for the repositoy " + repoName + " is being used.")
	}
	/*
			fatManifest, fatManifestErr := getFatManifest(repoName, token)
		    if fatManifestErr != nil {
		        fmt.Println(fatManifestErr)
		    }

		    fs.CreateDir("export")
		    fs.SaveJson(fatManifest, "export/manifestlist.json")

		    for _, manifestValue := range fatManifest.Manifests {
		        manifest, manifestErr := getManifest(repoName, manifestValue.Digest, token)
		        if manifestErr != nil {
		            fmt.Println(manifestErr)
		        }
		        fmt.Println(manifest)
		        fs.CreateDir("export/" + manifestValue.Platform.Os + "/" + manifestValue.Platform.Architecture + "/manifests")
		        fs.CreateDir("export/" + manifestValue.Platform.Os + "/" + manifestValue.Platform.Architecture + "/blobs")
		        fs.SaveJson(manifest, "export/" + manifestValue.Platform.Os + "/" + manifestValue.Platform.Architecture + "/manifests/latest")
		        stringManifest, stringManifestErr := json.Marshal(manifest)
		        if stringManifestErr != nil {
		            fmt.Println("Error")
		        }
		        fs.SaveJson(manifest, "export/" + manifestValue.Platform.Os + "/" + manifestValue.Platform.Architecture + "/manifests/" + fs.Sha256izeString(string(stringManifest)))

		        config, configErr := getConfig(repoName, manifest.Config.Digest, token)
		        if configErr != nil {
		            fmt.Println(configErr)
		        }
		        fmt.Println(config)

		        for _, layerValue := range manifest.Layers {
		            fmt.Println(layerValue)
		        }
		    }
	*/
}

func getToken(repoName string) string {
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
	if err := json.Unmarshal(body, &jsonResult); err != nil { // Parse []byte to the go struct pointer
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
	req.Header.Set("Authorization", "Bearer "+token)

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
	if err := json.Unmarshal(body, &fatManifest); err != nil { // Parse []byte to the go struct pointer
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
	req.Header.Set("Authorization", "Bearer "+token)

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
	if err := json.Unmarshal(body, &manifest); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	return manifest, nil
}

func getConfig(repoName string, digest string, token string) (Config, error) {
	acceptList := [7]string{"application/vnd.docker.distribution.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.docker.distribution.manifest.list.v2+json",
		"application/vnd.docker.container.image.v1+json",
		"application/vnd.docker.image.rootfs.diff.tar.gzip",
		"application/vnd.docker.image.rootfs.foreign.diff.tar.gzip",
		"application/vnd.docker.plugin.v1+json"}

	url := "https://registry-1.docker.io/v2/library/" + repoName + "/blobs/" + digest

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", strings.Join(acceptList[:], ", "))
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Non-OK HTTP status:", resp.StatusCode)
		return Config{}, ErrNonOKhttpStatus
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var config Config
	if err := json.Unmarshal(body, &config); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	return config, nil
}

/*
func getLayer(repoName string, digest string, token string) (string, error) {
	acceptList := [7]string{"application/vnd.docker.distribution.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.docker.distribution.manifest.list.v2+json",
		"application/vnd.docker.container.image.v1+json",
		"application/vnd.docker.image.rootfs.diff.tar.gzip",
		"application/vnd.docker.image.rootfs.foreign.diff.tar.gzip",
		"application/vnd.docker.plugin.v1+json"}

	url := "https://registry-1.docker.io/v2/library/" + repoName + "/blobs/" + digest

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", strings.Join(acceptList[:], ", "))
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Non-OK HTTP status:", resp.StatusCode)
		return Config{}, ErrNonOKhttpStatus
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var config Config
	if err := json.Unmarshal(body, &config); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	return config, nil
}
*/
