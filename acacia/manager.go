package acacia

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/highgrav/taproot/v1/common"
	"github.com/highgrav/taproot/v1/logging"
	"os"
	"path/filepath"
	"quamina.net/go/quamina"
	"sort"
	"strings"
)

type PolicyManager struct {
	patterns map[string]*quamina.Quamina
	policies map[string]Policy
}

func NewPolicyManager() *PolicyManager {
	pm := &PolicyManager{
		patterns: make(map[string]*quamina.Quamina),
		policies: make(map[string]Policy),
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
	logging.LogToDeck(context.Background(), "info", "ACAC", "info", "loading policy files from "+dirName)
	s, err := os.Stat(dirName)
	if err != nil {
		return err
	}
	if !s.IsDir() {
		return errors.New("Not a directory:  (" + dirName + ")")
	}
	err = filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".acacia") {
			// compile Acacia file
			input, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			p, _ := NewParser(string(input))
			policy, err := p.Parse()
			if err != nil {
				return err
			}
			logging.LogToDeck(context.Background(), "info", "ACAC", "info", "loading policy file "+info.Name())
			err = pm.AddPolicy(path, policy)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (pm *PolicyManager) AddPolicy(name string, policy Policy) error {
	for _, route := range policy.Routes {
		if _, ok := pm.patterns[route]; !ok {
			q, err := quamina.New(quamina.WithMediaType("application/json"), quamina.WithPatternDeletion(true))
			if err != nil {
				return err
			}
			pm.patterns[route] = q
		}
		logging.LogToDeck(context.Background(), "info", "ACAC", "info", "adding policy to route "+route)
		pm.policies[name] = policy
		pm.patterns[route].AddPattern(name, policy.Match)
	}
	return nil
}

func (pm *PolicyManager) Apply(ctx context.Context, route string, request *RightsRequest) (RightResponse, error) {
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
		logging.LogToDeck(ctx, "error", "ACAC", "error", "attempted to call Acacia on unbound route "+route)
		return rr, nil
	}
	js, err := json.Marshal(*request)
	if err != nil {
		return rr, err
	}

	resps, err := q.MatchesForEvent(js)
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

	for _, resp_id := range resps {
		resp, ok := pm.policies[resp_id.(string)]
		if !ok {
			return rr, errors.New("could not access policy ID " + resp_id.(string))
		}
		pri := resp.Manifest.Priority
		pris = append(pris, pri)

		// grab any return code (note that if there's a tie in priority, last-in wins)
		if resp.Rights.ReturnCode > 0 {
			if pri > topRespPri {
				topRespPri = pri
			}
			responses[pri] = RightCodeResponse{
				ReturnMsg:  resp.Rights.ReturnMsg,
				ReturnCode: resp.Rights.ReturnCode,
			}
		}

		// grab any redirection (note that if there's a tie in priority, last-in wins)
		if resp.Rights.Redirect != "" {
			if pri > topRedirPri {
				topRedirPri = pri
			}
			redirects[pri] = resp.Rights.Redirect
		}

		// get approvals/denials
		_, ok = approvals[pri]
		if !ok {
			approvals[pri] = make([]string, 0)
		}
		_, ok = denials[pri]
		if !ok {
			denials[pri] = make([]string, 0)
		}
		if len(resp.Rights.Allowed) > 0 && pri > topApprovalPri {
			topApprovalPri = pri
		}
		approvals[pri] = append(approvals[pri], resp.Rights.Allowed...)
		denials[pri] = append(denials[pri], resp.Rights.Denied...)
	}

	// if response pri >= redirect pri && response pri > approval pri
	if topRespPri >= topRedirPri && topRespPri > topApprovalPri && len(responses) > 0 {
		rr.Response = responses[topRespPri]
		rr.Type = RESP_TYPE_RESPONSE
		return rr, nil
	}

	// if redirect pri >= approval pri
	if topRedirPri >= topApprovalPri && len(redirects) > 0 {
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
