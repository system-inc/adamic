package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type AdamicModifier struct {
	Present bool   `json:"present"`
	Kind    string `json:"kind"`
	Value   string `json:"value"`
}
type AdamicRoot struct {
	Root    string `json:"root"`
	Value   string `json:"value"`
	Present bool   `json:"present"`
}
type AdamicPrediction struct {
	Op       string         `json:"op"`
	Input    string         `json:"input"`
	Extra    string         `json:"extra"`
	Text     string         `json:"text"`
	Bool     bool           `json:"bool"`
	Parts    []string       `json:"parts"`
	Roots    []AdamicRoot   `json:"roots"`
	Modifier AdamicModifier `json:"modifier"`
}

var adamicPredictions []AdamicPrediction

func adamicSegment(s string, c byte) []string {
	v := segment(s, c)
	adamicPredictions = append(adamicPredictions, AdamicPrediction{Op: "segment", Input: s, Extra: string(c), Parts: v})
	return v
}
func adamicDecode(s string) string {
	v := decodeArbitraryValue(s)
	adamicPredictions = append(adamicPredictions, AdamicPrediction{Op: "decode", Input: s, Text: v})
	return v
}
func adamicValidArbitrary(s string) bool {
	v := isValidArbitrary(s)
	adamicPredictions = append(adamicPredictions, AdamicPrediction{Op: "arbitrary", Input: s, Bool: v})
	return v
}
func adamicBlank(s string) bool {
	v := isBlank(s)
	adamicPredictions = append(adamicPredictions, AdamicPrediction{Op: "blank", Input: s, Bool: v})
	return v
}
func adamicValidNamed(s string) bool {
	v := isValidNamedValue(s)
	adamicPredictions = append(adamicPredictions, AdamicPrediction{Op: "named", Input: s, Bool: v})
	return v
}
func adamicModifier(s string) *ParsedModifier {
	v := parseModifier(s)
	a := AdamicModifier{}
	if v != nil {
		a = AdamicModifier{true, string(v.Kind), v.Value}
	}
	adamicPredictions = append(adamicPredictions, AdamicPrediction{Op: "modifier", Input: s, Modifier: a})
	return v
}
func adamicFindRoots(variant bool, s string, exists func(string) bool) []rootCandidate {
	v := findRoots(s, exists)
	a := []AdamicRoot{}
	for _, r := range v {
		a = append(a, AdamicRoot{r.root, r.value, r.hasValue})
	}
	adamicPredictions = append(adamicPredictions, AdamicPrediction{Op: "roots", Input: s, Extra: fmt.Sprint(variant), Roots: a})
	return v
}

type adamicSystem struct {
	system *LoadedDesignSystem
	prefix string
}

func (s adamicSystem) Prefix() string { return s.prefix }
func (s adamicSystem) HasUtility(root string, kind UtilityKind) bool {
	v := s.system.HasUtility(root, kind)
	adamicPredictions = append(adamicPredictions, AdamicPrediction{Op: "utility", Input: root, Extra: string(kind), Bool: v})
	return v
}
func (s adamicSystem) HasVariant(root string) bool { return s.system.HasVariant(root) }
func (s adamicSystem) VariantKind(root string) ParsedVariantKind {
	v := s.system.VariantKind(root)
	adamicPredictions = append(adamicPredictions, AdamicPrediction{Op: "kind", Input: root, Text: string(v)})
	return v
}
func (s adamicSystem) VariantCompoundsWith(root string, child ParsedVariant) bool {
	v := s.system.VariantCompoundsWith(root, child)
	adamicPredictions = append(adamicPredictions, AdamicPrediction{Op: "compounds", Input: root, Extra: string(child.Kind) + "|" + child.Root + "|" + child.Selector, Bool: v})
	if strings.HasPrefix(child.Selector, "&:is(") && strings.HasSuffix(child.Selector, ")") {
		changed := child
		changed.Selector = child.Selector[5 : len(child.Selector)-1]
		truth := s.system.VariantCompoundsWith(root, changed)
		adamicPredictions = append(adamicPredictions, AdamicPrediction{Op: "compounds", Input: root, Extra: string(changed.Kind) + "|" + changed.Root + "|" + changed.Selector, Bool: truth})
	}
	return v
}

type AdamicParserCase struct {
	Mode        string             `json:"mode"`
	Input       string             `json:"input"`
	Prefix      string             `json:"prefix"`
	Predictions []AdamicPrediction `json:"predictions"`
}

func AdamicParse(sample *AdamicParserCase, system *LoadedDesignSystem) string {
	adamicPredictions = []AdamicPrediction{}
	ds := adamicSystem{system, sample.Prefix}
	var out strings.Builder
	p := func(v any) { fmt.Fprintln(&out, v) }
	var printVariant func(*ParsedVariant)
	printMod := func(m *ParsedModifier) {
		p(m != nil)
		if m != nil {
			p(m.Kind)
			p(m.Value)
		}
	}
	printVariant = func(v *ParsedVariant) {
		p(v != nil)
		if v == nil {
			return
		}
		p(v.Kind)
		p(v.Root)
		p(v.Selector)
		p(v.Relative)
		p(v.Value != nil)
		if v.Value != nil {
			p(v.Value.Kind)
			p(v.Value.Value)
		}
		printMod(v.Modifier)
		printVariant(v.Variant)
	}
	if sample.Mode == "variant" {
		printVariant(ParseVariant(sample.Input, ds))
	} else {
		values := ParseCandidate(sample.Input, ds)
		p(len(values))
		for _, c := range values {
			p(c.Kind)
			p(c.Root)
			p(c.Property)
			p(c.PropertyValue)
			p(c.Value != nil)
			if c.Value != nil {
				p(c.Value.Kind)
				p(c.Value.Value)
				p(c.Value.DataType)
				p(c.Value.Fraction)
			}
			printMod(c.Modifier)
			p(len(c.Variants))
			for i := range c.Variants {
				printVariant(&c.Variants[i])
			}
			p(c.Important)
			p(c.Raw)
		}
	}
	sample.Predictions = adamicPredictions
	return out.String()
}
func AdamicParserSystem(path string) *LoadedDesignSystem {
	if e := os.WriteFile(path, []byte("@theme { --breakpoint-tablet: 48rem; }\n@utility dual { color: red; }\n@utility dual-* { color: --value(integer); }\n@custom-variant custom (&:hover);"), 0644); e != nil {
		panic(e)
	}
	s, e := LoadDesignSystem(LoadOptions{EntryPoint: path, Resolve: func(specifier, from string) (string, error) { return "", fmt.Errorf("unexpected import") }})
	if e != nil {
		panic(e)
	}
	return s
}

var _ = json.Marshal

func AdamicVariantInputs(input string) []string { return segment(input, ':') }
func AdamicVariantControls() []string {
	values := []string{}
	for _, r := range FrameworkVariantRegistrations {
		for _, suffix := range []string{"", "-red", "-[foo]", "-(--var)", "/name", "-hover", "-group-hover/name"} {
			values = append(values, r.Name+suffix)
		}
	}
	return values
}
