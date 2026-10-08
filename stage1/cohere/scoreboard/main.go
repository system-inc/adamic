package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	manifest := flag.String("manifest", "", "JSON corpus manifest")
	root := flag.String("root", ".", "repository root")
	fixtures := flag.Bool("fixtures", false, "identified available corpus run (not the full quiet hundred)")
	reduce := flag.Bool("reduce", false, "reduce successful lint byte divergences")
	native := flag.String("native", "", "native lint binary for checker recording")
	out := flag.String("out", "", "report destination; default stdout")
	limit := flag.Duration("limit", 30*time.Second, "per process timeout")
	full := flag.Bool("full-tree", false, "stream the pinned full-tree run using persistent workers")
	workers := flag.Int("workers", 4, "number of independent host worker pairs")
	maxFiles := flag.Int("max-files", 0, "explicit partial transport validation scope; zero means every file")
	reduceInput := flag.String("reduce-sweep", "", "receipt directory to reduce at original source paths")
	watch := flag.Bool("reduce-watch", false, "reduce newly completed receipt chunks until the full run completes")
	verifyInput := flag.String("verify-go-sweep", "", "verify every oracle stdout using the byte-preserving transport")
	flag.Parse()
	if *verifyInput != "" {
		abs, err := filepath.Abs(*root)
		if err == nil {
			err = verifyGo(abs, *manifest, *verifyInput, *out, *limit, *watch)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *reduceInput != "" {
		abs, err := filepath.Abs(*root)
		if err == nil {
			err = reduceSweep(abs, *manifest, *reduceInput, *out, *limit, *watch)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *full {
		abs, err := filepath.Abs(*root)
		if err == nil {
			err = sweep(abs, *manifest, *out, *workers, *maxFiles, *limit)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := runCLI(*manifest, *root, *native, *out, *fixtures, *reduce, *limit); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func runCLI(manifest, root, native, out string, fixtures, reduce bool, limit time.Duration) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(manifest)
	if err != nil {
		return err
	}
	var m Manifest
	if err = json.Unmarshal(data, &m); err != nil {
		return err
	}
	if err = validate(m, !fixtures, root); err != nil {
		return err
	}
	scratch, err := os.MkdirTemp("", "adamic-scoreboard-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)
	d := Driver{Root: root, Scratch: scratch, Native: native, Limit: limit}
	if err = d.build(); err != nil {
		return err
	}
	r, err := d.run(m, !fixtures, reduce)
	if err != nil {
		return err
	}
	data, err = json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if out != "" {
		if err = os.WriteFile(out, data, 0644); err != nil {
			return err
		}
	} else {
		fmt.Print(string(data))
	}
	for _, c := range r.Cells {
		if c.Status != "agree" {
			return fmt.Errorf("scoreboard contains divergences or blocked checks; report retained")
		}
	}
	if len(r.Blockers) > 0 {
		return fmt.Errorf("census blocked; report retained")
	}
	return nil
}
