package lower

import (
	"context"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/load"
)

// Test-only normalized call-site heap. A future fresh interpreter must produce
// this snapshot after operand evaluation, before transfer. This is not an AST
// recognizer or a compiler hook. Edges plus roots represent all paths finitely,
// including infinitely many walks around a cycle; Paths contains one witness
// per outside root, not an exhaustive expansion of those walks.
type ownershipObject struct {
	edges   map[string]int
	region  int
	unknown bool
}
type ownershipRoot struct {
	path                  string
	object                int
	after, weak, borrowed bool
}
type ownershipSnapshot struct {
	objects map[int]ownershipObject
	roots   []ownershipRoot
}
type ownershipPath struct {
	Path                      string
	UsedAfter, Weak, Borrowed bool
}
type ownershipAnswer struct {
	Members []int
	Paths   []ownershipPath
	Unique  bool
	Message string
}

func ownershipQuery(s ownershipSnapshot, owner string) ownershipAnswer {
	answer := ownershipAnswer{Unique: true}
	var selected *ownershipRoot
	for i := range s.roots {
		if s.roots[i].path == owner {
			selected = &s.roots[i]
			break
		}
	}
	if selected == nil {
		answer.Unique = false
		answer.Message = "cannot move " + owner + ": whole reachable ownership is not proven"
		return answer
	}
	// Close over strong edges AND whole non-splitting regions. Iterate because
	// additional region members can themselves own counted descendants.
	members := map[int]bool{}
	queue := []int{selected.object}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if members[id] {
			continue
		}
		members[id] = true
		object, ok := s.objects[id]
		if !ok || object.unknown {
			answer.Unique = false
			answer.Message = "cannot move " + owner + ": whole reachable ownership is not proven"
		}
		for _, target := range object.edges {
			queue = append(queue, target)
		}
		if object.region != 0 {
			for other, candidate := range s.objects {
				if candidate.region == object.region {
					queue = append(queue, other)
				}
			}
		}
	}
	for id := range members {
		answer.Members = append(answer.Members, id)
	}
	sort.Ints(answer.Members)
	// Find shortest deterministic ingress witness from each surviving root.
	// Dead but still owning roots remain blockers; liveness never drops a count.
	for _, root := range s.roots {
		pending := []struct {
			id   int
			path string
		}{{root.object, root.path}}
		seen := map[int]bool{}
		for len(pending) > 0 {
			at := pending[0]
			pending = pending[1:]
			if seen[at.id] {
				continue
			}
			seen[at.id] = true
			if members[at.id] {
				answer.Paths = append(answer.Paths, ownershipPath{at.path, root.after, root.weak, root.borrowed})
				if root.path != owner || root.weak || root.borrowed {
					answer.Unique = false
					if answer.Message == "" {
						answer.Message = "cannot move " + owner + ": another variable, field or closure may reach the graph via " + at.path
					}
				}
				break
			}
			object, known := s.objects[at.id]
			if !known || object.unknown {
				answer.Unique = false
				if answer.Message == "" {
					answer.Message = "cannot move " + owner + ": whole reachable ownership is not proven at " + at.path
				}
			}
			labels := []string{}
			for label := range object.edges {
				labels = append(labels, label)
			}
			sort.Strings(labels)
			for _, label := range labels {
				pending = append(pending, struct {
					id   int
					path string
				}{s.objects[at.id].edges[label], at.path + label})
			}
		}
	}
	sort.Slice(answer.Paths, func(i, j int) bool { return answer.Paths[i].Path < answer.Paths[j].Path })
	if selected.after {
		answer.Unique = false
		answer.Message = "use after move: " + owner + "; " + owner + " was moved into parallelMap"
	}
	return answer
}

