package acacia

import (
	"encoding/json"
	"highgrav/taproot/v1/common"
	"quamina.net/go/quamina"
	"sort"
	"strings"
)

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

func (pm *PolicyManager) Apply(route string, request *RightsRequest) (RightResponse, error) {
	rr := RightResponse{
		Response: RightCodeResponse{
			ReturnMsg:  "",
			ReturnCode: 0,
		},
		Redirect: "",
		Rights:   make([]string, 0),
		Metadata: make(map[string]string),
	}

	// route not registered
	q, ok := pm.patterns[route]
	if !ok {
		return rr, nil
	}
	js, err := json.Marshal(*request)
	if err != nil {
		return rr, err
	}

	resps, err := q.MatchesForEvent([]byte(js))
	if err != nil {
		return rr, err
	}

	responses := make(map[int]RightCodeResponse)
	redirects := make(map[int]string)
	approvals := make(map[int][]string)
	denials := make(map[int][]string)
	pris := make([]int, 0)
	topRespPri := -999999
	topRedirPri := -999999
	topApprovalPri := -999999

	for _, resp := range resps {
		pri := resp.(Policy).Manifest.Priority
		pris = append(pris, pri)

		// grab any return code (note that if there's a tie in priority, last-in wins)
		if resp.(Policy).Rights.ReturnCode > 0 {
			if pri > topRespPri {
				topRedirPri = pri
			}
			responses[pri] = RightCodeResponse{
				ReturnMsg:  resp.(Policy).Rights.ReturnMsg,
				ReturnCode: resp.(Policy).Rights.ReturnCode,
			}
		}

		// grab any redirection (note that if there's a tie in priority, last-in wins)
		if resp.(Policy).Rights.Redirect != "" {
			if pri > topRedirPri {
				topRedirPri = pri
			}
			redirects[pri] = resp.(Policy).Rights.Redirect
		}

		// get approvals/denials
		_, ok := approvals[pri]
		if !ok {
			approvals[pri] = make([]string, 0)
		}
		_, ok = denials[pri]
		if !ok {
			denials[pri] = make([]string, 0)
		}
		if len(resp.(Policy).Rights.Allowed) > 0 && pri > topApprovalPri {
			topApprovalPri = pri
		}
		approvals[pri] = append(approvals[pri], resp.(Policy).Rights.Allowed...)
		denials[pri] = append(denials[pri], resp.(Policy).Rights.Denied...)
	}

	// if response pri >= redirect pri && reponse pri > approval pri
	if topRespPri >= topRedirPri && topRespPri > topApprovalPri {
		rr.Response = responses[topRespPri]
		rr.Type = RESP_TYPE_RESPONSE
		return rr, nil
	}

	// if redirect pri > approval pri
	if topRedirPri > topApprovalPri {
		rr.Redirect = redirects[topRedirPri]
		rr.Type = RESP_TYPE_REDIRECT
		return rr, nil
	}

	// Default: determine approval rights. Note that we normalize rights to lower case
	sort.Ints(pris)
	approved := make([]string, 0)
	// apply by priority
	for _, i := range pris {
		if denials[i] != nil && len(denials[i]) > 0 {
			apps := make([]string, 0)

			// remove any denials
			for _, den := range denials[i] {
				// remove from approved
				for _, app := range approved {
					if strings.ToLower(app) != strings.ToLower(den) {
						apps = append(apps, strings.ToLower(app))
					}
				}
			}
			approved = apps
		}
		if approvals[i] != nil && len(approvals[i]) > 0 {
			for _, app := range approvals[i] {
				approved = append(approved, strings.ToLower(app))
			}
		}
	}

	approved = common.Dedupe[string](approved)
	rr.Type = RESP_TYPE_RIGHTS
	rr.Rights = approved
	return rr, nil
}

func New(dir string) (*PolicyManager, error) {
	pm := &PolicyManager{}
	return pm, nil
}
