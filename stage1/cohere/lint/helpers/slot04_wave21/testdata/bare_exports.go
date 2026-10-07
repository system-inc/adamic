package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

type AdamicPart struct {
	Text, Trimmed string
	Positive      bool
}
type AdamicBare struct {
	Argument, Candidate, Fraction, Resolved, Inferred, ParseSource, Printed, PercentSource string
	Parts                                                                                  []AdamicPart
	Spacing, PercentPositive                                                               bool
}

func AdamicBareAdapt(argument, candidate, fraction string) AdamicBare {
	c := AdamicBare{Argument: argument, Candidate: candidate, Fraction: fraction, Resolved: candidate, Parts: []AdamicPart{}}
	if argument == "ratio" {
		c.Resolved = fraction
	}
	c.Inferred = string(InferDataType(c.Resolved, []DataType{DataType(argument)}))
	for _, p := range segment(c.Resolved, '/') {
		c.Parts = append(c.Parts, AdamicPart{p, strings.TrimSpace(p), IsPositiveInteger(p)})
	}
	c.Spacing = isValidSpacingMultiplier(c.Resolved)
	c.PercentSource = strings.TrimSuffix(c.Resolved, "%")
	c.PercentPositive = IsPositiveInteger(c.PercentSource)
	c.ParseSource = c.Resolved
	if c.Inferred == "ratio" && len(c.Parts) == 2 {
		c.ParseSource = c.Parts[0].Trimmed + " / " + c.Parts[1].Trimmed
	}
	c.Printed = ValueToCss(ParseValue(c.ParseSource))
	return c
}
func AdamicBareObserve(c AdamicBare) {
	s := &utilityEvaluation{}
	nodes, ratio, ok := s.resolveBareArgument(c.Argument, &ParsedValue{Kind: ParsedValueKindNamed, Value: c.Candidate, Fraction: c.Fraction})
	fmt.Printf("bare|%t|%t|%s\n", ok, ratio, ValueToCss(nodes))
}

var adamicBareLock sync.Mutex

func AdamicRecordBare(argument string, value *ParsedValue) {
	path := os.Getenv("ADAMIC_SLOT04_BARE")
	if path == "" {
		return
	}
	row := struct{ Helper, Source, Argument, Candidate, Fraction string }{"resolveBareArgument", argument, argument, value.Value, value.Fraction}
	adamicBareLock.Lock()
	defer adamicBareLock.Unlock()
	f, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	if e = json.NewEncoder(f).Encode(row); e != nil {
		panic(e)
	}
}
