package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
)

type AdamicCase struct {
	Name, Source, Parsed, Helper string
	ParseOK                      bool
}

func AdamicAdapt(c AdamicCase) AdamicCase {
	v, e := strconv.ParseFloat(c.Source, 64)
	c.ParseOK = e == nil
	c.Parsed = strconv.FormatFloat(v, 'g', -1, 64)
	if c.Parsed == "+Inf" {
		c.Parsed = "Infinity"
	}
	if c.Parsed == "-Inf" {
		c.Parsed = "-Infinity"
	}
	return c
}
func AdamicExpand(c AdamicCase) []AdamicCase { return []AdamicCase{AdamicAdapt(c)} }
func AdamicObserve(c AdamicCase) {
	fmt.Printf("%t|%t|%t\n", IsPositiveInteger(c.Source), roundTripsAsJavaScriptNumber(c.Source), isValidSpacingMultiplier(c.Source))
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
