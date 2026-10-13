package typeaware

import (
	"strconv"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// A checkerMutant is the checker archive (./bridge/tsgo/archive as a c-archive) with one change planted, which a
// mutant check links its native binary against. Each is a buildcache.GoBuildInMutated product keyed by its change and
// built ahead by its TestProduct_checker_mutant_ declaration below, so a check runs no go. Until landable-16 each check
// ran its own go build -overlay, which a runner refuses (landable-15: seventeen of them, all in this table).
type checkerMutant struct {
	name     string
	sanitize bool
	buildcache.Mutation
}

const (
	checkerFacts   = "bridge/tsgo/checker/facts.go"
	checkerProgram = "bridge/tsgo/checker/program.go"
	checkerArchive = "bridge/tsgo/archive/main.go"
)

// A malformed frame replaces the answer the archive hands back with value, so the native decoder must refuse it.
func checkerFrameMutant(name, buffer, value string) checkerMutant {
	return checkerMutant{name: name, Mutation: buildcache.Mutation{File: checkerArchive, Before: "*" + buffer + " = buffer(answer)", After: "_ = answer; *" + buffer + " = buffer(" + strconv.Quote(value) + ")"}}
}

var checkerMutants = []checkerMutant{
	// TestInspectRequestRefusals.
	{name: "wrong-kind", Mutation: buildcache.Mutation{File: checkerFacts, Before: `candidate.Kind.String() == "Kind"+kind`, After: `kind != ""`}},
	{name: "unknown-question", Mutation: buildcache.Mutation{File: checkerFacts, Before: `return "", fmt.Errorf("unsupported checker question: %s", question)`, After: `return out.String(), nil`}},
	// TestSixRuleAgreementAndMutants.
	{name: "declaration-source", Mutation: buildcache.Mutation{File: checkerFacts, Before: "scanner.GetTokenPosOfNode(name, source, false)", After: "scanner.GetTokenPosOfNode(name, f, false)"}},
	{name: "type-name", Mutation: buildcache.Mutation{File: checkerFacts, Before: "out.text(c.TypeToString(p.typesByID[id-1]))", After: "out.text(c.TypeToString(p.typesByID[0]))"}},
	{name: "raw-shape-constraint", Mutation: buildcache.Mutation{File: checkerFacts, Before: `if mode != "raw-type" && mode != "raw-shape"`, After: `if mode != "raw-type"`}},
	{name: "assignability-direction", Mutation: buildcache.Mutation{File: checkerFacts, Before: "checker.Checker_isTypeAssignableTo(c, c.GetTypeAtLocation(node), c.GetTypeAtLocation(target))", After: "checker.Checker_isTypeAssignableTo(c, c.GetTypeAtLocation(target), c.GetTypeAtLocation(node))"}},
	{name: "resolved-signature", Mutation: buildcache.Mutation{File: checkerFacts, Before: "signature := c.GetResolvedSignature(node)", After: `var firstCall *ast.Node; var findCall func(*ast.Node) bool; findCall = func(n *ast.Node) bool { if n.Kind == ast.KindCallExpression { firstCall = n; return true }; return n.ForEachChild(findCall) }; findCall(source.AsNode()); selected := node; if node.Kind == ast.KindCallExpression && firstCall != nil { selected = firstCall }; signature := c.GetResolvedSignature(selected)`}},
	// TestTypeAwareAgreementAndMutants' stale-handle check plants this same change, so both read one product.
	{name: "retained-handle", Mutation: buildcache.Mutation{File: checkerArchive, Before: "delete(programs.live, uint64(handle))", After: "// Mutant retains released roots."}},
	{name: "facts-length", sanitize: true, Mutation: buildcache.Mutation{File: checkerArchive, Before: "*facts = buffer(answer)", After: "*facts = buffer(answer); facts.length++"}},
	checkerFrameMutant("facts-empty-bad", "facts", ""),
	checkerFrameMutant("facts-integer-bad", "facts", "1\nx"),
	checkerFrameMutant("facts-length-bad", "facts", "-1\n"),
	checkerFrameMutant("facts-version-bad", "facts", "1\n2"),
	// TestTypeAwareAgreementAndMutants.
	checkerFrameMutant("empty-frame", "parts", ""),
	checkerFrameMutant("missing-header", "parts", "x"),
	checkerFrameMutant("bad-length", "parts", "1\n-1\n"),
	{name: "ignored-kind", Mutation: buildcache.Mutation{File: checkerProgram, Before: `strings.TrimPrefix(candidate.Kind.String(), "Kind") == kind`, After: `kind != ""`}},
}

// checkerMutantArchive is the path of the named checker mutant's archive.
func checkerMutantArchive(t *testing.T, name string) string {
	t.Helper()
	for _, mutant := range checkerMutants {
		if mutant.name != name {
			continue
		}
		var environment []string
		if mutant.sanitize {
			environment = []string{"CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"}
		}
		return buildcache.GoBuildInMutated(t, "", name+"-checker.a", "./bridge/tsgo/archive", []string{"-buildmode=c-archive"}, []buildcache.Mutation{mutant.Mutation}, environment...)
	}
	t.Fatalf("no checker mutant %s", name)
	return ""
}

func TestProduct_checker_mutant_wrong_kind(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "wrong-kind")
}
func TestProduct_checker_mutant_unknown_question(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "unknown-question")
}
func TestProduct_checker_mutant_declaration_source(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "declaration-source")
}
func TestProduct_checker_mutant_type_name(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "type-name")
}
func TestProduct_checker_mutant_raw_shape_constraint(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "raw-shape-constraint")
}
func TestProduct_checker_mutant_assignability_direction(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "assignability-direction")
}
func TestProduct_checker_mutant_resolved_signature(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "resolved-signature")
}
func TestProduct_checker_mutant_retained_handle(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "retained-handle")
}
func TestProduct_checker_mutant_facts_length(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "facts-length")
}
func TestProduct_checker_mutant_facts_empty_bad(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "facts-empty-bad")
}
func TestProduct_checker_mutant_facts_integer_bad(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "facts-integer-bad")
}
func TestProduct_checker_mutant_facts_length_bad(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "facts-length-bad")
}
func TestProduct_checker_mutant_facts_version_bad(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "facts-version-bad")
}
func TestProduct_checker_mutant_empty_frame(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "empty-frame")
}
func TestProduct_checker_mutant_missing_header(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "missing-header")
}
func TestProduct_checker_mutant_bad_length(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "bad-length")
}
func TestProduct_checker_mutant_ignored_kind(t *testing.T) {
	t.Parallel()
	checkerMutantArchive(t, "ignored-kind")
}
