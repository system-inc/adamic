package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestChangedTrainingByteCannotReachProfileFlags(t *testing.T) {
	directory := t.TempDir()
	training := filepath.Join(directory, "training.json")
	data, err := os.ReadFile("../../stage1/profiles/training.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(training, data, 0o644); err != nil {
		t.Fatal(err)
	}
	record := native.Stage1ProfileManifest{Training: fmt.Sprintf("%x", sha256.Sum256(data))}
	manifest, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(directory, "manifest.json"), manifest, 0o644); err != nil {
		t.Fatal(err)
	}
	profile := filepath.Join(directory, "profile.txt")
	if got, diagnostic := profileForTraining(profile, training); got != profile || diagnostic != "" {
		t.Fatalf("matching corpus rejected: %q %q", got, diagnostic)
	}
	// A byte in the real list changes while the program and runtime stay identical.
	data[len(data)-1] ^= 1
	if err = os.WriteFile(training, data, 0o644); err != nil {
		t.Fatal(err)
	}
	got, diagnostic := profileForTraining(profile, training)
	for _, flags := range [][]string{native.Flags(native.Options{Release: true, Profile: got}), native.LinkFlags(native.Options{Release: true, Profile: got})} {
		for _, flag := range flags {
			if strings.HasPrefix(flag, "-fprofile-") {
				t.Fatalf("changed training list reached profile flags: %q", flags)
			}
		}
	}
	if got != "" || !strings.Contains(diagnostic, "using plain ThinLTO") || strings.Contains(diagnostic, "\n") {
		t.Fatalf("missing single-line training fallback: %q %q", got, diagnostic)
	}
	t.Log("one changed training-list byte clears profile flags at compile and link")
}
