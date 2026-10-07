package main

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"
)

// Profiling is opt in and written separately from verdicts and ordered progress.
// Offsets share a monotonic origin, so overlapping workers can be inspected directly.
type runProfile struct {
	origin  time.Time
	mutex   sync.Mutex
	Tests   []*testProfile `json:"tests"`
	Prepare []phaseProfile `json:"prepare"`
}
type testProfile struct {
	Path   string         `json:"path"`
	Worker int            `json:"worker"`
	Start  float64        `json:"start"`
	End    float64        `json:"end"`
	Phases []phaseProfile `json:"phases"`
	owner  *runProfile
}
type phaseProfile struct {
	Stage string  `json:"stage"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Hit   bool    `json:"hit,omitempty"`
}

func newRunProfile() *runProfile      { return &runProfile{origin: time.Now()} }
func (p *runProfile) offset() float64 { return time.Since(p.origin).Seconds() }
func (p *runProfile) preparing(stage string) func() {
	if p == nil {
		return func() {}
	}
	start := p.offset()
	return func() { p.Prepare = append(p.Prepare, phaseProfile{Stage: stage, Start: start, End: p.offset()}) }
}
func (p *runProfile) begin(path string, worker int) *testProfile {
	if p == nil {
		return nil
	}
	return &testProfile{Path: path, Worker: worker, Start: p.offset(), owner: p}
}
func (p *testProfile) finish() {
	if p == nil {
		return
	}
	p.End = p.owner.offset()
	p.owner.mutex.Lock()
	p.owner.Tests = append(p.owner.Tests, p)
	p.owner.mutex.Unlock()
}
func (p *testProfile) command(stage string, execute func() execution) execution {
	if p == nil {
		return execute()
	}
	start := p.owner.offset()
	result := execute()
	p.Phases = append(p.Phases, phaseProfile{Stage: stage, Start: start, End: p.owner.offset()})
	return result
}
func (p *testProfile) observation(stage string, cache *resultCache, key string, execute func() (execution, bool)) execution {
	if p == nil {
		return cache.observe(key, execute)
	}
	start := p.owner.offset()
	hit := true
	result := cache.observe(key, func() (execution, bool) { hit = false; return execute() })
	p.Phases = append(p.Phases, phaseProfile{Stage: stage, Start: start, End: p.owner.offset(), Hit: hit})
	return result
}
func (p *runProfile) write(path string) error {
	sort.Slice(p.Tests, func(i, j int) bool { return p.Tests[i].Path < p.Tests[j].Path })
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0600)
}
