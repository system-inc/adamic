package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// backendOutcome is one side of a ruled backend divergence: a run's exact stdout and stderr and its exit code.
type backendOutcome struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// ruledBackendDivergence is one row of the list internal/oracle's backend divergence test pins
// (internal/oracle/testdata/backend_divergences.json, #tqm95v6): a fixture whose backends are ruled to differ,
// from each other, or, where Node is set, both from the source on Node while they agree with each other. The
// tool reads the oracle's own rows, never a list of its own.
type ruledBackendDivergence struct {
	Fixture    string          `json:"fixture"`
	Name       string          `json:"name,omitempty"`
	Ruling     string          `json:"ruling"`
	Reason     string          `json:"reason"`
	Node       *backendOutcome `json:"node,omitempty"`
	Native     backendOutcome  `json:"native"`
	JavaScript backendOutcome  `json:"javascript"`
}
type ruledBackendDivergenceList struct {
	Comment     string                   `json:"_comment"`
	Divergences []ruledBackendDivergence `json:"divergences"`
}

// rulingTask is the task a row's ruling names, #id, anywhere in its words.
var rulingTask = regexp.MustCompile(`#[0-9a-z]+`)

// repositoryPath says a path is a clean one inside the repository, as a list names a program.
func repositoryPath(path string) bool {
	return path != "" && !filepath.IsAbs(path) && filepath.ToSlash(filepath.Clean(path)) == path && path != ".." && !strings.HasPrefix(path, "../")
}

// readRuledBackendDivergences reads the oracle's rows. Each names its fixture once, a ruling naming its task,
// and its reason: the tool refuses the list otherwise, not only review.
func readRuledBackendDivergences(path string) (ruledBackendDivergenceList, error) {
	var list ruledBackendDivergenceList
	data, err := os.ReadFile(path)
	if err != nil {
		return list, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&list); err != nil {
		return list, err
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return list, fmt.Errorf("ruled backend divergence list has trailing JSON")
	}
	if !strings.Contains(list.Comment, "Adding a row is a ruling") || !strings.Contains(list.Comment, "@system_adamic") || list.Divergences == nil {
		return list, fmt.Errorf("ruled backend divergence list needs its ruling header and divergences array")
	}
	fixtures := map[string]bool{}
	for index, row := range list.Divergences {
		if !repositoryPath(row.Fixture) {
			return list, fmt.Errorf("ruled backend divergence row %d needs its fixture's repository path", index+1)
		}
		if fixtures[row.Fixture] {
			return list, fmt.Errorf("ruled backend divergence fixture listed twice: %s", row.Fixture)
		}
		fixtures[row.Fixture] = true
		if strings.TrimSpace(row.Ruling) == "" || !rulingTask.MatchString(row.Ruling) {
			return list, fmt.Errorf("ruled backend divergence needs its ruling, naming its task: %s", row.Fixture)
		}
		if row.Reason == "" {
			return list, fmt.Errorf("ruled backend divergence needs its reason: %s", row.Fixture)
		}
	}
	return list, nil
}

// ran says a run is exactly this outcome: no execution error, and the same exit code, stdout and stderr.
func (outcome backendOutcome) ran(o observation) bool {
	return o.Error == "" && o.Exit == outcome.ExitCode && o.Stdout == outcome.Stdout && o.Stderr == outcome.Stderr
}

// acceptsRuledBackendDivergence says a run that disagrees with Node is a ruled divergence, when all three hold:
// native and JavaScript agree with each other, as outputsAgree reads agreement (no execution error, the same exit
// code and stdout); the program's path is a row's fixture, on a row that rules Node's side; and Node's run is
// exactly that row's Node outcome, and native's its native outcome. The row's JavaScript outcome is carried but not
// compared here: JavaScript is held to native's exit and stdout, as outputsAgree holds it, so the javascript column is
// the oracle test's to check, never this one's. A program on no row is never one, whatever it prints.
func acceptsRuledBackendDivergence(path string, list ruledBackendDivergenceList, node, javascript, native observation) bool {
	if javascript.Error != "" || native.Error != "" || javascript.Exit != native.Exit || javascript.Stdout != native.Stdout {
		return false
	}
	for _, row := range list.Divergences {
		if row.Fixture == path && row.Node != nil {
			return row.Node.ran(node) && row.Native.ran(native)
		}
	}
	return false
}
