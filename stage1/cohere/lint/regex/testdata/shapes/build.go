// Generate bounded per-site fixtures from the approved table, using Go regexp as oracle.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"regexp/syntax"
	"strings"
	"unicode/utf16"
)

type row struct {
	ID, Rule, File, Expression, Feature string
	Line                                int
	Go                                  *string `json:"go_pattern"`
	JS                                  string  `json:"js_pattern"`
	Flags                               string  `json:"js_flags"`
}
type span struct {
	Start int    `json:"start"`
	End   int    `json:"end"`
	Text  string `json:"text"`
}
type sample struct {
	Input string `json:"input"`
	Go    []span `json:"go_matches"`
	Node  []span `json:"node_matches"`
}
type fixture struct {
	ID         string   `json:"id"`
	Rule       string   `json:"rule"`
	Shape      string   `json:"shape"`
	Expression string   `json:"expression"`
	Go         string   `json:"go_pattern"`
	Pattern    string   `json:"pattern"`
	Flags      string   `json:"flags"`
	Binding    string   `json:"binding"`
	Outside    string   `json:"outside"`
	Samples    []sample `json:"samples"`
}

func witness(r *syntax.Regexp) string {
	switch r.Op {
	case syntax.OpLiteral:
		return string(r.Rune)
	case syntax.OpCharClass:
		if r.Rune[0] <= 65 && r.Rune[len(r.Rune)-1] >= 65 {
			for i := 0; i < len(r.Rune); i += 2 {
				if r.Rune[i] <= 65 && r.Rune[i+1] >= 65 {
					return "A"
				}
			}
		}
		return string(r.Rune[0])
	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		return "a"
	case syntax.OpConcat:
		s := ""
		for _, c := range r.Sub {
			s += witness(c)
		}
		return s
	case syntax.OpAlternate, syntax.OpCapture, syntax.OpPlus:
		return witness(r.Sub[0])
	case syntax.OpRepeat:
		return strings.Repeat(witness(r.Sub[0]), r.Min)
	}
	return ""
}
func translated(p string) (string, string) {
	flags := "gu"
	if strings.HasPrefix(p, "(?i)") {
		flags += "i"
		p = p[4:]
	}
	if strings.HasPrefix(p, "(?s)") {
		p = strings.ReplaceAll(p[4:], ".", "[\\s\\S]")
	}
	p = strings.ReplaceAll(p, `[\s`, `[ \t\n\f\r`)
	p = strings.ReplaceAll(p, `\s`, `[ \t\n\f\r]`)
	p = strings.ReplaceAll(p, "$", `(?![\s\S])`)
	return p, flags
}
func dynamic(id string) (string, string) {
	switch id {
	case "core/default_case.go:99":
		return "^no default$", "commentPattern=^no default$"
	case "core/id_length.go:93":
		return "^_", "exceptionPatterns=[^_]"
	case "core/id_match.go:111":
		return "^[a-z]+$", "pattern=^[a-z]+$"
	case "core/no_fallthrough.go:169":
		return "(?i)falls? through", "pattern=falls? through"
	case "core/no_inline_comments.go:102":
		return "TODO", "ignorePattern=TODO"
	case "core/no_param_reassign.go:553":
		return "^ignore", "ignorePropertyModificationsForRegex=[^ignore]"
	case "core/no_unused_vars.go:282":
		return "^_", "destructuredArrayIgnorePattern=^_"
	case "core/no_unused_vars.go:2276":
		return "^_", "vars/args/caughtErrorsIgnorePattern=^_"
	case "core/no_warning_comments.go:354":
		return `(?i)^[\s*]*todo\b`, "term=todo; location=start; decoration=[*]"
	case "core/object_shorthand.go:140":
		return "^ignore", "methodsIgnorePattern=^ignore"
	case "core/require_description.go:233":
		return `^(custom)(?:[\t\n\x0B\f\r \x{00A0}\x{1680}\x{2000}-\x{200A}\x{2028}\x{2029}\x{202F}\x{205F}\x{3000}\x{FEFF}]|$)`, "additionalDirectives=[custom]"
	case "nexus/abbreviation_vocabulary.go:162":
		return "^cfg[A-Z]", "abbreviation=cfg"
	case "nexus/abbreviation_vocabulary.go:176":
		return "(^|[^a-zA-Z])Cfg($|[A-Z0-9])", "word=Cfg"
	case "nexus/abbreviation_vocabulary.go:177":
		return "[a-z0-9]Cfg($|[A-Z0-9])", "word=Cfg"
	case "nexus/abbreviation_vocabulary.go:178":
		return "Cfg($|[A-Z0-9])", "word=Cfg"
	case "nexus/consistency_no_return_void.go:116":
		return `^return\s+void\s+value\s*;$`, "operandText=value"
	case "react/boolean_prop_naming.go:93", "react/boolean_prop_naming.go:182":
		return "^(is|has)[A-Z]", "rule=^(is|has)[A-Z]"
	case "react/exhaustive_deps.go:463":
		return "^useCustom", "additionalHooks=^useCustom"
	case "react/no_unstable_nested_components.go:893":
		return "^render.*$", "propNamePattern=render*"
	case "react/sort_comp.go:589":
		return "(?i)^render$", "regex descriptor=/^render$/i"
	case "tailwind/class_literals.go:167":
		return "^classes$", "variablePatterns=[^classes$]"
	case "tailwind/enforce_canonical_classes.go:345":
		return "^custom-", "ignorePatterns=[^custom-]"
	case "typescript/no_require_imports.go:378":
		return "^node:", "allow=[^node:]"
	case "typescript/switch_exhaustiveness_check.go:212":
		return "^no default$", "defaultCaseCommentPattern=^no default$"
	}
	panic(id)
}
func main() {
	root := os.Args[1]
	data, e := os.ReadFile(filepath.Join(root, "../../table.json"))
	if e != nil {
		panic(e)
	}
	var rows []row
	if e = json.Unmarshal(data, &rows); e != nil {
		panic(e)
	}
	var corpus []string
	data, e = os.ReadFile(filepath.Join(root, "../corpus.json"))
	if e != nil {
		panic(e)
	}
	if e = json.Unmarshal(data, &corpus); e != nil {
		panic(e)
	}
	fixtures := []fixture{}
	for _, r := range rows {
		f := fixture{ID: r.ID, Rule: r.Rule, Shape: r.Feature, Expression: r.Expression, Pattern: r.JS, Flags: r.Flags}
		if r.Go != nil {
			f.Go = *r.Go
		} else {
			f.Go, f.Binding = dynamic(r.ID)
			f.Pattern, f.Flags = translated(f.Go)
			f.Pattern = regexp.MustCompile(`\\x\{([0-9A-Fa-f]+)\}`).ReplaceAllString(f.Pattern, `\u{$1}`)
		}
		re := regexp.MustCompile(f.Go)
		tree, e := syntax.Parse(f.Go, syntax.Perl)
		if e != nil {
			panic(e)
		}
		w := witness(tree)
		texts := []string{"", "🌍", "TODO\n", "\u00a0TODO", w, "x" + w, w + "\n"}
		positive := false
		for _, s := range texts {
			positive = positive || re.MatchString(s)
		}
		added := 0
		for _, s := range corpus {
			if re.MatchString(s) {
				texts = append(texts, s)
				positive = true
				added++
				if added == 3 {
					break
				}
			}
		}
		if !positive {
			panic("no positive input: " + r.ID)
		}
		seen := map[string]bool{}
		for _, text := range texts {
			if seen[text] {
				continue
			}
			seen[text] = true
			matches := []span{}
			for _, v := range re.FindAllStringIndex(text, -1) {
				matches = append(matches, span{len(utf16.Encode([]rune(text[:v[0]]))), len(utf16.Encode([]rune(text[:v[1]]))), text[v[0]:v[1]]})
			}
			f.Samples = append(f.Samples, sample{Input: text, Go: matches, Node: []span{}})
		}
		fixtures = append(fixtures, f)
	}
	out, _ := json.MarshalIndent(fixtures, "", "  ")
	if e = os.WriteFile(filepath.Join(root, "fixtures.json"), append(out, '\n'), 0644); e != nil {
		panic(e)
	}
	fmt.Println("generated", len(fixtures), "per-site fixtures")
}
