package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GoCorrection struct {
	Index  int         `json:"index"`
	SHA    string      `json:"source_sha256"`
	Lint   *WorkAnswer `json:"go_lint,omitempty"`
	Format WorkAnswer  `json:"go_format"`
}

// encoding/json replaces malformed UTF-8 with U+FFFD. An old stdout string
// without that sentinel is provably lossless; an explicit byte field is lossless
// by construction. Replay every ambiguous string rather than trusting it.
func ambiguousBytes(a WorkAnswer) bool {
	return len(a.OutputRaw) == 0 && strings.Contains(a.Output, "\ufffd")
}

func verifyGo(root, manifest, in, out string, limit time.Duration, watch bool) error {
	m, err := readManifest(manifest)
	if err != nil {
		return err
	}
	if err = m.Dependencies.validate(); err != nil {
		return err
	}
	os.MkdirAll(out, 0755)
	scratch := filepath.Join(out, ".work")
	os.MkdirAll(scratch, 0755)
	d := Driver{Root: root, Scratch: scratch, Limit: limit}
	if err = d.build(); err != nil {
		return err
	}
	binary, _, _, err := d.buildWorkers()
	if err != nil {
		return err
	}
	worker := Worker{args: []string{binary}, dir: root, Limit: limit}
	defer worker.close()
	checkedFiles, changed := 0, 0
	for {
		files, _ := filepath.Glob(filepath.Join(in, "receipts-*.jsonl.gz"))
		for _, p := range files {
			dest := filepath.Join(out, "verified-"+filepath.Base(p))
			if _, err = os.Stat(dest); err == nil {
				continue
			}
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			g, err := gzip.NewReader(f)
			if err != nil {
				f.Close()
				continue
			}
			_, err = io.Copy(io.Discard, g)
			g.Close()
			f.Close()
			if err != nil {
				continue
			}
			f, err = os.Open(p)
			if err != nil {
				return err
			}
			g, err = gzip.NewReader(f)
			if err != nil {
				return err
			}
			decoder := json.NewDecoder(g)
			target, err := os.Create(dest + ".tmp")
			if err != nil {
				return err
			}
			z := gzip.NewWriter(target)
			encoder := json.NewEncoder(z)
			count, corrections := 0, 0
			for {
				var r struct {
					Index    int         `json:"index"`
					Path     string      `json:"path"`
					SHA      string      `json:"source_sha256"`
					GoLint   *WorkAnswer `json:"go_lint"`
					GoFormat WorkAnswer  `json:"go_format"`
				}
				err = decoder.Decode(&r)
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}
				source, err := inputBytes(m, r.Path)
				if err != nil {
					return err
				}
				if hash(source) != r.SHA {
					return fmt.Errorf("oracle verification source changed %s", r.Path)
				}
				q := WorkRequest{Path: r.Path, Source: source, Rule: "all"}
				different := false
				var lint *WorkAnswer
				if r.GoLint != nil {
					q.Op = "lint"
					a := *r.GoLint
					if ambiguousBytes(a) {
						a = worker.call(q)
					}
					if a.Error != "" && r.GoLint.Error == "" {
						oldLimit := worker.Limit
						worker.Limit = 180 * time.Second
						a = worker.call(q)
						worker.Limit = oldLimit
						if a.Error != "" {
							return fmt.Errorf("successful original oracle answer could not be verified: %s", r.Path)
						}
					}
					lint = &a
					different = a.Output != r.GoLint.Output || (a.Error == "") != (r.GoLint.Error == "") || a.Exit != r.GoLint.Exit
				}
				q.Op = "format"
				format := r.GoFormat
				if ambiguousBytes(format) {
					format = worker.call(q)
				}
				if format.Error != "" && r.GoFormat.Error == "" {
					oldLimit := worker.Limit
					worker.Limit = 180 * time.Second
					format = worker.call(q)
					worker.Limit = oldLimit
					if format.Error != "" {
						return fmt.Errorf("successful original formatter answer could not be verified: %s", r.Path)
					}
				}
				different = different || format.Output != r.GoFormat.Output || (format.Error == "") != (r.GoFormat.Error == "") || format.Exit != r.GoFormat.Exit
				if different {
					corrections++
					if err = encoder.Encode(GoCorrection{Index: r.Index, SHA: r.SHA, Lint: lint, Format: format}); err != nil {
						return err
					}
				}
				count++
				checkedFiles++
			}
			g.Close()
			f.Close()
			z.Close()
			target.Close()
			if err = os.Rename(dest+".tmp", dest); err != nil {
				return err
			}
			changed += corrections
			receipt := map[string]any{"files": count, "corrected": corrections, "method": "source SHA held; old stdout without U+FFFD is lossless; explicit byte fields are lossless; ambiguous strings replayed"}
			data, _ := json.Marshal(receipt)
			os.WriteFile(dest+".receipt.json", data, 0644)
			fmt.Fprintf(os.Stderr, "oracle byte verification: %d files, %d corrected answers\n", checkedFiles, changed)
		}
		data, _ := os.ReadFile(filepath.Join(in, "summary.json"))
		var status Sweep
		json.Unmarshal(data, &status)
		if !watch || status.Complete {
			break
		}
		time.Sleep(20 * time.Second)
	}
	return nil
}
