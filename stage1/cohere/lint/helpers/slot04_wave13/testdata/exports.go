package regexp

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

type AdamicCase struct {
	Name, Source, Word                              string
	Data                                            []int
	Width, Kind                                     int
	Unicode, Negated, IgnoreCase, Multiline, DotAll bool
}

func adamicOptions(c AdamicCase) rewriteOptions {
	return rewriteOptions{unicode: c.Unicode, ignoreCase: c.IgnoreCase, multiline: c.Multiline, dotAll: c.DotAll}
}
func adamicInts(s string) []int {
	out := []int{}
	for _, b := range []byte(s) {
		out = append(out, int(b))
	}
	return out
}
func AdamicAdapt(c AdamicCase) AdamicCase {
	if c.Data != nil {
		raw := []byte{}
		for _, b := range c.Data {
			raw = append(raw, byte(b))
		}
		c.Source = string(raw)
	}
	c.Data = adamicInts(c.Source)
	c.Width = quantifierWidth(c.Source)

	return c
}
func AdamicExpand(c AdamicCase) []AdamicCase {
	out := []AdamicCase{}
	for k := 0; k < 3; k++ {
		for flags := 0; flags < 16; flags++ {
			for _, n := range []bool{false, true} {
				v := c
				v.Kind = k
				v.Unicode = flags&1 != 0
				v.IgnoreCase = flags&2 != 0
				v.Multiline = flags&4 != 0
				v.DotAll = flags&8 != 0
				v.Negated = n
				out = append(out, AdamicAdapt(v))
			}
		}
	}
	return out
}
func adamicError(err error) string {
	if err == nil {
		return ""
	}
	parts := []string{}
	for _, b := range []byte(err.Error()) {
		parts = append(parts, fmt.Sprint(b))
	}
	return strings.Join(parts, ",")
}
func AdamicObserve(c AdamicCase) {
	fmt.Println(adamicError(errNothingToRepeat(c.Source)))
	fmt.Println(adamicError(checkGroupQuantifier(groupKind(c.Kind), c.Source, adamicOptions(c))))
	fmt.Println(adamicError(checkGroupConstruct(c.Source)))
}

var adamicLock sync.Mutex

func AdamicRecord(c AdamicCase) {
	path := os.Getenv("ADAMIC_SLOT04_ASSERTIONS")
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
