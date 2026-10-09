package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"

	"github.com/system-inc/cohere/internal/lint/configuration"
)

// outputBlock is the settings' "output" block: how a run prints, for this repository. It is the one output
// setting in 1.0, and the flags it stands in for still win when named.
type outputBlock struct {
	// Phases is `--phases` made the repository's default.
	Phases *bool `json:"phases"`
}

// readOutputBlock reads the "output" block from the settings file cohere reads first, the project's own. A
// file extending another may not carry one (the loader refuses it there, since this reads only here), and
// a key the block does not know is refused, naming the file, rather than ignored.
func readOutputBlock(settingsPath string) (outputBlock, error) {
	contents, err := configuration.SourceContents(settingsPath)
	if errors.Is(err, fs.ErrNotExist) {
		return outputBlock{}, nil
	}
	if err != nil {
		return outputBlock{}, nil // The lint config's own read reports a file it cannot read.
	}
	var keyed map[string]json.RawMessage
	if json.Unmarshal(contents, &keyed) != nil {
		return outputBlock{}, nil // And a file that does not parse.
	}
	raw, present := keyed["output"]
	if !present {
		return outputBlock{}, nil
	}
	var block outputBlock
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&block); err != nil {
		return outputBlock{}, fmt.Errorf("the \"output\" block in %s: %w; the only key it takes is \"phases\"", settingsPath, err)
	}
	return block, nil
}
