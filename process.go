package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var mapDoubles map[string][]string

func Process(baseDir string, dir string) error {
	resultsMutex.Lock()
	results = results[0:0]
	resultsMutex.Unlock()
	root := path.Join(baseDir, dir)
	_, err := os.Stat(root)
	if err != nil {
		return err
	}

	mapDoubles = make(map[string][]string)
	err = filepath.Walk(
		root,
		//listProcess,
		DoubleProcess,
	)

	return err
}

// ListProcess for testing purposes.
// global variable results hest updated regularly.
func ListProcess(path string, info fs.FileInfo, err error) error {
	time.Sleep(100 * time.Millisecond) // artificial slow down

	if info.IsDir() { // process dirs
		if strings.HasSuffix(path, ".git") {
			fmt.Println("Skipping .git directory : ", path)
			return fs.SkipDir
		}
	} else { // process files
		resultsMutex.Lock()
		results = append(results, path)
		resultsMutex.Unlock()
	}
	return nil
}

func DoubleProcess(path string, info fs.FileInfo, err error) error {

	if info == nil || err != nil {
		fmt.Println("WalkFunction called with a pre-existing error : ", err)
		return nil
	}

	if info.IsDir() { // process dirs
		if strings.HasSuffix(path, ".git") {
			fmt.Println("Skipping .git directory : ", path)
			return fs.SkipDir
		}
	} else { // fill double map

		if !info.Mode().IsRegular() {
			fmt.Println("Ignoring non regular file  : ", path)
			return nil
		}

		if info.Size() == 0 {
			fmt.Println("Ignoring empty file  : ", path)
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		h := sha1.New()
		_, err = io.Copy(h, file)
		if err != nil {
			return err
		}
		ss := string(h.Sum(nil))
		l := mapDoubles[ss]
		l = append(l, path)
		mapDoubles[ss] = l
	}

	// now, update result list

	rr := make([]string, 0, 50)
	for _, v := range mapDoubles {
		if len(v) >= 2 {
			rr = append(rr, fmt.Sprintf("------------- %4d files have identical content ---------", len(v)))
			rr = append(rr, v...)
		}
	}
	resultsMutex.Lock()
	results = rr
	resultsMutex.Unlock()
	return nil
}
