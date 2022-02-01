package main

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func Process(baseDir string, dir string) error {
	results = results[0:0]
	root := path.Join(baseDir, dir)
	_, err := os.Stat(root)
	if err != nil {
		return err
	}

	err = filepath.Walk(
		root,
		func(path string, info fs.FileInfo, err error) error {
			time.Sleep(100 * time.Millisecond) // artificial slow down

			if info.IsDir() { // process dirs
				if strings.HasSuffix(path, ".git") {
					return fs.SkipDir
				}
			} else { // process files
				results = append(results, path)
			}
			return nil
		},
	)

	return err
}
