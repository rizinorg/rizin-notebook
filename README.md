# Rizin Notebook

A notebook to write notes while using `rizin`.

If you want to compare it with something similar, you can call it as the rizin equivalent of [jupyter notebook](https://jupyter.org/)

## Requirements

Requires at least rizin version `0.2.0`

## Screenshot

![rizin-notebook](https://raw.githubusercontent.com/rizinorg/rizin-notebook/master/.rizin-notebook.png)

## Building

```bash
# required for go-assets-builder
go get -v github.com/jessevdk/go-assets-builder

go-assets-builder assets -o assets.go
go build -ldflags "-X main.NBVERSION=$(git rev-list -1 HEAD)"
```