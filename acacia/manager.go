package acacia

import "quamina.net/go/quamina"

type PolicyManager struct {
	patterns map[string]*quamina.Quamina
}

func (pm *PolicyManager) FlushAllFor(route string) {

}

func (pm *PolicyManager) LoadAllFor(dirName string, suffix string, route string) error {
	return nil
}

func (pm *PolicyManager) FlushAll() {

}

func (pm *PolicyManager) LoadAll(dirName string) error {
	return nil
}

func (pm *PolicyManager) AddPolicy(policy string) {

}

func New(dir string) (*PolicyManager, error) {
	pm := &PolicyManager{}

	return pm, nil
}
