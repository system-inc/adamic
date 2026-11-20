package lower

import (
	"errors"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"os/exec"
	"strings"
	"testing"
)

type kindHelperMutation struct {
	name, file, before, after, pin, witness string
	rows                                    []string
}

var kindHelperMutations = []kindHelperMutation{
	{"Depth", "internal/lower/enum_tag_kind_helper.go", "if r.depth >= kindHelperMaxDepth {", "if false && r.depth >= kindHelperMaxDepth {", "TestKindHelperFallbacksRejectWholeBody/depth", "kind-helper fallback depth", []string{}},
	{"Cycle", "internal/lower/enum_tag_kind_helper.go", "if r.active[function] {\n\t\treturn false\n\t}", "if r.active[function] {\n\t\treturn true\n\t}", "TestKindHelperFallbacksRejectWholeBody/cycle", "kind-helper fallback cycle", []string{}},
	{"NonHelper", "internal/lower/enum_tag_kind_helper.go", "if !child.kindHelperBody(function.Body()) {\n\t\treturn false\n\t}", "if !child.kindHelperBody(function.Body()) {\n r.enum=child.enum; for _,kind:=range child.members {r.add(kind)}; return true\n}", "TestKindHelperFallbacksRejectWholeBody/(non_helper|other_read)", "kind-helper fallback", []string{"non_helper", "other_read"}},
	{"WrittenParameter", "internal/lower/enum_tag_kind_helper.go", "return r.body(body)", "if len(statements)>1 && statements[len(statements)-1].Kind==ast.KindReturnStatement {return r.comparisons(statements[len(statements)-1].AsReturnStatement().Expression)}; return r.body(body)", "TestKindHelperFallbacksRejectWholeBody/written_parameter", "kind-helper fallback written_parameter", []string{}},
	{"WrongArgument", "internal/lower/enum_tag_kind_helper.go", "if !r.read(call.Arguments.Nodes[0]) {", "if false && !r.read(call.Arguments.Nodes[0]) {", "TestKindHelperFallbacksRejectWholeBody/(wrong_field|arithmetic_argument)", "kind-helper fallback", []string{"wrong_field", "arithmetic_argument"}},
	{"Call", "internal/lower/enum_tag_kind_set.go", "return r.kindHelperCall(n)", "return false", "TestKindHelperShapesAgreeWithNode/default_chain", "kind-helper grammar default_chain", []string{}},
	{"Scalar", "internal/lower/enum_tag_kind_set.go", "if r.scalar {\n\t\treturn r.parameterRead(n)\n\t}", "if r.scalar {\n\t\treturn false\n\t}", "TestKindHelperShapesAgreeWithNode/default_chain", "kind-helper grammar default_chain", []string{}},
	{"Default", "internal/lower/enum_tag_kind_helper.go", "} else if c.Kind == ast.KindDefaultClause && expression.Kind == ast.KindCallExpression {", "} else if false && c.Kind == ast.KindDefaultClause && expression.Kind == ast.KindCallExpression {", "TestKindHelperShapesAgreeWithNode/default_chain", "kind-helper grammar default_chain", []string{}},
	{"Tail", "internal/lower/enum_tag_kind_helper.go", "tail = fork.members", "return false", "TestKindHelperShapesAgreeWithNode/switch_tail", "kind-helper grammar switch_tail", []string{}},
	{"ReturnBody", "internal/lower/enum_tag_kind_helper.go", "return r.body(body)", "return false", "TestKindHelperShapesAgreeWithNode/return_chain", "kind-helper grammar return_chain", []string{}},
	{"Arrow", "internal/lower/enum_tag_kind_helper.go", "function.Parent.Flags&ast.NodeFlagsConst == 0", "true", "TestKindHelperShapesAgreeWithNode/arrow", "kind-helper grammar arrow", []string{}},
	{"DefaultMask", "internal/lower/enum_tag_kind_helper.go", "if !excluded {", "if true || !excluded {", "TestKindHelperDefaultExcludesExplicitFalse", "kind-helper default exclusion", []string{}},
	{"DefaultParameter", "internal/lower/enum_tag_kind_helper.go", "if p.Initializer != nil || p.QuestionToken != nil || p.DotDotDotToken != nil {", "if false && p.Initializer != nil || p.QuestionToken != nil || p.DotDotDotToken != nil {", "TestKindHelperFallbacksRejectWholeBody/default_parameter", "kind-helper fallback default_parameter", []string{}},
	{"Enum", "internal/lower/enum_tag_kind_set.go", "if identity == nil || r.enum != nil && r.enum != identity {", "if identity == nil {", "TestKindHelperFallbacksRejectWholeBody/two_enums", "kind-helper fallback two_enums", []string{}},
	{"Overload", "internal/lower/enum_tag_kind_helper.go", "if symbol == nil || len(symbol.Declarations) != 1 {\n\t\treturn false\n\t}\n\tfunction := symbol.Declarations[0]", "if symbol == nil || len(symbol.Declarations)==0 {return false}; function:=symbol.Declarations[len(symbol.Declarations)-1]", "TestKindHelperFallbacksRejectWholeBody/overload", "kind-helper fallback overload", []string{}},
	{"Let", "internal/lower/enum_tag_kind_helper.go", "function.Parent.Flags&ast.NodeFlagsConst == 0", "false", "TestKindHelperFallbacksRejectWholeBody/let_helper", "kind-helper fallback let_helper", []string{}},
	{"AliasArgument", "internal/lower/enum_tag_kind_helper.go", "if r.alias != nil {\n\t\treturn false\n\t}", "if false && r.alias != nil {\n\t\treturn false\n\t}", "TestKindHelperFallbacksRejectWholeBody/alias_argument", "kind-helper fallback alias_argument", []string{}},
	{"MixedDisjunction", "internal/lower/enum_tag_kind_set.go", "return r.comparisons(b.Left) && r.comparisons(b.Right)", "return r.comparisons(b.Left)", "TestKindHelperFallbacksRejectWholeBody/mixed_disjunction", "kind-helper fallback mixed_disjunction", []string{}},
	{"DuplicateCase", "internal/lower/enum_tag_kind_helper.go", "if checker.Checker_isTypeIdenticalTo(r.l.checker, held, kind) {", "if false && checker.Checker_isTypeIdenticalTo(r.l.checker, held, kind) {", "TestKindHelperFallbacksRejectWholeBody/h05", "kind-helper fallback h05", []string{}},
	{"EmptyDefaultTail", "internal/lower/enum_tag_kind_helper.go", "fallback = dynamic // An empty default inherits its successor, not the tail.", "// mutant: an empty default retains the unreachable tail", "TestKindHelperEmptyDefaultTailRoutes/empty_false", "kind-helper empty default resurrected an unreachable tail", []string{}},
}

