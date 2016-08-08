package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path"
	"path/filepath"
	"time"
)

type file struct {
	name             string
	data             string
	mime             string
	mtime            time.Time
	uncompressedSize int64
}

func processDir(c chan [2]string, dir string, parents []string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		n := filepath.Join(dir, file.Name())
		id := append(parents, file.Name())
		if file.IsDir() {
			processDir(c, n, id)
		} else {
			c <- [2]string{n, path.Join(id...)}
		}
	}
}

func main() {
	var outputFile, packageName string
	flag.StringVar(&outputFile, "o", "files.go", "File to write results to.")
	flag.StringVar(&packageName, "p", "", "Package name of the resulting file. Defaults to name of the resulting file directory")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Println("Please pass in a directory to process")
		flag.PrintDefaults()
		os.Exit(1)
		return
	}
	if packageName == "" {
		f, err := filepath.Abs(outputFile)
		if err != nil {
			log.Fatal(err)
		}
		packageName = filepath.Base(filepath.Dir(f))
	}
	c := make(chan [2]string, 128)
	go func() {
		for _, arg := range flag.Args() {
			processDir(c, arg, []string{})
		}
		close(c)
	}()

	files := make([]*file, 0, 32)
	var b bytes.Buffer
	for asset := range c {
		f, err := os.Open(asset[0])
		if err != nil {
			log.Fatal(err)
		}
		stat, err := f.Stat()
		if err != nil {
			log.Fatal(err)
		}
		writer, _ := gzip.NewWriterLevel(&b, gzip.BestCompression)

		if _, err = io.Copy(writer, f); err != nil {
			log.Fatal(err)
		}
		files = append(files, &file{
			name:             asset[1],
			data:             b.String(),
			mime:             mime.TypeByExtension(filepath.Ext(asset[0])),
			mtime:            stat.ModTime(),
			uncompressedSize: stat.Size(),
		})
		b.Reset()
	}
	if err := GenerateTemplate(&b, packageName, files); err != nil {
		log.Fatal(err)
	}
	res, err := format.Source(b.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write(res); err != nil {
		log.Fatal(err)
	}
}
