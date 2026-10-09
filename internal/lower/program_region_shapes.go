package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"sort"
	"strings"
)

type programShape struct {
	offered  map[string]bool
	required []string
	nodes    []cycleNode
}
type programShapeIndex struct {
	finder  *cycleFinder
	all     []cycleNode
	groups  []*programShape
	offered map[string][]*programShape
	cache   map[*checker.Type][]cycleNode
}

// Assignability requires every required target property to exist on the source.
// Index exact structural shapes once, then intersect offered-property postings and
// test subset shapes. The checker still decides compatibility within that shortlist.
func newProgramShapeIndex(f *cycleFinder, nodes []cycleNode) *programShapeIndex {
	index := &programShapeIndex{finder: f, all: nodes, offered: map[string][]*programShape{}, cache: map[*checker.Type][]cycleNode{}}
	groups := map[string]*programShape{}
	seen := map[cycleNode]bool{}
	for _, node := range nodes {
		if seen[node] || !index.eligible(node.proven) {
			continue
		}
		seen[node] = true
		offered, required := index.properties(node.proven)
		names := make([]string, 0, len(offered))
		for name := range offered {
			names = append(names, name)
		}
		sort.Strings(names)
		key := strings.Join(names, "\x00") + "\x01" + strings.Join(required, "\x00")
		group := groups[key]
		if group == nil {
			group = &programShape{offered: offered, required: required}
			groups[key] = group
			index.groups = append(index.groups, group)
			for name := range offered {
				index.offered[name] = append(index.offered[name], group)
			}
		}
		group.nodes = append(group.nodes, node)
	}
	return index
}
func (index *programShapeIndex) eligible(proven *checker.Type) bool {
	f := index.finder
	return proven != nil && proven.Flags()&checker.TypeFlagsObject != 0 && !f.weak(proven) && !f.template(proven) && !f.isFunction(proven) && !f.programScalarStorage(proven)
}
func (index *programShapeIndex) properties(proven *checker.Type) (map[string]bool, []string) {
	offered := map[string]bool{}
	required := []string{}
	for _, field := range index.finder.l.checker.GetPropertiesOfType(proven) {
		offered[field.Name] = true
		if field.Flags&ast.SymbolFlagsOptional == 0 {
			required = append(required, field.Name)
		}
	}
	sort.Strings(required)
	return offered, required
}
func (index *programShapeIndex) candidates(proven *checker.Type, indexed bool) []cycleNode {
	if !indexed {
		return index.all
	}
	if nodes, ok := index.cache[proven]; ok {
		return nodes
	}
	offered, required := index.properties(proven)
	selected := map[*programShape]bool{}
	// Target shapes whose required properties are a subset of this source shape.
	for _, group := range index.groups {
		subset := true
		for _, name := range group.required {
			if !offered[name] {
				subset = false
				break
			}
		}
		if subset {
			selected[group] = true
		}
	}
	// Source shapes that offer every required property of this target shape.
	sources := index.groups
	for _, name := range required {
		if len(index.offered[name]) < len(sources) {
			sources = index.offered[name]
		}
	}
	for _, group := range sources {
		superset := true
		for _, name := range required {
			if !group.offered[name] {
				superset = false
				break
			}
		}
		if superset {
			selected[group] = true
		}
	}
	result := []cycleNode{}
	for _, group := range index.groups {
		if selected[group] {
			result = append(result, group.nodes...)
		}
	}
	index.cache[proven] = result
	return result
}
