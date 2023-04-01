package taproot

import (
	"context"
	"errors"
	"fmt"
	"github.com/highgrav/taproot/v1/common"
	"github.com/highgrav/taproot/v1/languages/jsmltranspiler"
	"github.com/highgrav/taproot/v1/logging"
	"os"
	"path/filepath"
	"strings"
)

// A small struct for managing JSML scripts
type ScriptAccessor struct {
	srv *AppServer
}

// Gets a JSML file (used when attempting to compile a file directly or via inclusion)
func (sa ScriptAccessor) GetJSMLScriptByID(id string) (string, error) {
	// Make sure we're looking for a JSML file
	if strings.HasSuffix(id, ".js") {
		id = id + "ml"
	}
	basePath := sa.srv.Config.JSMLFilePath
	// use the common.FindRelocatedFile() to try to locate the errant uncompiled JSML file
	fileName, err := common.FindRelocatedFile(basePath, id)
	if err != nil {
		return "", err
	}
	script, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	return string(script), nil
}

// Gets the compiled JS for a JSML file.
func (sa ScriptAccessor) GetJSScriptByID(id string) (string, error) {
	return sa.srv.js.GetScriptText(id)
}

func (srv *AppServer) compileOne(fileName string, srcDirName string, dstDirName string) error {
	logging.LogToDeck(context.Background(), "info", "JSML", "info", fmt.Sprintf("transpilation %s, from %s to %s\n", fileName, srcDirName, dstDirName))
	var sa ScriptAccessor = ScriptAccessor{
		srv: srv,
	}
	gfSrc, err := os.ReadFile(fileName)

	if err != nil {
		return err
	}

	trans, err := jsmltranspiler.NewAndTranspile(strings.TrimPrefix(fileName, srcDirName), sa, string(gfSrc), false)
	if err != nil {
		return err
	}
	err = trans.ToJS()
	if err != nil {
		return err
	}

	// update the import list
	for _, imp := range trans.GetImports() {
		srv.js.Dependencies.AddDependency(imp, trans.ID)
	}

	jsSrc := trans.Builder().String()

	// Output the compiled file
	relativeFileName := strings.TrimSuffix(strings.TrimPrefix(fileName, srcDirName), ".jsml") + ".js"
	// TODO -- this is fragile if the user puts './' prefixes in their config file
	// TODO srv.Config.ScriptFilePath,
	jsFileName := filepath.Join(dstDirName, relativeFileName)
	logging.LogToDeck(context.Background(), "info", "JSML", "info", fmt.Sprintf("transpiled JSML %s, moving to %s\n", fileName, filepath.Dir(jsFileName)))
	// create directory path
	err = os.MkdirAll(filepath.Dir(jsFileName), 0777) // TODO -- fileperm
	if err != nil {
		return err
	}
	newFile, err := os.OpenFile(jsFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	_, err = newFile.Write([]byte(jsSrc))
	if err != nil {
		return err
	}
	newFile.Close()
	return nil
}

// This simply compiles all the JSML files at startup.
func (srv *AppServer) compileJSMLFiles(srcDirName, dstDirName string) error {
	var sa ScriptAccessor = ScriptAccessor{
		srv: srv,
	}
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
	var scripts []string = make([]string, 0)

	err = filepath.Walk(srcDirName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".jsml") {
			scripts = append(scripts, path)
		}
		return nil
	})
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "JSML", "error", "error finding JSML files: "+err.Error())
		os.Exit(-310)
	}
	for _, script := range scripts {
		gfSrc, err := os.ReadFile(script)

		if err != nil {
			logging.LogToDeck(context.Background(), "error", "JSML", "error", "error reading JSML JSScript "+script+": "+err.Error())
			if retainedError == nil {
				retainedError = err
			} else {
				errors.Join(retainedError, err)
			}
			continue
		}

		trans, err := jsmltranspiler.NewAndTranspile(script, sa, string(gfSrc), false)
		if err != nil {
			logging.LogToDeck(context.Background(), "error", "JSML", "error", "error compiling JSML to JSScript "+script+": "+err.Error())
			if retainedError == nil {
				retainedError = err
			} else {
				errors.Join(retainedError, err)
			}
			continue
		}
		err = trans.ToJS()
		if err != nil {
			logging.LogToDeck(context.Background(), "error", "JSML", "error", "error compiling JSML to JSScript "+script+": "+err.Error())
			if retainedError == nil {
				retainedError = err
			} else {
				errors.Join(retainedError, err)
			}
			continue
		}

		// update the import list
		for _, imp := range trans.GetImports() {
			srv.js.Dependencies.AddDependency(imp, trans.ID)
		}

		jsSrc := trans.Builder().String()

		// Output the compiled file
		relativeFileName := strings.TrimSuffix(strings.TrimPrefix(script, srcDirName), ".jsml") + ".js"
		// TODO -- this is fragile if the user puts './' prefixes in their config file
		// TODO srv.Config.ScriptFilePath,
		jsFileName := filepath.Join(srv.Config.ScriptFilePath, dstDirName, relativeFileName)
		logging.LogToDeck(context.Background(), "info", "JSML", "info", fmt.Sprintf("transpiled initial JSML %s, moving to %s\n", script, filepath.Dir(jsFileName)))
		// create directory path
		err = os.MkdirAll(filepath.Dir(jsFileName), 0777) // TODO -- fileperm
		if err != nil {
			return err
		}
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
