package main

import (
	"encoding/json"
	"errors"
	"fmt"
	engine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
	"path/filepath"
	"strings"
)

type Case struct {
	Name, Kind, Root, Entry, Package, Directory, Message, Path, Source string
	Failed                                                             bool
	Sheets                                                             []string
	System                                                             int
	Inputs                                                             []engine.AdamicUtilityInput
	Keys                                                               []string
	Adaptation                                                         engine.AdamicAdaptation
}

var current Case
var trace []string
var ErrNoTailwindEntryPoint = errors.New("no Tailwind entry point found in this project")

type system struct {
	ID          int
	Stylesheets []string
}
type table struct{ ID int }
type DesignSystemResult struct {
	System     *system
	Table      *table
	Err        error
	EntryPoint string
}
type recording struct{}
type adamicLoadOptions struct{ EntryPoint, TailwindPackageRoot string }

func projectRootOf(program string) string { trace = append(trace, "root"); return program }
func FindEntryPoint(root string, exists func(string) bool) string {
	trace = append(trace, "entry:"+root)
	return current.Entry
}
func findTailwindPackageRoot(dir string, exists func(string) bool) string {
	trace = append(trace, "package:"+dir)
	return current.Package
}
func (r *recording) FileExists(path string) bool { return true }
func (r *recording) Stat(path string)            { trace = append(trace, "stat:"+path) }
func adamicLoad(o adamicLoadOptions) (*system, error) {
	trace = append(trace, "load:"+o.EntryPoint+":"+o.TailwindPackageRoot)
	if current.Failed {
		return nil, errors.New(current.Message)
	}
	return &system{current.System, current.Sheets}, nil
}
func adamicTable(s *system) *table {
	trace = append(trace, "table:"+fmt.Sprint(s.ID))
	return &table{s.ID + 100}
}
func main() {
	cases := []Case{}
	data, err := os.ReadFile(os.Args[len(os.Args)-1])
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		panic(err)
	}
	adapted := len(os.Args) > 2 && os.Args[1] == "--cases"
	for i, c := range cases {
		if adapted {
			cases[i].Directory = filepath.Dir(c.Entry)
			for j, in := range c.Inputs {
				cases[i].Inputs[j].Trimmed = strings.TrimSpace(in.Params)
			}
			if c.Kind == "normalize" {
				cases[i].Adaptation = engine.AdamicAdapt(c.Source)
			}
			continue
		}
		switch c.Kind {
		case "load":
			current = c
			trace = []string{}
			result := loadDesignSystemThrough(c.Root, &recording{})
			message := ""
			sys, tab := 0, 0
			if result.Err != nil {
				message = result.Err.Error()
			}
			if result.System != nil {
				sys = result.System.ID
			}
			if result.Table != nil {
				tab = result.Table.ID
			}
			fmt.Printf("%t|%s|%s|%d|%d|%s\n", result.Err != nil, message, result.EntryPoint, sys, tab, strings.Join(trace, "|"))
		case "utility":
			engine.AdamicUtility(c.Inputs, c.Path, c.Keys)
		case "normalize":
			engine.AdamicNormalize(c.Source)
		default:
			panic("unknown case")
		}
	}
	if adapted {
		if err = json.NewEncoder(os.Stdout).Encode(cases); err != nil {
			panic(err)
		}
	}
}
