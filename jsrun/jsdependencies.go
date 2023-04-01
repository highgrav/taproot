package jsrun

type JSDependencies struct {
	importingScripts map[string][]string // Maps imported scripts to array of importing scripts
}

func NewJSDependencies() JSDependencies {
	return JSDependencies{
		importingScripts: make(map[string][]string),
	}
}

func (jsd *JSDependencies) AddDependency(importedScript, importingScript string) {
	if _, ok := jsd.importingScripts[importedScript]; !ok {
		jsd.importingScripts[importedScript] = make([]string, 0)
	}
	jsd.importingScripts[importedScript] = append(jsd.importingScripts[importedScript], importingScript)
}

func (jsd *JSDependencies) GetDependents(script string) []string {
	if _, ok := jsd.importingScripts[script]; !ok {
		return []string{}
	}
	return jsd.importingScripts[script]
}
