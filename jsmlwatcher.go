package taproot

import (
	"github.com/fsnotify/fsnotify"
	"github.com/google/deck"
	"github.com/highgrav/taproot/v1/common"
	"os"
	"strings"
)

// Starts monitoring of the JSML file directories to catch any updates and recompile accordingly.
// TODO -- this is having issues detecting changes
func (srv *AppServer) monitorJSMLDirectories(srcDirName, dstDirName string) {
	dirList := []string{srcDirName}

	// populate with initial subdirectories
	subdirs, err := common.GetDirs(srcDirName)
	if err != nil {
		deck.Error("JSML\terror\tjs monitoring could not be started: " + err.Error())
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
			deck.Error("JSML\terror\terror watching directory " + v + ": " + err.Error())
			return
		}
	}
	deck.Info("Watching script file directories")

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				if strings.HasSuffix(event.Name, ".jsml") {
					deck.Info("Recompiling " + event.Name)
					err := srv.compileOne(event.Name, srcDirName, dstDirName)
					if err != nil {
						deck.Error("JSML\terror\tjsml error transpiling " + event.Name + ": " + err.Error())
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
					if strings.HasSuffix(event.Name, ".jsml") {
						deck.Info("Compiling " + event.Name)
						err := srv.compileOne(event.Name, srcDirName, dstDirName)
						if err != nil {
							deck.Error("JSML\terror\tjsml error transpiling " + event.Name + ": " + err.Error())
						}
					}
				}
			}

			if event.Op&fsnotify.Remove == fsnotify.Remove {
				// We can't fstat a removed file, so...
				// Try to remove watcher from directory
				watcher.Remove(event.Name)
				// Try to delete the filename from destination directory

			}
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				// We can't fstat a renamed file either, so...
				// Try to remove watcher from directory
				watcher.Remove(event.Name)
				// Try to delete the filename from destination

				// A rename fires off a create event also, so it'll handle
				// watcher/compilation in that block
			}
			continue
		case err := <-watcher.Errors:
			deck.Error("Error in JSML filewatcher: " + err.Error())
			continue
		}
	}
}
