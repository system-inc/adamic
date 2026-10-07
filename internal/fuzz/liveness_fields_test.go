package fuzz

import (
	"context"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// liveLabel is the line a liveness probe prints before it runs: its name, shape, kind, prelude and
// assignment.
var liveLabel = regexp.MustCompile(`^console\.log\('live[0-9]+ ([A-Za-z]+) ([A-Za-z]+) ([A-Za-z]+) ([A-Za-z]+)'\);$`)

// fieldLabel is the line a field-representation probe prints before it runs: its name and shape.
var fieldLabel = regexp.MustCompile(`^console\.log\('field[0-9]+ ([A-Za-z]+)'\);$`)

// Within forty seeds the liveness scene reads a local in every place a throw lands, takes the old
// value every way, and throws from every kind of call, and the field-representation scene writes every
// probe. Every program they're in checks and lowers. Leaving either out drops it, and neither scene
// changes what the rest of a seed's program is.
func TestLivenessAndFieldShapes(t *testing.T) {
	t.Parallel()
	liveShapesSeen := map[string]bool{}
	livePreludesSeen := map[string]bool{}
	liveAssignsSeen := map[string]bool{}
	fieldShapesSeen := map[string]bool{}
	directory := t.TempDir()
	for seed := uint64(1); seed <= 40; seed++ {
		source := Generate(seed).Source()
		for _, line := range strings.Split(source, "\n") {
			if match := liveLabel.FindStringSubmatch(line); match != nil {
				liveShapesSeen[match[1]] = true
				livePreludesSeen[match[3]] = true
				liveAssignsSeen[match[2]+" "+match[4]] = true
			}
			if match := fieldLabel.FindStringSubmatch(line); match != nil {
				fieldShapesSeen[match[1]] = true
			}
		}
		path := filepath.Join(directory, "program.a")
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		program, err := load.Load([]string{path})
		if err != nil {
			t.Errorf("seed %d: the checker refused it: %v", seed, err)
			continue
		}
		if _, err := lower.Lower(context.Background(), program); err != nil {
			t.Errorf("seed %d: stage 0 didn't lower it: %v", seed, err)
		}
	}
	for _, shape := range liveShapes() {
		if !liveShapesSeen[shape.name] {
			t.Errorf("40 seeds never read a local after a throw in the %s shape", shape.name)
		}
	}
	for _, prelude := range (&generator{random: rand.New(rand.NewPCG(1, 2))}).livePreludes() {
		if !livePreludesSeen[prelude.name] {
			t.Errorf("40 seeds never took the old value with %s", prelude.name)
		}
	}
	for _, assign := range liveAssigns() {
		if !liveAssignsSeen[assign.kind+" "+assign.name] {
			t.Errorf("40 seeds never assigned a %s from %s", assign.kind, assign.name)
		}
	}
	for _, shape := range fieldShapes() {
		if !fieldShapesSeen[shape.name] {
			t.Errorf("40 seeds never wrote the %s field probe", shape.name)
		}
	}

	if strings.Contains(GenerateWithout(1, []string{"liveness"}).Source(), "LiveTree") {
		t.Error("leaving liveness out still wrote the scene")
	}
	if strings.Contains(GenerateWithout(1, []string{"field-representation"}).Source(), "FieldBox") {
		t.Error("leaving field-representation out still wrote the scene")
	}
	// The scenes draw from streams of their own: without them, a seed's program is the one it was.
	for _, seed := range []uint64{3, 8} {
		scenes := &generator{seed: seed}
		written := &Program{Block: Block{Statements: append(scenes.livenessProgram(), scenes.fieldRepresentationProgram()...)}}
		with := Generate(seed).Source()
		if !strings.Contains(with, written.Source()) {
			t.Fatalf("seed %d doesn't write the scenes in one piece", seed)
		}
		if strings.Replace(with, written.Source(), "", 1) != GenerateWithout(seed, []string{"liveness", "field-representation"}).Source() {
			t.Errorf("adding the scenes changed the rest of seed %d's program", seed)
		}
	}
}
