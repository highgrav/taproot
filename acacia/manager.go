package acacia

import "quamina.net/go/quamina"

type PolicyManager struct {
	patterns map[string]*quamina.Quamina
}

func NewPolicyManager() *PolicyManager {
	pm := &PolicyManager{
		patterns: make(map[string]*quamina.Quamina),
	}
	return pm
}

func (pm *PolicyManager) FlushAllFor(route string) {
	// TODO
}

func (pm *PolicyManager) FlushAll() {
	// TODO
}

func (pm *PolicyManager) LoadAllFrom(dirName string) error {
	// TODO
	return nil
}

func (pm *PolicyManager) AddPolicy(policy Policy) error {
	for _, route := range policy.Routes {
		if _, ok := pm.patterns[route]; !ok {
			q, err := quamina.New(quamina.WithMediaType("application/json"), quamina.WithPatternDeletion(true))
			if err != nil {
				return err
			}
			pm.patterns[route] = q
		}
		pm.patterns[route].AddPattern(policy, policy.Match)
	}
	return nil
}

func (pm *PolicyManager) Apply(request UserRightRequest) (RightResponse, error) {
	rr := RightResponse{}

	return rr, nil
}

func New(dir string) (*PolicyManager, error) {
	pm := &PolicyManager{}

	return pm, nil
}
