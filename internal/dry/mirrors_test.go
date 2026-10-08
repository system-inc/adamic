// Package dry holds the repository guard against new cohere mirrors.
package dry

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type mirror struct {
	Path            string `json:"path"`
	Kind            string `json:"kind"`
	RuleID          string `json:"rule_id,omitempty"`
	Counterpart     string `json:"counterpart"`
	Decision        string `json:"decision"`
	Reason          string `json:"reason"`
	RetiringTask    string `json:"retiring_task"`
	Evidence        string `json:"evidence,omitempty"`
	PendingFunction string `json:"pending_function,omitempty"`
}

func TestCohereMirrors(t *testing.T) {
	t.Parallel()
	problems, err := checkMirrors("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, problem := range problems {
		t.Error(problem)
	}
}

func parseGo(path string) (*ast.File, error) {
	return parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
}

// goFiles includes scratch files, generated files and tests, regardless of build tags.
func goFiles(root string, visit func(string, *ast.File) error) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path == filepath.Join(root, "cohere") || path == filepath.Join(root, "stage1") || entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		file, err := parseGo(path)
		if err != nil {
			return fmt.Errorf("%s: parse Go before checking mirrors: %w", path, err)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		return visit(filepath.ToSlash(relative), file)
	})
}

// Registration and definition are joined by package directory and variable name. An id
// merely mentioned in a cohere test or policy message is not a registered rule.
func cohereIDs(root string) (map[string]bool, error) {
	data, err := os.ReadFile(filepath.Join(root, "cohere/internal/lint/configuration/sets/adamic.json"))
	if err != nil {
		return nil, fmt.Errorf("read cohere adamic set (initialize submodules): %w", err)
	}
	var set struct {
		Rules map[string]json.RawMessage `json:"rules"`
	}
	if err := json.Unmarshal(data, &set); err != nil {
		return nil, err
	}
	if len(set.Rules) == 0 {
		return nil, fmt.Errorf("cohere adamic set has no rules")
	}
	definitions := map[string]string{}
	registrations := map[string]bool{}
	rulesRoot := filepath.Join(root, "cohere/internal/lint/rules")
	err = goFiles(rulesRoot, func(path string, file *ast.File) error {
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		directory := filepath.Dir(path)
		for _, decl := range file.Decls {
			general, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range general.Specs {
				value, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for index, expr := range value.Values {
					if index >= len(value.Names) {
						continue
					}
					literal, ok := expr.(*ast.CompositeLit)
					if !ok {
						continue
					}
					selector, ok := literal.Type.(*ast.SelectorExpr)
					if !ok || selector.Sel.Name != "Rule" {
						continue
					}
					for _, element := range literal.Elts {
						pair, ok := element.(*ast.KeyValueExpr)
						if !ok {
							continue
						}
						key, ok := pair.Key.(*ast.Ident)
						if !ok || key.Name != "Name" {
							continue
						}
						if name, ok := stringValue(pair.Value); ok {
							definitions[directory+"/"+value.Names[index].Name] = name
						}
					}
				}
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Register" {
				return true
			}
			for _, arg := range call.Args {
				literal, ok := arg.(*ast.CompositeLit)
				if !ok {
					continue
				}
				for _, element := range literal.Elts {
					pair, ok := element.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					key, ok := pair.Key.(*ast.Ident)
					if !ok || key.Name != "Rule" {
						continue
					}
					if name, ok := pair.Value.(*ast.Ident); ok {
						registrations[directory+"/"+name.Name] = true
					}
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	registered := map[string]bool{}
	for key := range registrations {
		if id := definitions[key]; id != "" {
			registered[id] = true
		}
	}
	ids := map[string]bool{}
	for id := range set.Rules {
		if !registered[id] {
			return nil, fmt.Errorf("cohere adamic set rule %s has no parsed registration; update the guard's registration parser", id)
		}
		ids[id] = true
	}
	return ids, nil
}

func stringValue(expr ast.Expr) (string, bool) {
	return resolvedString(expr, map[*ast.Object]bool{})
}

func resolvedString(expr ast.Expr, visiting map[*ast.Object]bool) (string, bool) {
	switch value := expr.(type) {
	case *ast.Ident:
		if value.Obj == nil || visiting[value.Obj] {
			return "", false
		}
		declaration, ok := value.Obj.Decl.(*ast.ValueSpec)
		if !ok {
			return "", false
		}
		visiting[value.Obj] = true
		defer delete(visiting, value.Obj)
		for index, name := range declaration.Names {
			if name.Name == value.Name && index < len(declaration.Values) {
				return resolvedString(declaration.Values[index], visiting)
			}
		}
	case *ast.BasicLit:
		if value.Kind == token.STRING {
			text, err := strconv.Unquote(value.Value)
			return text, err == nil
		}
	case *ast.BinaryExpr:
		if value.Op == token.ADD {
			left, lok := resolvedString(value.X, visiting)
			right, rok := resolvedString(value.Y, visiting)
			return left + right, lok && rok
		}
	case *ast.ParenExpr:
		return resolvedString(value.X, visiting)
	}
	return "", false
}

// Patterns apply only to parsed comments. Go expressions and registrations use ASTs.
var provenance = regexp.MustCompile(`(?is)lifted\s+from\s+cohere|(?:cohere.*?(?:\b[0-9a-f]{7,40}\b|/commit/))`)

func hasProvenance(file *ast.File) bool {
	for _, group := range file.Comments {
		if provenance.MatchString(group.Text()) {
			return true
		}
	}
	return false
}

func ruleMention(text, id string) bool {
	for offset := 0; offset < len(text); {
		index := strings.Index(text[offset:], id)
		if index < 0 {
			return false
		}
		index += offset
		end := index + len(id)
		word := func(b byte) bool {
			return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || b == '-' || b == '_' || b == '/'
		}
		if (index == 0 || !word(text[index-1])) && (end == len(text) || !word(text[end])) {
			return true
		}
		offset = end
	}
	return false
}

func checkMirrors(root string) ([]string, error) {
	ids, err := cohereIDs(root)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(root, "cohere-mirrors.json"))
	if err != nil {
		return nil, err
	}
	var entries []mirror
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&entries); err != nil {
		return nil, err
	}
	byPath := map[string][]int{}
	var problems []string
	seen := map[string]bool{}
	for index, entry := range entries {
		key := entry.Path + ":" + entry.Kind + ":" + entry.RuleID
		if seen[key] {
			problems = append(problems, fmt.Sprintf("%s: duplicate mirror %s; remove the duplicate entry", entry.Path, entry.RuleID))
		}
		seen[key] = true
		if filepath.ToSlash(filepath.Clean(entry.Path)) != entry.Path || filepath.IsAbs(entry.Path) || strings.HasPrefix(entry.Path, "../") || !strings.HasSuffix(entry.Path, ".go") {
			return nil, fmt.Errorf("invalid mirror path %q", entry.Path)
		}
		if entry.Kind != "lift" && entry.Kind != "rule" || entry.Decision != "a" && entry.Decision != "b" && entry.Decision != "c" || entry.Reason == "" || entry.RetiringTask == "" {
			return nil, fmt.Errorf("%s: mirror needs kind, decision, reason and retiring task", entry.Path)
		}
		if entry.Kind == "rule" && entry.RuleID == "" {
			return nil, fmt.Errorf("%s: rule mirror needs rule_id", entry.Path)
		}
		if !strings.HasPrefix(entry.Counterpart, "cohere/") || strings.Contains(entry.Counterpart, "..") {
			return nil, fmt.Errorf("%s: counterpart must name a cohere path", entry.Path)
		}
		if _, err := os.Stat(filepath.Join(root, entry.Counterpart)); err != nil {
			return nil, fmt.Errorf("%s: cohere counterpart %s: %w", entry.Path, entry.Counterpart, err)
		}
		byPath[entry.Path] = append(byPath[entry.Path], index)
	}
	matched := make([]bool, len(entries))
	err = goFiles(root, func(path string, file *ast.File) error {
		lifted := hasProvenance(file)
		literals := []string{}
		functions := map[string]bool{}
		// Imports identify dependencies, not diagnostics. Comments likewise do not emit ids.
		imports := map[*ast.BasicLit]bool{}
		for _, imp := range file.Imports {
			imports[imp.Path] = true
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if function, ok := node.(*ast.FuncDecl); ok {
				functions[function.Name.Name] = true
			}
			if literal, ok := node.(*ast.BasicLit); ok && imports[literal] {
				return false
			}
			if expr, ok := node.(ast.Expr); ok {
				if value, ok := stringValue(expr); ok {
					literals = append(literals, value)
				}
			}
			return true
		})
		used := map[string]bool{}
		for _, literal := range literals {
			for id := range ids {
				if ruleMention(literal, id) {
					used[id] = true
				}
			}
		}
		covered := map[string]bool{}
		for _, index := range byPath[path] {
			entry := entries[index]
			if entry.Kind == "lift" {
				matched[index] = lifted
				continue
			}
			for _, literal := range literals {
				if ruleMention(literal, entry.RuleID) || entry.Evidence != "" && literal == entry.Evidence {
					matched[index] = true
				}
			}
			if entry.PendingFunction != "" && functions[entry.PendingFunction] {
				matched[index] = true
			}
			if matched[index] {
				covered[entry.RuleID] = true
			}
		}
		if lifted && len(byPath[path]) == 0 {
			problems = append(problems, fmt.Sprintf("%s: cohere source provenance is unlisted; consume cohere by reference, or record the existing mirror's decision and retiring task in cohere-mirrors.json", path))
		}
		for id := range used {
			if !covered[id] {
				problems = append(problems, fmt.Sprintf("%s: emits cohere rule %s without a mirror entry; consume the rule by reference or add this file/id with its decision and retiring task to cohere-mirrors.json", path, id))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for index, entry := range entries {
		if !matched[index] {
			problems = append(problems, fmt.Sprintf("%s: stale %s mirror %s; remove its cohere-mirrors.json entry because its provenance, diagnostic or pending landing function no longer exists", entry.Path, entry.Kind, entry.RuleID))
		}
	}
	sort.Strings(problems)
	return problems, nil
}

func TestMirrorMutants(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../cohere-mirrors.json")
	if err != nil {
		t.Fatal(err)
	}
	var entries []mirror
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatal(err)
	}
	ruleID := ""
	removedIndex := -1
	for index, entry := range entries {
		if entry.Kind == "rule" && entry.Evidence == "" && entry.PendingFunction == "" {
			ruleID = entry.RuleID
			removedIndex = index
			break
		}
	}
	if removedIndex < 0 {
		t.Fatal("inventory has no active rule for the removal mutant")
	}
	for _, mutation := range []string{"lifted header", "commit source", "unlisted rule", "assembled rule", "removed entry", "stale entry", "retired landing"} {
		t.Run(mutation, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			cohere, err := filepath.Abs("../../cohere")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(cohere, filepath.Join(root, "cohere")); err != nil {
				t.Fatal(err)
			}
			write := func(path string, content []byte) {
				t.Helper()
				target := filepath.Join(root, path)
				if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(target, content, 0644); err != nil {
					t.Fatal(err)
				}
			}
			for _, entry := range entries {
				source, err := os.ReadFile(filepath.Join("../..", entry.Path))
				if err != nil {
					t.Fatal(err)
				}
				write(entry.Path, source)
			}
			write("cohere-mirrors.json", data)
			baseline, err := checkMirrors(root)
			if err != nil || len(baseline) != 0 {
				t.Fatalf("mutant baseline: %v, %v", baseline, err)
			}
			expectedFile := "scratch.go"
			expectedAction := "cohere-mirrors.json"
			switch mutation {
			case "lifted header":
				write(expectedFile, []byte("// lIfTeD FrOm CoHeRe\npackage scratch\n"))
			case "commit source":
				write(expectedFile, []byte("// Source: cohere at commit 715ba94\npackage scratch\n"))
			case "unlisted rule":
				write(expectedFile, []byte("package scratch\nconst diagnostic = "+strconv.Quote("refused ("+ruleID+")")+"\n"))
			case "assembled rule":
				parts := strings.SplitN(ruleID, "/", 2)
				write(expectedFile, []byte("package scratch\nconst prefix = "+strconv.Quote(parts[0]+"/")+"\nconst diagnostic = prefix + "+strconv.Quote(parts[1])+"\n"))
			case "removed entry":
				reduced := append([]mirror{}, entries[:removedIndex]...)
				reduced = append(reduced, entries[removedIndex+1:]...)
				changed, err := json.Marshal(reduced)
				if err != nil {
					t.Fatal(err)
				}
				write("cohere-mirrors.json", changed)
				expectedFile = entries[removedIndex].Path
			case "stale entry":
				for _, entry := range entries {
					if entry.Kind == "lift" {
						expectedFile = entry.Path
						break
					}
				}
				write(expectedFile, []byte("package retired\n"))
				expectedAction = "remove its"
			case "retired landing":
				// The actual optional-widening landing is implemented now. Exercise a
				// pending-only entry in scratch, so an active diagnostic cannot mask retirement.
				pending := entries[removedIndex]
				pending.Path = "pending_landing.go"
				pending.Evidence = ""
				pending.PendingFunction = "pendingLanding"
				expectedFile = pending.Path
				write(expectedFile, []byte("package scratch\nfunc (l *lowering) pendingLanding() {}\n"))
				withPending := append(append([]mirror{}, entries...), pending)
				changed, err := json.Marshal(withPending)
				if err != nil {
					t.Fatal(err)
				}
				write("cohere-mirrors.json", changed)
				baseline, err := checkMirrors(root)
				if err != nil || len(baseline) != 0 {
					t.Fatalf("pending mutant baseline: %v, %v", baseline, err)
				}
				write(expectedFile, []byte("package scratch\nfunc (l *lowering) retiredLanding() {}\n"))
				expectedAction = "remove its"
			}
			problems, err := checkMirrors(root)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, problem := range problems {
				if strings.Contains(problem, expectedFile) && strings.Contains(problem, expectedAction) {
					found = true
					t.Log(problem)
				}
			}
			if !found {
				t.Fatalf("%s survived or lacked file/action: %v", mutation, problems)
			}
		})
	}
}
