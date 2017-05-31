# Installing slick

## Prebuilt binaries
Prebuilt binaries are available on [our releases page](https://github.com/1egoman/slick/releases)
for linux and macos. These are built by ci on each push to `master`.

To install, make them executable `chmod +x slick*` and move to `/usr/bin`.

## Building yourself

Something like the below should work:

```bash
$ go version
go version go1.8.3 darwin/amd64 # must have go1.8.x or later!
$ go get github.com/1egoman/slick
$ go build -v
$ slick
```
