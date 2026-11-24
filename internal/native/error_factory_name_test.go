package native

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestErrorFactoryNameStates pins the Error layouts and their factory name
// states. The SystemError-unwritten plant removes the one explicit own state.
func TestErrorFactoryNameStates(t *testing.T) {
	t.Parallel()
	type factory struct {
		scope string
		make  string
		name  string
		own   bool
	}
	// Error and its named built-in constructors share adamic_error_new.
	// Buffer's coded helper receives the name from its two callers.
	// Both node_system_error callers inherit Error.prototype.name.
	rows := map[string][]factory{
		"exceptions.c  error_shape": {
			{"adamic_error_new", "adamic_object_new(message==NULL?&absent_message_shape:&error_shape)", "Error or built-in constructor name", false},
		},
		"exceptions.c  absent_message_shape": {
			{"adamic_error_new", "adamic_object_new(message==NULL?&absent_message_shape:&error_shape)", "Error or built-in constructor name", false},
		},
		"exceptions.c  coded_error_shape": {
			{"adamic_throw_coded", "adamic_object_new(&coded_error_shape)", "RangeError from node_buffer.c (ERR_OUT_OF_RANGE)", false},
		},
		"exceptions.c  plain_coded_error_shape": {
			{"adamic_throw_coded", "adamic_object_new(&plain_coded_error_shape)", "RangeError from node_buffer.c (ERR_MEMORY_ALLOCATION_FAILED)", false},
		},
		"node_fs_directory.c  error_shape": {
			{"adamic_node_fs_raise", "adamic_object_new(&error_shape)", "Error", false},
		},
		"node_fs_directory.c  coded_error_shape": {
			{"valid", "adamic_object_new(&coded_error_shape)", "TypeError", false},
		},
		"node_fs_file.c  error_shape": {
			{"raise_error", "adamic_object_new(shape)", "Error or RangeError from caller", false},
			{"adamic_fs_file_host_require_missing", "adamic_object_new(&error_shape)", "Error", false},
			{"system_error", "raise_error(&error_shape,\"Error\",code,message)", "Error", false},
			{"read_file", `raise_error(&error_shape,"Error","ERR_STRING_TOO_LONG",message)`, "Error", false},
			{"integer", `raise_error(shape,"RangeError","ERR_OUT_OF_RANGE",message)`, "RangeError", false},
			{"read_range", `raise_error(shape,"RangeError","ERR_OUT_OF_RANGE",message)`, "RangeError", false},
			{"remove_error", "raise_error(&error_shape,\"Error\",code,message)", "Error", false},
		},
		"node_fs_file.c  coded_error_shape": {
			{"raise_error", "adamic_object_new(shape)", "TypeError, RangeError, Error or SystemError from caller", false},
			{"invalid_value", "adamic_object_new(&coded_error_shape)", "TypeError", false},
			{"adamic_fs_file_watch_interval", `raise_error(&coded_error_shape,"TypeError","ERR_INVALID_ARG_TYPE",`, "TypeError", false},
			{"adamic_fs_file_read_sync", `raise_error(&coded_error_shape,"TypeError","ERR_INVALID_ARG_VALUE",`, "TypeError", false},
			{"fs_utimes", `raise_error(&coded_error_shape,"TypeError","ERR_INVALID_ARG_TYPE",message)`, "TypeError", false},
			{"adamic_fs_file_host_inspector_unavailable", `raise_error(&coded_error_shape,"Error","ERR_INSPECTOR_NOT_AVAILABLE",`, "Error", false},
			{"remove_path", "raise_error(&coded_error_shape,\"SystemError\",\"ERR_FS_EISDIR\",message)", "SystemError", true},
		},
		"node_fs_file.c realpath_system_error shape": {
			{"realpath_system_error", "node_system_error(error,operation,name,NULL,NULL,path,&shape)", "Error", false},
			{"node_system_error", "adamic_object_new(shape)", "Error", false},
		},
		"node_fs_watch_impl.h adamic_fs_file_host_watch watch_error_shape": {
			{"adamic_fs_file_host_watch", "node_system_error(errno,\"watch\",name,name,NULL,path,&watch_error_shape)", "Error", false},
			{"adamic_fs_file_host_watch", "node_system_error(error,\"watch\",name,name,message,path,&watch_error_shape)", "Error", false},
			{"adamic_fs_file_host_watch", "node_system_error(error,\"watch\",name,name,NULL,path,&watch_error_shape)", "Error", false},
		},
		"node_process.c host_directory_error shape": {
			{"host_directory_error", "adamic_object_new(&shape)", "Error", false},
		},
		"process.c valid_code shape": {
			{"valid_code", "adamic_object_new(&shape)", "RangeError", false},
		},
	}
	shapes, _ := scannedHostShapes(t)
	sources := map[string]map[string]string{}
	allocations := map[string]map[string]bool{}
	for _, shape := range shapes {
		if !shape.errorShape {
			continue
		}
		file := filepath.Base(shape.file)
		key := file + " " + shape.scope + " " + shape.name
		factories, listed := rows[key]
		if !listed {
			t.Errorf("%s:%d: unlisted Error shape %s", shape.file, shape.line, shape.name)
			continue
		}
		delete(rows, key)
		if sources[file] == nil {
			source, err := os.ReadFile(filepath.Join("..", "..", shape.file))
			if err != nil {
				t.Fatal(err)
			}
			tokens, err := cTokens(strings.ReplaceAll(string(source), "\\\n", "  "))
			if err != nil {
				t.Fatal(err)
			}
			sources[file] = map[string]string{}
			for i := 0; i < len(tokens); i++ {
				token := tokens[i]
				if token.text != "{" {
					continue
				}
				end := closingBrace(tokens, i)
				if end < 0 {
					t.Fatalf("%s: unbalanced braces", file)
				}
				if scope := braceScope(tokens, i); scope != "" {
					var body strings.Builder
					for _, token := range tokens[i+1 : end] {
						if token.text == "raise_error" || token.text == "node_system_error" {
							body.WriteByte(' ')
						}
						body.WriteString(token.text)
					}
					sources[file][scope] += body.String()
				}
				// The body includes all nested calls.
				i = end
			}
		}
		// A file-scope shape can be allocated by any function in its file.
		// A local shape is visible only in its declaring function.
		direct := "adamic_object_new(&" + shape.name + ")"
		for scope, body := range sources[file] {
			if shape.scope != "" && scope != shape.scope {
				continue
			}
			want := 0
			for _, factory := range factories {
				if factory.scope == scope && factory.make == direct {
					want++
				}
			}
			if got := strings.Count(body, direct); got != want {
				t.Errorf("%s %s: %s has %d direct allocations, audit lists %d", file, scope, shape.name, got, want)
			}
		}
		if allocations[file] == nil {
			allocations[file] = map[string]bool{}
		}
		for _, factory := range factories {
			allocations[file][factory.scope+" "+factory.make] = true
			body := sources[file][factory.scope]
			if factory.own {
				const lookup = `name_slot=adamic_object_optional_find(adamic_thrown,"name",&cache);`
				const write = "adamic_object_field_types(adamic_thrown)[name_slot-adamic_thrown->slots]=adamic_error_own_hidden;"
				if find, store := strings.Index(body, lookup), strings.Index(body, write); find < 0 || store <= find {
					t.Errorf("%s %s: factory for %s must mark the slot found by name", file, factory.scope, factory.name)
				}
			}
			if strings.Count(body, factory.make) != 1 ||
				strings.Contains(body, "=adamic_error_own_hidden") != factory.own {
				t.Errorf("%s %s: factory for %s must occur once and choose own hidden=%t: %s", file, factory.scope, factory.name, factory.own, factory.make)
			}
		}
	}
	// Count calls to the two forwarding helpers separately. A new caller has
	// to get its own row even when it uses an already audited shape.
	for file, bodies := range sources {
		for _, helper := range []string{"raise_error", "node_system_error"} {
			want := 0
			for call := range allocations[file] {
				if strings.Contains(call, " "+helper+"(") {
					want++
				}
			}
			got := 0
			for _, body := range bodies {
				got += strings.Count(body, " "+helper+"(")
			}
			if got != want {
				t.Errorf("%s: %s has %d calls, audit lists %d", file, helper, got, want)
			}
		}
	}
	buffer, err := os.ReadFile(filepath.Join("..", "..", hostShapeRuntime, "node_buffer.c"))
	if err != nil {
		t.Fatal(err)
	}
	bufferSources := string(buffer)
	if strings.Count(bufferSources, "adamic_throw_coded(") != 2 ||
		!strings.Contains(bufferSources, `adamic_throw_coded("RangeError", "ERR_OUT_OF_RANGE",`) ||
		!strings.Contains(bufferSources, `adamic_throw_coded("RangeError", "ERR_MEMORY_ALLOCATION_FAILED",`) {
		t.Error("node_buffer.c: coded RangeError factory callers changed")
	}
	for key := range rows {
		t.Errorf("factory audit lists an absent shape: %s", key)
	}
}
