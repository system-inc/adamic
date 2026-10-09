package oracle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

const rmOptionsFixture = "internal/oracle/testdata/node_fs_rm_options.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{rmOptionsFixture, true, false})
}

// Probe Node's actual validator before checking that unsupported options cannot
// reach either backend. Every row starts with a real directory and retains it.
func TestNodeFSRmOptionsValidation(t *testing.T) {
	t.Parallel()
	for _, key := range []string{"recursive", "force", "maxRetries", "retryDelay"} {
		values := []string{"undefined", "null", "'yes'"}
		if key == "recursive" || key == "force" {
			values = append(values, "0")
		} else {
			values = append(values, "true", "-1", "0.5")
		}
		for _, value := range values {
			t.Run(key+"/"+value, func(t *testing.T) {
				t.Parallel()
				directory := filepath.Join(t.TempDir(), "directory")
				options := "{" + strings.ReplaceAll("recursive:true,force:true,", key+":true,", "") + key + ":" + value + "}"
				code := `const fs=require('node:fs'); const dir=` + strconv.Quote(directory) + `; fs.mkdirSync(dir); try { fs.rmSync(dir,` + options + `); console.log('removed'); } catch(e) { console.log(e.name+' '+e.code+' '+fs.existsSync(dir)); }`
				result := execute(t, "node", "-e", code)
				class, codeName := "TypeError", "ERR_INVALID_ARG_TYPE"
				if value == "-1" || value == "0.5" {
					class, codeName = "RangeError", "ERR_OUT_OF_RANGE"
				}
				want := class + " " + codeName + " true\n"
				if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != want {
					t.Fatalf("Node validator: got exit %d stdout %q stderr %q, want %q", result.exitCode, result.stdout, result.stderr, want)
				}
				// Node's declarations reject wrong types before lowering. Undefined is
				// declared legal, so it must reach a named NotYet instead of a default.
				path := filepath.Join(t.TempDir(), "main.a")
				source := "import {rmSync} from 'node:fs'; rmSync(" + strconv.Quote(directory) + "," + options + ");"
				if err := os.WriteFile(path, []byte(source), 0644); err != nil {
					t.Fatal(err)
				}
				checked, err := load.Load([]string{path})
				if err != nil {
					if value == "undefined" || value == "-1" || value == "0.5" {
						t.Fatalf("expected declared options to reach lowering: %v", err)
					}
					if !strings.Contains(err.Error(), "not assignable") {
						t.Fatalf("want a type mismatch, not an unrelated checker failure: %v", err)
					}
					t.Logf("Node %s %s; checker rejects wrong type", class, codeName)
					return
				}
				_, err = lower.Lower(context.Background(), checked)
				var missing *lower.NotYet
				if !errors.As(err, &missing) || !strings.Contains(missing.What, "rmSync") {
					t.Fatalf("Node %s %s: want named rmSync NotYet, got %v", class, codeName, err)
				}
				if value == "undefined" && (!strings.Contains(missing.What, "options."+key) || !strings.Contains(missing.What, "ERR_INVALID_ARG_TYPE")) {
					t.Fatalf("want present-undefined refusal naming %s and Node's code, got %v", key, err)
				}
				t.Logf("Node %s %s; %s", class, codeName, missing.What)
			})
		}
	}
}

func TestNodeFSRmOptionsUndefinedBinding(t *testing.T) {
	t.Parallel()
	for _, key := range []string{"recursive", "force", "maxRetries", "retryDelay"} {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.a")
			options := "{" + strings.ReplaceAll("recursive:true,force:true,", key+":true,", "") + key + ":undefined}"
			source := fmt.Sprintf("import {rmSync} from 'node:fs'; const options=%s; rmSync('missing',options);", options)
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			_, err := lowered(t, path)
			var missing *lower.NotYet
			if !errors.As(err, &missing) || !strings.Contains(missing.What, "rmSync") {
				t.Fatalf("want named rmSync refusal, got %v", err)
			}
		})
	}
}

func TestNodeFSRmForceUndefinedReproducer(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/oct9_native_p32_fs_rm_force_undefined.a")
	_, err := lowered(t, path)
	var missing *lower.NotYet
	if !errors.As(err, &missing) || !strings.Contains(missing.What, "options.force present as undefined") {
		t.Fatalf("want force undefined refusal, got %v", err)
	}
	root := t.TempDir()
	result := onNodeWith(t, inputRun{arguments: []string{root}}, path)
	if result.exitCode != 70 || string(result.stdout) != "" || !strings.Contains(string(result.stderr), `"options.force"`) || !strings.Contains(string(result.stderr), "Received undefined") {
		t.Fatalf("Node reproducer: exit %d stdout %q stderr %q", result.exitCode, result.stdout, result.stderr)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 || !entries[0].IsDir() {
		t.Fatalf("Node must retain directory: entries=%v error=%v", entries, err)
	}
	t.Log("recreated oct9 reproducer: Node rejects before removal; lowering refuses before either backend")
}

func TestNodeFSRmOptionsCounts(t *testing.T) {
	t.Parallel()
	row := counted(t, rmOptionsFixture, false, nil, false, false)
	t.Log(row)
	recorded, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(recorded), row+"\n") {
		t.Fatalf("rm options counts not recorded: %s", row)
	}
}
