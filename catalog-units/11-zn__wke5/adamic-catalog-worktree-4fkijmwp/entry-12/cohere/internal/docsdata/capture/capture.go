// Package capture is the record the lint test harness writes for the docs generator: one case a rule's
// test asserted, with the source it ran on and what the rule did there.
//
// It is a leaf, importing nothing of cohere's, because both ends need it and they cannot import each
// other: the harness (internal/lint/testing) is imported by every rule package's tests, and the
// generator (internal/docsdata) imports every rule package through the registry.
//
// Nothing is written unless COHERE_DOCS_CAPTURE names a directory, so an ordinary test run pays one
// string comparison per assertion and touches no file.
package capture

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Variable is the environment variable naming the directory records are written to.
const Variable = "COHERE_DOCS_CAPTURE"

// The outcomes a record can carry: the assertion that passed.
const (
	OutcomeFindings = "Findings"
	OutcomeClean    = "Clean"
	OutcomeFixed    = "Fixed"
)

// Record is one asserted case.
type Record struct {
	Rule string `json:"rule"`
	// File is the subject file's name as the test gave it, and Source its text.
	File   string `json:"file"`
	Source string `json:"source"`
	// OtherFiles counts the files the program held beside the subject, which the record does not carry.
	OtherFiles int `json:"otherFiles,omitempty"`
	// Options is the options value the test handed the rule, as encoding/json renders it, and absent when
	// it handed none.
	Options json.RawMessage `json:"options,omitempty"`
	Outcome string          `json:"outcome"`
	// Findings is what the rule reported, in order. FixedSource is the text the test asserted the fixes
	// produce, present only for OutcomeFixed.
	Findings    []Finding `json:"findings,omitempty"`
	FixedSource string    `json:"fixedSource,omitempty"`
}

// Finding is one diagnostic, positioned as cohere prints it: one-based line and column.
type Finding struct {
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	EndLine     int    `json:"endLine"`
	EndColumn   int    `json:"endColumn"`
	MessageId   string `json:"messageId"`
	Message     string `json:"message"`
	Fix         bool   `json:"fix,omitempty"`
	Suggestions int    `json:"suggestions,omitempty"`
}

// Directory is where records go, empty when capture is off. Read once per process.
var Directory = sync.OnceValue(func() string { return os.Getenv(Variable) })

var (
	writeMutex sync.Mutex
	output     *os.File
)

// Write appends a record to this process's file in Directory, one JSON object per line. A test binary
// is one process and packages run as separate binaries, so a file per process is never shared.
func Write(record Record) error {
	directory := Directory()
	if directory == "" {
		return nil
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		return err
	}
	writeMutex.Lock()
	defer writeMutex.Unlock()
	if output == nil {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return err
		}
		file, err := os.OpenFile(filepath.Join(directory, fmt.Sprintf("capture-%d.jsonl", os.Getpid())), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		output = file
	}
	_, err = output.Write(append(encoded, '\n'))
	return err
}
