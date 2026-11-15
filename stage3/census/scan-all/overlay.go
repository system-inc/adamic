// Package scanall generates a scratch-only, no-output measurement overlay.
// Compiler consumer marker: github.com/system-inc/adamic/internal/lower.
// The measurement driver imports it from generated .go.txt text, which the
// tracked-Go dependency census otherwise cannot see.
package scanall

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/system-inc/adamic/stage3/census/latent/refusalrewrite"
)

//go:embed measure.go.txt
var measure string

//go:embed main.go.txt
var driver string

//go:embed bridge.go.txt
var bridge string

//go:embed heads/*.go.txt
var ruleHeads embed.FS

// Rewrite returns copies only. The whole-file anchors deliberately fail on any
// drift, including new pipeline stops that would otherwise inherit a stale label.
func Rewrite(lower, refusals []byte) (map[string]string, error) {
	for _, a := range []struct {
		name, hash string
		data       []byte
	}{
		{"lower.go snapshot", "9eb309df48b07d2e828eab1a8fc6689bb016902a1a0bd98224b62fb6c093e39e", lower},
		{"refusals.go snapshot", "4d417bf10e6a7d304abbae97a4db087806ffdaf0b17e7706a02f82211c70f05e", refusals},
	} {
		if fmt.Sprintf("%x", sha256.Sum256(a.data)) != a.hash {
			return nil, fmt.Errorf("scan-all anchor %s: source drift; review before updating", a.name)
		}
	}
	s := string(lower)
	var err error
	once := func(name, old, new string) {
		if err != nil {
			return
		}
		n := strings.Count(s, old)
		if n != 1 {
			err = fmt.Errorf("scan-all anchor %s: want exactly one match, got %d", name, n)
			return
		}
		s = strings.Replace(s, old, new, 1)
	}
	once("graph record field", "graphNodeArrays []graphNodeArraySite", "graphNodeArrays []graphNodeArraySite\n graphRecords []graphRecordSite")
	once("Lower entry", "func Lower(", "func scanAllLower(")
	once("LowerWithOptions entry", "func LowerWithOptions(", "func scanAllLowerWithOptions(")
	once("Lower delegation", "return LowerWithOptions(ctx, program, Options{", "return scanAllLowerWithOptions(ctx, program, Options{")
	once("underlying abort", "defer func() { err = lowering.hostMemberPending(err) }()", "defer func() { if err != nil { stop := scanAllDescribe(err); scanAllUnderlying = &stop }; err = lowering.hostMemberPending(err) }()")
	once("refusal scan", "lowering.refuse(module)", "lowering.scanAllRefuse(module)")
	// Every matched boundary is also a unique literal anchor. The snapshot above
	// rejects added/deleted/reshaped boundaries before this instrumentation runs.
	re := regexp.MustCompile(`(?m)^(\t*)(if err := ([^;\n]+); err != nil \{)`)
	matches := re.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		once("phase "+m[3], m[0], m[1]+"scanAllPhase = "+fmt.Sprintf("%q", strings.Split(m[3], "(")[0])+"\n"+m[0])
	}
	for _, a := range []struct{ old, phase string }{
		{"modules, err := lowering.moduleOrder(entry)", "moduleOrder"},
		{"body, err := lowering.moduleStatements(module)", "moduleStatements"},
		{"if lowering.unlowerable != nil {", "unlowerable"},
		{"if len(lowering.asyncPayloadFailures) > 0 {", "asyncPayloadFailures"},
	} {
		once(a.phase, a.old, "scanAllPhase = "+fmt.Sprintf("%q", a.phase)+"\n"+a.old)
	}
	if err != nil {
		return nil, err
	}
	s += `
func Lower(ctx context.Context, p *load.Program) (*ir.Program,error) { return nil, fmt.Errorf("scan-all: measurement only; no IR output") }
func LowerWithOptions(ctx context.Context, p *load.Program, o Options) (*ir.Program,error) { return Lower(ctx,p) }
`
	generated, e := refusalrewrite.Rewrite(refusals)
	if e != nil {
		return nil, e
	}
	f := string(generated)
	for _, a := range []struct{ visitor, binding string }{{"contracts", "contractError"}, {"visit", "found"}} {
		startAnchor := a.visitor + " = func(node *ast.Node) (latentStop bool) {"
		endAnchor := "\t\tdefer func() {\n\t\t\tlatentRecord(" + a.binding + ")"
		if strings.Count(f, startAnchor) != 1 || strings.Count(f, endAnchor) != 1 {
			return nil, fmt.Errorf("scan-all anchor %s visitor/body-exclusion: want exactly one match", a.visitor)
		}
		start := strings.Index(f, startAnchor) + len(startAnchor)
		end := strings.Index(f, endAnchor)
		if end <= start || !strings.Contains(f[start:end], "latentFullEnabled()") {
			return nil, fmt.Errorf("scan-all anchor %s body-exclusion: unsupported shape", a.visitor)
		}
		f = f[:start] + "\n" + f[end:]
	}
	anchor := "prune, err := rule.check(l, node)"
	// The original production refuse and its clone each contain this call.
	split := "func (l *lowering) latentRefuse("
	if strings.Count(f, split) != 1 {
		return nil, fmt.Errorf("scan-all anchor latentRefuse: want exactly one match")
	}
	pos := strings.Index(f, split)
	clone := f[pos:]
	if strings.Count(clone, anchor) != 1 {
		return nil, fmt.Errorf("scan-all anchor registry call: want exactly one match")
	}
	f = f[:pos] + strings.Replace(clone, anchor, "scanAllPhase = \"refusal_scan/\" + rule.name\n"+anchor, 1)
	sources := map[string]string{
		"internal/lower/lower.go":               s,
		"internal/lower/refusals.go":            f,
		"internal/lower/scan_all.go":            measure,
		"stage3/census/scan-all/tool/main.go":   driver,
		"stage3/census/scan-all/bridge_test.go": bridge,
	}
	for _, name := range []string{"function_bind.go", "cycles.go", "graph_records.go", "records.go"} {
		data, e := ruleHeads.ReadFile("heads/" + name + ".txt")
		if e != nil {
			return nil, e
		}
		sources["internal/lower/"+name] = string(data)
	}
	return sources, nil
}

// OutsideRepository keeps all generated source and the overlay map outside the
// checkout. Refuse symlink aliases too; the caller resolves existing parents.
func OutsideRepository(repo, out string) bool {
	rel, err := filepath.Rel(repo, out)
	return err == nil && rel != "." && (rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// ValidateRuleInputs anchors the current-main files replaced wholesale by the
// scout heads. Never silently discard a changed production rule implementation.
func ValidateRuleInputs(repo string) error {
	for name, hash := range map[string]string{
		"function_bind.go": "8bb7b8b72c9cc3317e4cc5c578f8a6f141c445b7d560a62ccabe26a54f4eaf2f",
		"cycles.go":        "b133d7d1f97c570efcce7d67263411ee1306ef319e93d80cb22ad607d7065180",
		"graph_records.go": "8ca27443477ac37f8b093151b01a70c8bc378511b8c5bbc555a13a75aff933f1",
		"records.go":       "52d3c14644e5128b18d60a27b30663ada477c8edf431d8f7e3d76d75b51c5cac",
	} {
		data, err := os.ReadFile(filepath.Join(repo, "internal/lower", name))
		if err != nil {
			return fmt.Errorf("scan-all anchor %s snapshot: %w", name, err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != hash {
			return fmt.Errorf("scan-all anchor %s snapshot: source drift; review before replacing with scout head", name)
		}
	}
	return nil
}
