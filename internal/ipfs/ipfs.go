package ipfs

import (
	"fmt"

	shell "github.com/ipfs/go-ipfs-api"
)

// Add adds a directory to IPFS. If willPin is true, the added item is pinned.
func Add(dirPath string, willPin bool) (string, error) {
	// Where your local node is running on localhost:5001
	sh := shell.NewShell("localhost:5001")
	cid, err := sh.AddDir(dirPath, shell.CidVersion(1), shell.Pin(willPin))
	if err != nil {
		return "", err
	}
	fmt.Printf("added %s \n", cid)
	return cid, nil
}
