package main

import (
	"bufio"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

//go:embed typescript.cjs
var typescriptHelper string

// One stock TypeScript process shares parsed libraries across requests. Results are
// cached by the exact adapted source bytes for this process's fixed options and prelude.
type typescriptOracle struct {
	mutex   sync.Mutex
	command *exec.Cmd
	input   io.WriteCloser
	output  *bufio.Reader
	cache   map[[32]byte][]string
	stats   oracleStats
}
type oracleStats struct {
	Version      string `json:"version"`
	Checks       int    `json:"checks"`
	Hits         int    `json:"cacheHits"`
	Milliseconds int64  `json:"milliseconds"`
}

func startTypescript(root, work string) (*typescriptOracle, error) {
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		return nil, fmt.Errorf("TypeScript oracle requires ADAMIC_TYPESCRIPT_SOURCE set to the pinned TypeScript 6.0.3 checkout")
	}
	module, err := filepath.Abs(filepath.Join(source, "lib/typescript.js"))
	if err != nil {
		return nil, fmt.Errorf("TypeScript oracle ADAMIC_TYPESCRIPT_SOURCE: %w", err)
	}
	if _, err := os.Stat(module); err != nil {
		return nil, fmt.Errorf("TypeScript oracle requires lib/typescript.js in ADAMIC_TYPESCRIPT_SOURCE (%s): %w", module, err)
	}
	helper := filepath.Join(work, "typescript.cjs")
	if err := os.WriteFile(helper, []byte(typescriptHelper), 0644); err != nil {
		return nil, err
	}
	prelude, err := filepath.Abs(filepath.Join(root, "internal/load/prelude.d.ts"))
	if err != nil {
		return nil, err
	}
	programPath, err := filepath.Abs(filepath.Join(work, "program.ts"))
	if err != nil {
		return nil, err
	}
	command := exec.Command("node", helper, module, prelude, programPath)
	input, err := command.StdinPipe()
	if err != nil {
		return nil, err
	}
	output, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		return nil, err
	}
	oracle := &typescriptOracle{command: command, input: input, output: bufio.NewReader(output), cache: map[[32]byte][]string{}}
	line, err := oracle.output.ReadBytes('\n')
	if err != nil {
		oracle.close()
		return nil, fmt.Errorf("starting TypeScript oracle: %w", err)
	}
	if err := json.Unmarshal(line, &oracle.stats.Version); err != nil {
		oracle.close()
		return nil, err
	}
	if oracle.stats.Version != "6.0.3" {
		oracle.close()
		return nil, fmt.Errorf("TypeScript oracle requires version 6.0.3 from ADAMIC_TYPESCRIPT_SOURCE; %s reports %q", module, oracle.stats.Version)
	}
	return oracle, nil
}
func (o *typescriptOracle) close() { o.input.Close(); o.command.Process.Kill(); o.command.Wait() }
func (o *typescriptOracle) check(source string) ([]string, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	hash := sha256.Sum256([]byte(source))
	if codes, ok := o.cache[hash]; ok {
		o.stats.Hits++
		return codes, nil
	}
	started := time.Now()
	defer func() { o.stats.Milliseconds += time.Since(started).Milliseconds() }()
	if err := json.NewEncoder(o.input).Encode(source); err != nil {
		return nil, err
	}
	type response struct {
		codes []string
		err   error
	}
	done := make(chan response, 1)
	go func() {
		line, err := o.output.ReadBytes('\n')
		var codes []string
		if err == nil {
			err = json.Unmarshal(line, &codes)
		}
		done <- response{codes, err}
	}()
	select {
	case got := <-done:
		if got.err != nil {
			return nil, got.err
		}
		o.stats.Checks++
		o.cache[hash] = got.codes
		return got.codes, nil
	case <-time.After(2 * time.Minute):
		o.command.Process.Kill()
		return nil, fmt.Errorf("TypeScript oracle timed out")
	}
}

var diagnosticCode = regexp.MustCompile(`\berror (TS[0-9]+):`)

func checkerCode(reason string) string {
	match := diagnosticCode.FindStringSubmatch(reason)
	if len(match) == 2 {
		return match[1]
	}
	return ""
}
func typescriptVerdict(reason string, codes []string) verdict {
	code := checkerCode(reason)
	for _, candidate := range codes {
		if code != "" && candidate == code {
			return verdict{outcomeNotTypescript, code}
		}
	}
	if code != "" {
		other := "accepted"
		if len(codes) > 0 {
			other = strings.Join(codes, ",")
		}
		reason += " (tsc: " + other + ")"
	}
	return verdict{outcomeRefused, reason}
}
