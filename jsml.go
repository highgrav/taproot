package taproot

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/google/deck"
	"github.com/highgrav/taproot/v1/common"
	"github.com/highgrav/taproot/v1/languages/jsmltranspiler"
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
		return "", nil
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
	fmt.Sprintf("jsml transpilation %s, from %s to %s\n", fileName, srcDirName, dstDirName)
	var sa ScriptAccessor = ScriptAccessor{
		srv: srv,
	}
	gfSrc, err := os.ReadFile(fileName)

	if err != nil {
		return err
	}

	trans, err := jsmltranspiler.NewAndTranspile(sa, string(gfSrc), false)
	if err != nil {
		return err
	}
	err = trans.ToJS()
	if err != nil {
		return err
	}

	jsSrc := trans.Builder().String()

	// Output the compiled file
	relativeFileName := strings.TrimSuffix(strings.TrimPrefix(fileName, srcDirName), ".jsml") + ".js"
	// TODO -- this is fragile if the user puts './' prefixes in their config file
	jsFileName := filepath.Join(srv.Config.ScriptFilePath, dstDirName, relativeFileName)
	deck.Info(fmt.Sprintf("Transpiled JSML %s, moving to %s\n", fileName, filepath.Dir(jsFileName)))
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
		deck.Error("Error finding JSML files: " + err.Error())
		os.Exit(-310)
	}
	for _, script := range scripts {
		gfSrc, err := os.ReadFile(script)

		if err != nil {
			deck.Error("Error reading JSML JSScript " + script + ": " + err.Error())
			if retainedError == nil {
				retainedError = err
			} else {
				errors.Join(retainedError, err)
			}
			continue
		}

		trans, err := jsmltranspiler.NewAndTranspile(sa, string(gfSrc), false)
		if err != nil {
			deck.Error("Error compiling JSML to JSScript " + script + ": " + err.Error())
			if retainedError == nil {
				retainedError = err
			} else {
				errors.Join(retainedError, err)
			}
			continue
		}
		err = trans.ToJS()
		if err != nil {
			deck.Error("Error compiling JSML to JSScript " + script + ": " + err.Error())
			if retainedError == nil {
				retainedError = err
			} else {
				errors.Join(retainedError, err)
			}
			continue
		}

		jsSrc := trans.Builder().String()

		// Output the compiled file
		relativeFileName := strings.TrimSuffix(strings.TrimPrefix(script, srcDirName), ".jsml") + ".js"
		// TODO -- this is fragile if the user puts './' prefixes in their config file
		jsFileName := filepath.Join(srv.Config.ScriptFilePath, dstDirName, relativeFileName)
		deck.Info(fmt.Sprintf("Transpiled JSML %s, moving to %s\n", script, filepath.Dir(jsFileName)))
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

	// go srv.monitorJSMLDirectories(srcDirName, dstDirName)
	return retainedError
}

// Starts monitoring of the JSML file directories to catch any updates and recompile accordingly.
// TODO -- this is having issues detecting changes
func (srv *AppServer) monitorJSMLDirectories(srcDirName, dstDirName string) {
	dirList := []string{srcDirName}
	subdirs, err := common.GetDirs(srcDirName)
	if err != nil {
		deck.Error("jsml file monitoring cannot be started")
		deck.Error(err.Error())
		return
	}
	dirList = append(dirList, subdirs...)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		deck.Error("jsml file monitoring cannot be started")
		deck.Error(err.Error())
		return
	}
	for _, d := range dirList {
		watcher.Add(d)
	}
	deck.Info("Watching jsml file directories: ")
	for {
		select {
		case exitFlag := <-srv.ExitServerCh:
			if exitFlag {
				deck.Info("shutting down jsml filewatcher")
				return
			}
		case err := <-watcher.Errors:
			deck.Error("error in jsml filewatcher")
			deck.Error(err.Error())
			return
		case event := <-watcher.Events:
			if event.Has(fsnotify.Write) {
				if strings.HasSuffix(event.Name, ".jsml") {
					err := srv.compileOne(event.Name, srcDirName, dstDirName)
					if err != nil {
						deck.Error("jsml error transpiling " + event.Name + ": " + err.Error())
					}
				}
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				// TODO
				s, err := os.Stat(event.Name)
				if err != nil {
					deck.Error("jsml error handling created file " + event.Name + ": " + err.Error())
				} else {
					if s.IsDir() {
						// if a directory, add a watcher
						deck.Info("Caught create event on JSML directory " + event.Name)
						watcher.Add(event.Name)
					} else {
						// if a file, compile it

						if strings.HasSuffix(event.Name, ".jsml") {
							deck.Info("Caught create event on JSML file " + event.Name)
							err := srv.compileOne(event.Name, srcDirName, dstDirName)
							if err != nil {
								deck.Error("jsml error transpiling " + event.Name + ": " + err.Error())
							}
						}
					}
				}
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				deck.Info("Caught remove event on JSML file " + event.Name)
				watcher.Remove(event.Name)
				// delete the compiled JS file
				// TODO -- find relocated file and delete it
				fileName, err := common.FindRelocatedFile(dstDirName, event.Name[:len(event.Name)-2])
				if err != nil {
					deck.Error("jsml error dealing with deleted file " + event.Name + ": " + err.Error())
				} else {
					// TODO
					deck.Error("TODO: Delete " + fileName)
				}
			}
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				watcher.Remove(event.Name)
				// TODO
				// rename the compiled JS file
				fileName, err := common.FindRelocatedFile(dstDirName, event.Name[:len(event.Name)-2])
				if err != nil {
					deck.Error("jsml error dealing with deleted file " + event.Name + ": " + err.Error())
				} else {
					// TODO
					deck.Error("TODO: Delete " + fileName)
				}
			}
		}
		return
	}
}
