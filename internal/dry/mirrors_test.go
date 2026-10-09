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
	Path         string `json:"path"`
	Kind         string `json:"kind"`
	RuleID       string `json:"rule_id,omitempty"`
	Counterpart  string `json:"counterpart"`
	Decision     string `json:"decision"`
	Reason       string `json:"reason"`
	RetiringTask string `json:"retiring_task"`
	Evidence     string `json:"evidence,omitempty"`
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
			if path == filepath.Join(root, "cohere") || path == filepath.Join(root, "stage1") || entry.Name() == ".git" || entry.Name() == "review" || entry.Name() == "node_modules" {
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
	for id := range set.Rules {
		if !registered[id] {
			return nil, fmt.Errorf("cohere adamic set rule %s has no parsed registration; update the guard's registration parser", id)
		}
	}
	// The soundness set alone misses ordinary registered cohere rules.
	if len(registered) == 0 {
		return nil, fmt.Errorf("no cohere registrations parsed")
	}
	return registered, nil
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
var provenance = regexp.MustCompile(`(?is)lifted\s+from\s+cohere`)

func hasProvenance(file *ast.File) bool {
	for _, group := range file.Comments {
		if provenance.MatchString(group.Text()) {
			return true
		}
	}
	return false
}

// Check leading source comments in every language, including stage1 ports. Prose
// documents and string literals are not source provenance headers.
func sourceHeaders(root string) (map[string]bool, error) {
	headers := map[string]bool{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "cohere" || entry.Name() == "review" || entry.Name() == "node_modules" || entry.Name() == ".cache" {
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(path) {
		case ".go", ".a", ".ts", ".c", ".h", ".js", ".mjs", ".py", ".sh":
		default:
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var comments []string
		block := false
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "/*") {
				block = true
			}
			if !block && !strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "#") {
				break
			}
			comments = append(comments, strings.TrimLeft(line, "/*# "))
			if strings.Contains(line, "*/") {
				block = false
			}
		}
		if provenance.MatchString(strings.Join(comments, "\n")) {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			headers[filepath.ToSlash(relative)] = true
		}
		return nil
	})
	return headers, err
}

