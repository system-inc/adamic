package compilercensus

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// nativeBuilds are internal/native's functions that run clang under the process's jobserver (native.UseJobserver).
var nativeBuilds = []string{"Build", "BuildFlags", "BuildTSGo", "BuildSplitTSGo", "RuntimeLibrary", "RuntimeLibraryForSource", "RuntimeRequestFor"}

const (
	nativePath      = "github.com/system-inc/adamic/internal/native"
	ccjobserverPath = "github.com/system-inc/adamic/internal/ccjobserver"
)

// Every package that builds C through native imports internal/ccjobserver, so its builds compile under the jobserver
// Loom names (#yzwm4rj): native reads no environment for it, so a package that calls a native build without the import
// would build unbounded under Loom. A call from a package's own code needs the import in its own code, since its
// binary links no test file; a call from its tests needs it in any of its files. A package whose code imports a
// package that imports it is bounded too, by that package's init, so only direct callers are held. Mutant: the
// import taken out of a package that builds (internal/oracle's).
func TestEveryNativeBuildIsBoundedByTheProcessJobserver(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	type directory struct {
		callsCode, callsTest     []string
		importsCode, importsTest bool
	}
	directories := map[string]*directory{}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(root, path)
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			if relative == "cohere" || relative == "internal/ccjobserver" || entry.Name() == "testdata" || entry.Name() == "node_modules" || strings.HasPrefix(entry.Name(), ".") && relative != "." {
				return filepath.SkipDir
			}
			return nil
		}
		// native's own files build through its own functions, under the jobserver they're handed or the process's.
		if !strings.HasSuffix(relative, ".go") || filepath.ToSlash(filepath.Dir(relative)) == "internal/native" {
			return nil
		}
		files := token.NewFileSet()
		parsed, err := parser.ParseFile(files, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		test := strings.HasSuffix(relative, "_test.go")
		place := filepath.ToSlash(filepath.Dir(relative))
		if directories[place] == nil {
			directories[place] = &directory{}
		}
		found := directories[place]
		nativeName := ""
		for _, imported := range parsed.Imports {
			importPath, _ := strconv.Unquote(imported.Path.Value)
			switch importPath {
			case ccjobserverPath:
				if test {
					found.importsTest = true
				} else {
					found.importsCode = true
				}
			case nativePath:
				nativeName = "native"
				if imported.Name != nil {
					nativeName = imported.Name.Name
				}
			}
		}
		if nativeName == "" {
			return nil
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if name, ok := selector.X.(*ast.Ident); ok && name.Name == nativeName && slices.Contains(nativeBuilds, selector.Sel.Name) {
				site := files.Position(selector.Pos()).String()
				site = strings.TrimPrefix(site, root+string(filepath.Separator)) + " native." + selector.Sel.Name
				if test {
					found.callsTest = append(found.callsTest, site)
				} else {
					found.callsCode = append(found.callsCode, site)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	places := make([]string, 0, len(directories))
	for place := range directories {
		places = append(places, place)
	}
	sort.Strings(places)
	builders := 0
	for _, place := range places {
		found := directories[place]
		switch {
		case len(found.callsCode) > 0 && !found.importsCode:
			t.Errorf("%s builds C through native in its own code (%s) and its code doesn't import %s, so Loom's jobserver doesn't bound it", place, found.callsCode[0], ccjobserverPath)
		case len(found.callsTest) > 0 && !found.importsCode && !found.importsTest:
			t.Errorf("%s's tests build C through native (%s) and none of its files imports %s, so Loom's jobserver doesn't bound them", place, found.callsTest[0], ccjobserverPath)
		}
		if len(found.callsCode)+len(found.callsTest) > 0 {
			builders++
		}
	}
	// The walk must find the builders it holds, or the census holds nothing.
	if builders < 10 {
		t.Fatalf("only %d packages build C through native: the census isn't reading the tree", builders)
	}
}
