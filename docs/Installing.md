# Installing slick

## Prebuilt binaries
Prebuilt binaries are available on [our releases page](https://github.com/1egoman/slick/releases)
for linux and macos. These are built by ci on each push to `master`.

To install, make them executable `chmod +x slick*` and move to `/usr/bin`.

## Building yourself

Something like the below should work:

**TODO: verify this works**

```bash
$ go get github.com/1egoman/slick
$ go build -v
$ slick
```
