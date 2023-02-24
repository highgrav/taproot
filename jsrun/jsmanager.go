package jsrun

import (
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/fsnotify/fsnotify"
	"github.com/google/deck"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ErrScriptNotFound = errors.New("Script not found!")
)

type JSManager struct {
	watchDir        chan bool
	fileDir         string
	fileDirs        []string
	compileMu       sync.Mutex
	compiledScripts map[string]*goja.Program
}

func (jsm *JSManager) GetScript(key string) (*goja.Program, error) {
	fullKey := filepath.Join(jsm.fileDir, key)
	v, ok := jsm.compiledScripts[fullKey]
	if !ok {
		deck.Error("Script '" + key + "' not found for path " + fullKey + "!")
		return nil, ErrScriptNotFound
	}
	return v, nil
}

func New(dir string) (*JSManager, error) {
	s, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !s.IsDir() {
		return nil, errors.New(dir + " is not a directory")
	}

	jsm := &JSManager{}
	jsm.fileDir = dir
	jsm.watchDir = make(chan bool)
	jsm.compiledScripts = make(map[string]*goja.Program)
	err = jsm.CompileAll()
	go jsm.watchDirAndRecompile()
	if err != nil {
		return nil, err
	}
	return jsm, nil
}

// Compiles all the scripts under the source directory
func (jsm *JSManager) CompileAll() error {
	dirList := []string{jsm.fileDir}

	// populate with initial subdirectories
	subdirs, err := jsm.getDirs(jsm.fileDir)
	if err != nil {
		return err
	}
	dirList = append(dirList, subdirs...)
	for _, v := range dirList {
		jsm.compileDir(v)
	}

	return nil
}

// Recursively finds all directories under a path
func (jsm *JSManager) getDirs(dirName string) ([]string, error) {
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
		subDir, err := jsm.getDirs(ds[currIdx])
		if err != nil {
			return nil, err
		}
		ds = append(ds, subDir...)
		currIdx++
	}

	return ds, nil
}

// Compiles all JS files in a directory
func (jsm *JSManager) compileDir(dirName string) error {
	jsm.compileMu.Lock()
	defer jsm.compileMu.Unlock()
	// get scripts
	scripts, err := filepath.Glob(filepath.Join(dirName, "*.js"))
	if err != nil {
		deck.Error("Error reading JS files: " + err.Error())
		os.Exit(-310)
	}
	for _, script := range scripts {
		src, err := os.ReadFile(script)

		if err != nil {
			deck.Error("Error reading JSScript " + script + ": " + err.Error())
			continue
		}

		comp, err := goja.Compile(filepath.Join(dirName, script), string(src), false)
		if err != nil {
			deck.Error("Error compiling JSScript " + script + ": " + err.Error())
			continue
		}
		deck.Info(fmt.Sprintf("Loaded JSScript '%s'\n", script))
		jsm.compiledScripts[script] = comp
	}
	return nil
}

func (jsm *JSManager) CompileOne(script string) error {
	src, err := os.ReadFile(script)

	if err != nil {
		deck.Error("Error reading JSScript " + script + ": " + err.Error())
		return err
	}

	comp, err := goja.Compile(filepath.Join(jsm.fileDir, script), string(src), false)
	if err != nil {
		deck.Error("Error compiling JSScript " + script + ": " + err.Error())
		return err
	}
	deck.Info(fmt.Sprintf("Loaded JSScript '%s'\n", script))
	jsm.compiledScripts[script] = comp
	return nil
}

func (jsm *JSManager) watchDirAndRecompile() {
	dirList := []string{jsm.fileDir}

	// populate with initial subdirectories
	subdirs, err := jsm.getDirs(jsm.fileDir)
	if err != nil {
		deck.Error(err.Error())
		return
	}
	dirList = append(dirList, subdirs...)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		deck.Error(err.Error())
		return
	}
	defer watcher.Close()
	for _, v := range dirList {
		err = watcher.Add(v)
		if err != nil {
			deck.Error(err.Error())
			return
		}
	}
	deck.Info("Watching script file directories")

	for {
		select {
		case exitFlag := <-jsm.watchDir:
			if exitFlag {
				return
			}
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				// If JS file, compile
				// TODO -- Error: Note that for some reason updating a file with 'vi' doesn't trigger an FS notification
				// TODO -- Works as expected with other editors
				if strings.HasSuffix(event.Name, ".js") {
					deck.Info("Recompiling " + event.Name)
					err := jsm.CompileOne(event.Name)
					if err != nil {
						deck.Error("Error when recompiling " + event.Name + ": " + err.Error())
					}
				}
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				// Directory or file?
				// Check to see if this is a directory. If so, need to add watcher to directory
				fileInfo, err := os.Stat(event.Name)
				if err != nil {
					deck.Error("Error when reading created file " + event.Name + ": " + err.Error())
				} else {
					if fileInfo.IsDir() {
						deck.Info("Watching new directory " + event.Name)
						err = watcher.Add(event.Name)
						if err != nil {
							deck.Error("Error when adding watcher to created dir " + event.Name + ": " + err.Error())
						}
					}

					// It's a file, so try compiling it
					if strings.HasSuffix(event.Name, ".js") {
						deck.Info("Recompiling " + event.Name)
						err := jsm.CompileOne(event.Name)
						if err != nil {
							deck.Error("Error when recompiling " + event.Name + ": " + err.Error())
						}
					}
				}
			}

			if event.Op&fsnotify.Remove == fsnotify.Remove {
				// We can't fstat a removed file, so...
				// Try to remove watcher from directory
				watcher.Remove(event.Name)
				// Try to remove the filename from cache
				_, ok := jsm.compiledScripts[event.Name]
				if ok {
					delete(jsm.compiledScripts, event.Name)
					deck.Info("Removed compiled script " + event.Name)
				}

			}
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				// We can't fstat a renamed file either, so...
				// Try to remove watcher from directory
				watcher.Remove(event.Name)
				// Try to remove the filename from cache
				_, ok := jsm.compiledScripts[event.Name]
				if ok {
					delete(jsm.compiledScripts, event.Name)
					deck.Info("Removed compiled script " + event.Name)
				}
				// A rename fires off a create event also, so it'll handle
				// watcher/compilation in that block
			}
			continue
		case err := <-watcher.Errors:
			deck.Error("Error in JS filewatcher: " + err.Error())
			continue
		}
	}
}
