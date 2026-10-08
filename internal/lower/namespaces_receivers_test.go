package lower

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNamespaceReceiverImportedEscape(t *testing.T) {
	folder := t.TempDir()
	for name, source := range map[string]string{
		"values.a": "export namespace Debug {export const x:number=1;export function read(this:{x:number}):number{return this.x;}} const escaped=Debug.read;",
		"main.a":   "import {Debug} from './values.a';console.log(`${Debug.read()}`);",
	} {
		if err := os.WriteFile(filepath.Join(folder, name), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	program, err := load.Load([]string{filepath.Join(folder, "main.a")})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(context.Background(), program)
	var ny *NotYet
	if !errors.As(err, &ny) || !strings.Contains(err.Error(), "closed-world receiver") {
		t.Fatalf("imported receiver escape lost: %v", err)
	}
}
