// checker-bridge-scout measures the existing production ABI, not a simulated crossing.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	bridge "github.com/system-inc/adamic/bridge/tsgo/checker"
)

type request struct {
	File                  string
	Start, End            uint64
	Kind, Question, Value string
}
type measurement struct {
	ProgramFiles      int              `json:"program_files"`
	File              string           `json:"file"`
	SHA256            string           `json:"sha256"`
	Questions         int              `json:"questions"`
	UniqueQuestions   int              `json:"unique_questions"`
	Repetitions       int              `json:"repetitions"`
	Round             int              `json:"round"`
	GoLoadNS          int64            `json:"go_load_ns"`
	GoQueryNS         int64            `json:"go_query_ns"`
	GoFirstNS         int64            `json:"go_first_ns"`
	Native            map[string]int64 `json:"native"`
	Pilot             map[string]int64 `json:"pilot"`
	Findings          int              `json:"findings"`
	GoOracleProcessNS int64            `json:"go_oracle_process_ns"`
}

func frame(text string) string { return fmt.Sprintf("%d\n%s", len(utf16.Encode([]rune(text))), text) }
func key(q request) string {
	return frame(q.File) + frame(fmt.Sprint(q.Start)) + frame(fmt.Sprint(q.End)) + frame(q.Kind) + frame(q.Question)
}
func unwrapped(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// This is an ask planner only. Node replay must reproduce every key, order and answer.
// It contains no lint predicate or finding implementation.
func plan(p *bridge.Program, path string) ([]request, error) {
	file := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(path)))
	if file == nil {
		return nil, fmt.Errorf("file absent from program: %s", path)
	}
	if len(file.Diagnostics()) != 0 {
		return nil, fmt.Errorf("parse diagnostics in %s", path)
	}
	asks := []request{}
	add := func(node *ast.Node, question string) {
		asks = append(asks, request{File: path, Start: uint64(node.Pos()), End: uint64(node.End()), Kind: strings.TrimPrefix(node.Kind.String(), "Kind"), Question: question})
	}
	add(file.AsNode(), "options")
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindBinaryExpression {
			b := node.AsBinaryExpression()
			switch b.OperatorToken.Kind {
			case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken:
				for _, pair := range [][2]*ast.Node{{b.Right, b.Left}, {b.Left, b.Right}} {
					literal := unwrapped(pair[0])
					if literal != nil && (literal.Kind == ast.KindTrueKeyword || literal.Kind == ast.KindFalseKeyword) {
						add(unwrapped(pair[1]), "type-shape")
						break
					}
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(file.AsNode())
	return asks, nil
}
func answers(p *bridge.Program, asks []request, repetitions int) (int64, int64, error) {
	var total, first int64
	for pass := 0; pass < repetitions; pass++ {
		for i := range asks {
			q := &asks[i]
			started := time.Now()
			value, err := p.Inspect(q.File, q.Start, q.End, q.Kind, q.Question)
			elapsed := time.Since(started).Nanoseconds()
			total += elapsed
			if pass == 0 && i == 0 {
				first = elapsed
			}
			if err != nil {
				return 0, 0, err
			}
			if pass == 0 {
				q.Value = value
			} else if value != q.Value {
				return 0, 0, fmt.Errorf("unstable answer")
			}
		}
	}
	return total, first, nil
}
func encodeRequests(asks []request) string {
	var out strings.Builder
	out.WriteString(frame(fmt.Sprint(len(asks))))
	for _, q := range asks {
		out.WriteString(key(q))
		out.WriteString(frame(q.Value))
	}
	return out.String()
}
func transcript(asks []request) string {
	var out strings.Builder
	for _, q := range asks {
		out.WriteString(frame(key(q)) + frame("Value") + frame(q.Value))
	}
	return out.String()
}
func expectedOutput(asks []request, repetitions int) []byte {
	var out strings.Builder
	for _, q := range asks {
		out.WriteString(frame(q.Value) + "\n")
	}
	fmt.Fprintf(&out, "queries %d\n", len(asks)*repetitions)
	return []byte(out.String())
}

var metricPattern = regexp.MustCompile(`([a-z_]+)=([0-9]+)`)

func metrics(stderr []byte) map[string]int64 {
	out := map[string]int64{}
	for _, pair := range metricPattern.FindAllSubmatch(stderr, -1) {
		n, _ := strconv.ParseInt(string(pair[2]), 10, 64)
		out[string(pair[1])] = n
	}
	return out
}
func run(config, path, native, pilot, oracle, directory string, repetitions, round int, profile bool) (measurement, error) {
	started := time.Now()
	p, err := bridge.Open(config, nil)
	load := time.Since(started).Nanoseconds()
	if err != nil {
		return measurement{}, err
	}
	asks, err := plan(p, path)
	if err != nil {
		return measurement{}, err
	}
	total, first, err := answers(p, asks, repetitions)
	if err != nil {
		return measurement{}, err
	}
	name := fmt.Sprintf("file-%x-r%d", sha256.Sum256([]byte(path)), round)
	requestPath := filepath.Join(directory, name+".requests")
	if err = os.WriteFile(requestPath, []byte(encodeRequests(asks)), 0600); err != nil {
		return measurement{}, err
	}
	if err = os.WriteFile(filepath.Join(directory, name+".transcript"), []byte(transcript(asks)), 0600); err != nil {
		return measurement{}, err
	}
	command := exec.Command(native, config, requestPath, strconv.Itoa(repetitions))
	command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
	if profile {
		command.Env = append(command.Env, "ADAMIC_TSGO_PROFILE="+filepath.Join(directory, name+".pprof"))
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err = command.Run()
	for suffix, data := range map[string][]byte{".stdout": stdout.Bytes(), ".stderr": stderr.Bytes()} {
		if e := os.WriteFile(filepath.Join(directory, name+suffix), data, 0600); e != nil {
			return measurement{}, e
		}
	}
	if err != nil {
		return measurement{}, fmt.Errorf("native: %w: %s", err, stderr.Bytes())
	}
	if !bytes.Equal(stdout.Bytes(), expectedOutput(asks, repetitions)) {
		return measurement{}, fmt.Errorf("native facts differ from direct Go")
	}
	m := metrics(stderr.Bytes())
	if m["queries"] != int64(len(asks)*repetitions) {
		return measurement{}, fmt.Errorf("timing question count differs")
	}
	unique := map[string]bool{}
	for _, q := range asks {
		unique[key(q)] = true
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return measurement{}, err
	}
	result := measurement{ProgramFiles: len(p.Compiler.GetSourceFiles()), File: path, SHA256: fmt.Sprintf("%x", sha256.Sum256(source)), Questions: len(asks), UniqueQuestions: len(unique), Repetitions: repetitions, Round: round, GoLoadNS: load, GoQueryNS: total, GoFirstNS: first, Native: m}
	oracleCommand := exec.Command(oracle, config, path)
	started = time.Now()
	want, e := oracleCommand.Output()
	result.GoOracleProcessNS = time.Since(started).Nanoseconds()
	if e != nil {
		return measurement{}, fmt.Errorf("Go oracle: %w", e)
	}
	typed := exec.Command(pilot, config, path, strconv.Itoa(repetitions))
	typed.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
	if profile {
		typed.Env = append(typed.Env, "ADAMIC_TSGO_PROFILE="+filepath.Join(directory, name+"-pilot.pprof"))
	}
	stdout.Reset()
	stderr.Reset()
	typed.Stdout = &stdout
	typed.Stderr = &stderr
	e = typed.Run()
	for suffix, data := range map[string][]byte{"-pilot.stdout": stdout.Bytes(), "-pilot.stderr": stderr.Bytes(), "-oracle.stdout": want} {
		if e := os.WriteFile(filepath.Join(directory, name+suffix), data, 0600); e != nil {
			return measurement{}, e
		}
	}
	if e != nil {
		return measurement{}, fmt.Errorf("production pilot: %w: %s", e, stderr.Bytes())
	}
	if !bytes.Equal(stdout.Bytes(), want) {
		return measurement{}, fmt.Errorf("production pilot finding/repair bytes differ from Go for %s", path)
	}
	result.Pilot = metrics(stderr.Bytes())
	if result.Pilot["queries"] != int64(len(asks)*repetitions) {
		return measurement{}, fmt.Errorf("production pilot ask count differs")
	}
	if _, e = fmt.Sscanf(string(want), "findings %d", &result.Findings); e != nil {
		return measurement{}, e
	}
	return result, nil
}
func main() {
	config := flag.String("config", "", "existing tsconfig; its roots are retained")
	manifest := flag.String("manifest", "", "one absolute source path per line")
	native := flag.String("native", "", "native driver binary built from driver.ts")
	pilot := flag.String("pilot", "", "native pilot.ts binary using production Checker")
	oracle := flag.String("oracle", "", "independent Go pilot oracle")
	out := flag.String("out", "", "new artifact directory")
	repetitions := flag.Int("repetitions", 100, "repeat identical ordered questions on the warm checker")
	rounds := flag.Int("rounds", 3, "measured rounds per file")
	profile := flag.Bool("profile", false, "separate instrumented run including CPU profiling")
	quiet := flag.Bool("quiet-hundred", false, "label a complete supplied Kirk corpus; public samples must leave this false")
	flag.Parse()
	if *config == "" || *manifest == "" || *native == "" || *pilot == "" || *oracle == "" || *out == "" || *repetitions < 1 || *rounds < 1 {
		panic("config, manifest, native, out and positive repetitions/rounds required")
	}
	data, err := os.ReadFile(*manifest)
	if err != nil {
		panic(err)
	}
	paths := []string{}
	seen := map[string]bool{}
	for _, path := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if path == "" || !filepath.IsAbs(path) || seen[path] {
			panic("manifest needs distinct absolute paths")
		}
		seen[path] = true
		paths = append(paths, path)
	}
	if err = os.Mkdir(*out, 0755); err != nil {
		panic(err)
	}
	records := []measurement{}
	for round := 1; round <= *rounds; round++ {
		for _, path := range paths {
			m, err := run(*config, path, *native, *pilot, *oracle, *out, *repetitions, round, *profile)
			if err != nil {
				panic(err)
			}
			records = append(records, m)
		}
	}
	metadata := struct {
		GOOS, GOARCH, GoVersion string
		CPUs, GOMAXPROCS        int
		Profile, QuietHundred   bool
		Records                 []measurement
	}{runtime.GOOS, runtime.GOARCH, runtime.Version(), runtime.NumCPU(), runtime.GOMAXPROCS(0), *profile, *quiet, records}
	encoded, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(filepath.Join(*out, "results.json"), append(encoded, '\n'), 0644); err != nil {
		panic(err)
	}
}
