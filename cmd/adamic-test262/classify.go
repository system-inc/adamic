package main

import (
	"regexp"
	"sort"
	"strings"
)

// harnessAdamicCanLoad is the harness files whose job the prelude does. compareArray.js is empty
// upstream (the function lives in assert.js). Anything else needs defineProperty, getters, Reflect,
// proxies, or another host the prelude does not stand in for, and the test is skipped.
var harnessAdamicCanLoad = map[string]bool{
	"assert.js":       true,
	"sta.js":          true,
	"compareArray.js": true,
	"regExpUtils.js":  true,
}

// classified is one test, either skipped with a reason or ready to run.
type classified struct {
	Path                 string
	Directory            string
	Skip                 string
	NegativePhase        string
	NegativeType         string
	Program              string
	Adaptations          map[string]int
	Original             string
	Includes             []string
	ConstructorAssertion bool
}

// classify decides whether a test is attempted. Skipped tests are the ones Adamic will not learn
// from by compiling: a feature 0.1 refuses or has not built (Proxy, Symbol, eval, with, getters,
// async, and the same kind of thing), a negative parse or early error (test262's phase "parse"
// covers both, per INTERPRETING.md; an older "early" is kept), a negative resolution, or a harness
// include the prelude cannot stand in for. A missing library method is not skipped. It is attempted,
// and the refusal reason is what the library waves measure.
//
// adapt applies the in-memory test262-style rewrites (var to let, callback parameters, strict
// equality, Test262Error). It never writes the checkout. Without it, the body is only the harness
// rename the prelude needs.
func classify(path string, source string, adapt bool) classified {
	result := classified{Path: path, Directory: directoryOf(path)}
	meta := parseFrontmatter(source)
	if reason := negativeSkip(meta); reason != "" {
		result.Skip = reason
		return result
	}
	result.NegativePhase = meta.NegativePhase
	result.NegativeType = meta.NegativeType
	if strings.HasPrefix(path, "built-ins/RegExp/") {
		result.Original = bodyAfterFrontmatter(source)
		result.Includes = meta.Includes
		tokens := tokenize(result.Original)
		for i := 0; i+3 < len(tokens); i++ {
			if tokens[i].text == "assert" && tokens[i+1].text == "." && tokens[i+2].text == "throws" && tokens[i+3].text == "(" {
				result.ConstructorAssertion = true
			}
		}
	}
	if reason := flagSkip(meta.Flags); reason != "" {
		result.Skip = reason
		return result
	}
	if reason := includeSkip(meta.Includes); reason != "" {
		result.Skip = reason
		return result
	}
	if reason := featureSkip(meta.Features); reason != "" {
		result.Skip = reason
		return result
	}
	body := bodyAfterFrontmatter(source)
	if adapt && result.Original != "" {
		literalAdapted := adaptRegExpLiterals(body)
		body = literalAdapted.Source
		result.Adaptations = literalAdapted.Counts
	}
	if adapt && result.Original != "" {
		legacy := adaptRegExpConstructorErrors(body)
		body = legacy.Source
		for kind, count := range legacy.Counts {
			result.Adaptations[kind] += count
		}
	}
	if reason := syntaxSkip(codeOnly(body)); reason != "" {
		result.Skip = reason
		return result
	}
	if adapt {
		rewritten := adaptSource(body)
		body = rewritten.Source
		if result.Adaptations == nil {
			result.Adaptations = map[string]int{}
		}
		for kind, count := range rewritten.Counts {
			result.Adaptations[kind] += count
		}
		if result.Original != "" {
			regexAdapted := adaptRegExp(body)
			body = regexAdapted.Source
			if result.Adaptations == nil {
				result.Adaptations = map[string]int{}
			}
			for kind, count := range regexAdapted.Counts {
				result.Adaptations[kind] += count
			}
		}
	}
	if adapt && result.Original != "" {
		checked := adaptRegExpSyntaxAssertions(body)
		body = checked.Source
		for kind, count := range checked.Counts {
			result.Adaptations[kind] += count
		}
		// Only assertions discharged by the intrinsic-only proof can bypass
		// the generic harness constructor-identity refusal.
		result.ConstructorAssertion = false
		tokens := tokenize(body)
		for i := 0; i+3 < len(tokens); i++ {
			if tokens[i].text == "assert" && tokens[i+1].text == "." && tokens[i+2].text == "throws" && tokens[i+3].text == "(" {
				result.ConstructorAssertion = true
			}
		}
	}
	result.Program = program(rewriteHarnessCalls(body))
	if result.Adaptations["regex-intrinsic-syntax-assertion"] > 0 {
		result.Program = regexpSyntaxPrelude + "\n" + result.Program
	}
	for _, include := range meta.Includes {
		if include == "regExpUtils.js" {
			result.Program = regexpPrelude + "\n" + result.Program
			break
		}
	}
	return result
}

func directoryOf(path string) string {
	slash := strings.LastIndex(path, "/")
	if slash < 0 {
		return "."
	}
	return path[:slash]
}

