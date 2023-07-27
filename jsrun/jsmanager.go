package jsrun

import (
	"context"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/fsnotify/fsnotify"
	"github.com/highgrav/taproot/logging"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ErrScriptNotFound = errors.New("Script not found!")
)

type JSRunOnFileEventFn func(string)

type JSManager struct {
	watchDir         chan bool
	fileDir          string
	fileDirs         []string
	compileMu        sync.Mutex
	scripts          map[string]string
	compiledScripts  map[string]*goja.Program
	fnRunOnRecompile JSRunOnFileEventFn
	fnRunOnDelete    JSRunOnFileEventFn
	Dependencies     JSDependencies
}

func New(dir string, runOnRecompile, runOnDelete JSRunOnFileEventFn) (*JSManager, error) {
	s, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !s.IsDir() {
		return nil, errors.New(dir + " is not a directory")
	}

	jsm := &JSManager{}
	jsm.Dependencies = JSDependencies{
		importingScripts: make(map[string][]string),
	}
	jsm.fnRunOnDelete = runOnDelete
	jsm.fnRunOnRecompile = runOnRecompile
	jsm.fileDir = dir
	jsm.watchDir = make(chan bool)
	jsm.compiledScripts = make(map[string]*goja.Program)
	jsm.scripts = make(map[string]string)
	logging.LogToDeck(context.Background(), "info", "JS", "-", "jsmanager compiling scripts")
	err = jsm.CompileAll()
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "JS", "error", err.Error())
		return nil, err
	}
	logging.LogToDeck(context.Background(), "info", "JS", "-", "jsmanager done compiling")
	go jsm.watchDirAndRecompile()
	if err != nil {
		return nil, err
	}
	logging.LogToDeck(context.Background(), "info", "JS", "-", "jsmanager ready")
	return jsm, nil
}

func (jsm *JSManager) GetScriptText(key string) (string, error) {
	if strings.HasSuffix(key, "jsml") {
		key = key[:len(key)-2]
	}
	fullKey := filepath.Join(jsm.fileDir, key)
	script, ok := jsm.scripts[fullKey]
	if !ok {
		err := jsm.CompileOne(fullKey)
		if err != nil {
			return "", errors.New("could not locate script '" + key + "': " + err.Error())
		}
	}
	return script, nil
}

func (jsm *JSManager) GetScript(key string) (*goja.Program, error) {
	fullKey := filepath.Join(jsm.fileDir, key)
	v, ok := jsm.compiledScripts[fullKey]
	if !ok {
		return nil, errors.New("Script '" + key + "' not found for path " + fullKey + "!")
	}
	return v, nil
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
		logging.LogToDeck(context.Background(), "error", "JS", "error", "getDirs(): Error reading "+dirName)
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
		logging.LogToDeck(context.Background(), "fatal", "JS", "fatal", "error reading JS files: "+err.Error())
		os.Exit(-310)
	}
	for _, script := range scripts {
		src, err := os.ReadFile(script)

		if err != nil {
			logging.LogToDeck(context.Background(), "error", "JS", "error", "error reading JSScript "+script+": "+err.Error())
			continue
		}

		comp, err := goja.Compile(filepath.Join(dirName, script), string(src), false)
		if err != nil {
			logging.LogToDeck(context.Background(), "error", "JS", "error", "error compiling JSScript "+script+": "+err.Error())
			continue
		}
		logging.LogToDeck(context.Background(), "info", "JS", "info", fmt.Sprintf("loaded JSScript '%s'\n", script))
		jsm.compiledScripts[script] = comp
		jsm.scripts[script] = string(src) // TODO -- make sure this preserves unicode
	}
	return nil
}

