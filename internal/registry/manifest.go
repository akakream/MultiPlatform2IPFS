package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/akakream/MultiPlatform2IPFS/internal/fs"
)

func storeFatManifest(fatManifestRaw []byte, dir_manifests string) error {
	err := fs.WriteBytesToFile(filepath.Join(dir_manifests, "latest"), fatManifestRaw)
	if err != nil {
		return err
	}

	fatManifestSha256, err := fs.Sha256File(filepath.Join(dir_manifests, "latest"))
	if err != nil {
		return err
	}
	err = fs.WriteBytesToFile(
		filepath.Join(dir_manifests, "sha256:"+fatManifestSha256),
		fatManifestRaw,
	)
	if err != nil {
		return err
	}
	return nil
}

func getFatManifest(imageName string, imageTag string, token string) (*FatManifest, []byte, error) {
	url := registryEndpoint + imageName + "/manifests/" + imageTag

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

	if resp.Header.Get(
		"content-type",
	) != "application/vnd.docker.distribution.manifest.list.v2+json" &&
		resp.Header.Get("content-type") != "application/vnd.oci.image.index.v1+json" {
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

func getManifest(imageName string, digest string, token string) (Manifest, []byte, error) {
	url := registryEndpoint + imageName + "/manifests/" + digest

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

func getConfig(imageName string, digest string, token string) ([]byte, error) {
	url := registryEndpoint + imageName + "/blobs/" + digest

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
