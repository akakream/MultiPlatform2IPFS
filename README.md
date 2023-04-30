# MultiPlatform2IPFS

MultiPlatform2IPFS is a tool

The folder structure of the multi-platform Docker image that is pushed to IPFS can be seen below. It is created in a way that is compatible with [IPDR](https://github.com/ipdr/ipdr). For each platform there is a folder and the folder name is created by concatenating the os, architecture and variant. The folder structure is kept as flat as possible to avoid IPFS Dag traversals.

```
./manifestlist.json
./linuxamd64
    ./blobs
        ./sha256:...
        ./sha256:...
        ./sha256:...
    ./manifests
        ./latest
        ./sha256:...
./linuxarmv5
    ./blobs
        ./sha256:...
        ./sha256:...
        ./sha256:...
    ./manifests
        ./latest
        ./sha256:...
```

## How to build

`make build`

## How to run

`make run copy busybox`

or

`go run main.go copy busybox`

## TODO

- Add guards checking if the layer already exists on IPFS.
  - If it exists, do not even pull the layer?
  - If it exists, do not try to push obviously? Or if I try to push, it does the right thing and not push?
