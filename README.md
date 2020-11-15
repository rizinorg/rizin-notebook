# Rizin Notebook

A notebook to write notes while using `rizin`.

## Screenshot

![rizin-notebook](https://raw.githubusercontent.com/rizinorg/rizin-notebook/master/.rizin-notebook.png)

## Building

```bash
go get -v github.com/gin-gonic/gin
go get -v github.com/jessevdk/go-assets-builder

go-assets-builder assets -o assets.go
go build
```