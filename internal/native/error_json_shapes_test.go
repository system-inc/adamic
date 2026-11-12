package native

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

type errorJSONRow struct {
	file, scope, name     string
	complete              bool
	keys, measured, probe string
}

// Every error shape is either complete or stops, with the Node measurement that names its keys.
// Factory variants sharing an incomplete shape remain stopped even if this one probe looks simple.
var errorJSONRows = []errorJSONRow{
	{"exceptions.c", "", "error_shape", true, "", "", "throw new Error('x')"},
	{"exceptions.c", "", "absent_message_shape", true, "", "", "throw new Error()"},
	{"process.c", "valid_code", "shape", true, "code", "code", "process.exitCode=NaN"},
	{"node_fs_directory.c", "", "coded_error_shape", true, "code", "code", "require('node:fs').symlinkSync('x\\0y','/__adamic_error_json_missing__/link')"},
	{"node_fs_file.c", "realpath_system_error", "shape", true, "errno,code,syscall,path", "errno,code,syscall,path", "require('node:fs').realpathSync.native('/__adamic_error_json_missing__/entry')"},
	{"node_fs_watch_impl.h", "adamic_fs_file_host_watch", "watch_error_shape", true, "errno,syscall,code,path,filename", "errno,syscall,code,path,filename", "require('node:fs').watch('/__adamic_error_json_missing__/entry',()=>{})"},
	{"node_fs_directory.c", "", "error_shape", false, "", "errno,code,syscall,path", "require('node:fs').readdirSync('/__adamic_error_json_missing__/entry')"},
	{"node_fs_file.c", "", "error_shape", false, "", "errno,code,syscall,path", "require('node:fs').openSync('/__adamic_error_json_missing__/entry')"},
	{"node_fs_file.c", "", "coded_error_shape", false, "", "code", "require('node:fs').openSync('/tmp','bad-flag')"},
	{"node_process.c", "host_directory_error", "shape", false, "", "errno,code,syscall,path,dest", "process.chdir('/__adamic_error_json_missing__/entry')"},
}

// A new error shape, a missing registration, a changed key list, or a Node order change must fail here.
func TestErrorJSONShapeClosure(t *testing.T) {
	t.Parallel()
	shapes, unknown := scannedHostShapes(t)
	for _, problem := range unknown {
		t.Error(problem)
	}
	rows := map[string]errorJSONRow{}
	for _, row := range errorJSONRows {
		key := fmt.Sprintf("%s%s %s %s[0]", hostShapeRuntime, row.file, row.scope, row.name)
		if _, duplicate := rows[key]; duplicate {
			t.Fatalf("duplicate row %s", key)
		}
		rows[key] = row
	}
	seen := map[string]bool{}
	for _, shape := range shapes {
		if !shape.errorShape {
			continue
		}
		row, ok := rows[shape.key()]
		if !ok {
			t.Errorf("%s has no error key row", shape.key())
			continue
		}
		seen[shape.key()] = true
		t.Run(shape.key(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("runtime", row.file))
			if err != nil {
				t.Fatal(err)
			}
			source := string(data)
			pattern := `static\s+adamic_error_json\s+` + regexp.QuoteMeta(row.name) + `_json\s*=\s*\{\s*&` + regexp.QuoteMeta(row.name) + `\s*,\s*([^,]+),\s*([0-9]+),\s*NULL,\s*false\s*\}`
			match := regexp.MustCompile(pattern).FindStringSubmatch(source)
			if !row.complete {
				if match != nil {
					t.Fatal("stopping shape now has an entry; remeasure and classify it")
				}
			} else {
				if match == nil {
					t.Fatal("complete shape has no key entry")
				}
				if !strings.Contains(source, "adamic_register_error_json(&"+row.name+"_json)") {
					t.Fatal("entry is not registered")
				}
				keys := []string{}
				if row.keys != "" {
					keys = strings.Split(row.keys, ",")
				}
				if match[2] != strconv.Itoa(len(keys)) {
					t.Fatalf("entry count %s, want %d", match[2], len(keys))
				}
				if len(keys) == 0 {
					if strings.TrimSpace(match[1]) != "NULL" {
						t.Fatal("empty entry has a key array")
					}
				} else {
					keyPattern := regexp.QuoteMeta(strings.TrimSpace(match[1])) + `\[\]\s*=\s*\{([^}]+)\}`
					array := regexp.MustCompile(keyPattern).FindStringSubmatch(source)
					if array == nil {
						t.Fatal("no key array")
					}
					actual := []string{}
					for _, quoted := range regexp.MustCompile(`"[^"\\]*"`).FindAllString(array[1], -1) {
						key, err := strconv.Unquote(quoted)
						if err != nil {
							t.Fatal(err)
						}
						actual = append(actual, key)
					}
					if strings.Join(actual, ",") != row.keys {
						t.Fatalf("entry keys %q, want %q", actual, keys)
					}
				}
			}
			script := `if(process.version!=='v24.19.0')throw Error('unexpected Node version');try {` + row.probe + `;console.log('no error')} catch(e) {console.log(Object.keys(e).join(','))}`
			if got := runWithInput(t, "", "node", "--eval", script); got != row.measured+"\n" {
				t.Fatalf("Node keys %q, measured %q", got, row.measured+"\n")
			}
		})
	}
	for key := range rows {
		if !seen[key] {
			t.Errorf("error key row %s has no shape", key)
		}
	}
}
