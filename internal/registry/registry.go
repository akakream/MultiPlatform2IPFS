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

var acceptList = [9]string{"application/vnd.docker.distribution.manifest.v1+json",
	"application/vnd.docker.distribution.manifest.v2+json",
	"application/vnd.docker.distribution.manifest.list.v2+json",
	"application/vnd.docker.container.image.v1+json",
	"application/vnd.docker.image.rootfs.diff.tar.gzip",
	"application/vnd.docker.image.rootfs.foreign.diff.tar.gzip",
	"application/vnd.docker.plugin.v1+json",
	"application/vnd.oci.image.index.v1+json",
	"application/vnd.oci.image.manifest.v1+json"}

var (
	// ErrManifestIsNotFat is error for when the repository is not multi-platform
	ErrManifestIsNotFat = errors.New("the repository is not multi-platform")
	// ErrNonOKhttpStatus is error for when the http status is not OK.
	ErrNonOKhttpStatus = errors.New("the http status is not OK")
)

const registryEndpoint = "https://index.docker.io/v2/library/"

// const registryEndpoint = "https://registry-1.docker.io/v2/library/"

func CopyImage(repoName string) {
	fmt.Println("Downloading the image...")
	downloadImage(repoName)
	fmt.Println("Uploading the image...")
	uploadImage()
	fmt.Println("The multi-arch image is uploaded to the IPFS!")
}

func downloadImage(repoName string) {
	token := getCachedOrNewToken(repoName)

	fatManifest, fatManifestRaw, err := getFatManifest(repoName, token)
	if err != nil {
		log.Fatalln(err)
	}

	dir_manifests := "export/manifests/"
	dir_blobs := "export/blobs/"
	fs.CreateDirs([]string{dir_manifests, dir_blobs})
	/*
		err = fs.SaveJson(fatManifest, dir_manifests+"latest")
		if err != nil {
			log.Fatalln(err)
		}
	*/
	err = fs.WriteBytesToFile(dir_manifests+"latest", fatManifestRaw)
	if err != nil {
		fmt.Println(err)
	}

	fatManifestSha256, err := fs.Sha256File(dir_manifests + "latest")
	if err != nil {
		log.Fatalln(err)
	}
	err = fs.WriteBytesToFile(dir_manifests+"sha256:"+fatManifestSha256, fatManifestRaw)
	if err != nil {
		fmt.Println(err)
	}

	for _, manifestValue := range fatManifest.Manifests {
		manifest, manifestRaw, err := getManifest(repoName, manifestValue.Digest, token)
		if err != nil {
			log.Fatalln(err)
		}

		/*
			err = fs.SaveJson(manifest, dir_manifests+manifestValue.Digest)
			if err != nil {
				log.Fatalln(err)
			}
		*/
		err = fs.WriteBytesToFile(dir_manifests+manifestValue.Digest, manifestRaw)
		if err != nil {
			fmt.Println(err)
		}

		config, err := getConfig(repoName, manifest.Config.Digest, token)
		if err != nil {
			fmt.Println(err)
		}

		err = fs.WriteBytesToFile(dir_blobs+manifest.Config.Digest, config)
		if err != nil {
			fmt.Println(err)
		}

		for _, layerValue := range manifest.Layers {
			downloadLayer(repoName, layerValue.Digest, token, dir_blobs+layerValue.Digest)
		}
	}
}

func uploadImage() {
	fmt.Println("uploadImage")
	ipfs.Add("/Users/ahmetkerem/projects/MultiPlatform2IPFS/export", true)
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

func getFatManifest(repoName string, token string) (FatManifest, []byte, error) {
	url := registryEndpoint + repoName + "/manifests/latest"

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
		return FatManifest{}, nil, ErrNonOKhttpStatus
	}

	if resp.Header.Get("content-type") != "application/vnd.docker.distribution.manifest.list.v2+json" && resp.Header.Get("content-type") != "application/vnd.oci.image.index.v1+json" {
		return FatManifest{}, nil, ErrManifestIsNotFat
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var fatManifest FatManifest
	if err := json.Unmarshal(body, &fatManifest); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	return fatManifest, body, nil
}

func getManifest(repoName string, digest string, token string) (Manifest, []byte, error) {
	url := registryEndpoint + repoName + "/manifests/" + digest

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
		return Manifest{}, nil, ErrNonOKhttpStatus
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	return manifest, body, nil
}

func getConfig(repoName string, digest string, token string) ([]byte, error) {
	url := registryEndpoint + repoName + "/blobs/" + digest

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
		return nil, ErrNonOKhttpStatus
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// var config Config
	// if err := json.Unmarshal(body, &config); err != nil { // Parse []byte to the go struct pointer
	// 	fmt.Println("Can not unmarshal JSON")
	// }

	return body, nil
}

func downloadLayer(repoName string, digest string, token string, destination string) error {
	url := registryEndpoint + repoName + "/blobs/" + digest

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
