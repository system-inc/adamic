package tailwind

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Slot02B6Variant struct {
	Params, Trimmed string
	Parts           []string
}
type Slot02B6Import struct {
	Params, Path, Directory, Resolved string
	Parts                             []string
	Trim, Quote                       map[string]string
	Error                             []string
}
type Slot02B6Step struct {
	Kind, Input, Value string
	Error              []string
}
type Slot02B6Load struct {
	Path, Absolute           string
	Visiting, InitialPresent bool
	Steps                    []Slot02B6Step
}
type Slot02B6Corpus struct {
	Variants []Slot02B6Variant
	Imports  []Slot02B6Import
	Loads    []Slot02B6Load
	Want     string
}

var Slot02B6Steps []Slot02B6Step
var Slot02B6Nodes []*Node

func Slot02B6Error(err error) []string {
	result := []string{}
	for err != nil {
		result = append(result, err.Error())
		err = errors.Unwrap(err)
	}
	return result
}
func AdamicSlot02Batch6Abs(path string) (string, error) {
	value, err := filepath.Abs(path)
	Slot02B6Steps = append(Slot02B6Steps, Slot02B6Step{Kind: "abs", Input: path, Value: value, Error: Slot02B6Error(err)})
	return value, err
}
func AdamicSlot02Batch6Read(path string) ([]byte, error) {
	value, err := os.ReadFile(path)
	Slot02B6Steps = append(Slot02B6Steps, Slot02B6Step{Kind: "read", Input: path, Value: string(value), Error: Slot02B6Error(err)})
	return value, err
}
func AdamicSlot02Batch6Parse(content string) ([]*Node, error) {
	value, err := ParseCSS(content)
	Slot02B6Nodes = value
	Slot02B6Steps = append(Slot02B6Steps, Slot02B6Step{Kind: "parse", Input: content, Value: "1", Error: Slot02B6Error(err)})
	return value, err
}
func AdamicSlot02Batch6Ingest(collector *stylesheetCollector, nodes []*Node, path string) error {
	same := len(nodes) == len(Slot02B6Nodes) && (len(nodes) == 0 || &nodes[0] == &Slot02B6Nodes[0])
	state := fmt.Sprintf("%t:%t:%t", collector.visiting[path], len(collector.stylesheets) > 0 && collector.stylesheets[len(collector.stylesheets)-1] == path, same)
	err := collector.ingest(nodes, path)
	Slot02B6Steps = append(Slot02B6Steps, Slot02B6Step{Kind: "ingest", Input: path, Value: state, Error: Slot02B6Error(err)})
	return err
}

var Slot02B6SegmentInput string
var Slot02B6SegmentSeparator byte
var Slot02B6SegmentCalls int

