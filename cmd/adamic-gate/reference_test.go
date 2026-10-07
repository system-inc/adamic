package main

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type plainMember struct {
	name, text string
	kind       byte
}

func writePlainArchive(t *testing.T, archive string, members []plainMember) {
	t.Helper()
	f, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	z := gzip.NewWriter(f)
	w := tar.NewWriter(z)
	for _, member := range members {
		kind := member.kind
		if kind == 0 {
			kind = tar.TypeReg
		}
		size := int64(len(member.text))
		if kind != tar.TypeReg {
			size = 0
		}
		if err := w.WriteHeader(&tar.Header{Name: member.name, Typeflag: kind, Mode: 0600, Size: size, Linkname: "outside"}); err != nil {
			t.Fatal(err)
		}
		if size > 0 {
			if _, err := w.Write([]byte(member.text)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

// Not parallel: compare fetches relative to the process working directory.
func TestCompareGateLogsReference(t *testing.T) {
	remote := t.TempDir()
	gitFixture(t, remote, "init", "-q")
	gitFixture(t, remote, "config", "user.name", "Gate reference fixture")
	gitFixture(t, remote, "config", "user.email", "gate@example.invalid")
	gitFixture(t, remote, "commit", "--allow-empty", "-qm", "Tested tree")
	commit := gitFixture(t, remote, "rev-parse", "HEAD")
	branch := "gate-logs/" + commit[:12] + "/plain"
	gitFixture(t, remote, "checkout", "-qb", branch)
	log := `{"Action":"run","Package":"p","Test":"TestPass"}
{"Action":"pass","Package":"p","Test":"TestPass"}
{"Action":"run","Package":"p","Test":"TestSkip"}
{"Action":"skip","Package":"p","Test":"TestSkip"}
{"Action":"pass","Package":"p"}
`
	update := func(notes, contents string) {
		t.Helper()
		writePlainArchive(t, filepath.Join(remote, "plain.tgz"), []plainMember{{name: "gate-out/test.jsonl", text: contents}, {name: "gate-out/run-notes.txt", text: notes}})
		gitFixture(t, remote, "add", "plain.tgz")
		gitFixture(t, remote, "commit", "-qm", "Reference evidence")
	}
	update("commit="+commit+"\ngo=go version go1.27.1 linux/amd64\nnode=v24.19.0\n", log)
	checkout := t.TempDir()
	gitFixture(t, checkout, "init", "-q")
	gitFixture(t, checkout, "remote", "add", "origin", remote)
	t.Chdir(checkout)
	root := t.TempDir()
	local := filepath.Join(root, "test.jsonl")
	if err := os.WriteFile(local, []byte(log), 0600); err != nil {
		t.Fatal(err)
	}
	results, _, _, err := readLog(local)
	if err != nil {
		t.Fatal(err)
	}
	m := merged{Green: true, Plan: plan{Commit: commit, GoVersion: "go version go1.27.1 linux/amd64", NodeVersion: "v24.19.0"}, Results: results, Pass: 1, Skip: 1, TestEvents: 2}
	if err := saveJSON(filepath.Join(root, "merged.json"), m); err != nil {
		t.Fatal(err)
	}
	reference := "git:origin:" + branch
	if err := run([]string{"compare", root, reference}); err != nil {
		t.Fatal(err)
	}
	// The same ref now contains a wrongly labelled archive. A fresh fetch and
	// the independently recorded tested SHA must refuse it.
	update("commit="+strings.Repeat("f", 40)+"\n", log)
	if err := compare(root, reference); err == nil || !strings.Contains(err.Error(), "plain tested commit") {
		t.Fatal("different tested SHA accepted", err)
	}
	update("commit="+commit+"\ngo=go version go1.27.1 linux/amd64\nnode=v24.19.0\n", strings.Replace(log, `"Action":"pass","Package":"p","Test":"TestPass"`, `"Action":"fail","Package":"p","Test":"TestPass"`, 1))
	if err := compare(root, reference); err == nil || !strings.Contains(err.Error(), "compare failed") {
		t.Fatal("different terminal verdict accepted", err)
	}
	if _, _, err := fetchPlainReference("git:origin:gate-logs/0000000/plain", commit); err == nil {
		t.Fatal("different branch SHA accepted")
	}
}

func TestPlainArchiveRefusesMissingDuplicateAndLinkedEvidence(t *testing.T) {
	t.Parallel()
	commit := strings.Repeat("a", 40)
	log := plainMember{name: "gate-out/test.jsonl", text: "{}\n"}
	notes := plainMember{name: "gate-out/run-notes.txt", text: "commit=" + commit + "\n"}
	for _, fixture := range []struct {
		name    string
		members []plainMember
	}{
		{"missing-notes", []plainMember{log}},
		{"missing-log", []plainMember{notes}},
		{"duplicate-log", []plainMember{log, log, notes}},
		{"symlink-log", []plainMember{{name: log.name, kind: tar.TypeSymlink}, notes}},
		{"duplicate-commit", []plainMember{log, {name: notes.name, text: notes.text + notes.text}}},
		{"wrong-commit", []plainMember{log, {name: notes.name, text: "commit=" + strings.Repeat("b", 40) + "\n"}}},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			archive := filepath.Join(root, "plain.tgz")
			writePlainArchive(t, archive, fixture.members)
			if _, err := extractPlainReference(archive, root, commit); err == nil {
				t.Fatal("invalid plain evidence accepted")
			}
		})
	}
}
