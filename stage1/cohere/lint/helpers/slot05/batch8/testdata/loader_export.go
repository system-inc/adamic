package tailwind

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type AdamicFramework struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}
type AdamicPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type AdamicLoaderCase struct {
	Mode             string            `json:"mode"`
	EntryPoint       string            `json:"entryPoint"`
	PackageRoot      string            `json:"packageRoot"`
	ResolverPresent  bool              `json:"resolverPresent"`
	FrameworkPresent bool              `json:"frameworkPresent"`
	Framework        []AdamicFramework `json:"framework"`
	Defaults         []AdamicFramework `json:"defaults"`
	Version          string            `json:"version"`
	Absolute         string            `json:"absolute"`
	LoadError        string            `json:"loadError"`
	Stylesheets      []string          `json:"stylesheets"`
	Skipped          []string          `json:"skipped"`
	Statics          []AdamicPair      `json:"statics"`
	Roots            []AdamicPair      `json:"roots"`
	Custom           []string          `json:"custom"`
	Definitions      []string          `json:"definitions"`
	Count            int               `json:"count"`
	Trace            []string          `json:"trace"`
}

var adamicLoader *AdamicLoaderCase

func adamicTrace(s string) {
	if adamicLoader != nil {
		adamicLoader.Trace = append(adamicLoader.Trace, s)
	}
}
func adamicLoaderResolver(root string) StylesheetResolver {
	adamicTrace("resolver " + root)
	return NodeStylesheetResolver(root)
}
func adamicLoaderAbsolute(entry string) (string, error) {
	adamicTrace("absolute " + entry)
	p, e := filepath.Abs(entry)
	if adamicLoader != nil {
		adamicLoader.Absolute = p
	}
	return p, e
}
func adamicLoaderTheme() *Theme { adamicTrace("newTheme"); return NewTheme() }
func adamicLoaderFile(c *stylesheetCollector, path string) error {
	adamicTrace("load " + path)
	e := c.loadFile(path)
	if adamicLoader != nil {
		a := adamicLoader
		if e != nil {
			a.LoadError = e.Error()
		}
		a.Stylesheets = append([]string{}, c.stylesheets...)
		a.Skipped = []string{}
		for _, v := range c.skipped {
			a.Skipped = append(a.Skipped, v.Name+" "+v.Params+" "+v.Path)
		}
		a.Statics = []AdamicPair{}
		for _, k := range adamicSortedKeys(c.staticUtilityNodes) {
			a.Statics = append(a.Statics, AdamicPair{k, fmt.Sprint(len(c.staticUtilityNodes[k]))})
		}
		a.Roots = []AdamicPair{}
		for _, k := range adamicSortedKeys(c.utilityRoots) {
			a.Roots = append(a.Roots, AdamicPair{k, fmt.Sprint(c.utilityRoots[k][UtilityKindStatic]) + " " + fmt.Sprint(c.utilityRoots[k][UtilityKindFunctional])})
		}
		a.Custom = adamicSortedKeys(c.customVariants)
		a.Definitions = []string{}
		for _, d := range c.utilityDefinitions {
			a.Definitions = append(a.Definitions, d.Name)
		}
	}
	return e
}
func adamicLoaderVariants() *VariantRegistry { adamicTrace("newVariants"); return NewVariantRegistry() }
func adamicLoaderFramework(v *VariantRegistry) {
	adamicTrace("framework")
	v.RegisterFrameworkVariants(FrameworkVariantRegistrations)
}
func adamicLoaderBreakpoints(v *VariantRegistry, t *Theme) {
	adamicTrace("breakpoints")
	registerThemeBreakpointVariants(v, t)
}
func adamicLoaderComparisons(v *VariantRegistry, t *Theme) {
	adamicTrace("comparisons")
	attachFrameworkVariantComparisons(v, t)
}
func adamicLoaderRegister(v *VariantRegistry, n string, k ParsedVariantKind) {
	adamicTrace("register " + n + " " + string(k))
	v.Register(n, k)
}
func adamicLoaderEvaluator(t *Theme, ds []*UtilityDefinition) *UtilityEvaluator {
	adamicTrace("evaluator")
	return NewUtilityEvaluator(t, ds)
}
func adamicLoaderCount() int { adamicTrace("count"); return nextBuildCount() }
func AdamicLoad(a *AdamicLoaderCase) string {
	a.Defaults = []AdamicFramework{}
	for _, r := range FrameworkVariantRegistrations {
		a.Defaults = append(a.Defaults, AdamicFramework{r.Name, string(r.Kind)})
	}
	a.Version = TailwindVersion
	a.Count = BuildsSoFar() + 1
	a.Trace = []string{}
	adamicLoader = a
	opts := LoadOptions{EntryPoint: a.EntryPoint, TailwindPackageRoot: a.PackageRoot}
	if a.ResolverPresent {
		opts.Resolve = func(specifier, from string) (string, error) { return "", fmt.Errorf("unexpected import") }
	}
	if a.FrameworkPresent {
		opts.FrameworkVariants = []FrameworkVariant{}
		for _, r := range a.Framework {
			opts.FrameworkVariants = append(opts.FrameworkVariants, FrameworkVariant{r.Name, ParsedVariantKind(r.Kind)})
		}
	}
	s, e := LoadDesignSystem(opts)
	adamicLoader = nil
	var out strings.Builder
	p := func(v any) { fmt.Fprintln(&out, v) }
	for _, line := range a.Trace {
		p(line)
	}
	p("result")
	p(e == nil)
	if e != nil {
		p(e.Error())
		return out.String()
	}
	p("")
	p(s.EntryPoint)
	p(s.TailwindVersion)
	p(s.BuildCount)
	p(s.theme != nil)
	p(s.variants != nil)
	p(s.utility != nil)
	p(len(s.Stylesheets))
	for _, v := range s.Stylesheets {
		p(v)
	}
	p(len(s.SkippedDirectives))
	for _, v := range s.SkippedDirectives {
		p(v.Name + " " + v.Params + " " + v.Path)
	}
	for _, keys := range [][]string{adamicSortedKeys(s.staticUtilityNodes), adamicSortedKeys(s.utilityRoots), adamicSortedKeys(s.customVariants), adamicSortedKeys(s.frameworkVariants)} {
		p(len(keys))
		for _, v := range keys {
			p(v)
		}
	}
	return out.String()
}
func AdamicLoaderFileFixture(dir, source string, variant int) string {
	path := filepath.Join(dir, fmt.Sprintf("loader-%d.css", variant))
	css := "@theme { --breakpoint-tablet: 48rem; } @custom-variant z (&:hover); @custom-variant a (&:focus); @custom-variant hover (&:active); @custom-variant é (&:hover); @custom-variant 😀 (&:hover); @plugin 'example';"
	if variant == 1 {
		css += "@utility sample-* { color: --value(integer); } @utility sample { color: red; }"
	}
	if variant == 2 {
		css = "a { color: \"unterminated"
	}
	if e := os.WriteFile(path, []byte(css+"\n/* "+strings.ReplaceAll(source, "*/", "* /")+" */"), 0644); e != nil {
		panic(e)
	}
	return path
}

var _ = sort.Strings

func adamicSortedKeys[V any](values map[string]V) []string {
	keys := []string{}
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
