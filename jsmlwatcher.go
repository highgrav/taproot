package taproot

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/google/deck"
	"github.com/highgrav/taproot/v1/common"
	"github.com/highgrav/taproot/v1/logging"
	"os"
	"strings"
)

// Starts monitoring of the JSML file directories to catch any updates and recompile accordingly.
// TODO -- this is having issues detecting changes
func (srv *AppServer) monitorJSMLDirectories(srcDirName, dstDirName string) {
	logging.LogToDeck("info", "JSML\tinfo\tmonitoring directory from "+srcDirName)
	dirList := []string{srcDirName}
	subdirs, err := common.GetDirs(srcDirName)
	if err != nil {
		deck.Error("JSML\terror\tjsml file monitoring cannot be started: " + err.Error())
		return
	}
	dirList = append(dirList, subdirs...)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		deck.Error("JSML\terror\tjsml file monitoring cannot be started: " + err.Error())
		return
	}
	defer watcher.Close()
	for _, d := range dirList {
		err := watcher.Add(d)
		if err != nil {
			deck.Error("JSML\terror\terror watching directory " + d + ": " + err.Error())
			return
		}
		logging.LogToDeck("info", "JSML\tinfo\twatching directory "+d)
	}
	for {
		select {
		case exitFlag := <-srv.ExitServerCh:
			if exitFlag {
				deck.Info("JSML\tinfo\tshutting down jsml filewatcher")
				return
			}
		case err := <-watcher.Errors:
			deck.Error("JSML\terror\terror in jsml filewatcher")
			deck.Error(err.Error())
			return
		case event := <-watcher.Events:
			fmt.Println("CAUGHT EVENT " + event.Name)
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Println("CAUGHT WRITE " + event.Name)
				if strings.HasSuffix(event.Name, ".jsml") {
					fmt.Println("CAUGHT WRITE " + event.Name)
					err := srv.compileOne(event.Name, srcDirName, dstDirName)
					if err != nil {
						deck.Error("JSML\terror\tjsml error transpiling " + event.Name + ": " + err.Error())
					}
				}
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				// TODO
				s, err := os.Stat(event.Name)
				if err != nil {
					deck.Error("JSML\terror\tjsml error handling created file " + event.Name + ": " + err.Error())
				} else {
					if s.IsDir() {
						// if a directory, add a watcher
						deck.Info("JSML\tinfo\tCaught create event on JSML directory " + event.Name)
						watcher.Add(event.Name)
					} else {
						// if a file, compile it
						if strings.HasSuffix(event.Name, ".jsml") {
							deck.Info("JSML\tinfo\tCaught create event on JSML file " + event.Name)
							err := srv.compileOne(event.Name, srcDirName, dstDirName)
							if err != nil {
								deck.Error("JSML\terror\tjsml error transpiling " + event.Name + ": " + err.Error())
							}
						}
					}
				}
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				deck.Info("JSML\tinfo\tCaught remove event on JSML file " + event.Name)
				watcher.Remove(event.Name)
				// delete the compiled JS file
				// TODO -- find relocated file and delete it
				fileName, err := common.FindRelocatedFile(dstDirName, event.Name[:len(event.Name)-2])
				if err != nil {
					deck.Error("JSML\terror]tjsml error dealing with deleted file " + event.Name + ": " + err.Error())
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
					deck.Error("JSML\terror\tjsml error dealing with deleted file " + event.Name + ": " + err.Error())
				} else {
					// TODO
					deck.Error("TODO: Delete " + fileName)
				}
			}
		}
		return
	}
}
