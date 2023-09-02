package registry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"

	"github.com/akakream/MultiPlatform2IPFS/internal/fs"
	"github.com/akakream/MultiPlatform2IPFS/internal/ipfs"
	"github.com/akakream/MultiPlatform2IPFS/utils"
)

var acceptList = [9]string{
	"application/vnd.docker.distribution.manifest.v1+json",
	"application/vnd.docker.distribution.manifest.v2+json",
	"application/vnd.docker.distribution.manifest.list.v2+json",
	"application/vnd.docker.container.image.v1+json",
	"application/vnd.docker.image.rootfs.diff.tar.gzip",
	"application/vnd.docker.image.rootfs.foreign.diff.tar.gzip",
	"application/vnd.docker.plugin.v1+json",
	"application/vnd.oci.image.index.v1+json",
	"application/vnd.oci.image.manifest.v1+json",
}

var (
	// ErrManifestIsNotFat is error for when the repository is not multi-platform
	ErrManifestIsNotFat = errors.New("the repository is not multi-platform")
	// ErrNonOKhttpStatus is error for when the http status is not OK.
	ErrNonOKhttpStatus = errors.New("the http status is not OK")
)

const registryEndpoint = "https://index.docker.io/v2/library/"

// const registryEndpoint = "https://registry-1.docker.io/v2/library/"

func CopyImage(ctx context.Context, imageName string, imageTag string) (string, error) {
	fmt.Println("Removing existing files under the export directory...")
	clearExportPath()

	fmt.Println("Downloading the image...")
	err := downloadImage(imageName, imageTag)
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

func createFolderStructure() (string, string, error) {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	exportPath, err := utils.GetEnv("EXPORT_PATH", "./export")
	if err != nil {
		panic(err)
	}

	dir_manifests := filepath.Join(exportPath, "manifests")
	dir_blobs := filepath.Join(exportPath, "blobs")
	err = fs.CreateDirs([]string{dir_manifests, dir_blobs})
	if err != nil {
		return "", "", err
	}

	return dir_manifests, dir_blobs, err
}

func downloadImage(imageName string, imageTag string) error {
	token, err := getCachedOrNewToken(imageName, imageTag)
	if err != nil {
		return err
	}

	dir_manifests, dir_blobs, err := createFolderStructure()
	if err != nil {
		return err
	}

	fatManifest, fatManifestRaw, err := getFatManifest(imageName, imageTag, token)
	downloadWG := sync.WaitGroup{}

	if err != nil {
		log.Println("For the provided repository name, there is no Fat Manifest.")
		log.Print(err)
		err = getManifestWithLayers(
			imageName,
			"latest",
			dir_manifests,
			dir_blobs,
			token,
			&downloadWG,
		)
		if err != nil {
			return err
		}
	} else {
		err = storeFatManifest(fatManifestRaw, dir_manifests)
		if err != nil {
			return err
		}

		for _, manifestValue := range fatManifest.Manifests {
			getManifestWithLayers(imageName, manifestValue.Digest, dir_manifests, dir_blobs, token, &downloadWG)
		}
	}

	downloadWG.Wait()
	return nil
}

func getManifestWithLayers(
	imageName string,
	manifestDigest string,
	dir_manifests string,
	dir_blobs string,
	token string,
	downloadWG *sync.WaitGroup,
) error {
	manifest, manifestRaw, err := getManifest(imageName, manifestDigest, token)
	if err != nil {
		return err
	}

	err = fs.WriteBytesToFile(filepath.Join(dir_manifests, manifestDigest), manifestRaw)
	if err != nil {
		return err
	}

	config, err := getConfig(imageName, manifest.Config.Digest, token)
	if err != nil {
		return err
	}

	err = fs.WriteBytesToFile(filepath.Join(dir_blobs, manifest.Config.Digest), config)
	if err != nil {
		return err
	}

	for _, layerValue := range manifest.Layers {
		// TODO: ADD RETRY HERE
		downloadWG.Add(1)
		go downloadLayer(
			imageName,
			layerValue.Digest,
			token,
			filepath.Join(dir_blobs, layerValue.Digest),
			downloadWG,
		)
	}
	return nil
}

func uploadImage() (string, error) {
	fmt.Println("uploadImage")
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	exportPath, err := utils.GetEnv("EXPORT_PATH", "./export")
	if err != nil {
		panic(err)
	}
	cid, err := ipfs.Add(exportPath, true)
	if err != nil {
		return "", err
	}
	return cid, nil
}

func clearExportPath() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	exportPath, err := utils.GetEnv("EXPORT_PATH", "./export")
	if err != nil {
		panic(err)
	}

	os.RemoveAll(exportPath)
}
