package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	z, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(z)
	texts := map[string]bool{}
	pattern := regexp.MustCompile(`[[:alnum:]_@.\-]+`)
	for {
		var row struct{ Rule, File, Source string }
		err = decoder.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		texts[row.Source] = true
		for _, s := range pattern.FindAllString(row.Source, -1) {
			texts[s] = true
		}
	}
	z.Close()
	f.Close()
	for _, value := range []string{"", "+", "-", ".", ".5", "5.", "0", "01", "+1", "-0", "-1", "1e", "1e+", "1e-2", "1E+3", "1em", "NaN", "Infinity", "0x10", "😀", "1😀", "1é", "é1", "١", "１２", "calc(1px+2px)", "CALC(1)", "xcalc(", "url(round(x))"} {
		for _, suffix := range []string{"", "%", "deg", "rad", "grad", "turn", "DEG", "px", " ", "\n", "\x00"} {
			texts[value+suffix] = true
		}
	}
	for _, name := range []string{"calc", "min", "max", "clamp", "mod", "rem", "sin", "cos", "tan", "asin", "acos", "atan", "atan2", "pow", "sqrt", "hypot", "log", "exp", "round"} {
		for _, value := range []string{name + "(", strings.ToUpper(name) + "(", name + " (", "x" + name + "(", "😀" + name + "(", "'" + name + "('", name + "\x00("} {
			texts[value] = true
		}
	}
	ordered := []string{}
	for s := range texts {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	type input struct {
		Value      string `json:"value"`
		Consumed   int    `json:"consumed"`
		Math       bool   `json:"math"`
		Angle      bool   `json:"angle"`
		Percentage bool   `json:"percentage"`
	}
	inputs := []input{}
	outputs := []string{}
	for _, value := range ordered {
		consumed, math, angle, percentage := collapse.AdamicDependencies(value)
		inputs = append(inputs, input{value, consumed, math, angle, percentage})
		results, traces := collapse.AdamicPredicates(value)
		for i, result := range results {
			outputs = append(outputs, fmt.Sprintf("%t/%s", result, traces[i]))
		}
	}
	out, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(out).Encode(inputs); err != nil {
		panic(err)
	}
	out.Close()
	for _, output := range outputs {
		fmt.Println(output)
	}
}
