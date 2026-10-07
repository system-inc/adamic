package regexp

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

type AdamicReply struct {
	Text    string `json:"text"`
	Error   string `json:"error"`
	Number  int    `json:"number"`
	Width   int    `json:"width"`
	Kind    int    `json:"kind"`
	Set     int    `json:"set"`
	Options int    `json:"options"`
	Yes     bool   `json:"yes"`
	Negated bool   `json:"negated"`
}
type AdamicOp struct {
	Key   string      `json:"key"`
	Reply AdamicReply `json:"reply"`
}

var adamicOps map[string]AdamicOp
var adamicTrace strings.Builder
var adamicKeyInputs = map[string]string{}

func adamicKey(s string) string {
	v := adamicHex(s)
	a, b := 0, 0
	for _, r := range v {
		a = (a*131 + int(r)) % 65521
		b = (b*137 + int(r)) % 65519
	}
	key := fmt.Sprintf("%d,%d,%d", len(v), a, b)
	if prior, ok := adamicKeyInputs[key]; ok && prior != v {
		panic("dependency key collision")
	}
	adamicKeyInputs[key] = v
	return key
}
func adamicHex(s string) string { return hex.EncodeToString([]byte(s)) }
func adamicError(e error) string {
	if e != nil {
		return adamicHex(e.Error())
	}
	return ""
}
func adamicBits(o rewriteOptions) int {
	b := 0
	if o.unicode {
		b |= 1
	}
	if o.ignoreCase {
		b |= 2
	}
	if o.multiline {
		b |= 4
	}
	if o.dotAll {
		b |= 8
	}
	return b
}
func adamicOptions(b int) rewriteOptions {
	return rewriteOptions{unicode: b&1 != 0, ignoreCase: b&2 != 0, multiline: b&4 != 0, dotAll: b&8 != 0}
}
func adamicCtx(c escapeContext) int {
	b := c.groups * 2
	if c.named {
		b++
	}
	return b
}
func adamicRecord(name, s string, a, b, c int, v AdamicReply) {
	key := fmt.Sprintf("%s:%s:%d:%d:%d", name, adamicKey(s), a, b, c)
	adamicOps[key] = AdamicOp{key, v}
	adamicTrace.WriteString(key + ";")
}
func adamicCount(s string) (int, bool) {
	n, b := countGroups(s)
	adamicRecord("count", s, 0, 0, 0, AdamicReply{Number: n, Yes: b})
	return n, b
}
func adamicRune(s string) (rune, int) {
	r, w := utf8.DecodeRuneInString(s)
	adamicRecord("rune", s, 0, 0, 0, AdamicReply{Number: int(r), Width: w, Text: adamicHex(string(r))})
	return r, w
}
func adamicBack(s string) (string, int, bool) {
	v, w, b := namedBackreference(s)
	adamicRecord("back", s, 0, 0, 0, AdamicReply{Text: adamicHex(v), Width: w, Yes: b})
	return v, w, b
}
func adamicOpener(s string) (string, int, bool) {
	v, w, b := namedGroupOpener(s)
	adamicRecord("opener", s, 0, 0, 0, AdamicReply{Text: adamicHex(v), Width: w, Yes: b})
	return v, w, b
}
func adamicEscape(s string, i int, c escapeContext) (decodedEscape, error) {
	v, e := decodeEscape(s, i, c)
	u := 0
	if c.unicode {
		u = 1
	}
	adamicRecord("escape", s, i, u, adamicCtx(c), AdamicReply{Number: int(v.r), Kind: int(v.kind), Set: int(v.set), Width: v.width, Negated: v.negated, Error: adamicError(e)})
	return v, e
}
func adamicCase(r rune, u bool) (string, bool) {
	v, b := CaseClass(r, u)
	ub := 0
	if u {
		ub = 1
	}
	adamicRecord("case", "", int(r), ub, 0, AdamicReply{Text: adamicHex(v), Yes: b})
	return v, b
}
func adamicLiteral(r rune) string {
	v := literalRune(r)
	adamicRecord("literal", "", int(r), 0, 0, AdamicReply{Text: adamicHex(v)})
	return v
}
func adamicQuantifier(s string) int {
	v := quantifierWidth(s)
	adamicRecord("quantifier", s, 0, 0, 0, AdamicReply{Number: v})
	return v
}
func adamicRepeat(s string) error {
	e := errNothingToRepeat(s)
	adamicRecord("repeat", s, 0, 0, 0, AdamicReply{Error: adamicError(e)})
	return e
}
func adamicBoundary(n bool, o rewriteOptions) string {
	v := wordBoundary(n, o)
	b := 0
	if n {
		b = 1
	}
	adamicRecord("boundary", "", b, adamicBits(o), 0, AdamicReply{Text: adamicHex(v)})
	return v
}
func adamicAtomText(a []classAtom) string {
	var s strings.Builder
	for _, x := range a {
		fmt.Fprintf(&s, "%d,%d,%d,%s;", x.kind, x.lo, x.hi, adamicHex(x.text))
	}
	return s.String()
}
func adamicWord(o rewriteOptions) []classAtom {
	v := wordClassAtoms(o)
	adamicRecord("word", "", adamicBits(o), 0, 0, AdamicReply{Text: adamicHex(adamicAtomText(v))})
	return v
}
func adamicWrite(a []classAtom, n bool, o rewriteOptions) string {
	v := writeClass(a, n, o)
	b := 0
	if n {
		b = 1
	}
	adamicRecord("write", adamicAtomText(a), b, adamicBits(o), 0, AdamicReply{Text: adamicHex(v)})
	return v
}
func adamicClass(s string) (string, bool, int, error) {
	v, n, w, e := readClass(s)
	adamicRecord("class", s, 0, 0, 0, AdamicReply{Text: adamicHex(v), Negated: n, Width: w, Error: adamicError(e)})
	return v, n, w, e
}
func adamicAtoms(s string, o rewriteOptions, c escapeContext) ([]classAtom, bool, error) {
	v, b, e := classAtoms(s, o, c)
	u := 0
	if c.unicode {
		u = 1
	}
	adamicRecord("atoms", s, adamicBits(o), u, adamicCtx(c), AdamicReply{Text: adamicHex(adamicAtomText(v)), Yes: b, Error: adamicError(e)})
	return v, b, e
}
func adamicModifier(s string, o rewriteOptions) (rewriteOptions, int, bool) {
	v, w, b := modifierGroup(s, o)
	adamicRecord("modifier", s, adamicBits(o), 0, 0, AdamicReply{Options: adamicBits(v), Width: w, Yes: b})
	return v, w, b
}
func adamicConstruct(s string) error {
	e := checkGroupConstruct(s)
	adamicRecord("construct", s, 0, 0, 0, AdamicReply{Error: adamicError(e)})
	return e
}
func adamicGroup(s string) groupKind {
	v := groupKindOf(s)
	adamicRecord("group", s, 0, 0, 0, AdamicReply{Kind: int(v)})
	return v
}
func adamicGroupQuantifier(k groupKind, s string, o rewriteOptions) error {
	e := checkGroupQuantifier(k, s, o)
	adamicRecord("groupQuantifier", s, int(k), adamicBits(o), 0, AdamicReply{Error: adamicError(e)})
	return e
}
func AdamicRewrite(s string, b int) ([]AdamicOp, string) {
	adamicOps = map[string]AdamicOp{}
	adamicTrace.Reset()
	v, x, e := rewrite(s, adamicOptions(b))
	ops := []AdamicOp{}
	for _, op := range adamicOps {
		ops = append(ops, op)
	}
	sort.Slice(ops, func(i, j int) bool { return ops[i].Key < ops[j].Key })
	return ops, fmt.Sprintf("%s|%t|%s|%s", adamicHex(v), x, adamicError(e), adamicTrace.String())
}
