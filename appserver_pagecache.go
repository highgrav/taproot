package taproot

import "strings"

func (svr *AppServer) removePageCacheEntry(id string) {
	scriptId := strings.TrimPrefix(id, svr.Config.ScriptFilePath)

	// Remove leading '/' from the pathname
	svr.PageCache.Flush(scriptId[1:])
}
