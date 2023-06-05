package taproot

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/highgrav/taproot/v1/common"
	"github.com/highgrav/taproot/v1/logging"
	"os"
	"strings"
)

// Starts monitoring of the JSML file directories to catch any updates and recompile accordingly.
func (srv *AppServer) monitorJSMLDirectories(srcDirName, dstDirName string) {
	dirList := []string{srcDirName}

	// populate with initial subdirectories
	subdirs, err := common.GetDirs(srcDirName)
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "JSML", "error", "JSML\terror\tjs monitoring could not be started: "+err.Error())
		return
	}
	dirList = append(dirList, subdirs...)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "JSML", "error", "error creating JSML watcher: "+err.Error())
		return
	}
	defer watcher.Close()
	for _, v := range dirList {
		err = watcher.Add(v)
		if err != nil {
			logging.LogToDeck(context.Background(), "error", "JSML", "error", "error watching directory "+v+": "+err.Error())
			return
		}
	}
	logging.LogToDeck(context.Background(), "info", "JSML", "info", "watching script file directories from "+srcDirName+" to "+dstDirName)

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				if strings.HasSuffix(event.Name, ".jsml") {
					logging.LogToDeck(context.Background(), "info", "JSML", "info", "recompiling "+event.Name)
					err := srv.compileOne(event.Name, srcDirName, dstDirName)
					if err != nil {
						logging.LogToDeck(context.Background(), "error", "JSML", "error", "JSML error transpiling "+event.Name+": "+err.Error())
					}
				}
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				// Directory or file?
				// Check to see if this is a directory. If so, need to add watcher to directory
				fileInfo, err := os.Stat(event.Name)
				if err != nil {
					logging.LogToDeck(context.Background(), "error", "JSML", "error", "error when reading created file "+event.Name+": "+err.Error())
				} else {
					if fileInfo.IsDir() {
						logging.LogToDeck(context.Background(), "info", "JSML", "info", "watching new directory "+event.Name)
						err = watcher.Add(event.Name)
						if err != nil {
							logging.LogToDeck(context.Background(), "info", "JSML", "error", "error when adding watcher to created dir "+event.Name+": "+err.Error())
						}
					}

					// It's a file, so try compiling it
					if strings.HasSuffix(event.Name, ".jsml") {
						logging.LogToDeck(context.Background(), "info", "JSML", "info", "compiling "+event.Name)
						err := srv.compileOne(event.Name, srcDirName, dstDirName)
						if err != nil {
							logging.LogToDeck(context.Background(), "error", "JSML", "error", "JSML error transpiling "+event.Name+": "+err.Error())
						}
					}
				}
			}

			if event.Op&fsnotify.Remove == fsnotify.Remove {
				// We can't fstat a removed file, so...
				// Try to remove watcher from directory
				watcher.Remove(event.Name)
				// TODO handle directories

				if strings.HasSuffix(event.Name, ".jsml") {
					// Try to delete the filename from destination directory
					fileName, err := common.FindRelocatedFile(dstDirName, event.Name[:len(event.Name)-2])
					if err != nil {
						logging.LogToDeck(context.Background(), "error", "JSML", "error", "could not locate file to delete "+event.Name[:len(event.Name)-2])
						return
					}
					st, err := os.Stat(fileName)
					if err != nil {
						logging.LogToDeck(context.Background(), "error", "JSML", "error", "could not stat file "+event.Name[:len(event.Name)-2])
						return
					}
					if st.IsDir() {
						logging.LogToDeck(context.Background(), "error", "JSML", "error", "tried to recompile directory "+event.Name[:len(event.Name)-2])
						return
					}
					logging.LogToDeck(context.Background(), "info", "JSML", "info", "deleting file "+fileName)
					err = os.Remove(fileName)
					if err != nil {
						logging.LogToDeck(context.Background(), "error", "JSML", "error", "could not delete "+event.Name[:len(event.Name)-2]+": "+err.Error())
						return
					}
				}
			}
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				// We can't fstat a renamed file either, so...
				// Try to remove watcher from directory
				watcher.Remove(event.Name)
				// TODO -- handle directory

				if strings.HasSuffix(event.Name, ".jsml") {
					// Try to delete the filename from destination
					fileName, err := common.FindRelocatedFile(dstDirName, event.Name[:len(event.Name)-2])
					if err != nil {
						logging.LogToDeck(context.Background(), "error", "JSML", "error", "could not locate file to delete for rename "+event.Name[:len(event.Name)-2])
						return
					}
					st, err := os.Stat(fileName)
					if err != nil {
						logging.LogToDeck(context.Background(), "error", "JSML", "error", "could not stat file "+event.Name[:len(event.Name)-2])
						return
					}
					if st.IsDir() {
						logging.LogToDeck(context.Background(), "error", "JSML", "error", "tried to recompile directory "+event.Name[:len(event.Name)-2])
						return
					}
					logging.LogToDeck(context.Background(), "info", "JSML", "info", "deleting file "+fileName)
					err = os.Remove(fileName)
					if err != nil {
						logging.LogToDeck(context.Background(), "error", "JSML", "error", "could not delete "+event.Name[:len(event.Name)-2]+": "+err.Error())
						return
					}
					// A rename fires off a create event also, so it'll handle
					// watcher/compilation in that block
				}
			}
			continue
		case err := <-watcher.Errors:
			logging.LogToDeck(context.Background(), "error", "JSML", "error", "error in JSML filewatcher: "+err.Error())
			continue
		}
	}
}
