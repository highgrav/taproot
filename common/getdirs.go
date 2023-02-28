package common

import (
	"github.com/google/deck"
	"os"
	"path/filepath"
)

func GetDirs(dirName string) ([]string, error) {
	fs, err := os.ReadDir(dirName)
	if err != nil {
		deck.Error("Error reading " + dirName)
		return nil, err
	}
	ds := []string{}
	currIdx := 0
	for _, v := range fs {
		if v.IsDir() {
			ds = append(ds, filepath.Join(dirName, v.Name()))
		}
	}

	for currIdx < len(ds) {
		subDir, err := GetDirs(ds[currIdx])
		if err != nil {
			return nil, err
		}
		ds = append(ds, subDir...)
		currIdx++
	}

	return ds, nil
}
