package lower

// Probe-only instrumentation for #s8gy89r. Never for landing.
import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"os"
	"strings"
	"sync"
)

type enumTagProbeRecord struct {
	why    string
	at     *ast.Node
	detail string
}

var (
	enumTagProbePath  = os.Getenv("ADAMIC_PROBE_ENUMTAG")
	enumTagProbeState sync.Map
	enumTagProbeMutex sync.Mutex
)

func (l *lowering) probeReset()                       { enumTagProbeState.Delete(l) }
func (l *lowering) probeSet(why string, at *ast.Node) { l.probeSetDetail(why, at, "") }
func (l *lowering) probeSetDetail(why string, at *ast.Node, detail string) {
	if enumTagProbePath == "" {
		return
	}
	if _, ok := enumTagProbeState.Load(l); ok {
		return
	}
	enumTagProbeState.Store(l, enumTagProbeRecord{why, at, detail})
}
func (l *lowering) probeHazardDetail(union *checker.Type, field string) string {
	if enumTagProbePath == "" {
		return ""
	}
	if union.Flags()&checker.TypeFlagsUnion == 0 {
		return "not a union"
	}
	for _, member := range union.Types() {
		for _, property := range l.checker.GetPropertiesOfType(member) {
			if field != "" && property.Name != field {
				continue
			}
			tag := l.checker.GetTypeOfSymbol(property)
			if l.openNumericEnumType(tag) && !sharedByEveryMember(l.checker, union, property.Name, tag) && l.enumTagClosedTag(union, member, property.Name, tag) == nil {
				_, closed, reason := l.closedKinds(member, property.Name)
				return "open " + property.Name + " on " + l.checker.TypeToString(member) + ": " + l.checker.TypeToString(tag) + fmt.Sprintf(" | closed=%v %s", closed, reason)
			}
		}
	}
	return "shared value"
}
func probeAppend(row map[string]string) {
	line, _ := json.Marshal(row)
	enumTagProbeMutex.Lock()
	defer enumTagProbeMutex.Unlock()
	file, err := os.OpenFile(enumTagProbePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.Write(append(line, '\n'))
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
	at, atKind := "", ""
	if record.at != nil {
		at = l.program.Where(record.at)
		atKind = record.at.Kind.String()
	}
	declared := ""
	if symbol := l.flagValueSymbol(node); symbol != nil {
		declared = l.checker.TypeToString(l.checker.GetTypeOfSymbol(symbol))
	}
	probeAppend(map[string]string{"site": l.program.Where(node), "reference": text(reference), "declared": declared, "why": record.why, "at": at, "atKind": atKind, "atText": text(record.at), "detail": record.detail})
}
func (l *lowering) probeFitFail(of, kind, target, source *checker.Type, at *ast.Node, reason string) {
	if enumTagProbePath == "" {
		return
	}
	where, sourceText := "", ""
	if at != nil {
		where = l.program.Where(at)
	}
	if source != nil {
		sourceText = l.checker.TypeToString(source)
	}
	probeAppend(map[string]string{"site": where, "why": "fit-fail", "at": where, "atText": reason, "detail": l.checker.TypeToString(of) + " kind " + l.checker.TypeToString(kind) + " target " + l.checker.TypeToString(target) + " source " + sourceText})
}

// Recognizer misses are tentative: another operand or grammar can still succeed.
// Keep their labels on the walk/parser, publishing only a final failed walk step.
func (w *enumTagWalk) probeMiss(why string, at *ast.Node) {
	if enumTagProbePath != "" {
		w.probeWhy = why
		w.probeAt = at
	}
}
func (w *enumTagWalk) probeReject(why string, at *ast.Node) bool  { w.probeMiss(why, at); return false }
func (w *enumTagWalk) probeField(why string, at *ast.Node) string { w.probeMiss(why, at); return "" }
func (r *kindSetBody) probeReject(why string, at *ast.Node) bool {
	if enumTagProbePath != "" {
		r.probeWhy = why
		r.probeAt = at
	}
	return false
}
func (r *kindSetBody) probeResult(why string, at *ast.Node, result bool) bool {
	if !result {
		r.probeReject(why+":"+r.probeWhy, at)
	}
	return result
}
func (r *kindSetBody) probeMember(why string, at *ast.Node) *checker.Type {
	r.probeReject(why, at)
	return nil
}

// The free source-order helper has no lowering receiver. Carry its tentative
// miss by the alias node until kindAliasField consumes it, never as a walk failure.
var enumTagAliasMiss sync.Map

func probeAliasNo(why string, alias *ast.Node) bool {
	if enumTagProbePath != "" {
		enumTagAliasMiss.Store(alias, why)
	}
	return false
}
