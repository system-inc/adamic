package regexp

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dlclark/regexp2/v2"
	"strings"
	"time"
	"unicode/utf8"
)

type AdamicAtom struct {
	Kind int    `json:"kind"`
	Lo   int    `json:"lo"`
	Hi   int    `json:"hi"`
	Text string `json:"text"`
}
type AdamicValue struct {
	Kind    int    `json:"kind"`
	Set     int    `json:"set"`
	Rune    int    `json:"rune"`
	Width   int    `json:"width"`
	Negated bool   `json:"negated"`
	Error   string `json:"error"`
}
type AdamicOp struct {
	Key      string       `json:"key"`
	Value    AdamicValue  `json:"value"`
	Atoms    []AdamicAtom `json:"atoms"`
	HasAtoms bool         `json:"hasAtoms"`
	Error    string       `json:"error"`
	Matched  bool         `json:"matched"`
	Failed   bool         `json:"failed"`
}
type AdamicSource struct {
	Bytes []int    `json:"bytes"`
	Runes [][2]int `json:"runes"`
}

var adamicFull string
var adamicTrace strings.Builder
var adamicOps map[string]AdamicOp
var adamicMuted int
var adamicEngine *regexp2.Regexp
var adamicEngineID int
var adamicForced bool

func adamicContext(ctx escapeContext) string {
	return fmt.Sprintf("%t:%t:%d:%t", ctx.unicode, ctx.inClass, ctx.groups, ctx.named)
}
func adamicOptions(o rewriteOptions) string {
	return fmt.Sprintf("%t:%t:%t:%t", o.unicode, o.ignoreCase, o.multiline, o.dotAll)
}
func adamicValue(v decodedEscape, e error) AdamicValue {
	s := ""
	if e != nil {
		s = e.Error()
	}
	return AdamicValue{int(v.kind), int(v.set), int(v.r), v.width, v.negated, s}
}
func adamicPublic(atoms []classAtom) []AdamicAtom {
	if atoms == nil {
		return nil
	}
	out := []AdamicAtom{}
	for _, a := range atoms {
		out = append(out, AdamicAtom{int(a.kind), int(a.lo), int(a.hi), hex.EncodeToString([]byte(a.text))})
	}
	return out
}
func adamicAtomText(atoms []classAtom) string {
	var s strings.Builder
	for _, a := range atoms {
		fmt.Fprintf(&s, "%d,%d,%d,%s;", a.kind, a.lo, a.hi, hex.EncodeToString([]byte(a.text)))
	}
	return s.String()
}
func adamicRecord(key string, op AdamicOp) {
	if adamicMuted > 0 {
		return
	}
	op.Key = key
	adamicOps[key] = op
	adamicTrace.WriteString(key + ";")
}
func adamicRune(s string) (rune, int) {
	r, w := utf8.DecodeRuneInString(s)
	if adamicMuted == 0 {
		adamicTrace.WriteString(fmt.Sprintf("rune:%d;", len(adamicFull)-len(s)))
	}
	return r, w
}
func adamicUnicode(s string, i, size int, ctx escapeContext) (decodedEscape, error) {
	v, e := decodeUnicodeEscape(s, i, size, ctx)
	adamicRecord(fmt.Sprintf("unicode:%d:%d:%s", i, size, adamicContext(ctx)), AdamicOp{Value: adamicValue(v, e)})
	return v, e
}
func adamicHex(s string, i, size, digits int, ctx escapeContext) (decodedEscape, error) {
	v, e := decodeFixedHex(s, i, size, digits, ctx)
	adamicRecord(fmt.Sprintf("hex:%d:%d:%d:%s", i, size, digits, adamicContext(ctx)), AdamicOp{Value: adamicValue(v, e)})
	return v, e
}
func adamicControl(s string, i, size int, ctx escapeContext) (decodedEscape, error) {
	v, e := decodeControlEscape(s, i, size, ctx)
	adamicRecord(fmt.Sprintf("control:%d:%d:%s", i, size, adamicContext(ctx)), AdamicOp{Value: adamicValue(v, e)})
	return v, e
}
func adamicProperty(s string, i, size int, ctx escapeContext) (decodedEscape, error) {
	v, e := decodePropertyEscape(s, i, size, ctx)
	adamicRecord(fmt.Sprintf("property:%d:%d:%s", i, size, adamicContext(ctx)), AdamicOp{Value: adamicValue(v, e)})
	return v, e
}
func adamicIdentity(r rune, size int, ctx escapeContext) (decodedEscape, error) {
	v, e := identityEscape(r, size, ctx)
	adamicRecord(fmt.Sprintf("identity:%d:%d:%s", r, size, adamicContext(ctx)), AdamicOp{Value: adamicValue(v, e)})
	return v, e
}
func adamicNumeric(s string, i int, ctx escapeContext) (decodedEscape, error) {
	v, e := decodeNumericEscape(s, i, ctx)
	adamicRecord(fmt.Sprintf("numeric:%d:%s", i, adamicContext(ctx)), AdamicOp{Value: adamicValue(v, e)})
	return v, e
}
func adamicEscape(s string, i int, ctx escapeContext) (decodedEscape, error) {
	adamicMuted++
	v, e := decodeEscape(s, i, ctx)
	adamicMuted--
	adamicRecord(fmt.Sprintf("escape:%d:%s", i, adamicContext(ctx)), AdamicOp{Value: adamicValue(v, e)})
	return v, e
}
func adamicWord(o rewriteOptions) []classAtom {
	a := wordClassAtoms(o)
	adamicRecord("word:"+adamicOptions(o), AdamicOp{Atoms: adamicPublic(a), HasAtoms: a != nil})
	return a
}
func adamicNonword(o rewriteOptions) []classAtom {
	a := nonWordClassAtoms(o)
	adamicRecord("nonword:"+adamicOptions(o), AdamicOp{Atoms: adamicPublic(a), HasAtoms: a != nil})
	return a
}
func adamicJoin(a []classAtom, o rewriteOptions) ([]classAtom, error) {
	v, e := joinRanges(a, o)
	message := ""
	if e != nil {
		message = e.Error()
	}
	adamicRecord("join:"+adamicOptions(o)+":"+adamicAtomText(a), AdamicOp{Atoms: adamicPublic(v), HasAtoms: v != nil, Error: message})
	return v, e
}
func adamicMatch(engine *regexp2.Regexp, subject string) (bool, error) {
	if engine != adamicEngine {
		panic("engine identity drift")
	}
	m, e := engine.MatchString(subject)
	if adamicForced {
		m = true
		e = errors.New("controlled partial match failure")
	}
	adamicRecord(fmt.Sprintf("match:%d:%s", adamicEngineID, hex.EncodeToString([]byte(subject))), AdamicOp{Matched: m, Failed: e != nil})
	return m, e
}
func AdamicSourceOf(s string) AdamicSource {
	v := AdamicSource{Bytes: []int{}, Runes: [][2]int{}}
	for i := 0; i < len(s); i++ {
		v.Bytes = append(v.Bytes, int(s[i]))
		r, w := utf8.DecodeRuneInString(s[i:])
		v.Runes = append(v.Runes, [2]int{int(r), w})
	}
	v.Runes = append(v.Runes, [2]int{int(utf8.RuneError), 0})
	return v
}
func AdamicObserve(mode, source string, i int, u, c, named bool, groups int, ignore bool) ([]AdamicOp, string) {
	adamicFull = source
	adamicTrace.Reset()
	adamicOps = map[string]AdamicOp{}
	ctx := escapeContext{unicode: u, inClass: c, named: named, groups: groups}
	var result string
	if mode == "decode" {
		v, e := decodeEscape(source, i, ctx)
		a := adamicValue(v, e)
		result = fmt.Sprintf("%d,%d,%d,%d,%t|%s", a.Kind, a.Set, a.Rune, a.Width, a.Negated, a.Error)
	} else {
		a, exact, e := classAtoms(source, rewriteOptions{unicode: u, ignoreCase: ignore}, ctx)
		message := ""
		if e != nil {
			message = e.Error()
		}
		result = fmt.Sprintf("%t|%t|%s|%s", a == nil, exact, adamicAtomText(a), message)
	}
	ops := []AdamicOp{}
	for _, op := range adamicOps {
		ops = append(ops, op)
	}
	return ops, result + "|" + adamicTrace.String()
}
func AdamicTest(pattern, subject string, id, state int, forced, timeout bool) ([]AdamicOp, string, bool) {
	adamicTrace.Reset()
	adamicOps = map[string]AdamicOp{}
	adamicEngineID = id
	adamicForced = forced
	var r *RegExp
	if state == 1 {
		r = &RegExp{}
	} else if state == 2 {
		adamicMuted++
		compiled, e := Compile(pattern, "")
		adamicMuted--
		if e != nil {
			return nil, "", false
		}
		r = compiled
		adamicEngine = r.re
		if timeout {
			r.re.MatchTimeout = time.Nanosecond
		}
	}
	value := r.Test(subject)
	ops := []AdamicOp{}
	for _, op := range adamicOps {
		ops = append(ops, op)
	}
	return ops, fmt.Sprintf("%t|%s", value, adamicTrace.String()), true
}