func negativeSkip(meta frontmatter) string {
	switch meta.NegativePhase {
	case "", "runtime":
		// A runtime negative test is attempted. Its expectation is applied when the two runs are
		// compared, not by skipping it.
		return ""
	case "parse":
		// test262 folded early errors into phase parse. Both are "the program is not a program".
		return "negative parse or early error"
	case "early":
		return "negative early error"
	case "resolution":
		return "negative resolution"
	default:
		return "negative " + meta.NegativePhase
	}
}

func flagSkip(flags []string) string {
	for _, flag := range flags {
		switch flag {
		case "noStrict":
			return "noStrict (sloppy mode only; Adamic modules are always strict)"
		case "async":
			return "async"
		case "module":
			return "module"
		case "CanBlockIsTrue", "CanBlockIsFalse":
			return "shared memory"
		}
	}
	return ""
}

func includeSkip(includes []string) string {
	var missing []string
	for _, include := range includes {
		if !harnessAdamicCanLoad[include] {
			missing = append(missing, include)
		}
	}
	if len(missing) == 0 {
		return ""
	}
	sort.Strings(missing)
	if len(missing) == 1 {
		return "harness include " + missing[0]
	}
	return "harness includes " + strings.Join(missing, ", ")
}

func featureSkip(features []string) string {
	var refused []string
	seen := map[string]bool{}
	for _, feature := range features {
		if _, skip := unsupportedFeature(feature); skip && !seen[feature] {
			seen[feature] = true
			refused = append(refused, feature)
		}
	}
	if len(refused) == 0 {
		return ""
	}
	sort.Strings(refused)
	if len(refused) == 1 {
		return "feature " + refused[0]
	}
	return "features " + strings.Join(refused, ", ")
}

// unsupportedFeature reports whether a test262 feature is one Adamic refuses or has not built as a
// language feature, as opposed to a library method stage 0 may simply not lower yet.
func unsupportedFeature(feature string) (string, bool) {
	switch feature {
	case "Proxy", "Reflect", "Symbol", "BigInt", "generators", "async-functions",
		"async-iteration", "top-level-await", "Atomics", "SharedArrayBuffer", "ArrayBuffer",
		"resizable-arraybuffer", "TypedArray", "DataView", "cross-realm", "eval", "Temporal",
		"WeakMap", "WeakSet", "WeakRef", "FinalizationRegistry", "decorators", "import.meta",
		"dynamic-import", "json-modules", "source-phase-imports", "explicit-resource-management",
		"tail-call-optimization", "hashbang", "new.target", "with", "getters", "setters":
		return feature, true
	}
	if strings.HasPrefix(feature, "Symbol.") || strings.HasPrefix(feature, "Reflect.") {
		return feature, true
	}
	if strings.HasPrefix(feature, "async") {
		return feature, true
	}
	if strings.HasSuffix(feature, "Array") && (strings.HasPrefix(feature, "Float") || strings.HasPrefix(feature, "Int") || strings.HasPrefix(feature, "Uint") || strings.HasPrefix(feature, "Big")) {
		return feature, true
	}
	return "", false
}

var (
	syntaxEval    = regexp.MustCompile(`(?:^|[^\w$])eval\s*\(|(?:^|[^\w$])new\s+Function\b`)
	syntaxWith    = regexp.MustCompile(`(?:^|[^\w$])with\s*\(`)
	syntaxGetter  = regexp.MustCompile(`(?:^|[^\w$.])get\s+(?:[A-Za-z_$]|\[)`)
	syntaxSetter  = regexp.MustCompile(`(?:^|[^\w$.])set\s+(?:[A-Za-z_$]|\[)`)
	syntaxAsync   = regexp.MustCompile(`(?:^|[^\w$])async\s+(?:function|\()|(?:^|[^\w$])await\b`)
	syntaxYield   = regexp.MustCompile(`(?:^|[^\w$])yield\b`)
	syntaxProxy   = regexp.MustCompile(`(?:^|[^\w$])Proxy\b`)
	syntaxSymbol  = regexp.MustCompile(`(?:^|[^\w$])Symbol\b`)
	syntaxReflect = regexp.MustCompile(`(?:^|[^\w$])Reflect\b`)
	syntaxHost    = regexp.MustCompile(`(?:^|[^\w$])\$262\b`)
	syntaxBigInt  = regexp.MustCompile(`(?:^|[^\w$])BigInt\b`)
)

// syntaxSkip catches the same refusals when the test uses them without listing a feature. The scan
// sees code only, not comments or strings.
func syntaxSkip(code string) string {
	switch {
	case syntaxHost.MatchString(code):
		return "host $262"
	case syntaxEval.MatchString(code):
		return "eval"
	case syntaxWith.MatchString(code):
		return "with"
	case syntaxAsync.MatchString(code):
		return "async"
	case syntaxYield.MatchString(code):
		return "yield"
	case syntaxGetter.MatchString(code):
		return "getters"
	case syntaxSetter.MatchString(code):
		return "setters"
	case syntaxProxy.MatchString(code):
		return "Proxy"
	case syntaxSymbol.MatchString(code):
		return "Symbol"
	case syntaxReflect.MatchString(code):
		return "Reflect"
	case syntaxBigInt.MatchString(code):
		return "BigInt"
	default:
		return ""
	}
}
