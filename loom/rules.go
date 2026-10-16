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
	// Sizing: each block's size from the red rate on the queue's log, the budget standing only without it. Optional
	// (Loom 0.8.19, #hr7v2wc).
	Sizing *SizingRule `json:"sizing"`
	// Depth: the most blocks in flight, each pipelined on the last prefix of the one below it, from DefaultBlocksDepth to
	// MaximumBlocksDepth. Optional, and never null; absent reads as 2 (Loom 0.8.71, #wpcbqpd).
	Depth *int `json:"depth"`
}

// DefaultBlocksDepth and MaximumBlocksDepth are the queue's: one standing block and one on it, and the deepest tower
// (Loom's lander/rules.go).
const (
	DefaultBlocksDepth = 2
	MaximumBlocksDepth = 32
)

// SizingRule is a block-size policy: the largest block whose chance of coming back all green, at the red rate of the last
// Window decided branches, is at least Target, clamped to Min and Max.
type SizingRule struct {
	Min    *int     `json:"min"`
	Max    *int     `json:"max"`
	Target *float64 `json:"target"`
	Window *int     `json:"window"`
}

// MaximumSizingWindow is the queue's: the most decided branches a window reads (Loom's lander/rules.go).
const MaximumSizingWindow = 200

type RestackRule struct {
	On *bool `json:"on"`
}

// ReadRules is rules.json as landed with this package.
func ReadRules() (Rules, error) {
	return ParseRules(rulesJSON)
}

// ParseRules reads a rules file strictly: a rule the queue doesn't know, a field a rule doesn't take, a missing field,
// a budget that isn't a whole number of 1 or more, a sizing or a depth the lander would refuse, or anything after the object is refused, so a typo can't land as a
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
	if sizing := rules.Blocks.Sizing; sizing != nil {
		if sizing.Min == nil || sizing.Max == nil || sizing.Target == nil || sizing.Window == nil {
			return Rules{}, errors.New("rules.json: blocks' sizing needs min, max, target and window")
		}
		if *sizing.Min < 1 || *sizing.Max < *sizing.Min || !(*sizing.Target > 0 && *sizing.Target < 1) || *sizing.Window < 1 || *sizing.Window > MaximumSizingWindow {
			return Rules{}, fmt.Errorf("rules.json: blocks' sizing needs 1 <= min <= max, target above 0 and below 1, and window from 1 to %d", MaximumSizingWindow)
		}
	}
	// Present or absent, never null, as the lander reads it: a null decodes as absent, so it's read from the raw object.
	var raw struct {
		Blocks map[string]json.RawMessage `json:"blocks"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return Rules{}, fmt.Errorf("rules.json: %w", err)
	}
	if depth, present := raw.Blocks["depth"]; present {
		if bytes.Equal(bytes.TrimSpace(depth), []byte("null")) || rules.Blocks.Depth == nil {
			return Rules{}, fmt.Errorf("rules.json: blocks' depth is %s, and it's a whole number from %d to %d", depth, DefaultBlocksDepth, MaximumBlocksDepth)
		}
		if *rules.Blocks.Depth < DefaultBlocksDepth || *rules.Blocks.Depth > MaximumBlocksDepth {
			return Rules{}, fmt.Errorf("rules.json: blocks' depth is %d, and it's from %d to %d", *rules.Blocks.Depth, DefaultBlocksDepth, MaximumBlocksDepth)
		}
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
