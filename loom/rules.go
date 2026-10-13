// Package loom holds the rules Loom's queue runs main's line by. A rule changes only by landing a change to rules.json
// here: the queue's rule.changed event names the landed commit that changed it and carries the file's value exactly,
// so the log never says a rule moved without the code that moved it (Loom's contract v1, section 4).
package loom

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
)

//go:embed rules.json
var rulesJSON []byte

// Rules is rules.json: every rule the file sets, each in the shape the queue's POST /rules takes for it.
type Rules struct {
	// Blocks: with them on, the queue opens a block of every checked change waiting, in line order up to the budget,
	// the instant no block is in flight, and the longest green prefix lands.
	Blocks *BlocksRule `json:"blocks"`
	// Restack: with it on, a change whose base main moved past is merged onto main's tip and tested there, not parked.
	Restack *RestackRule `json:"restack"`
	// First: changes the queue puts at the head of the next block to open, ahead of the line, as it does a change to these
	// rules. Optional; each a change id, once (Loom 0.8.15).
	First []string `json:"first"`
}

type BlocksRule struct {
	On     *bool `json:"on"`
	Budget *int  `json:"budget"`
}

type RestackRule struct {
	On *bool `json:"on"`
}

// ReadRules is rules.json as landed with this package.
func ReadRules() (Rules, error) {
	return ParseRules(rulesJSON)
}

// ParseRules reads a rules file strictly: a rule the queue doesn't know, a field a rule doesn't take, a missing field,
// a budget that isn't a whole number of 1 or more, or anything after the object is refused, so a typo can't land as a
// rule nobody meant.
func ParseRules(data []byte) (Rules, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var rules Rules
	if err := decoder.Decode(&rules); err != nil {
		return Rules{}, fmt.Errorf("rules.json: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return Rules{}, errors.New("rules.json: something after the rules object")
	}
	if rules.Blocks == nil || rules.Blocks.On == nil || rules.Blocks.Budget == nil {
		return Rules{}, errors.New("rules.json: blocks needs on and budget")
	}
	if *rules.Blocks.Budget < 1 {
		return Rules{}, fmt.Errorf("rules.json: blocks' budget is %d, and a block holds at least one change", *rules.Blocks.Budget)
	}
	if rules.Restack == nil || rules.Restack.On == nil {
		return Rules{}, errors.New("rules.json: restack needs on")
	}
	seen := map[string]bool{}
	for _, change := range rules.First {
		if !changeIDPattern.MatchString(change) || seen[change] {
			return Rules{}, fmt.Errorf("rules.json: first names change ids (chg_ and 26 base32 characters), each once, not %q", change)
		}
		seen[change] = true
	}
	return rules, nil
}

// changeIDPattern is a change id as the queue mints it, the lander's own test (Loom's lander/rules.go).
var changeIDPattern = regexp.MustCompile(`^chg_[0-9a-hjkmnp-tv-z]{26}$`)
