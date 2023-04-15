# MultiPlatform2IPFS

## How to build

`make build`

## How to run

`make run copy busybox`

or

`go run main.go copy busybox`

## TODO

Add guards checking if the layer already exists on IPFS.
If it exists, do not even pull the layer?
If it exists, do not try to push obviously? Or if I try to push, it does the right thing and not push?
