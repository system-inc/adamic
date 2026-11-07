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

func newRunProfile() *runProfile            { return &runProfile{origin: time.Now()} }
func (profile *runProfile) offset() float64 { return time.Since(profile.origin).Seconds() }
func (profile *runProfile) preparing(stage string) func() {
	if profile == nil {
		return func() {}
	}
	start := profile.offset()
	return func() {
		profile.Prepare = append(profile.Prepare, phaseProfile{Stage: stage, Start: start, End: profile.offset()})
	}
}
func (profile *runProfile) begin(path string, worker int) *testProfile {
	if profile == nil {
		return nil
	}
	return &testProfile{Path: path, Worker: worker, Start: profile.offset(), owner: profile}
}
func (profile *testProfile) finish() {
	if profile == nil {
		return
	}
	profile.End = profile.owner.offset()
	profile.owner.mutex.Lock()
	profile.owner.Tests = append(profile.owner.Tests, profile)
	profile.owner.mutex.Unlock()
}
func (profile *testProfile) command(stage string, execute func() execution) execution {
	if profile == nil {
		return execute()
	}
	start := profile.owner.offset()
	result := execute()
	profile.Phases = append(profile.Phases, phaseProfile{Stage: stage, Start: start, End: profile.owner.offset()})
	return result
}
func (profile *testProfile) observation(
	stage string,
	cache *resultCache,
	key string,
	execute func() (execution, bool),
) execution {
	if profile == nil {
		return cache.observe(key, execute)
	}
	start := profile.owner.offset()
	hit := true
	result := cache.observe(key, func() (execution, bool) { hit = false; return execute() })
	profile.Phases = append(
		profile.Phases,
		phaseProfile{Stage: stage, Start: start, End: profile.owner.offset(), Hit: hit},
	)
	return result
}
func (profile *runProfile) write(path string) error {
	sort.Slice(profile.Tests, func(i, j int) bool { return profile.Tests[i].Path < profile.Tests[j].Path })
	data, encodeError := json.MarshalIndent(profile, "", "  ")
	if encodeError != nil {
		return encodeError
	}
	return os.WriteFile(path, append(data, '\n'), 0600)
}
