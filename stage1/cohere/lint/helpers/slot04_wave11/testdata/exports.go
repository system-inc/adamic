package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type AdamicBytesCase struct {
	Name, Helper string
	Data         []int
	Prefixes     [][]int
}

func adamicBytes(s string) []int {
	out := []int{}
	for _, b := range []byte(s) {
		out = append(out, int(b))
	}
	return out
}
func adamicString(v []int) string {
	out := []byte{}
	for _, b := range v {
		out = append(out, byte(b))
	}
	return string(out)
}
func AdamicCase(value string, prefixes []string) AdamicBytesCase {
	c := AdamicBytesCase{Data: adamicBytes(value), Prefixes: [][]int{}}
	for _, p := range prefixes {
		c.Prefixes = append(c.Prefixes, adamicBytes(p))
	}
	return c
}
func AdamicObserve(c AdamicBytesCase) {
	value := adamicString(c.Data)
	prefixes := []string{}
	for _, p := range c.Prefixes {
		prefixes = append(prefixes, adamicString(p))
	}
	first, last := byte(0), byte(0)
	if len(value) > 0 {
		first = value[0]
		last = value[len(value)-1]
	}
	fmt.Printf("%t|%t|%t|%t\n", isDigit(first), isDigit(last), isBracketed(value), hasAnyPrefix(value, prefixes))
}

var adamicLock sync.Mutex

func AdamicRecord(helper, value string, prefixes []string) {
	path := os.Getenv("ADAMIC_SLOT04_BYTES")
	if path == "" {
		return
	}
	adamicLock.Lock()
	defer adamicLock.Unlock()
	c := AdamicCase(value, prefixes)
	c.Helper = helper
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(c); err != nil {
		panic(err)
	}
}
func AdamicFull() {
	raw := func(data []int) {
		for _, prefixes := range [][][]int{{}, {[]int{}}, {{91}}, {{91, 93}}, {{48}, {255}}} {
			AdamicObserve(AdamicBytesCase{Data: data, Prefixes: prefixes})
		}
	}
	raw([]int{})
	for a := 0; a < 256; a++ {
		raw([]int{a})
		for b := 0; b < 256; b++ {
			raw([]int{a, b})
		}
	}
}
