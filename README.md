# staticfiles

Staticfiles allows you to embed a directory of files into your Go binary. It is optimized for performance and file size, and automatically compresses everything before embedding it.

It has some clever tricks, like only compressing a file if it actually makes the binary smaller (PNG files won't be compressed, as they already are and compressing them again will make them bigger).

I recommend creating a separate package inside your project to serve as the container for the embedded files.

## Example

For an example of how to use the resulting package, check out `example/example.go`. You can also see the API it generates at [godoc.org](https://godoc.org/github.com/bouk/staticfiles/files).

## Usage

Simply run the following command (it will create the result directory if it doesn't exist yet):

```
staticfiles -o files/files.go static/
```

I recommend putting it into a `Makefile` as follows:

```
files/files.go: static/*
	staticfiles -o files/files.go static/
```

The `staticfiles` command accept the following arguments:

```
-build-tags string
      Build tags to write to the file
-o string
      File to write results to. (default "staticfiles.go")
-p string
      Package name of the resulting file. Defaults to name of the resulting file directory
```

## API

The resulting file will contain the following functions an variables:

### `func ServeHTTP(http.ResponseWriter, *http.Request)`

`ServeHTTP` will attempt to serve an embedded file, responding with gzip compression if the clients supports it and the embedded file is compressed.

### `func Open(name string) (io.Reader, error)`

`Open` allows you to read an embedded file directly. It will return a decompressing `Reader` if the file is embedded in compressed format.

### `func ModTime(name string) time.Time`

`ModTime` returns the modification file of the original file. This can be useful for caching purposes.

### `NotFound http.Handler`

`NotFound` is used to respond to a request when no file was found that matches the request. It default to `http.NotFound`, but can be overwritten.

### `Server http.Handler`

`Server` is simply `ServeHTTP` but wrapped in `http.HandlerFunc` so it can be passed into `net/http` functions directly.
