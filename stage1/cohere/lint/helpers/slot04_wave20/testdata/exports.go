package structure

import (
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"sync"
)

func AdamicSearch(node *ast.Node, depth int) bool { return searchForJsxOrHook(node, depth) }

func AdamicDescends(node *ast.Node) bool { return descendsForJsxSearch(node) }

var adamicLock sync.Mutex

func AdamicRecord(node *ast.Node, helper string, depth int) {
	path := os.Getenv("ADAMIC_SLOT04_SEARCH")
	if path == "" {
		return
	}
	kind := "nil"
	if node != nil {
		kind = node.Kind.String()
	}
	row := struct {
		Helper, Kind string
		Depth        int
	}{helper, kind, depth}
	adamicLock.Lock()
	defer adamicLock.Unlock()
	f, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	if e = json.NewEncoder(f).Encode(row); e != nil {
		panic(e)
	}
}
