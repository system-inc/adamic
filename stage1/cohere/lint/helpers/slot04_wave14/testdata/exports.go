package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
)

type AdamicCase struct {
	Name, Source, Number, Divisor, Parsed string
	ParseOK                               bool
}

func adamicText(x float64) string {
	switch {
	case x == 0:
		return strconv.FormatFloat(x, 'g', -1, 64)
	default:
		s := strconv.FormatFloat(x, 'g', -1, 64)
		if s == "+Inf" {
			return "Infinity"
		}
		if s == "-Inf" {
			return "-Infinity"
		}
		return s
	}
}
func adamicFloat(s string) float64 {
	if s == "Infinity" {
		s = "+Inf"
	}
	if s == "-Infinity" {
		s = "-Inf"
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
func AdamicAdapt(c AdamicCase) AdamicCase {
	v, e := strconv.ParseFloat(c.Source, 64)
	c.ParseOK = e == nil
	c.Parsed = adamicText(v)
	if c.Number == "" {
		c.Number = c.Parsed
	}
	c.Number = adamicText(adamicFloat(c.Number))
	c.Divisor = adamicText(adamicFloat(c.Divisor))
	return c
}
func AdamicExpand(c AdamicCase) []AdamicCase {
	if c.Divisor != "" {
		return []AdamicCase{AdamicAdapt(c)}
	}
	out := []AdamicCase{}
	for _, d := range []string{"0.25", "1", "-0.25", "0", "-0", "+Inf", "-Inf", "NaN"} {
		v := c
		v.Divisor = d
		out = append(out, AdamicAdapt(v))
	}
	return out
}
func AdamicObserve(c AdamicCase) {
	v, d := adamicFloat(c.Number), adamicFloat(c.Divisor)
	fmt.Println(formatJavaScriptNumber(v))
	fmt.Println(formatJavaScriptNumber(modFloat(v, d)))
	fmt.Println(isMultipleOf(c.Source, d))
}

var adamicLock sync.Mutex

func AdamicRecord(c AdamicCase) {
	path := os.Getenv("ADAMIC_SLOT04_NUMERIC")
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
