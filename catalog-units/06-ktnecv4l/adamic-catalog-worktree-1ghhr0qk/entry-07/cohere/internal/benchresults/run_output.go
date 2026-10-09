package benchresults

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// RunOutput is what one measured run said about itself, read from its `--json` summary line, the contract
// cohere versions (command/cohere/json_output.go, schema/CohereOutput.schema.json). The human footer is
// free to change, and when it moved to one line every reader of it read nothing, so the benchmark reads
// the contract instead.
type RunOutput struct {
	// EngineSeconds is the engine's own wall clock for the run.
	EngineSeconds float64
	// Findings is every finding the run reported, whichever phase found it: lint findings, type errors,
	// and files the formatter would change.
	Findings int
	// Cache is what the run took from the cache: "replay" for a run answered whole, "files N/M" for N of
	// the M files in scope taken from the cache, "off" under --no-cache, and "none" when it checked every
	// file. N is the summary's `cached`, the count the footer prints, which splits the files in scope with
	// `checked`; `cache.filesReplayed` is narrower, and read 0 on a run that took every file from the cache.
	Cache string
}

// outputSchemaVersion is the `--json` contract version this reader knows.
const outputSchemaVersion = 1

// ParseRunOutput reads a run's `--json` output: the last line whose kind is summary. Output with no
// summary line, or one of another contract version, is refused, so an instrument that cannot read a run
// says so rather than recording zeroes.
func ParseRunOutput(output []byte) (RunOutput, error) {
	var summary struct {
		Kind          string  `json:"kind"`
		SchemaVersion int     `json:"schemaVersion"`
		Seconds       float64 `json:"seconds"`
		Cache         struct {
			Replayed bool `json:"replayed"`
			Off      bool `json:"off"`
		} `json:"cache"`
		TypeErrors   int `json:"typeErrors"`
		Findings     int `json:"findings"`
		WouldChange  int `json:"wouldChange"`
		FilesInScope int `json:"filesInScope"`
		Cached       int `json:"cached"`
	}
	found := false
	for _, line := range bytes.Split(output, []byte("\n")) {
		if !bytes.HasPrefix(bytes.TrimSpace(line), []byte("{")) {
			continue
		}
		var kind struct {
			Kind string `json:"kind"`
		}
		if json.Unmarshal(line, &kind) != nil || kind.Kind != "summary" {
			continue
		}
		if err := json.Unmarshal(line, &summary); err != nil {
			return RunOutput{}, fmt.Errorf("reading the summary line: %w", err)
		}
		found = true
	}
	if !found {
		return RunOutput{}, fmt.Errorf("no --json summary line in the run's output")
	}
	if summary.SchemaVersion != outputSchemaVersion {
		return RunOutput{}, fmt.Errorf("the summary is --json schema %d, and this reader knows %d", summary.SchemaVersion, outputSchemaVersion)
	}
	result := RunOutput{
		EngineSeconds: summary.Seconds,
		Findings:      summary.Findings + summary.TypeErrors + summary.WouldChange,
		Cache:         "none",
	}
	switch {
	case summary.Cache.Off:
		result.Cache = "off"
	case summary.Cache.Replayed:
		result.Cache = "replay"
	case summary.Cached > 0:
		result.Cache = fmt.Sprintf("files %d/%d", summary.Cached, summary.FilesInScope)
	}
	return result, nil
}
