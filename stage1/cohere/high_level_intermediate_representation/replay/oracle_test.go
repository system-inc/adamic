//go:build lintoracle

// Overlay this beside Go HIR with the shared testdata/oracle_test.go printer.
// Pass workers supply their own payload encoders; this file owns only framing.
package high_level_intermediate_representation

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

type OracleExtraIdentity struct {
	FunctionPath, Kind string
	Id                 int
}
type OracleSidecarRow struct {
	Namespace, FunctionPath, AnchorKind string
	AnchorId                            *int
	Key, Payload                        string
}
type OracleCheckpoint struct {
	Key, Pass, Graph string
	Identities       []OracleExtraIdentity
	Sidecars         []OracleSidecarRow
}

// UTF-16 escapes; invalid Go string bytes use U+DC80..DCFF surrogate escape.
// This preserves TypeScript-go's internal byte-prefixed symbol names losslessly.
func oracleFrameField(s string) string {
	var out strings.Builder
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError && size == 1 {
			r = 0xdc00 + rune(s[0])
		}
		s = s[size:]
		units := []rune{r}
		if r > 0xffff {
			hi, lo := utf16.EncodeRune(r)
			units = []rune{hi, lo}
		}
		for _, unit := range units {
			if unit == 37 || unit == 9 || unit == 10 || unit == 13 || unit >= 128 {
				fmt.Fprintf(&out, "%%%04x", unit)
			} else {
				out.WriteRune(unit)
			}
		}
	}
	return out.String()
}

func OracleWriteCheckpoint(c OracleCheckpoint) string {
	if !strings.HasSuffix(c.Graph, "\n") {
		panic("graph must end with a newline")
	}
	var out strings.Builder
	fmt.Fprintf(&out, "hir-checkpoint-v1\ncase\t%s\t%s\ngraph-lines %d\n%s", oracleFrameField(c.Key), oracleFrameField(c.Pass), strings.Count(c.Graph, "\n"), c.Graph)
	identities := append([]OracleExtraIdentity(nil), c.Identities...)
	sort.Slice(identities, func(i, j int) bool {
		a, b := identities[i], identities[j]
		if a.FunctionPath != b.FunctionPath {
			return a.FunctionPath < b.FunctionPath
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Id < b.Id
	})
	fmt.Fprintf(&out, "identities %d\n", len(identities))
	for _, id := range identities {
		fmt.Fprintf(&out, "identity\t%s\t%s\t%d\n", oracleFrameField(id.FunctionPath), id.Kind, id.Id)
	}
	rows := append([]OracleSidecarRow(nil), c.Sidecars...)
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.Namespace != b.Namespace {
			return a.Namespace < b.Namespace
		}
		if a.FunctionPath != b.FunctionPath {
			return a.FunctionPath < b.FunctionPath
		}
		if a.AnchorKind != b.AnchorKind {
			return a.AnchorKind < b.AnchorKind
		}
		x, y := -1, -1
		if a.AnchorId != nil {
			x = *a.AnchorId
		}
		if b.AnchorId != nil {
			y = *b.AnchorId
		}
		if x != y {
			return x < y
		}
		return a.Key < b.Key
	})
	fmt.Fprintf(&out, "sidecars %d\n", len(rows))
	for _, row := range rows {
		id := "-"
		if row.AnchorId != nil {
			id = fmt.Sprint(*row.AnchorId)
		}
		fmt.Fprintf(&out, "sidecar\t%s\t%s\t%s\t%s\t%s\t%s\n", oracleFrameField(row.Namespace), oracleFrameField(row.FunctionPath), row.AnchorKind, id, oracleFrameField(row.Key), oracleFrameField(row.Payload))
	}
	return out.String() + "end-checkpoint\n"
}

// Deliberately synthetic sidecars exercise every anchor on real Go construction graphs.
// They are not analysis answers. Pass adapters replace this namespace with their own facts.
func OracleIdentityFixture(key, graph string) OracleCheckpoint {
	c := OracleCheckpoint{Key: key, Pass: "construction", Graph: graph}
	stack := []string{}
	pending := ""
	seen := map[string]bool{}
	add := func(path, kind string, id *int) {
		name := path + "/" + kind + "/" + fmt.Sprint(id)
		if id != nil {
			name = path + "/" + kind + "/" + fmt.Sprint(*id)
		}
		if seen[name] {
			return
		}
		seen[name] = true
		c.Sidecars = append(c.Sidecars, OracleSidecarRow{Namespace: "replay-fixture", FunctionPath: path, AnchorKind: kind, AnchorId: id, Key: "identity", Payload: "%\tline\n\r🙂"})
	}
	for _, line := range strings.Split(strings.TrimSuffix(graph, "\n"), "\n") {
		words := strings.Split(line, " ")
		if words[0] == "function" {
			path := "$"
			if len(stack) > 0 {
				path = stack[len(stack)-1] + "/" + pending
			}
			pending = ""
			stack = append(stack, path)
			add(path, "function", nil)
			for _, kind := range []string{"scope", "reactive"} {
				id := 17
				if kind == "reactive" {
					id = 23
				}
				c.Identities = append(c.Identities, OracleExtraIdentity{FunctionPath: path, Kind: kind, Id: id})
				add(path, kind, &id)
			}
		} else if words[0] == "nested" {
			pending = words[1]
		} else if words[0] == "end" {
			stack = stack[:len(stack)-1]
		} else if len(stack) > 0 {
			path := stack[len(stack)-1]
			kind := words[0]
			if kind == "orphan" {
				words = words[1:]
				kind = words[0]
			}
			if kind == "identifier" || kind == "block" || kind == "instruction" {
				var id int
				if _, err := fmt.Sscan(words[1], &id); err != nil {
					panic(err)
				}
				add(path, kind, &id)
				if kind == "identifier" {
					var d int
					if _, err := fmt.Sscan(words[2], &d); err != nil {
						panic(err)
					}
					add(path, "declaration", &d)
				}
			}
		}
	}
	for _, name := range []string{"😀", "\ue000"} {
		c.Sidecars = append(c.Sidecars, OracleSidecarRow{Namespace: "replay-fixture", FunctionPath: "$", AnchorKind: "function", Key: name, Payload: "unknown-schema/v99: preserved"})
	}
	return c
}
