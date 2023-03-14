package acacia

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Policy struct {
	Manifest PolicyManifest `json:"manifest"`
	Routes   []string       `json:"paths"`
	Rights   PolicyRights   `json:"rights"`
	Logging  PolicyLogging  `json:"log"`
	Match    string         `json:"match"`
}

type PolicyManifest struct {
	ID          string `json:"id"`
	Priority    int    `json:"pri"`
	Namespace   string `json:"ns"`
	Version     string `json:"version"`
	Name        string `json:"name"`
	Description string `json:"desc"`
}

type PolicyRights struct {
	Allowed    []string `json:"allowed"`
	Denied     []string `json:"denied"`
	Redirect   string   `json:"redirect"`
	ReturnCode int      `json:"returnCode"`
	ReturnMsg  string   `json:"returnMsg"`
}

type PolicyLogging struct {
	OnPermit []PolicyLog `json:"permit"`
	OnDeny   []PolicyLog `json:"deny"`
	OnAny    []PolicyLog `json:"any"`
}

type PolicyLog struct {
	Source   string `json:"src"`
	Priority string `json:"pri"`
	Message  string `json:"msg"`
}

// Walk a directory and attempt to generate policies for every file.
func readAllPolicies(dirName string, suffix string) ([]Policy, error) {
	var policies []Policy
	err := filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, suffix) {
			policy, err := readPolicyFile(path)
			if err != nil {
				return errors.Join(err, errors.New("Failed to parse policy "+path))
			}
			policies = append(policies, policy)
		}
		return nil
	})
	if err != nil {
		return policies, err
	}
	return policies, nil
}

func readPolicyFile(fileName string) (Policy, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return Policy{}, err
	}
	policy := &Policy{}
	err = json.NewDecoder(file).Decode(&policy)
	if err != nil {
		return *policy, err
	}

	return *policy, nil
}
