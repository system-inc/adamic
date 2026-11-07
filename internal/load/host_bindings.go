package load

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"sync"
	"unicode"
)

// A .ts file's ambient value declaration is a host binding only where this table names it (#00qn5az,
// docs/rulings.md's entry "A `.ts` file's ambient value declaration is a host binding"). Each row is
// an axiom under Language's audit, docs/axioms/host-bindings.md, which TestHostBindingsAudit holds to
// these rows both ways. A row says, per target, whether the host gives the name a value, what Node
// answers for its typeof under the oracle's module mode (internal/oracle/host_bindings_test.go holds
// every column to that), and the ruling or task that made it. Lowering reads the rows; it decides what
// a row admits (internal/lower/host_bindings.go).
//
//go:embed host_bindings.json
var hostBindingsJSON []byte

// HostBindingAnswer is what one target's host gives a host binding's name.
type HostBindingAnswer string

const (
	// HostBindingProvided is a name the host defines, read through its host model.
	HostBindingProvided HostBindingAnswer = "Provided"
	// HostBindingAbsent is a name nothing defines: typeof answers "undefined".
	HostBindingAbsent HostBindingAnswer = "Absent"
	// HostBindingPending is a column its owner hasn't cleared yet.
	HostBindingPending HostBindingAnswer = "Pending"
)

// HostBindingDeclaration is the declaration a row names: its file's basename, its keyword and its
// declared type's exact text, so a program's `declare const process: number` is never the row's.
type HostBindingDeclaration struct {
	File    string `json:"file"`
	Keyword string `json:"keyword"`
	Type    string `json:"type"`
}

// HostBindingNode is Node's answer for the name's typeof, measured on the version named, running the
// declaring file as an ECMAScript module, as oracle/node.mjs runs every program.
type HostBindingNode struct {
	Version string `json:"version"`
	Module  string `json:"module"`
	TypeOf  string `json:"typeof"`
}

// HostBindingColumn is one target's answer; a pending column names the owner who clears it.
type HostBindingColumn struct {
	Answer HostBindingAnswer `json:"answer"`
	Owner  string            `json:"owner,omitempty"`
}

// HostBindingModel is where a Provided row's reads go: the host path that models the value, or, while
// none is built, the task that builds it.
type HostBindingModel struct {
	Path string `json:"path"`
	Task string `json:"task"`
}

// HostBinding is one row of the table.
type HostBinding struct {
	Name        string                 `json:"name"`
	Declaration HostBindingDeclaration `json:"declaration"`
	Node        HostBindingNode        `json:"node"`
	Native      HostBindingColumn      `json:"native"`
	JavaScript  HostBindingColumn      `json:"javascript"`
	WASM32      HostBindingColumn      `json:"wasm32"`
	Model       HostBindingModel       `json:"model"`
	// Divergence names a ruled backend divergence (internal/oracle's ruledBackendDivergences) under
	// which a column may differ from Node; it is empty for every row today.
	Divergence string `json:"divergence"`
	Ruling     string `json:"ruling"`
}

// HostBindingTarget is one column of a row, by the target's name.
type HostBindingTarget struct {
	Target string
	Column HostBindingColumn
}

// Targets are the row's three columns, in the table's order: native, javascript and wasm32.
func (binding HostBinding) Targets() []HostBindingTarget {
	return []HostBindingTarget{
		{"native", binding.Native},
		{"javascript", binding.JavaScript},
		{"wasm32", binding.WASM32},
	}
}

// hostBindingTypeOfAnswers are the answers JavaScript's typeof gives.
var hostBindingTypeOfAnswers = map[string]bool{
	"undefined": true,
	"object":    true,
	"boolean":   true,
	"number":    true,
	"bigint":    true,
	"string":    true,
	"symbol":    true,
	"function":  true,
}

// HostBindings is the table, parsed and checked once.
var HostBindings = sync.OnceValues(func() ([]HostBinding, error) {
	return parseHostBindings(hostBindingsJSON)
})

func parseHostBindings(data []byte) ([]HostBinding, error) {
	var rows []HostBinding
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decodeError := decoder.Decode(&rows); decodeError != nil {
		return nil, fmt.Errorf("load: host_bindings.json: %w", decodeError)
	}
	if decoder.Decode(new(any)) != io.EOF {
		return nil, fmt.Errorf("load: host_bindings.json: trailing data")
	}
	seen := map[string]bool{}
	for index, row := range rows {
		if reason := hostBindingProblem(row); reason != "" {
			return nil, fmt.Errorf("load: host binding row %d (%s): %s", index+1, row.Name, reason)
		}
		if seen[row.Name] {
			return nil, fmt.Errorf("load: host binding row %d (%s): a second row for the name", index+1, row.Name)
		}
		seen[row.Name] = true
	}
	return rows, nil
}

// hostBindingProblem is what's wrong with a row, or "" when nothing is.
func hostBindingProblem(row HostBinding) string {
	if row.Name == "" || strings.IndexFunc(
		row.Name,
		func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '$' },
	) >= 0 || unicode.IsDigit(rune(row.Name[0])) {
		return "the name isn't an identifier"
	}
	declaration := row.Declaration
	if declaration.File != path.Base(declaration.File) ||
		path.Ext(declaration.File) != ".ts" ||
		strings.HasSuffix(declaration.File, ".d.ts") {
		return "the declaration's file isn't the basename of a .ts file"
	}
	if declaration.Keyword != "const" && declaration.Keyword != "let" {
		return "the declaration's keyword is neither const nor let"
	}
	if strings.TrimSpace(declaration.Type) != declaration.Type ||
		declaration.Type == "" ||
		strings.ContainsAny(declaration.Type, "\r\n") {
		return "the declared type isn't one line of exact text"
	}
	if !strings.HasPrefix(row.Node.Version, "v") ||
		row.Node.Module != "ES" ||
		!hostBindingTypeOfAnswers[row.Node.TypeOf] {
		return "Node's answer needs its version, module mode ES and a typeof answer"
	}
	provided := false
	for _, target := range row.Targets() {
		switch target.Column.Answer {
		case HostBindingProvided:
			provided = true
		case HostBindingAbsent:
		case HostBindingPending:
			if target.Column.Owner == "" {
				return "the " + target.Target + " column is pending with no owner to clear it"
			}
			continue
		default:
			return "the " + target.Target + " column answers neither Provided, Absent nor Pending"
		}
		if target.Column.Owner != "" {
			return "the " + target.Target + " column names an owner but isn't pending"
		}
	}
	if provided != (row.Model.Path != "" || row.Model.Task != "") {
		return "a model is named exactly when a column is Provided"
	}
	if strings.TrimSpace(row.Ruling) == "" {
		return "no ruling or task names what made the row"
	}
	return ""
}
