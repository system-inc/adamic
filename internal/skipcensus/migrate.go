package skipcensus

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// conversionKey is independent of historical joined-caller identities and lines.
func conversionKey(r Row) string {
	return r.File + "\x00" + r.Test + "\x00" + r.Condition + "\x00" + r.Message
}

// Annotate converts a historical table through AST-discovered call positions.
// It verifies every declaration before writing and proves class/provider equality
// afterward. No regular expression locates or edits a Go call.
func Annotate(root string, declared []Row) (int, error) {
	actual, err := Inventory(root)
	if err != nil {
		return 0, err
	}
	previous := map[string]Row{}
	for _, r := range declared {
		k := conversionKey(r)
		if _, ok := previous[k]; ok {
			return 0, fmt.Errorf("duplicate migration row %s", r.File)
		}
		previous[k] = r
	}
	groups := map[string][]Row{}
	for _, r := range actual {
		d, ok := previous[conversionKey(r)]
		if !ok {
			return 0, fmt.Errorf("no table declaration for %s:%d (%s)", r.File, r.Line, r.Test)
		}
		if r.Class != "" {
			return 0, fmt.Errorf("already annotated %s:%d", r.File, r.Line)
		}
		r.Class = d.Class
		r.Provides = d.Provides
		if err := Audit([]Row{r}); err != nil {
			return 0, err
		}
		groups[r.File] = append(groups[r.File], r)
	}
	if len(actual) != len(declared) {
		return 0, fmt.Errorf("table has %d rows; source has %d", len(declared), len(actual))
	}
	changes := map[string][]byte{}
	for file, rows := range groups {
		path := filepath.Join(root, file)
		data, err := os.ReadFile(path)
		if err != nil {
			return 0, err
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].offset > rows[j].offset })
		for _, r := range rows {
			start := strings.LastIndex(string(data[:r.offset]), "\n") + 1
			prefix := string(data[start:r.offset])
			indent := prefix[:len(prefix)-len(strings.TrimLeft(prefix, "\t "))]
			text := "// census: " + r.Class
			if r.Provides != "" {
				text += " " + strings.Join(strings.Fields(r.Provides), " ")
			}
			position := start
			insertion := indent + text + "\n"
			if strings.TrimSpace(prefix) != "" {
				position = r.offset
				insertion = "\n" + indent + text + "\n" + indent
			}
			data = append(append(append([]byte{}, data[:position]...), []byte(insertion)...), data[position:]...)
		}
		formatted, err := format.Source(data)
		if err != nil {
			return 0, err
		}
		changes[path] = formatted
	}
	for path, data := range changes {
		info, err := os.Stat(path)
		if err != nil {
			return 0, err
		}
		if err := os.WriteFile(path, data, info.Mode().Perm()); err != nil {
			return 0, err
		}
	}
	after, err := Scan(root)
	if err != nil {
		return 0, err
	}
	if len(after) != len(declared) {
		return 0, fmt.Errorf("conversion changed row count")
	}
	for _, r := range after {
		old, ok := previous[conversionKey(r)]
		if !ok || old.Class != r.Class || strings.Join(strings.Fields(old.Provides), " ") != r.Provides {
			return 0, fmt.Errorf("lossy conversion at %s:%d", r.File, r.Line)
		}
	}
	return len(after), nil
}
