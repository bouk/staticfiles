# staticfiles

Staticfiles allows you to embed a directory of files into your Go binary. It is optimized for performance and file size, and automatically compresses everything before embedding it. Here are some of its features:

* Compresses files, to make sure the resulting binary isn't bloated. It only compresses files that are actually smaller when `gzip`ped.
* Serves files `gzip`ped (while still allowing clients that don't support it to be served).
* Ignores hidden files (anything that starts with `.`).
* Fast. The command-line tool reads and compresses files in parallel, and the resulting Go file serves files very quickly, avoiding unnecessary allocations.
* No built-in development mode, but makes it very easy to implement one (see [local development mode](#local-development-mode)).

It has some clever tricks, like only compressing a file if it actually makes the binary smaller (PNG files won't be compressed, as they already are and compressing them again will make them bigger).

I recommend creating a separate package inside your project to serve as the container for the embedded files.

## Example

For an example of how to use the resulting package, check out `example/example.go`. You can also see the API it generates at [godoc.org](https://godoc.org/github.com/bouk/staticfiles/files).

## Installation

Install with

```
go get github.com/bouk/staticfiles
```

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
--build-tags string
      Build tags to write to the file
-o string
      File to write results to. (default "staticfiles.go")
--package string
      Package name of the resulting file. Defaults to name of the resulting file directory
```

## Local development mode

While Staticfiles doesn't have a built-in local development mode, it does support build tags which makes implementing one very easy. Simply run `staticfiles` with `--build-tags="!dev"` and add a file in the same directory that implements the same API, but with `//+build dev` at the that and using `http.FileServer` under the hood. You can find an example in `files/files_dev.go`. Once you have that set up you can simply do `go build --tags="dev"` to compile the development version. In the way I set it up, you could even do `go build --tags="dev" -ldflags="-X github.com/bouk/staticfiles/files.staticDir=$(pwd)/static"` to set the static file directory to a specific path.

## API

The resulting file will contain the following functions and variables:

### `func ServeHTTP(http.ResponseWriter, *http.Request)`

`ServeHTTP` will attempt to serve an embedded file, responding with gzip compression if the clients supports it and the embedded file is compressed.

### `func Open(name string) (io.ReadCloser, error)`

`Open` allows you to read an embedded file directly. It will return a decompressing `Reader` if the file is embedded in compressed format. You should close the `Reader` after you're done with it.

### `func ModTime(name string) time.Time`

`ModTime` returns the modification time of the original file. This can be useful for caching purposes.

### `NotFound http.Handler`

`NotFound` is used to respond to a request when no file was found that matches the request. It defaults to `http.NotFound`, but can be overwritten.

### `Server http.Handler`

`Server` is simply `ServeHTTP` but wrapped in `http.HandlerFunc` so it can be passed into `net/http` functions directly.