func AdamicSlot02Batch6Segment(input string, separator byte) []string {
	Slot02B6SegmentInput = input
	Slot02B6SegmentSeparator = separator
	Slot02B6SegmentCalls++
	return segment(input, separator)
}
func AdamicSlot02Batch6Observe(inputs []string) []byte {
	corpus := Slot02B6Corpus{}
	var want strings.Builder
	texts := map[string]bool{"preexisting": true, "": true, "-*": true, "foo-*-*": true, "dark (&:where(.dark *))": true, "foo[bar baz]-* selector": true, "foo\\ bar-* rest": true, "foo\tbar-* selector": true, "foo-*\n selector": true, "\ufeffdark-* ": true, "\u0085dark-*\u0085": true, "\u200bdark-* ": true}
	for _, text := range inputs {
		texts[text] = true
		for _, field := range strings.Fields(text) {
			texts[field] = true
		}
		if nodes, err := ParseCSS(text); err == nil {
			var walk func([]*Node)
			walk = func(nodes []*Node) {
				for _, n := range nodes {
					if n.Name == "@custom-variant" {
						texts[n.Params] = true
					}
					walk(n.Nodes)
				}
			}
			walk(nodes)
		}
	}
	// Every ASCII control and each Go/JS whitespace boundary, including non-members.
	for _, r := range []rune{0, 9, 10, 11, 12, 13, 32, 0x85, 0xa0, 0x1680, 0x180e, 0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200a, 0x200b, 0x2028, 0x2029, 0x202f, 0x205f, 0x3000, 0xfeff, 0x1f600} {
		texts[string(r)+"dark-*"+string(r)] = true
	}
	ordered := []string{}
	for text := range texts {
		ordered = append(ordered, text)
	}
	sort.Strings(ordered)
	for _, text := range ordered {
		trimmed := strings.TrimSpace(text)
		parts := segment(trimmed, ' ')
		if parts == nil {
			parts = []string{}
		}
		corpus.Variants = append(corpus.Variants, Slot02B6Variant{text, trimmed, parts})
		collector := &stylesheetCollector{customVariants: map[string]bool{"preexisting": false}}
		Slot02B6SegmentCalls = 0
		collector.ingestCustomVariant(&Node{Params: text})
		keys := []string{}
		for key := range collector.customVariants {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fmt.Fprintf(&want, "variant:%d:%d:%s:%d\n", len(keys), Slot02B6SegmentCalls, Slot02B6SegmentInput, Slot02B6SegmentSeparator)
		// The only added key is derived from the real resulting map, not guessed from params.
		for _, key := range keys {
			if key != "preexisting" {
				fmt.Fprintf(&want, "added:%s:%t\n", key, collector.customVariants[key])
			}
		}
		fmt.Fprintf(&want, "preexisting:%t\n", collector.customVariants["preexisting"])
	}

	root, err := os.MkdirTemp("", "slot02-b6-graphs-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(root)
	params := append([]string{"", `"foo.css" source(none)`, `"foo.css" theme(reference)`, `"" source(none)`, `'"foo"' source(none) source(never)`, `foo source(unfinished`, `foo source`, `foosource(none)`}, ordered...)
	for i, param := range params {
		path := "/repo/styles/theme.css"
		parts := segment(strings.TrimSpace(param), ' ')
		if parts == nil {
			parts = []string{}
		}
		c := Slot02B6Import{Params: param, Path: path, Directory: filepath.Dir(path), Parts: parts, Trim: map[string]string{}, Quote: map[string]string{}, Error: []string{}}
		c.Trim[param] = strings.TrimSpace(param)
		for _, part := range parts {
			trim := strings.TrimSpace(part)
			c.Trim[part] = trim
			c.Quote[trim] = strconv.Quote(trim)
		}
		spec := ""
		if len(parts) > 0 {
			spec = strings.Trim(parts[0], `"'`)
		}
		c.Quote[spec] = strconv.Quote(spec)
		c.Resolved = "/resolved/" + spec
		if i%3 == 1 {
			c.Error = []string{"fixture resolver unavailable"}
		}
		calls := 0
		requested, dir := "", ""
		collector := &stylesheetCollector{resolve: func(value, directory string) (string, error) {
			calls++
			requested = value
			dir = directory
			if len(c.Error) > 0 {
				return "", errors.New(c.Error[0])
			}
			return "/resolved/" + value, nil
		}}
		Slot02B6SegmentCalls = 0
		value, err := collector.resolveImport(&Node{Params: param}, path)
		corpus.Imports = append(corpus.Imports, c)
		fmt.Fprintf(&want, "import-segment:%d:%s:%d\n", Slot02B6SegmentCalls, Slot02B6SegmentInput, Slot02B6SegmentSeparator)
		fmt.Fprintf(&want, "resolve:%d:%s:%s\n", calls, requested, dir)
		fmt.Fprintf(&want, "resolved:%s\n", value)
		for _, message := range Slot02B6Error(err) {
			fmt.Fprintf(&want, "error:%s\n", message)
		}
		fmt.Fprintln(&want, "end-error")
	}
	contents := append([]string{"", "a { color:red; }", "@theme {color:red;}", "a {", "@config 'ignored.js';"}, inputs...)
	for i, content := range contents {
		path := filepath.Join(root, fmt.Sprintf("sheet-%d.css", i))
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			panic(err)
		}
		for mode := 0; mode < 4; mode++ {
			current := path
			if mode == 1 {
				current = path + "-missing"
			}
			collector := &stylesheetCollector{resolve: func(string, string) (string, error) { return "", errors.New("fixture import unavailable") }, theme: NewTheme(), visiting: map[string]bool{"unrelated": true}, customVariants: map[string]bool{}, utilityRoots: map[string]map[UtilityKind]bool{}, staticUtilityNodes: map[string][]*Node{}}
			if mode == 2 {
				collector.visiting[current] = true
			}
			if mode == 3 {
				collector.visiting[current] = false
			}
			Slot02B6Steps = []Slot02B6Step{}
			err := collector.loadFile(current)
			c := Slot02B6Load{Path: current, Absolute: current, Visiting: mode == 2, InitialPresent: mode == 2 || mode == 3, Steps: Slot02B6Steps}
			corpus.Loads = append(corpus.Loads, c)
			for _, step := range Slot02B6Steps {
				fmt.Fprintf(&want, "load-call:%s:%s\n", step.Kind, step.Input)
				if step.Kind == "ingest" {
					fmt.Fprintf(&want, "ingest-state:%s\n", step.Value)
				}
			}
			for _, message := range Slot02B6Error(err) {
				fmt.Fprintf(&want, "error:%s\n", message)
			}
			fmt.Fprintln(&want, "end-error")
			_, present := collector.visiting[current]
			fmt.Fprintf(&want, "state:%t:%t:%t:%d\n", collector.visiting[current], present, collector.visiting["unrelated"], len(collector.stylesheets))
			for _, sheet := range collector.stylesheets {
				fmt.Fprintf(&want, "stylesheet:%s\n", sheet)
			}
		}
	}
	// Real filepath.Abs failure: Linux getcwd fails inside this deleted owned directory.
	saved, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	removed, err := os.MkdirTemp(root, "removed-cwd-")
	if err != nil {
		panic(err)
	}
	if err = os.Chdir(removed); err != nil {
		panic(err)
	}
	if err = os.Remove(removed); err != nil {
		panic(err)
	}
	collector := &stylesheetCollector{visiting: map[string]bool{"unrelated": true}}
	Slot02B6Steps = []Slot02B6Step{}
	failure := collector.loadFile("relative.css")
	if err = os.Chdir(saved); err != nil {
		panic(err)
	}
	if failure == nil {
		panic("deleted cwd did not fail filepath.Abs")
	}
	corpus.Loads = append(corpus.Loads, Slot02B6Load{Path: "relative.css", Absolute: "relative.css", Steps: Slot02B6Steps})
	for _, step := range Slot02B6Steps {
		fmt.Fprintf(&want, "load-call:%s:%s\n", step.Kind, step.Input)
	}
	for _, message := range Slot02B6Error(failure) {
		fmt.Fprintf(&want, "error:%s\n", message)
	}
	fmt.Fprintln(&want, "end-error")
	fmt.Fprintln(&want, "state:false:false:true:0")
	corpus.Want = want.String()
	data, err := json.Marshal(corpus)
	if err != nil {
		panic(err)
	}
	return data
}
