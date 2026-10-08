package main

import (
	"bufio"
	"bytes"
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
	"unicode/utf8"
)

type WorkRequest struct {
	Op, Path, Root, Rule, Options, Family string
	Source                                []byte
}
type WorkAnswer struct {
	OutputRaw []byte    `json:"stdout_bytes,omitempty"`
	Phase     string    `json:"phase,omitempty"`
	WallNS    int64     `json:"wall_elapsed_ns"`
	Spans     [][2]int  `json:"spans,omitempty"`
	Reusable  bool      `json:"reusable"`
	Output    string    `json:"stdout"`
	Error     string    `json:"error,omitempty"`
	Stderr    string    `json:"stderr,omitempty"`
	Stack     string    `json:"stack,omitempty"`
	Exit      int       `json:"exit_code"`
	NS        int64     `json:"elapsed_ns"`
	Valid     bool      `json:"valid"`
	Rows      []Missing `json:"rows,omitempty"`
}

func (a WorkAnswer) execution() Execution {
	return Execution{Output: a.Output, Error: a.Error, Stderr: a.Stderr}
}

type lockedBuffer struct {
	sync.Mutex
	Buffer bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()
	return b.Buffer.Write(p)
}
func (b *lockedBuffer) take() string {
	b.Lock()
	defer b.Unlock()
	s := b.Buffer.String()
	b.Buffer.Reset()
	return s
}

type Worker struct {
	args    []string
	dir     string
	cmd     *exec.Cmd
	in      io.WriteCloser
	decoder *json.Decoder
	stderr  lockedBuffer
	Limit   time.Duration
}

func (w *Worker) start() error {
	w.cmd = exec.Command(w.args[0], w.args[1:]...)
	w.cmd.Dir = w.dir
	w.cmd.Stderr = &w.stderr
	var err error
	w.in, err = w.cmd.StdinPipe()
	if err != nil {
		return err
	}
	out, err := w.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	w.decoder = json.NewDecoder(bufio.NewReader(out))
	return w.cmd.Start()
}
func (w *Worker) close() {
	if w.cmd != nil {
		w.in.Close()
		w.cmd.Process.Kill()
		w.cmd.Wait()
		w.cmd = nil
	}
}
func (w *Worker) call(q WorkRequest) WorkAnswer {
	if w.cmd == nil {
		if err := w.start(); err != nil {
			return WorkAnswer{Error: err.Error(), Exit: -1}
		}
	}
	start := time.Now()
	type result struct {
		a   WorkAnswer
		err error
	}
	done := make(chan result, 1)
	go func() {
		err := json.NewEncoder(w.in).Encode(q)
		var a WorkAnswer
		if err == nil {
			err = w.decoder.Decode(&a)
		}
		done <- result{a, err}
	}()
	select {
	case r := <-done:
		trace := w.stderr.take()
		applyTrace(&r.a, trace, false)
		r.a.WallNS = time.Since(start).Nanoseconds()
		if r.err != nil {
			r.a.Error = r.err.Error()
			r.a.Exit = -1
			w.close()
			applyTrace(&r.a, w.stderr.take(), false)
			if r.a.Output == "" {
				applyTrace(&r.a, trace, true)
			}
		}
		if r.a.Exit != 0 && !r.a.Reusable {
			w.close()
		}
		return r.a
	case <-time.After(w.Limit):
		w.close()
		<-done
		a := WorkAnswer{Error: "worker timeout", Exit: -1, NS: time.Since(start).Nanoseconds(), WallNS: time.Since(start).Nanoseconds()}
		applyTrace(&a, w.stderr.take(), true)
		return a
	}
}

