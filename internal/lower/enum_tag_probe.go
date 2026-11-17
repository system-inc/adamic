package lower

// A probe for #x2cjvy9's split, not for landing: with ADAMIC_PROBE_ENUMTAG set to a file, every
// open-enum literal-tag refusal appends one line naming its site, the walk's first failing step and
// the node that step is written at.

import (
	"encoding/json"
	"os"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

type enumTagProbeRecord struct {
	why string
	at  *ast.Node
}

var (
	enumTagProbePath  = os.Getenv("ADAMIC_PROBE_ENUMTAG")
	enumTagProbeState sync.Map
	enumTagProbeMutex sync.Mutex
)

func (l *lowering) probeReset() {
	enumTagProbeState.Delete(l)
}

func (l *lowering) probeSet(why string, at *ast.Node) {
	if _, ok := enumTagProbeState.Load(l); ok {
		return
	}
	enumTagProbeState.Store(l, enumTagProbeRecord{why: why, at: at})
}

func (l *lowering) probeWrite(node, reference *ast.Node) {
	if enumTagProbePath == "" {
		return
	}
	record := enumTagProbeRecord{why: "proven-or-unset"}
	if value, ok := enumTagProbeState.Load(l); ok {
		record = value.(enumTagProbeRecord)
	}
	text := func(n *ast.Node) string {
		if n == nil {
			return ""
		}
		file := ast.GetSourceFileOfNode(n)
		if file == nil {
			return ""
		}
		source := file.Text()
		start, end := n.Pos(), n.End()
		if start < 0 || end > len(source) || start > end {
			return ""
		}
		return strings.Join(strings.Fields(source[start:end]), " ")
	}
	at := ""
	atKind := ""
	if record.at != nil {
		at = l.program.Where(record.at)
		atKind = record.at.Kind.String()
	}
	declared := ""
	if symbol := l.flagValueSymbol(node); symbol != nil {
		declared = l.checker.TypeToString(l.checker.GetTypeOfSymbol(symbol))
	}
	line, _ := json.Marshal(map[string]string{
		"site":      l.program.Where(node),
		"reference": text(reference),
		"declared":  declared,
		"why":       record.why,
		"at":        at,
		"atKind":    atKind,
		"atText":    text(record.at),
	})
	enumTagProbeMutex.Lock()
	defer enumTagProbeMutex.Unlock()
	file, err := os.OpenFile(enumTagProbePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.Write(append(line, '\n'))
}
