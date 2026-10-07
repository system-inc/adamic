package tailwind

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	engine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
	"sync"
)

type AdamicCase struct {
	Name, Source, Value, Helper, Argument, Candidate, Fraction string
	Start, End                                                 int
	RangePresent, ValuePresent, BytesPresent                   bool
	SourceBytes, ValueBytes, Spaces                            []int
	Bare                                                       engine.AdamicBare
}

func adamicBytes(s string) []int {
	a := []int{}
	for _, b := range []byte(s) {
		a = append(a, int(b))
	}
	return a
}
func adamicString(a []int) string {
	b := make([]byte, len(a))
	for i, v := range a {
		b[i] = byte(v)
	}
	return string(b)
}
func AdamicAdapt(c AdamicCase) AdamicCase {
	if !c.BytesPresent {
		c.SourceBytes = adamicBytes(c.Source)
		value := c.Value
		if !c.ValuePresent {
			value = c.Source
		}
		c.ValueBytes = adamicBytes(value)
	}
	if c.SourceBytes == nil {
		c.SourceBytes = []int{}
	}
	if c.ValueBytes == nil {
		c.ValueBytes = []int{}
	}
	if !c.RangePresent {
		c.End = c.Start + len(c.ValueBytes)
	}
	c.Spaces = []int{}
	for b := 0; b < 256; b++ {
		if isSpace(rune(b)) {
			c.Spaces = append(c.Spaces, b)
		}
	}
	c.Bare = engine.AdamicBareAdapt(c.Argument, c.Candidate, c.Fraction)
	return c
}
func AdamicExpand(c AdamicCase) []AdamicCase { return []AdamicCase{AdamicAdapt(c)} }
func adamicPrint(label string, tokens []classToken, ok bool) {
	fmt.Printf("%s|%t|%t|%d\n", label, ok, tokens != nil, len(tokens))
	for _, t := range tokens {
		fmt.Printf("%t|%d|%d|", t.Separator, t.Range.Pos(), t.Range.End())
		for _, b := range []byte(t.Text) {
			fmt.Printf("%d,", b)
		}
		fmt.Println()
	}
}
func AdamicObserve(c AdamicCase) {
	value := adamicString(c.ValueBytes)
	source := adamicString(c.SourceBytes)
	adamicPrint("of", classTokensOf(value, c.Start), true)
	tokens, ok := classTokensIn(source, core.NewTextRange(c.Start, c.End), value)
	adamicPrint("in", tokens, ok)
	engine.AdamicBareObserve(c.Bare)
}

var adamicLock sync.Mutex

func AdamicRecord(c AdamicCase) {
	path := os.Getenv("ADAMIC_SLOT04_TOKENS")
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
