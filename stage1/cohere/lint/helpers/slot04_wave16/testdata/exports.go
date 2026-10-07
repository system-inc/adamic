package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type AdamicCase struct {
	Name, Source, Helper, Argument, Hint, ModifierKind, Key, Prefix, Inferred, Parsed, Escaped, Prefixed string
	ModifierPresent, EntryPresent                                                                        bool
	Options                                                                                              int
}

func AdamicAdapt(c AdamicCase) AdamicCase {
	if c.Argument == "" {
		c.Argument = "[*]"
	}
	c.Inferred = string(InferDataType(c.Source, []DataType{DataType(c.Argument[1 : len(c.Argument)-1])}))
	c.Parsed = ValueToCss(ParseValue(c.Source))
	theme := NewTheme()
	theme.Prefix = c.Prefix
	c.Prefixed = theme.PrefixKey(c.Key)
	c.Escaped = escapeCSSIdentifier(c.Prefixed)
	return c
}
func AdamicExpand(c AdamicCase) []AdamicCase { return []AdamicCase{AdamicAdapt(c)} }
func AdamicObserve(c AdamicCase) {
	state := &utilityEvaluation{}
	if c.ModifierPresent {
		state.modifier = &ParsedModifier{Kind: ParsedModifierKind(c.ModifierKind), Value: c.Source}
	}
	m := state.modifierAsValue()
	if m == nil {
		fmt.Print("false|||||")
	} else {
		fmt.Printf("true|%s|%s|%s|%s|", m.Kind, m.Value, m.Fraction, m.DataType)
	}
	nodes, ok := state.resolveArbitraryArgument(c.Argument, &ParsedValue{Value: c.Source, DataType: c.Hint})
	fmt.Printf("%t|%s|", ok, ValueToCss(nodes))
	theme := NewTheme()
	theme.Prefix = c.Prefix
	if c.EntryPresent {
		theme.values[c.Key] = themeValue{value: c.Source, options: ThemeOptions(c.Options)}
	}
	v, ok := theme.variableReference(c.Key)
	fmt.Printf("%t|%s\n", ok, v)
}

var adamicLock sync.Mutex

func AdamicRecord(c AdamicCase) {
	path := os.Getenv("ADAMIC_SLOT04_INTEGER")
	if path == "" {
		return
	}
	adamicLock.Lock()
	defer adamicLock.Unlock()
	f, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	if e = json.NewEncoder(f).Encode(c); e != nil {
		panic(e)
	}
}