func TestOwnershipQueryShapes(t *testing.T) {
	cases := []struct {
		name    string
		objects map[int]ownershipObject
		roots   []ownershipRoot
		members []int
		unique  bool
		paths   []ownershipPath
	}{
		{"flat literal", map[int]ownershipObject{1: {}}, []ownershipRoot{{path: "item", object: 1}}, []int{1}, true, []ownershipPath{{Path: "item"}}},
		{"nested object", map[int]ownershipObject{1: {edges: map[string]int{".child": 2}}, 2: {}}, []ownershipRoot{{path: "item", object: 1}}, []int{1, 2}, true, []ownershipPath{{Path: "item"}}},
		{"array of objects", map[int]ownershipObject{1: {edges: map[string]int{"[0]": 2, "[1]": 3}}, 2: {}, 3: {}}, []ownershipRoot{{path: "items", object: 1}}, []int{1, 2, 3}, true, []ownershipPath{{Path: "items"}}},
		{"local ring", map[int]ownershipObject{1: {edges: map[string]int{".next": 2}, region: 1}, 2: {edges: map[string]int{".next": 3}, region: 1}, 3: {edges: map[string]int{".next": 1}, region: 1}}, []ownershipRoot{{path: "ring", object: 1}}, []int{1, 2, 3}, true, []ownershipPath{{Path: "ring"}}},
		{"ring with second outside reference", map[int]ownershipObject{1: {edges: map[string]int{".next": 2}, region: 1}, 2: {edges: map[string]int{".next": 1}, region: 1}, 3: {edges: map[string]int{".saved": 2}}}, []ownershipRoot{{path: "ring", object: 1}, {path: "other", object: 3, after: true}}, []int{1, 2}, false, []ownershipPath{{Path: "other.saved", UsedAfter: true}, {Path: "ring"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ownershipQuery(ownershipSnapshot{c.objects, c.roots}, c.roots[0].path)
			if got.Unique != c.unique || !reflect.DeepEqual(got.Members, c.members) || !reflect.DeepEqual(got.Paths, c.paths) {
				t.Fatalf("unexpected answer: %+v", got)
			}
			if !got.Unique && got.Message != "cannot move ring: another variable, field or closure may reach the graph via other.saved" {
				t.Fatalf("unexpected refusal: %q", got.Message)
			}
			t.Logf("members=%v unique=%v paths=%+v message=%q", got.Members, got.Unique, got.Paths, got.Message)
		})
	}
}

func TestOwnershipQueryRefusals(t *testing.T) {
	for _, variant := range []string{"dead alias", "weak observer", "borrowed owner", "later source use", "unknown", "opaque outside root", "disconnected region member"} {
		t.Run(variant, func(t *testing.T) {
			s := ownershipSnapshot{map[int]ownershipObject{1: {}}, []ownershipRoot{{path: "item", object: 1}}}
			switch variant {
			case "dead alias":
				s.roots = append(s.roots, ownershipRoot{path: "alias", object: 1})
			case "weak observer":
				s.roots = append(s.roots, ownershipRoot{path: "weak", object: 1, weak: true})
			case "borrowed owner":
				s.roots[0].borrowed = true
			case "later source use":
				s.roots[0].after = true
			case "unknown":
				s.objects[1] = ownershipObject{unknown: true}
			case "opaque outside root":
				s.objects[2] = ownershipObject{unknown: true}
				s.roots = append(s.roots, ownershipRoot{path: "opaque", object: 2})
			case "disconnected region member":
				s.objects[1] = ownershipObject{region: 7}
				s.objects[2] = ownershipObject{region: 7}
				s.roots = append(s.roots, ownershipRoot{path: "alias", object: 2})
			}
			got := ownershipQuery(s, "item")
			if got.Unique || got.Message == "" {
				t.Fatalf("unsafe answer: %+v", got)
			}
		})
	}
}

// Measures actual reusable flow/liveness preparation on the largest available
// stage-1 ports, not the unimplemented heap snapshot extraction or whole query.
func BenchmarkOwnershipQueryPreparation(b *testing.B) {
	path := os.Getenv("ADAMIC_OWNERSHIP_BENCH_SOURCE")
	if path == "" {
		path = "../../stage1/typescript/parser/main.ts"
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		b.Fatal(err)
	}
	if os.Getenv("ADAMIC_OWNERSHIP_BENCH_TSGO") == "1" {
		checked.EnableTSGo()
	}
	program, err := Lower(context.Background(), checked)
	if err != nil {
		b.Fatal(err)
	}
	b.Logf("input=%s functions=%d locals=%d", path, len(program.Functions), len(program.Locals))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for function := -1; function < len(program.Functions); function++ {
			graph := flow.Build(program, function)
			live := flow.LiveOut(graph)
			if live == nil {
				b.Fatalf("missing liveness: %d", function)
			}
		}
	}
}
