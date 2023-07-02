package registry

import (
	"context"
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

func CopyImage(ctx context.Context, repoName string) (string, error) {
	fmt.Println("Downloading the image...")
	err := downloadImage(repoName)
	if err != nil {
		return "", err
	}
	fmt.Println("Uploading the image...")
	cid, err := uploadImage()
	if err != nil {
		return "", err
	}
	fmt.Println("The multi-arch image is uploaded to the IPFS!")
	return cid, nil
}

func downloadImage(repoName string) error {
	token, err := getCachedOrNewToken(repoName)
	if err != nil {
		return err
	}

	fatManifest, fatManifestRaw, err := getFatManifest(repoName, token)
	if err != nil {
		return err
	}

	dir_manifests := "export/manifests/"
	dir_blobs := "export/blobs/"
	err = fs.CreateDirs([]string{dir_manifests, dir_blobs})
	if err != nil {
		return err
	}
	/*
		err = fs.SaveJson(fatManifest, dir_manifests+"latest")
		if err != nil {
			log.Fatalln(err)
		}
	*/
	err = fs.WriteBytesToFile(dir_manifests+"latest", fatManifestRaw)
	if err != nil {
		return err
	}

	fatManifestSha256, err := fs.Sha256File(dir_manifests + "latest")
	if err != nil {
		return err
	}
	err = fs.WriteBytesToFile(dir_manifests+"sha256:"+fatManifestSha256, fatManifestRaw)
	if err != nil {
		return err
	}

	for _, manifestValue := range fatManifest.Manifests {
		manifest, manifestRaw, err := getManifest(repoName, manifestValue.Digest, token)
		if err != nil {
			return err
		}

		/*
			err = fs.SaveJson(manifest, dir_manifests+manifestValue.Digest)
			if err != nil {
				log.Fatalln(err)
			}
		*/
		err = fs.WriteBytesToFile(dir_manifests+manifestValue.Digest, manifestRaw)
		if err != nil {
			return err
		}

		config, err := getConfig(repoName, manifest.Config.Digest, token)
		if err != nil {
			return err
		}

		err = fs.WriteBytesToFile(dir_blobs+manifest.Config.Digest, config)
		if err != nil {
			return err
		}

		for _, layerValue := range manifest.Layers {
			if err := downloadLayer(repoName, layerValue.Digest, token, dir_blobs+layerValue.Digest); err != nil {
				return err
			}
		}
	}

	return nil
}

func uploadImage() (string, error) {
	fmt.Println("uploadImage")
	cid, err := ipfs.Add("export", true)
	if err != nil {
		return "", err
	}
	return cid, nil
}

func getCachedOrNewToken(repoName string) (string, error) {
	cacheHit, token, err := fs.GetCachedToken(repoName)
	if err != nil {
		return "", err
	}

	if !cacheHit {
		token, err = getToken(repoName)
		if err != nil {
			return "", err
		}
		err = fs.FillCache(repoName, token)
		if err != nil {
			return "", err
		}
		fmt.Println("New token for the repositoy " + repoName + " is fetched and stored.")
	} else {
		fmt.Println("Cached token for the repositoy " + repoName + " is being used.")
	}

	return token, nil
}

func getToken(repoName string) (string, error) {
	url := "https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/" + repoName + ":pull"

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

func getFatManifest(repoName string, token string) (*FatManifest, []byte, error) {
	url := registryEndpoint + repoName + "/manifests/latest"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", strings.Join(acceptList[:], ", "))
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Non-OK HTTP status:", resp.StatusCode)
		return &FatManifest{}, nil, ErrNonOKhttpStatus
	}

	if resp.Header.Get("content-type") != "application/vnd.docker.distribution.manifest.list.v2+json" && resp.Header.Get("content-type") != "application/vnd.oci.image.index.v1+json" {
		return &FatManifest{}, nil, ErrManifestIsNotFat
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var fatManifest FatManifest
	if err := json.Unmarshal(body, &fatManifest); err != nil { // Parse []byte to the go struct pointer
		return nil, nil, err
	}
	return &fatManifest, body, nil
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

	err = os.WriteFile(destination, body, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
