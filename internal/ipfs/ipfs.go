package ipfs

import (
	"fmt"
	"os"

	shell "github.com/ipfs/go-ipfs-api"
)

// Add adds a directory to IPFS. If willPin is true, the added item is pinned.
func Add(dirPath string, willPin bool) {
	// Where your local node is running on localhost:5001
	sh := shell.NewShell("localhost:5001")
	cid, err := sh.AddDir(dirPath, shell.CidVersion(1), shell.Pin(willPin))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("added %s \n", cid)
}