func kindHelperMutantBinary(t *testing.T, m kindHelperMutation) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "kind-helper-"+m.name+".test", "./internal/lower", nil, []buildcache.Mutation{{File: m.file, Before: m.before, After: m.after}})
}
func TestProduct_KindHelperDepthMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[0])
}
func TestProduct_KindHelperCycleMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[1])
}
func TestProduct_KindHelperNonHelperMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[2])
}
func TestProduct_KindHelperWrittenParameterMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[3])
}
func TestProduct_KindHelperWrongArgumentMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[4])
}
func TestProduct_KindHelperCallMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[5])
}
func TestProduct_KindHelperScalarMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[6])
}
func TestProduct_KindHelperDefaultMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[7])
}
func TestProduct_KindHelperTailMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[8])
}
func TestProduct_KindHelperReturnBodyMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[9])
}
func TestProduct_KindHelperArrowMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[10])
}
func TestProduct_KindHelperDefaultMaskMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[11])
}
func TestProduct_KindHelperDefaultParameterMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[12])
}
func TestProduct_KindHelperEnumMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[13])
}
func TestProduct_KindHelperOverloadMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[14])
}
func TestProduct_KindHelperLetMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[15])
}
func TestProduct_KindHelperAliasArgumentMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[16])
}
func TestProduct_KindHelperMixedDisjunctionMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[17])
}

func TestProduct_KindHelperDuplicateCaseMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[18])
}

func TestProduct_KindHelperEmptyDefaultTailMutant(t *testing.T) {
	t.Parallel()
	kindHelperMutantBinary(t, kindHelperMutations[19])
}

func TestKindHelperMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, m := range kindHelperMutations {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			binary := kindHelperMutantBinary(t, m)
			command := exec.CommandContext(t.Context(), binary, "-test.run", "^"+strings.ReplaceAll(m.pin, "/", "$/^")+"$", "-test.count=1", "-test.v")
			stdout, stderr, err := childguard.Capture(t.Context(), command)
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(stdout), m.witness) {
				t.Fatalf("%s helper mutant missed %s: %v\n%s\n%s", m.name, m.pin, err, stdout, stderr)
			}
			for _, row := range m.rows {
				if !strings.Contains(string(stdout), "kind-helper fallback "+row+" was recognized") {
					t.Fatalf("%s helper mutant missed %s:\n%s", m.name, row, stdout)
				}
			}
			t.Logf("%s helper mutant caught by %s:\n%s", m.name, m.pin, stdout)
		})
	}
}
