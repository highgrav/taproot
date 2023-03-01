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
	Match    any            `json:"match"`
}

type PolicyManifest struct {
	Namespace   string `json:"ns"`
	Version     string `json:"v"`
	Name        string `json:"name"`
	Description string `json:"desc"`
}

type PolicyRights struct {
	Allowed []string `json:"allowed"`
	Denied  []string `json:"denied"`
}

type PolicyLogging struct {
	OnPermit []string `json:"permit"`
	OnDeny   []string `json:"deny"`
	OnAny    []string `json:"any"`
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
