// Package meterdata holds adamic-meter recorded observations without loading a compiler.
package meterdata

type Finding struct {
	Kind     string          `json:"kind"`
	Reason   string          `json:"reason"`
	Count    int             `json:"count"`
	Files    []string        `json:"files"`
	Example  string          `json:"example"`
	FileSeen map[string]bool `json:"-"`
}

type Report struct {
	Root                  string       `json:"root"`
	FilesExamined         int          `json:"files_examined"`
	JavaScriptSkipped     int          `json:"javascript_files_skipped"`
	FilesReachingLowering int          `json:"files_reaching_lowering"`
	Reasons               []*Finding   `json:"reasons"`
	Adaptations           []Adaptation `json:"adaptations,omitempty"`
}

type Adaptation struct {
	Rewrite             string `json:"rewrite"`
	Removed             int    `json:"diagnostics_removed"`
	DeclarationsChanged int    `json:"declarations_changed,omitempty"`
	FunctionsChanged    int    `json:"functions_changed,omitempty"`
}