func ruleMention(text, id string) bool {
	// Bare upstream names also occur as English words and corpus metadata.
	// Compiler diagnostics mark them in brackets or parentheses.
	if !strings.Contains(id, "/") {
		return strings.Contains(text, "("+id+")") || strings.Contains(text, "["+id+"]")
	}
	if strings.HasPrefix(id, "@typescript-eslint/") && (strings.Contains(text, "("+strings.TrimPrefix(id, "@typescript-eslint/")+")") || strings.Contains(text, "["+strings.TrimPrefix(id, "@typescript-eslint/")+"]")) {
		return true
	}

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
	headers, err := sourceHeaders(root)
	if err != nil {
		return nil, err
	}
	ids, err := cohereIDs(root)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(root, "internal/dry/testdata/cohere-mirrors.json"))
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
		if filepath.ToSlash(filepath.Clean(entry.Path)) != entry.Path || filepath.IsAbs(entry.Path) || strings.HasPrefix(entry.Path, "../") || entry.Path == "." {
			return nil, fmt.Errorf("invalid mirror path %q", entry.Path)
		}
		if entry.Kind != "lift" && entry.Kind != "rule" || entry.Decision != "keep" && entry.Decision != "delete" || entry.Reason == "" || entry.RetiringTask == "" {
			return nil, fmt.Errorf("%s: mirror needs kind, decision, reason and retiring task", entry.Path)
		}
		if entry.Kind == "rule" && !ids[entry.RuleID] {
			return nil, fmt.Errorf("%s: rule mirror needs a registered cohere rule_id", entry.Path)
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
	for index, entry := range entries {
		if entry.Kind == "lift" && headers[entry.Path] {
			matched[index] = true
			delete(headers, entry.Path)
		}
	}
	err = goFiles(root, func(path string, file *ast.File) error {
		lifted := hasProvenance(file)
		literals := []string{}
		// Imports identify dependencies, not diagnostics. Comments likewise do not emit ids.
		imports := map[*ast.BasicLit]bool{}
		for _, imp := range file.Imports {
			imports[imp.Path] = true
		}
		// A rule registration or local Refused diagnostic in a test is still an
		// implementation. Ordinary assertions and source witnesses are references.
		if strings.HasSuffix(path, "_test.go") {
			ast.Inspect(file, func(node ast.Node) bool {
				literal, ok := node.(*ast.CompositeLit)
				if !ok {
					return true
				}
				name := ""
				switch kind := literal.Type.(type) {
				case *ast.Ident:
					name = kind.Name
				case *ast.SelectorExpr:
					name = kind.Sel.Name
				}
				if name == "Rule" || name == "Refused" {
					ast.Inspect(literal, func(node ast.Node) bool {
						if expr, ok := node.(ast.Expr); ok {
							if value, ok := stringValue(expr); ok {
								literals = append(literals, value)
							}
						}
						return true
					})
				}
				return true
			})
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if strings.HasSuffix(path, "_test.go") {
				return false
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
				matched[index] = matched[index] || lifted
				continue
			}
			for _, literal := range literals {
				if ruleMention(literal, entry.RuleID) || entry.Evidence != "" && literal == entry.Evidence {
					matched[index] = true
				}
			}
			if matched[index] {
				covered[entry.RuleID] = true
			}
		}
		coveredLift := false
		for _, index := range byPath[path] {
			if entries[index].Kind == "lift" && matched[index] {
				coveredLift = true
			}
		}
		if lifted && !coveredLift {
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
	for path := range headers {
		// Go comment groups have already produced their detailed diagnostic.
		if !strings.HasSuffix(path, ".go") || strings.HasPrefix(path, "stage1/") {
			problems = append(problems, path+": cohere source provenance is unlisted; record its decision in cohere-mirrors.json")
		}
	}
	for index, entry := range entries {
		if !matched[index] {
			problems = append(problems, fmt.Sprintf("%s: stale %s mirror %s; remove its cohere-mirrors.json entry because its provenance or diagnostic no longer exists", entry.Path, entry.Kind, entry.RuleID))
		}
	}
	sort.Strings(problems)
	return problems, nil
}

// Each mutant changes real guard input in its own directory. No repository source
// is mutated, so these leaves can run in parallel with the repository guard.
func mirrorMutant(t *testing.T, mutation string) {
	t.Helper()
	root := t.TempDir()
	cohere, err := filepath.Abs("../../cohere")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(cohere, filepath.Join(root, "cohere")); err != nil {
		t.Fatal(err)
	}
	write := func(path string, data []byte) {
		t.Helper()
		target := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	inventory := "internal/dry/testdata/cohere-mirrors.json"
	data, err := os.ReadFile("testdata/cohere-mirrors.json")
	if err != nil {
		t.Fatal(err)
	}
	var entries []mirror
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		source, err := os.ReadFile(filepath.Join("../..", entry.Path))
		if err != nil {
			t.Fatal(err)
		}
		write(entry.Path, source)
	}
	write(inventory, data)
	baseline, err := checkMirrors(root)
	if err != nil || len(baseline) != 0 {
		t.Fatalf("baseline: %v, %v", baseline, err)
	}
	expected := "scratch.go:"
	action := "unlisted"
	switch mutation {
	case "header":
		write("scratch.go", []byte("// lIfTeD FrOm CoHeRe\npackage scratch\n"))
	case "adamic header":
		expected = "scratch.a:"
		write("scratch.a", []byte("// Lifted from cohere\nexport const value = 1;\n"))
	case "test header":
		expected = "scratch_test.go:"
		write("scratch_test.go", []byte("// Lifted from cohere\npackage scratch\n"))
	case "rule":
		action = "emits cohere rule adamic/invariant-mutable"
		write("scratch.go", []byte("package scratch\nconst diagnostic = \"refused (adamic/invariant-mutable)\"\n"))
	case "test rule":
		expected = "scratch_test.go:"
		action = "emits cohere rule adamic/invariant-mutable"
		write("scratch_test.go", []byte("package scratch\ntype Rule struct { Name string }\nvar localRule = Rule{Name: \"adamic/invariant-mutable\"}\n"))
	case "core rule":
		action = "emits cohere rule no-debugger"
		write("scratch.go", []byte("package scratch\nconst diagnostic = \"refused (no-debugger)\"\n"))
	case "assembled rule":
		action = "emits cohere rule adamic/invariant-mutable"
		write("scratch.go", []byte("package scratch\nconst prefix = \"adamic/\"\nconst diagnostic = prefix + \"invariant-mutable\"\n"))
	case "stale header":
		expected = entries[0].Path + ":"
		action = "stale lift"
		write(entries[0].Path, []byte("package retired\n"))
	case "missing file":
		expected = entries[0].Path + ":"
		action = "stale lift"
		if err := os.Remove(filepath.Join(root, entries[0].Path)); err != nil {
			t.Fatal(err)
		}
	case "stale rule":
		expected = "internal/lower/cast_proof.go:"
		action = "stale rule mirror adamic/no-unchecked-cast"
		write("internal/lower/cast_proof.go", []byte("package retired\n"))
	default:
		t.Fatalf("unknown mutant %s", mutation)
	}
	problems, err := checkMirrors(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, problem := range problems {
		if strings.Contains(problem, expected) && strings.Contains(problem, action) {
			t.Logf("caught %s: %s", mutation, problem)
			return
		}
	}
	t.Fatalf("mutant %s survived: %v", mutation, problems)
}

func TestUnlistedMirrorHeader(t *testing.T)     { t.Parallel(); mirrorMutant(t, "header") }
func TestUnlistedTestMirrorHeader(t *testing.T) { t.Parallel(); mirrorMutant(t, "test header") }
func TestUnlistedMirrorRule(t *testing.T)       { t.Parallel(); mirrorMutant(t, "rule") }
func TestAssembledMirrorRule(t *testing.T)      { t.Parallel(); mirrorMutant(t, "assembled rule") }
func TestStaleMirrorHeader(t *testing.T)        { t.Parallel(); mirrorMutant(t, "stale header") }
func TestMissingMirrorFile(t *testing.T)        { t.Parallel(); mirrorMutant(t, "missing file") }
func TestStaleMirrorRule(t *testing.T)          { t.Parallel(); mirrorMutant(t, "stale rule") }

func TestUnlistedAdamicMirrorHeader(t *testing.T) { t.Parallel(); mirrorMutant(t, "adamic header") }

func TestUnlistedTestMirrorRule(t *testing.T) { t.Parallel(); mirrorMutant(t, "test rule") }
func TestUnlistedCoreMirrorRule(t *testing.T) { t.Parallel(); mirrorMutant(t, "core rule") }
