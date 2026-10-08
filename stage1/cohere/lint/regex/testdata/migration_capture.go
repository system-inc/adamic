package core

import (
	"encoding/json"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"os"
	"sync"
)

type waveRegexCase struct {
	Rule       string   `json:"rule"`
	Input      string   `json:"input"`
	Pattern    string   `json:"pattern"`
	Term       string   `json:"term"`
	Location   string   `json:"location"`
	Decoration []string `json:"decoration"`
	Matched    bool     `json:"matched"`
}

var waveRegexLock sync.Mutex
var waveWarningConfigs = map[*esregexp.RegExp]waveRegexCase{}

func waveRegexRecord(c waveRegexCase) {
	path := os.Getenv("ADAMIC_REGEX_CAPTURE")
	if path == "" {
		return
	}
	waveRegexLock.Lock()
	defer waveRegexLock.Unlock()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(c); err != nil {
		panic(err)
	}
}
func waveRegexTest(kind string, pattern *esregexp.RegExp, input string) bool {
	matched := pattern.TestOrTimeout(input)
	waveRegexRecord(waveRegexCase{Rule: kind, Input: input, Pattern: pattern.Source(), Matched: matched})
	return matched
}
func waveCompileWarning(source, term string, location NoWarningCommentsLocation, decoration []string) (*esregexp.RegExp, error) {
	pattern, err := esregexp.Compile(source, "iu")
	normalized := "anywhere"
	if location == NoWarningCommentsStart {
		normalized = "start"
	}
	c := waveRegexCase{Rule: "no-warning-comments", Pattern: source, Term: term, Location: normalized, Decoration: decoration}
	if err != nil {
		waveRegexRecord(c)
		return nil, err
	}
	waveRegexLock.Lock()
	waveWarningConfigs[pattern] = c
	waveRegexLock.Unlock()
	return pattern, nil
}
func waveWarningTest(pattern *esregexp.RegExp, input string) bool {
	waveRegexLock.Lock()
	c, ok := waveWarningConfigs[pattern]
	waveRegexLock.Unlock()
	if !ok {
		panic("missing captured warning construction")
	}
	matched := pattern.Test(input)
	c.Input = input
	c.Matched = matched
	waveRegexRecord(c)
	return matched
}
