package tailwind

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	"os"
	"path/filepath"
	"sync"
)

var adamicLoaderLock sync.Mutex

func loadDesignSystemThrough(program rule.Program, fs *rule.RecordingFS) DesignSystemResult {
	result := adamicOriginalLoadDesignSystemThrough(program, fs)
	output := os.Getenv("ADAMIC_SLOT04_WAVE9")
	if output == "" {
		return result
	}
	root := projectRootOf(program)
	entry := result.EntryPoint
	pkg := ""
	failed := false
	message := ""
	sheets := []string{}
	system := 0
	if entry != "" {
		pkg = findTailwindPackageRoot(filepath.Dir(entry), fs.FileExists)
	}
	if pkg != "" && result.Err != nil {
		failed = true
		message = result.Err.Error()
	}
	if result.System != nil {
		system = 1
		sheets = append(sheets, result.System.Stylesheets...)
	}
	row := map[string]any{"Kind": "load", "Root": root, "Entry": entry, "Package": pkg, "Failed": failed, "Message": message, "System": system, "Sheets": sheets}
	adamicLoaderLock.Lock()
	defer adamicLoaderLock.Unlock()
	f, err := os.OpenFile(output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(row); err != nil {
		panic(err)
	}
	return result
}
