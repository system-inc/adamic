package lower

// Throwaway probe for #ztpcwrr's slice 1, never to land. With ADAMIC_VIEW_PAIRS_LOG set to a file,
// each lowering appends one JSON line per view pair the finder would have to relate: each
// viewSources entry's source and target, and, recursively, each same-named field the view checks
// (its target field's type against its source field's). A pair records whether deriveRelated
// already relates it, so the count of unrelated pairs says how many links slice 1 would add.

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

var viewPairsLock sync.Mutex

type viewPairRecord struct {
	Program string `json:"program"`
	Depth   int    `json:"depth"`
	Source  string `json:"source"`
	Target  string `json:"target"`
	Field   string `json:"field,omitempty"`
	Related bool   `json:"related"`
	Cast    bool   `json:"cast"`
	Narrows bool   `json:"narrows"`
	Skipped string `json:"skipped,omitempty"`
}

func (f *cycleFinder) logViewPairs() {
	path := os.Getenv("ADAMIC_VIEW_PAIRS_LOG")
	if path == "" || len(f.l.viewSources) == 0 {
		return
	}
	name := ""
	if files := f.l.program.Files(); len(files) != 0 {
		name = files[0].FileName()
	}
	records := []viewPairRecord{}
	seen := map[[2]*checker.Type]bool{}
	var pair func(source, target *checker.Type, field string, depth int, cast, narrows bool)
	pair = func(source, target *checker.Type, field string, depth int, cast, narrows bool) {
		if source == nil || target == nil {
			records = append(records, viewPairRecord{Program: name, Depth: depth, Field: field, Cast: cast, Narrows: narrows, Skipped: "no source"})
			return
		}
		source, target = f.l.concrete(source), f.l.concrete(target)
		key := [2]*checker.Type{source, target}
		if seen[key] || depth > 6 {
			return
		}
		seen[key] = true
		record := viewPairRecord{Program: name, Depth: depth, Source: f.l.checker.TypeToString(source), Target: f.l.checker.TypeToString(target), Field: field, Cast: cast, Narrows: narrows}
		switch {
		case source.Flags()&checker.TypeFlagsObject == 0 && source.Flags()&checker.TypeFlagsUnion == 0:
			record.Skipped = "source not an object"
		case target.Flags()&checker.TypeFlagsObject == 0 && target.Flags()&checker.TypeFlagsUnion == 0:
			record.Skipped = "target not an object"
		default:
			record.Related = f.deriveRelated(source, target)
		}
		records = append(records, record)
		if record.Skipped != "" {
			return
		}
		sourceNonNull, targetNonNull := f.l.checker.GetNonNullableType(source), f.l.checker.GetNonNullableType(target)
		for _, property := range f.fields(targetNonNull) {
			counterpart := f.l.checker.GetPropertyOfType(sourceNonNull, property.Name)
			if counterpart == nil {
				continue
			}
			pair(f.l.checker.GetTypeOfSymbol(counterpart), f.l.checker.GetTypeOfSymbol(property), property.Name, depth+1, cast, narrows)
		}
	}
	for _, view := range f.l.viewSources {
		pair(view.source, view.target, "", 0, view.cast, view.narrows)
	}
	viewPairsLock.Lock()
	defer viewPairsLock.Unlock()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, record := range records {
		_ = encoder.Encode(record)
	}
}
