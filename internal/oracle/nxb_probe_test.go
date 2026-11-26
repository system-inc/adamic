package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Scratch probe for #nxb3xh1, never for main: whether an ordinary lowering makes tsc's checked view read
// of an optional array field, jsDoc.tags, that native emit panics on in trap mode.
func TestNxbProbe(t *testing.T) {
	directory := filepath.Join(repository, "internal/oracle/testdata/nxb_probe")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(directory, name))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			t.Logf("PROBE %s node exit %d stdout %q stderr %.300q", name, truth.exitCode, truth.stdout, truth.stderr)
			program, err := lowered(t, path)
			if err != nil {
				t.Logf("PROBE %s lowering: %v", name, err)
				return
			}
			t.Logf("PROBE %s ir: %s", name, nxbTagReads(program))
			native, _ := natively(t, program)
			t.Logf("PROBE %s sanitized exit %d stdout %q stderr %.600q", name, native.exitCode, native.stdout, native.stderr)
			release := released(t, program)
			t.Logf("PROBE %s release exit %d stdout %q stderr %.300q", name, release.exitCode, release.stdout, release.stderr)
			backend := onJavaScriptBackend(t, program)
			t.Logf("PROBE %s javascript exit %d stdout %q stderr %.300q", name, backend.exitCode, backend.stdout, backend.stderr)
		})
	}
}

// nxbTagReads names each read of tags in the program as lowering left it.
func nxbTagReads(program *ir.Program) string {
	found := []string{}
	var visit func(reflect.Value)
	visit = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}
		if value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return
			}
			if value.Kind() == reflect.Interface {
				if read, ok := value.Elem().Interface().(ir.Property); ok && read.Name == "tags" {
					found = append(found, fmt.Sprintf("Property{Of %d, Absent %t, Optional %t, View %q, ViewType %q, ViewContract %d, Object %T}", read.Of, read.Absent, read.Optional, read.View, read.ViewType, read.ViewContract, read.Object))
				}
			}
			visit(value.Elem())
			return
		}
		switch value.Kind() {
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				if value.Type().Field(index).IsExported() {
					visit(value.Field(index))
				}
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				visit(value.Index(index))
			}
		}
	}
	visit(reflect.ValueOf(program.Main))
	for _, function := range program.Functions {
		visit(reflect.ValueOf(function.Body))
	}
	return strings.Join(found, "; ")
}
