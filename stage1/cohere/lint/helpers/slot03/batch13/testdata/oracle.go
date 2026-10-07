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
	for _, value := range []string{"var()", "var(--x)", "var(calc(1))", "VAR(calc(1))", " var(--x)", "var(--x),serif", "serif,var(--x)", "var(--x),var(--y)", "", ",", ",serif", "serif,", "0font", "1font", "١font", "😀", "é", "0", "serif", "sans-serif", "red", "#fff", "10px", "1%", "1/2", "1 2 3", "thin", "medium", "larger", "url(x)", "top", "cover", "1deg", "5turn", "image(x)", "0font,var(--x)", "'1font'", "a,(x,y)", "a\\,b", "a)b,c", `a"b,c"`} {
		texts[value] = true
	}
	ordered := []string{}
	for s := range texts {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	type input struct {
		Value      string   `json:"value"`
		Predicates []bool   `json:"predicates"`
		Parts      []string `json:"parts"`
	}
	kinds := []string{"color", "length", "percentage", "ratio", "number", "integer", "url", "position", "bg-size", "line-width", "image", "family-name", "generic-name", "absolute-size", "relative-size", "angle", "vector", "", "unknown", "COLOR", "integer "}
	orders := [][]string{{}, append([]string{}, kinds[:17]...)}
	reverse := []string{}
	for i := 16; i >= 0; i-- {
		reverse = append(reverse, kinds[i])
	}
	orders = append(orders, reverse)
	for _, kind := range kinds[:17] {
		orders = append(orders, []string{kind})
	}
	orders = append(orders, []string{"unknown", "integer", "number"}, []string{"number", "length"}, []string{"length", "number"}, []string{"family-name", "generic-name"}, []string{"generic-name", "family-name"}, []string{"number", "number", "unknown"})
	inputs := []input{}
	outputs := []string{}
	for _, value := range ordered {
		inputs = append(inputs, input{value, collapse.AdamicPredicates(value), collapse.AdamicParts(value)})
		for _, kind := range kinds {
			result, trace := collapse.AdamicMatch(value, kind)
			outputs = append(outputs, fmt.Sprintf("%t/%s", result, trace))
		}
		for _, order := range orders {
			result, trace := collapse.AdamicInfer(value, order)
			outputs = append(outputs, result+"/"+trace)
		}
		result, trace := collapse.AdamicFamily(value)
		outputs = append(outputs, fmt.Sprintf("%t/%s", result, trace))
	}
	out, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(out).Encode(map[string]any{"rows": inputs, "kinds": kinds, "orders": orders}); err != nil {
		panic(err)
	}
	out.Close()
	for _, output := range outputs {
		fmt.Println(output)
	}
}
