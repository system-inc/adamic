package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/system-inc/adamic/internal/boundedrun"
)

var plainBranch = regexp.MustCompile(`^gate-logs/([0-9a-f]{7,40})/plain$`)
var referenceRemote = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.-]*$`)
var fullCommit = regexp.MustCompile(`^[0-9a-f]{40}$`)

// Fetch fresh evidence without checking out the branch or reusing an archive.
// The archive must record the tested SHA independently of its branch name.
func fetchPlainReference(reference, commit string) (string, func(), error) {
	parts := strings.SplitN(reference, ":", 3)
	if len(parts) != 3 || parts[0] != "git" || !referenceRemote.MatchString(parts[1]) {
		return "", nil, errors.New("plain reference must be git:<remote>:gate-logs/<sha>/plain")
	}
	match := plainBranch.FindStringSubmatch(parts[2])
	if match == nil || !fullCommit.MatchString(commit) || !strings.HasPrefix(commit, match[1]) {
		return "", nil, fmt.Errorf("plain reference %q does not name merged commit %s", reference, commit)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", nil, err
	}
	if err := os.MkdirAll(cache, 0755); err != nil {
		return "", nil, err
	}
	root, err := os.MkdirTemp(cache, "adamic-plain-reference-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { os.RemoveAll(root) }
	failed := true
	defer func() {
		if failed {
			cleanup()
		}
	}()
	cmd, release := boundedrun.Command(boundedrun.Build, "git", "fetch", "--no-tags", "--", parts[1], "refs/heads/"+parts[2])
	var diagnostic commandBuffer
	cmd.Stdout, cmd.Stderr = &diagnostic, &diagnostic
	err = cmd.Run()
	release()
	if err != nil {
		return "", nil, fmt.Errorf("fetch plain reference: %w\n%s", err, diagnostic.String())
	}
	object, err := output("git", "rev-parse", "FETCH_HEAD^{commit}")
	if err != nil {
		return "", nil, err
	}
	archive := filepath.Join(root, "plain.tgz")
	f, err := os.Create(archive)
	if err != nil {
		return "", nil, err
	}
	cmd, release = boundedrun.Command(boundedrun.Build, "git", "show", object+":plain.tgz")
	cmd.Stdout, cmd.Stderr = f, &diagnostic
	err = errors.Join(cmd.Run(), f.Close())
	release()
	if err != nil {
		return "", nil, fmt.Errorf("read plain.tgz: %w\n%s", err, diagnostic.String())
	}
	log, err := extractPlainReference(archive, root, commit)
	if err != nil {
		return "", nil, err
	}
	fmt.Printf("plain reference=%s evidence_commit=%s tested_commit=%s\n", reference, object, commit)
	failed = false
	return log, cleanup, nil
}

// Only copy the known regular files, never arbitrary archive paths or links.
func extractPlainReference(archive, root, commit string) (string, error) {
	f, err := os.Open(archive)
	if err != nil {
		return "", err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer z.Close()
	reader := tar.NewReader(z)
	log := filepath.Join(root, "test.jsonl")
	seen := map[string]bool{}
	notes := ""
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		name := strings.TrimPrefix(header.Name, "./")
		if name != "gate-out/test.jsonl" && name != "gate-out/run-notes.txt" {
			continue
		}
		if seen[name] || header.Typeflag != tar.TypeReg || header.Size < 0 || header.Size > 512*1024*1024 {
			return "", fmt.Errorf("invalid or duplicate plain archive member %s", name)
		}
		seen[name] = true
		if name == "gate-out/run-notes.txt" {
			if header.Size > 1024*1024 {
				return "", errors.New("plain run notes exceed 1 MiB")
			}
			data, err := io.ReadAll(reader)
			if err != nil {
				return "", err
			}
			notes = string(data)
		} else {
			out, err := os.Create(log)
			if err != nil {
				return "", err
			}
			_, copyErr := io.Copy(out, reader)
			if err := errors.Join(copyErr, out.Close()); err != nil {
				return "", err
			}
		}
	}
	// Check the gzip trailer too, rather than stopping at tar's end markers.
	if _, err := io.Copy(io.Discard, z); err != nil {
		return "", err
	}
	if !seen["gate-out/test.jsonl"] || !seen["gate-out/run-notes.txt"] {
		return "", errors.New("plain archive requires gate-out/test.jsonl and gate-out/run-notes.txt with commit=<full SHA>")
	}
	recorded := ""
	for _, line := range strings.Split(notes, "\n") {
		if strings.HasPrefix(line, "commit=") {
			if recorded != "" {
				return "", errors.New("plain run notes have duplicate commit records")
			}
			recorded = strings.TrimPrefix(line, "commit=")
		}
	}
	if !fullCommit.MatchString(recorded) || recorded != commit {
		return "", fmt.Errorf("plain tested commit %q differs from merged commit %s", recorded, commit)
	}
	if err := os.WriteFile(filepath.Join(root, "run-notes.txt"), []byte(notes), 0600); err != nil {
		return "", err
	}
	return log, nil
}
