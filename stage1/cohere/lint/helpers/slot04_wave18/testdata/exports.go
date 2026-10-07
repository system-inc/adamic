package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

type AdamicPart struct {
	Source             string
	Length, Percentage bool
}
type AdamicGroup struct {
	Source string
	Parts  []AdamicPart
}
type AdamicEntry struct {
	Key, Value, Reference string
	Options               int
	ReferenceOK           bool
}
type AdamicCase struct {
	Name, Source, Helper, Argument, Candidate, Prefix, BaseKey, ArgKey, Resolved string
	Keys, Nested, ArgKeys                                                        []string
	Entries                                                                      []AdamicEntry
	Groups                                                                       []AdamicGroup
	BaseOK, ArgOK, ResolvedOK                                                    bool
}

func adamicTheme(c AdamicCase) *Theme {
	t := NewTheme()
	t.Prefix = c.Prefix
	for _, e := range c.Entries {
		t.values[e.Key] = themeValue{value: e.Value, options: ThemeOptions(e.Options)}
	}
	return t
}
func adamicEntries(t *Theme) []AdamicEntry {
	keys := []string{}
	for k := range t.values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rows := []AdamicEntry{}
	for _, k := range keys {
		v := t.values[k]
		ref, ok := t.variableReference(k)
		rows = append(rows, AdamicEntry{Key: k, Value: v.value, Options: int(v.options), Reference: ref, ReferenceOK: ok})
	}
	return rows
}
func AdamicAdapt(c AdamicCase) AdamicCase {
	if c.Keys == nil {
		c.Keys = []string{}
	}
	if c.Nested == nil {
		c.Nested = []string{}
	}
	t := adamicTheme(c)
	c.BaseKey, c.BaseOK = t.resolveKey(c.Candidate, true, c.Keys)
	c.ArgKeys = []string{}
	if strings.HasSuffix(c.Argument, "-*") {
		c.ArgKeys = []string{strings.TrimSuffix(c.Argument, "-*")}
	} else {
		parts := strings.Split(c.Argument, "-*")
		if len(parts) > 1 {
			c.ArgKeys = []string{parts[0]}
		}
	}
	c.ArgKey, c.ArgOK = t.resolveKey(c.Candidate, true, c.ArgKeys)
	c.Resolved, c.ResolvedOK = t.Resolve(c.Candidate, true, c.ArgKeys, ThemeOptionNone)
	c.Entries = adamicEntries(t)
	c.Groups = []AdamicGroup{}
	for _, g := range segment(c.Source, ',') {
		parts := []AdamicPart{}
		for _, s := range segment(g, ' ') {
			parts = append(parts, AdamicPart{Source: s, Length: IsLength(s), Percentage: isPercentage(s)})
		}
		c.Groups = append(c.Groups, AdamicGroup{Source: g, Parts: parts})
	}
	return c
}
func AdamicExpand(c AdamicCase) []AdamicCase { return []AdamicCase{AdamicAdapt(c)} }
func AdamicObserve(c AdamicCase) {
	t := adamicTheme(c)
	v, extra, ok := t.ResolveWith(c.Candidate, c.Keys, c.Nested)
	fmt.Printf("%t|%t|%t|%s|", isBackgroundSize(c.Source), ok, extra != nil, v)
	for _, name := range c.Nested {
		value, present := extra[name]
		fmt.Printf("%t:%s;", present, value)
	}
	state := &utilityEvaluation{evaluator: &UtilityEvaluator{Theme: t}}
	r, yes := state.resolveThemeArgument(c.Argument, c.Candidate)
	fmt.Printf("|%t|%s\n", yes, r)
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
func AdamicRecordTheme(c AdamicCase, t *Theme) {
	if os.Getenv("ADAMIC_SLOT04_INTEGER") == "" {
		return
	}
	c.Prefix = t.Prefix
	c.Entries = adamicEntries(t)
	AdamicRecord(c)
}
