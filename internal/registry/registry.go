package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/akakream/MultiPlatform2IPFS/internal/fs"
	"github.com/akakream/MultiPlatform2IPFS/internal/ipfs"
)

var (
	// ErrManifestIsNotFat is error for when the repository is not multi-platform
	ErrManifestIsNotFat = errors.New("the repository is not multi-platform")
	// ErrNonOKhttpStatus is error for when the http status is not OK.
	ErrNonOKhttpStatus = errors.New("the http status is not OK")
)

func CopyImage(repoName string) {
	fmt.Println("Downloading the image...")
	downloadImage(repoName)
	fmt.Println("Uploading the image...")
	uploadImage()
	fmt.Println("The multi-arch image is uploaded to the IPFS!")
}

func downloadImage(repoName string) {
	token := getCachedOrNewToken(repoName)

	fatManifest, err := getFatManifest(repoName, token)
	if err != nil {
		log.Fatalln(err)
	}

	fs.CreateDir("export")
	fs.SaveJson(fatManifest, "export/manifestlist.json")

	for _, manifestValue := range fatManifest.Manifests {
		manifest, err := getManifest(repoName, manifestValue.Digest, token)
		if err != nil {
			log.Fatalln(err)
		}
		manifestsFolderPath := "export/" + manifestValue.Platform.Os + "/" + manifestValue.Platform.Architecture + "/manifests"
		blobsFolderPath := "export/" + manifestValue.Platform.Os + "/" + manifestValue.Platform.Architecture + "/blobs"
		fs.CreateDir(manifestsFolderPath)
		fs.CreateDir(blobsFolderPath)
		fs.SaveJson(manifest, manifestsFolderPath+"/latest")

		/*
		   // I DO NOT THINK THAT THIS PART IS NECESSARY.
		   // ALSO REMOVE IT FROM IPDR.
		   manifestSha256, err := fs.Sha256File(manifestsFolderPath + "/latest")
		   if err != nil {
		       log.Fatalln(err)
		   }
		   fs.SaveJson(manifest, manifestsFolderPath+"/sha256"+manifestSha256)
		*/

		config, err := getConfig(repoName, manifest.Config.Digest, token)
		if err != nil {
			fmt.Println(err)
		}
		fs.SaveJson(config, blobsFolderPath+"/"+manifest.Config.Digest)

		for _, layerValue := range manifest.Layers {
			destinationFolder := "export/" + manifestValue.Platform.Os + "/" + manifestValue.Platform.Architecture + "/blobs"
			tempDestination := destinationFolder + "/layer.tmp.tgz"
			downloadLayer(repoName, layerValue.Digest, token, tempDestination)
			layerSha256, err := fs.Sha256File(tempDestination)
			if err != nil {
				os.Exit(1)
			}
			destination := destinationFolder + "/sha256:" + layerSha256
			if err := os.Rename(tempDestination, destination); err != nil {
				os.Exit(1)
			}
		}
	}
}

func uploadImage() {
	fmt.Println("uploadImage")
	ipfs.Add("/Users/ahmetkerem/projects/MultiPlatform2IPFS/export")
}

func getCachedOrNewToken(repoName string) string {
	cacheHit, token, err := fs.GetCachedToken(repoName)
	if err != nil {
		log.Fatalln(err)
	}
	if !cacheHit {
		token = getToken(repoName)
		fs.FillCache(repoName, token)
		fmt.Println("New token for the repositoy " + repoName + " is fetched and stored.")
	} else {
		fmt.Println("Cached token for the repositoy " + repoName + " is being used.")
	}

	return token
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var jsonResult TokenResponse
	if err := json.Unmarshal(body, &jsonResult); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var fatManifest FatManifest
	if err := json.Unmarshal(body, &fatManifest); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
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

	body, err := io.ReadAll(resp.Body)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var config Config
	if err := json.Unmarshal(body, &config); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	return config, nil
}

func downloadLayer(repoName string, digest string, token string, destination string) error {
	url := "https://registry-1.docker.io/v2/library/" + repoName + "/blobs/" + digest

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Non-OK HTTP status:", resp.StatusCode)
		return ErrNonOKhttpStatus
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	os.WriteFile(destination, body, os.ModePerm)
	return nil
}