func replaceOnce(source, anchor, replacement string) (string, error) {
	if strings.Count(source, anchor) != 1 {
		return "", fmt.Errorf("host adapter anchor occurs %d times: %q", strings.Count(source, anchor), anchor)
	}
	return strings.Replace(source, anchor, replacement, 1), nil
}
func importsAbsolute(source, base string) string {
	r := regexp.MustCompile(`from '([^']+)'`)
	return r.ReplaceAllStringFunc(source, func(s string) string {
		m := r.FindStringSubmatch(s)
		if !strings.HasPrefix(m[1], ".") {
			return s
		}
		return "from '" + filepath.ToSlash(filepath.Join(base, m[1])) + "'"
	})
}
func (d *Driver) buildWorkers() (goBinary, lintModule, formatModule string, err error) {
	lintRoot := filepath.Join(d.Root, "stage1/cohere/lint")
	source, e := os.ReadFile(filepath.Join(lintRoot, "main.ts"))
	if e != nil {
		err = e
		return
	}
	s := string(source)
	at := strings.Index(s, "const args = programArguments();")
	if at < 0 {
		err = fmt.Errorf("missing lint footer")
		return
	}
	s = s[:at]
	s, err = replaceOnce(s, "caseNumber = 0): number {", "caseNumber = 0, hostSource = ''): number {")
	if err != nil {
		return
	}
	s, err = replaceOnce(s, "const source = readTextFile(path);", "const source = { kind: 'Ok', text: hostSource, message: '' };")
	if err != nil {
		return
	}
	s = importsAbsolute(s, lintRoot) + "\nexport { run };\n"
	lintModule = filepath.Join(d.Scratch, "lint-host.ts")
	if err = os.WriteFile(lintModule, []byte(s), 0600); err != nil {
		return
	}
	source, err = os.ReadFile(filepath.Join(d.Root, "stage1/cohere/scoreboard/format.a"))
	if err != nil {
		return
	}
	s = string(source)
	start := strings.Index(s, "const args = programArguments();")
	end := strings.Index(s, "let text = '';")
	if start < 0 || end < start {
		err = fmt.Errorf("missing formatter transport anchors")
		return
	}
	s = s[:start] + "export function formatHost(path: string, family: string, source: string): void {\nconst input = { kind: 'Ok', text: source };\n" + s[end:] + "\n}\n"
	s = importsAbsolute(s, filepath.Join(d.Root, "stage1/cohere/scoreboard"))
	formatModule = filepath.Join(d.Scratch, "format-host.ts")
	if err = os.WriteFile(formatModule, []byte(s), 0600); err != nil {
		return
	}
	replacements := map[string]string{}
	var files []string
	add := func(name, path string) {
		v := filepath.Join(d.Root, "cohere", "scoreboard_worker_"+name+".go")
		replacements[v] = path
		files = append(files, v)
	}
	source, err = os.ReadFile(filepath.Join(lintRoot, "testdata/oracle.go"))
	if err != nil {
		return
	}
	s, err = replaceOnce(string(source), "func main() {", "func ordinaryOracleMain() {")
	if err != nil {
		return
	}
	s, err = replaceOnce(s, "data, err := os.ReadFile(path)", "data, err := scoreboardReadFile(path)")
	if err != nil {
		return
	}
	p := filepath.Join(d.Scratch, "oracle-worker.go")
	if err = os.WriteFile(p, []byte(s), 0600); err != nil {
		return
	}
	add("oracle", p)
	source, err = os.ReadFile(filepath.Join(d.Root, "stage1/cohere/scoreboard/testdata/catalog.go"))
	if err != nil {
		return
	}
	s, err = replaceOnce(string(source), "func main() {", "func ordinaryCatalogMain() {")
	if err != nil {
		return
	}
	p = filepath.Join(d.Scratch, "catalog-worker.go")
	if err = os.WriteFile(p, []byte(s), 0600); err != nil {
		return
	}
	add("catalog", p)
	add("transport", filepath.Join(d.Root, "stage1/cohere/scoreboard/testdata/worker.go"))
	add("registry", filepath.Join(lintRoot, ".generated/registry.go"))
	for _, desc := range d.Descriptors {
		add(strings.ReplaceAll(desc.Slug, "-", "_"), filepath.Join(lintRoot, "rules", desc.Slug, "oracle.go"))
	}
	overlay := filepath.Join(d.Scratch, "workers-overlay.json")
	data, _ := json.Marshal(map[string]any{"Replace": replacements})
	if err = os.WriteFile(overlay, data, 0600); err != nil {
		return
	}
	goBinary = filepath.Join(d.Scratch, "worker")
	args := append([]string{"build", "-overlay=" + overlay, "-o", goBinary}, files...)
	_, err = checked(filepath.Join(d.Root, "cohere"), "go", args...)
	return
}

func applyTrace(a *WorkAnswer, trace string, recoverOutput bool) {
	hasPhase := a.Phase != ""
	for _, line := range strings.SplitAfter(trace, "\n") {
		if strings.HasPrefix(line, "scoreboard-phase\t") {
			if !hasPhase {
				a.Phase = strings.TrimSpace(strings.TrimPrefix(line, "scoreboard-phase\t"))
			}
			continue
		}
		if strings.HasPrefix(line, "scoreboard-log\t") {
			if recoverOutput {
				var text string
				if json.Unmarshal([]byte(strings.TrimPrefix(line, "scoreboard-log\t")), &text) == nil {
					a.Output += text
				}
			}
			continue
		}
		a.Stderr += line
	}
}

// JSON text is lossless only for UTF-8. An oracle may print arbitrary bytes;
// keep those in base64 rather than letting encoding/json replace them.
func (a *WorkAnswer) UnmarshalJSON(data []byte) error {
	type plain WorkAnswer
	var p plain
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	*a = WorkAnswer(p)
	if len(a.OutputRaw) > 0 {
		a.Output = string(a.OutputRaw)
	}
	return nil
}
func (a WorkAnswer) MarshalJSON() ([]byte, error) {
	type plain WorkAnswer
	if !utf8.ValidString(a.Output) {
		a.OutputRaw = []byte(a.Output)
	}
	return json.Marshal(plain(a))
}
