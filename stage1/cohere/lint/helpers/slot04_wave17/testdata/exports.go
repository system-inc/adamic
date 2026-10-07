package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

type AdamicCase struct {
	Name, Source, Helper, Lowered, TrimLowered, NumberTail string
	NumberPresent, Math                                    bool
}

func AdamicAdapt(c AdamicCase) AdamicCase {
	c.Lowered = strings.ToLower(c.Source)
	c.TrimLowered = strings.ToLower(strings.TrimFunc(c.Source, isJavaScriptSpace))
	n := scanNumber(c.Source)
	c.NumberPresent = n > 0
	c.NumberTail = c.Source[n:]
	c.Math = hasMathFunction(c.Source)
	return c
}
func AdamicExpand(c AdamicCase) []AdamicCase {
	if c.Name == "all-named" {
		keys := make([]string, 0, len(namedColors))
		for key := range namedColors {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		rows := []AdamicCase{}
		for _, key := range keys {
			for _, source := range []string{key, strings.ToUpper(key), " " + key} {
				rows = append(rows, AdamicAdapt(AdamicCase{Name: "Go named table", Source: source}))
			}
		}
		return rows
	}
	return []AdamicCase{AdamicAdapt(c)}
}
func AdamicObserve(c AdamicCase) {
	fmt.Printf("%t|%t|%t\n", IsColor(c.Source), IsNamedColor(c.Source), IsLength(c.Source))
}

var adamicLock sync.Mutex

func AdamicRecord(helper, source string) {
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
	if e = json.NewEncoder(f).Encode(AdamicCase{Source: source, Helper: helper}); e != nil {
		panic(e)
	}
}
