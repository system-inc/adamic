package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestComplementRunsInitRegisteredFixture(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "go.mod"), []byte("module probe\n\ngo 1.27.1\n"), 0600)
	os.WriteFile(filepath.Join(root, "parent_test.go"), []byte(`package probe
import "testing"
var fixtures=[]string{"left/a", "right/b"}
func TestParent(t *testing.T) {for _,name:=range fixtures {t.Run(name,func(t *testing.T){})}}
`), 0600)
	os.WriteFile(filepath.Join(root, "extra_test.go"), []byte(`package probe
func init() {for _,name:=range []string{"left/b", "right/a", "new/fixture"} {fixtures=append(fixtures,name)}}
`), 0600)
	p := plan{Count: 2, Units: []unit{{Package: "probe", Test: "TestParent/left/a", Shard: 0}, {Package: "probe", Test: "TestParent/right/b", Shard: 0}}}
	p.Complements = complements(p)
	for index := 0; index < 2; index++ {
		selections := selections(p, index, "probe")
		if len(selections) != 1 {
			t.Fatal(selections)
		}
		args := selectionArgs(".", selections[0])
		cmd := exec.Command("go", args...)
		cmd.Dir = root
		log := filepath.Join(root, "shard-"+string(rune('0'+index))+".jsonl")
		file, err := os.Create(log)
		if err != nil {
			t.Fatal(err)
		}
		cmd.Stdout = file
		cmd.Stderr = file
		err = cmd.Run()
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
		rows, _, _, err := readLog(log)
		if err != nil {
			t.Fatal(err)
		}
		var children []string
		for _, r := range rows {
			if strings.Contains(r.Test, "/") {
				children = append(children, r.Test)
			}
		}
		expected := 2
		if index == 1 {
			expected = 3
		}
		if len(children) != expected {
			t.Fatalf("shard %d: %v", index, children)
		}
		if problems := validateResults(p, index, rows, map[string]bool{}, map[string]int{}); len(problems) > 0 {
			t.Fatal(problems)
		}
		if index == 1 && !strings.Contains(strings.Join(children, ","), "new/fixture") {
			t.Fatal("init fixture vanished")
		}
	}
}

func TestComplementRequiredAndDuplicatesRefused(t *testing.T) {
	if schedulingIdentity("context", 1) == schedulingIdentity("context", 4) {
		t.Fatal("package concurrency omitted from resume identity")
	}
	p := plan{Count: 2, Units: []unit{{Package: "probe", Test: "TestParent/known", Shard: 0}}}
	p.Complements = complements(p)
	r := result{Package: "probe", Test: "TestParent/new", Action: "pass"}
	seen := map[string]int{}
	if problems := validateResults(p, 1, []result{r}, map[string]bool{}, seen); len(problems) != 0 {
		t.Fatal(problems)
	}
	if problems := validateResults(p, 1, []result{r}, map[string]bool{}, seen); len(problems) == 0 {
		t.Fatal("duplicate complement child accepted")
	}
	if problems := validateResults(p, 0, []result{r}, map[string]bool{}, map[string]int{}); len(problems) == 0 {
		t.Fatal("unplanned child accepted outside complement")
	}
	keys := selectorKeys(p, 1, "probe")
	q := p
	q.Units = append(append([]unit{}, p.Units...), unit{Package: "probe", Test: "TestParent/other", Shard: 0})
	if strings.Join(keys, "") == strings.Join(selectorKeys(q, 1, "probe"), "") {
		t.Fatal("skip exclusions not keyed")
	}
	data, _ := json.Marshal(selections(p, 1, "probe"))
	if !strings.Contains(string(data), "Skip") {
		t.Fatal("complement not in plan commands")
	}
}

func TestComplementCheckpointRequiresParentTerminal(t *testing.T) {
	dir := t.TempDir()
	p := plan{Count: 2, Units: []unit{{Package: "probe", Test: "TestParent/known", Shard: 0}}}
	p.Complements = complements(p)
	log := `{"Action":"run","Package":"probe","Test":"TestParent/new"}
{"Action":"pass","Package":"probe","Test":"TestParent/new"}
{"Action":"pass","Package":"probe"}
`
	if err := os.WriteFile(filepath.Join(dir, "test.jsonl"), []byte(log), 0600); err != nil {
		t.Fatal(err)
	}
	e := packageEvidence{Package: "probe", Invocations: []invocation{{Package: "probe", Args: selectionArgs("probe", selections(p, 1, "probe")[0]), Uncached: true}}}
	err := completePackage(dir, e, p, 1, "probe")
	if err == nil || !strings.Contains(err.Error(), "complement parent has no terminal event") {
		t.Fatal("complement without parent terminal accepted", err)
	}
}