func (jsm *JSManager) CompileOne(script string) error {
	src, err := os.ReadFile(script)

	if err != nil {
		logging.LogToDeck(context.Background(), "error", "JS", "error", "error reading JSScript "+script+": "+err.Error())
		return err
	}

	comp, err := goja.Compile(filepath.Join(jsm.fileDir, script), string(src), false)
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "JS", "error", "error compiling JSScript "+script+": "+err.Error())
		return err
	}
	logging.LogToDeck(context.Background(), "info", "JS", "info", fmt.Sprintf("loaded JSScript '%s'\n", script))
	jsm.compiledScripts[script] = comp
	jsm.scripts[script] = string(src)
	return nil
}

func (jsm *JSManager) CompileOneAs(key string, script string) error {
	comp, err := goja.Compile(key, script, false)
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "JS", "error", "error compiling JSScript "+script+": "+err.Error())
		return err
	}
	logging.LogToDeck(context.Background(), "info", "JS", "info", fmt.Sprintf("loaded JSScript '%s'\n", key))
	jsm.compiledScripts[key] = comp
	jsm.scripts[key] = string(script)
	return nil
}

func (jsm *JSManager) watchDirAndRecompile() {
	dirList := []string{jsm.fileDir}

	// populate with initial subdirectories
	subdirs, err := jsm.getDirs(jsm.fileDir)
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "JS", "error", "JS file watcher could not be started: "+err.Error())
		return
	}
	dirList = append(dirList, subdirs...)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "JS", "error", "error creating JS filewatcher: "+err.Error())
		return
	}
	defer watcher.Close()
	for _, v := range dirList {
		err = watcher.Add(v)
		if err != nil {
			logging.LogToDeck(context.Background(), "error", "JS", "error", "error watching directory "+v+": "+err.Error())
			return
		}
	}
	logging.LogToDeck(context.Background(), "info", "JS", "info", "watching script file directories")

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
					logging.LogToDeck(context.Background(), "info", "JS", "info", "recompiling "+event.Name)
					err := jsm.CompileOne(event.Name)
					if err != nil {
						logging.LogToDeck(context.Background(), "error", "JS", "error", "error when recompiling "+event.Name+": "+err.Error())
					}
					jsm.fnRunOnRecompile(event.Name)
				}
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				// Directory or file?
				// Check to see if this is a directory. If so, need to add watcher to directory
				fileInfo, err := os.Stat(event.Name)
				if err != nil {
					logging.LogToDeck(context.Background(), "error", "JS", "error", "error when reading created file "+event.Name+": "+err.Error())
				} else {
					if fileInfo.IsDir() {
						logging.LogToDeck(context.Background(), "info", "JS", "info", "watching new directory "+event.Name)
						err = watcher.Add(event.Name)
						if err != nil {
							logging.LogToDeck(context.Background(), "error", "JS", "error", "error when adding watcher to created dir "+event.Name+": "+err.Error())
						}
					}

					// It's a file, so try compiling it
					if strings.HasSuffix(event.Name, ".js") {
						logging.LogToDeck(context.Background(), "info", "JS", "info", "recompiling "+event.Name)
						err := jsm.CompileOne(event.Name)
						if err != nil {
							logging.LogToDeck(context.Background(), "error", "JS", "error", "error when recompiling "+event.Name+": "+err.Error())
						}
						jsm.fnRunOnRecompile(event.Name)
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
					delete(jsm.scripts, event.Name)
					logging.LogToDeck(context.Background(), "info", "JS", "info", "removed compiled script "+event.Name)
					jsm.fnRunOnDelete(event.Name)
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
					delete(jsm.scripts, event.Name)
					logging.LogToDeck(context.Background(), "info", "JS", "info", "removed compiled script "+event.Name)
					jsm.fnRunOnDelete(event.Name)
				}
				// A rename fires off a create event also, so it'll handle
				// watcher/compilation in that block
			}
			continue
		case err := <-watcher.Errors:
			logging.LogToDeck(context.Background(), "error", "JS", "error", "error in JS filewatcher: "+err.Error())
			continue
		}
	}
}
