package taproot

import (
	"errors"
	"fmt"
	"github.com/google/deck"
	"highgrav/taproot/v1/languages/naive"
	"os"
	"path/filepath"
	"strings"
)

// This simply compiles all the files at startup.
func (srv *Server) compileGoldfusionFiles(srcDirName, dstDirName string) error {
	var retainedError error = nil

	fileOutDir := filepath.Join(srv.Config.ScriptFilePath, dstDirName)
	s, err := os.Stat(filepath.Join(srv.Config.ScriptFilePath, dstDirName))
	if err != nil {
		return errors.New(err.Error() + " (" + fileOutDir + ")")
	}

	if !s.IsDir() {
		return errors.New("Not a directory:  (" + fileOutDir + ")")
	}

	// get scripts
	scripts, err := filepath.Glob(filepath.Join(srcDirName, "*.gf"))
	if err != nil {
		deck.Error("Error reading GF files: " + err.Error())
		os.Exit(-310)
	}
	for _, script := range scripts {
		gfSrc, err := os.ReadFile(script)

		if err != nil {
			deck.Error("Error reading GF JSScript " + script + ": " + err.Error())
			if retainedError == nil {
				retainedError = err
			} else {
				errors.Join(retainedError, err)
			}
			continue
		}

		jsSrc, err := naive.ParseGoldfusionToJS(string(gfSrc))
		if err != nil {
			deck.Error("Error compiling GF to JSScript " + script + ": " + err.Error())
			if retainedError == nil {
				retainedError = err
			} else {
				errors.Join(retainedError, err)
			}
			continue
		}

		// Output the compiled file
		relativeFileName := strings.TrimSuffix(strings.TrimPrefix(script, srcDirName), ".gf") + ".js"
		// TODO -- this is fragile if the user puts './' prefixes in their config file
		jsFileName := filepath.Join(srv.Config.ScriptFilePath, dstDirName, relativeFileName)
		deck.Info(fmt.Sprintf("Transpiled GF %s, moving to %s\n", script, jsFileName))
		newFile, err := os.OpenFile(jsFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		_, err = newFile.Write([]byte(jsSrc))
		if err != nil {
			return err
		}
		newFile.Close()
	}
	return retainedError
}
