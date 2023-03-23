package common

import (
	"github.com/google/deck"
	"os"
	"path/filepath"
	"strings"
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

func SliceDirectory(basePath string) []string {
	elems := make([]string, 0)

	res := filepath.Base(basePath)
	basePath = strings.TrimSuffix(basePath, res)
	elems = append([]string{res}, elems...)

	for res != string(os.PathSeparator) && res != "." {
		res = filepath.Base(basePath)
		elems = append([]string{res}, elems...)
		basePath = strings.TrimSuffix(basePath, res+string(os.PathSeparator))
	}
	return elems
}

// For use primarily to find the original source of compiled and moved JSML files.
// Takes the base path to search, then
func FindRelocatedFile(toPath string, fromPath string) (string, error) {
	elems := SliceDirectory(fromPath)
	for x := 0; x < len(elems); x++ {
		fPath := toPath + string(os.PathSeparator) + strings.Join(elems[x:], string(os.PathSeparator))
		_, err := os.Stat(fPath)
		if err == nil {
			return fPath, nil
		}
	}
	return "", os.ErrNotExist
}
