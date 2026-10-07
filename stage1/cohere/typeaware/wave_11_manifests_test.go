package typeaware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type wave11Manifest struct {
	Name  string `json:"name"`
	Kinds []int  `json:"kinds"`
}

func wave11ManifestBytes(directory string) ([]byte, error) {
	paths, err := filepath.Glob(filepath.Join(directory, "*", "rule.json"))
	if err != nil {
		return nil, err
	}
	if len(paths) != 15 {
		return nil, fmt.Errorf("expected fifteen rule manifests, got %d", len(paths))
	}
	var lines []string
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var manifest wave11Manifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, err
		}
		sort.Ints(manifest.Kinds)
		values := make([]string, len(manifest.Kinds))
		for index, kind := range manifest.Kinds {
			values[index] = strconv.Itoa(kind)
		}
		lines = append(lines, manifest.Name+"\t"+strings.Join(values, ","))
	}
	sort.Strings(lines)
	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

func wave11CheckManifests(h *harness, truth []byte) {
	h.t.Helper()
	directory := filepath.Join(h.repository, "stage1/cohere/typeaware/wave-11-rules")
	actual, err := wave11ManifestBytes(directory)
	if err != nil {
		h.t.Fatal(err)
	}
	if !bytes.Equal(actual, truth) {
		h.t.Fatal("rule.json kinds differ from production Go registrations")
	}
	mutantDirectory := filepath.Join(h.directory, "manifest-mutant")
	paths, err := filepath.Glob(filepath.Join(directory, "*", "rule.json"))
	if err != nil {
		h.t.Fatal(err)
	}
	for index, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		if index == 0 {
			var manifest wave11Manifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				h.t.Fatal(err)
			}
			manifest.Kinds = nil
			data, err = json.Marshal(manifest)
			if err != nil {
				h.t.Fatal(err)
			}
		}
		target := filepath.Join(mutantDirectory, filepath.Base(filepath.Dir(path)))
		if err := os.MkdirAll(target, 0755); err != nil {
			h.t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(target, "rule.json"), data, 0644); err != nil {
			h.t.Fatal(err)
		}
	}
	changed, err := wave11ManifestBytes(mutantDirectory)
	if err != nil {
		h.t.Fatal(err)
	}
	if bytes.Equal(changed, truth) {
		h.t.Fatal("missing JSON kinds mutant survived")
	}
	h.t.Logf("manifest-missing-kinds: valid JSON, production Go comparison catches byte %d", firstDifference(changed, truth))
}
