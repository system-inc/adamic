package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"sort"
	"strconv"
	"strings"
)

// Property-name compatibility only filters impossible assignments. The checker
// still decides every surviving relation, including optional fields and variance.
type programShape struct {
	all, required map[string]bool
	key           string
	candidates    []cycleNode
}
type programShapeIndex struct {
	finder *cycleFinder
	groups []*programShape
	types  map[*checker.Type]*programShape
	cache  map[string][]cycleNode
}

func (f *cycleFinder) programShapes(candidates []cycleNode) *programShapeIndex {
	index := &programShapeIndex{finder: f, types: map[*checker.Type]*programShape{}, cache: map[string][]cycleNode{}}
	groups := map[string]*programShape{}
	for _, candidate := range candidates {
		t := candidate.proven
		if t == nil || t.Flags()&checker.TypeFlagsObject == 0 || f.weak(t) || f.template(t) || f.isFunction(t) || f.programScalarStorage(t) {
			continue
		}
		shape := index.shape(t)
		group := groups[shape.key]
		if group == nil {
			group = shape
			groups[shape.key] = group
			index.groups = append(index.groups, group)
		}
		group.candidates = append(group.candidates, candidate)
	}
	return index
}
func (index *programShapeIndex) shape(t *checker.Type) *programShape {
	if shape := index.types[t]; shape != nil {
		return shape
	}
	shape := &programShape{all: map[string]bool{}, required: map[string]bool{}}
	all, required := []string{}, []string{}
	for _, field := range index.finder.l.checker.GetPropertiesOfType(t) {
		shape.all[field.Name] = true
		all = append(all, field.Name)
		if field.Flags&ast.SymbolFlagsOptional == 0 {
			shape.required[field.Name] = true
			required = append(required, field.Name)
		}
	}
	sort.Strings(all)
	sort.Strings(required)
	shape.key = programShapeKey(all, required)
	index.types[t] = shape
	return shape
}
func programShapeSubset(required, available map[string]bool) bool {
	for name := range required {
		if !available[name] {
			return false
		}
	}
	return true
}
func (index *programShapeIndex) relatedCandidates(t *checker.Type) []cycleNode {
	shape := index.shape(t)
	if candidates, known := index.cache[shape.key]; known {
		return candidates
	}
	candidates := []cycleNode{}
	for _, group := range index.groups {
		if programShapeSubset(shape.required, group.all) || programShapeSubset(group.required, shape.all) {
			candidates = append(candidates, group.candidates...)
		}
	}
	index.cache[shape.key] = candidates
	return candidates
}
func programShapeKey(all, required []string) string {
	quote := func(names []string) string {
		encoded := make([]string, len(names))
		for i, name := range names {
			encoded[i] = strconv.Quote(name)
		}
		return strings.Join(encoded, ",")
	}
	return quote(all) + "|" + quote(required)
}
