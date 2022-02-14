package utils

import (
	"io/fs"
	"io/ioutil"
)

func ListFiles(directory string) ([]fs.FileInfo, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	return files, nil
}
