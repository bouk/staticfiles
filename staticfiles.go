package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"flag"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type file struct {
	name  string
	data  string
	mime  string
	mtime time.Time
	// size is the result size before compression. If 0, it means the data is uncompressed
	size int64
	hash []byte
}

func skipFile(file string, excludeSlice []string) bool {
	for _, pattern := range excludeSlice {
		matched, err := path.Match(pattern, file)
		if err != nil {
			log.Fatal(err)
		}
		if matched {
			return true
		}
	}
	return false
}

func processDir(c chan [2]string, dir string, parents []string, excludeSlice []string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		n := filepath.Join(dir, file.Name())
		id := append(parents, file.Name())
		if skipFile(path.Join(id...), excludeSlice) {
			continue
		}

		if file.IsDir() {
			processDir(c, n, id, excludeSlice)
		} else {
			c <- [2]string{n, path.Join(id...)}
		}
	}
}

type fileSlice []*file

func (a fileSlice) Len() int           { return len(a) }
func (a fileSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a fileSlice) Less(i, j int) bool { return a[i].name < a[j].name }

func main() {
	var outputFile, packageName, buildTags, excludeList string
	flag.StringVar(&outputFile, "o", "staticfiles.go", "File to write results to.")
	flag.StringVar(&packageName, "package", "", "Package name of the resulting file. Defaults to name of the resulting file directory")
	flag.StringVar(&buildTags, "build-tags", "", "Build tags to write to the file")
	flag.StringVar(&excludeList, "exclude", "", "Comma-separated patterns to exclude. (e.g. '*.scss,templates/')")
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
	excludeSlice := strings.Split(excludeList, ",")

	c := make(chan [2]string, 128)
	go func() {
		for _, arg := range flag.Args() {
			processDir(c, arg, []string{}, excludeSlice)
		}
		close(c)
	}()

	// Used to communicate back all the files that were read and compressed.
	results := make(chan fileSlice)
	process := func() {
		var b bytes.Buffer
		var b2 bytes.Buffer
		hash := sha256.New()
		files := make(fileSlice, 0, 32)
		for asset := range c {
			f, err := os.Open(asset[0])
			if err != nil {
				log.Fatal(err)
			}
			stat, err := f.Stat()
			if err != nil {
				log.Fatal(err)
			}
			if _, err := b.ReadFrom(f); err != nil {
				log.Fatal(err)
			}
			f.Close()
			compressedWriter, _ := gzip.NewWriterLevel(&b2, gzip.BestCompression)
			writer := io.MultiWriter(compressedWriter, hash)
			if _, err := writer.Write(b.Bytes()); err != nil {
				log.Fatal(err)
			}
			compressedWriter.Close()
			if b2.Len() < b.Len() {
				files = append(files, &file{
					name:  asset[1],
					data:  b2.String(),
					mime:  mime.TypeByExtension(filepath.Ext(asset[0])),
					mtime: stat.ModTime(),
					size:  stat.Size(),
					hash:  hash.Sum(nil),
				})
			} else {
				files = append(files, &file{
					name:  asset[1],
					data:  b.String(),
					mime:  mime.TypeByExtension(filepath.Ext(asset[0])),
					mtime: stat.ModTime(),
					hash:  hash.Sum(nil),
				})
			}
			b.Reset()
			b2.Reset()
			hash.Reset()
		}
		results <- files
	}
	// Concurrency! Read and compress in parallel, and combine all the slices at the end.
	concurrency := runtime.NumCPU()
	for i := 0; i < concurrency; i++ {
		go process()
	}
	var files fileSlice
	for i := 0; i < concurrency; i++ {
		files = append(files, (<-results)...)
	}

	// Should sort to make sure the output is stable
	sort.Sort(files)

	var b bytes.Buffer
	if err := GenerateTemplate(&b, packageName, files, buildTags); err != nil {
		log.Fatal(err)
	}
	res, err := format.Source(b.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
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
