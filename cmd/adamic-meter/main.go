// Command adamic-meter measures the distance between ordinary TypeScript and stage 0.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

type finding struct {
	Kind     string   `json:"kind"`
	Reason   string   `json:"reason"`
	Count    int      `json:"count"`
	Files    []string `json:"files"`
	Example  string   `json:"example"`
	fileSeen map[string]bool
}

type report struct {
	Root                  string       `json:"root"`
	FilesExamined         int          `json:"files_examined"`
	JavaScriptSkipped     int          `json:"javascript_files_skipped"`
	FilesReachingLowering int          `json:"files_reaching_lowering"`
	Reasons               []*finding   `json:"reasons"`
	Adaptations           []adaptation `json:"adaptations,omitempty"`
}

type adaptation struct {
	Rewrite             string `json:"rewrite"`
	Removed             int    `json:"diagnostics_removed"`
	DeclarationsChanged int    `json:"declarations_changed,omitempty"`
	FunctionsChanged    int    `json:"functions_changed,omitempty"`
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("adamic-meter", flag.ContinueOnError)
	flags.SetOutput(stderr)
	asJSON := flags.Bool("json", false, "write JSON")
	adapt := flags.Bool("adapt", false, "apply semantics-preserving rewrites in memory")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: adamic-meter [--adapt] [--json] <directory>")
		return 2
	}
	r, err := measure(flags.Arg(0), *adapt)
	if err != nil {
		fmt.Fprintf(stderr, "adamic-meter: %v\n", err)
		return 1
	}
	if *asJSON {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(r); err != nil {
			fmt.Fprintf(stderr, "adamic-meter: writing JSON: %v\n", err)
			return 1
		}
		return 0
	}
	fmt.Fprintf(stdout, "COUNT  FILES  KIND        REASON  EXAMPLE\n")
	for _, reason := range r.Reasons {
		fmt.Fprintf(stdout, "%5d  %5d  %-10s  %s  %s\n", reason.Count, len(reason.Files), reason.Kind, reason.Reason, reason.Example)
	}
	for _, adapted := range r.Adaptations {
		fmt.Fprintf(stdout, "adapted: %s removed %d diagnostics\n", adapted.Rewrite, adapted.Removed)
	}
	fmt.Fprintf(stdout, "\n%d TypeScript files examined", r.FilesExamined)
	if r.JavaScriptSkipped != 0 {
		fmt.Fprintf(stdout, "; %d JavaScript files skipped (stage 0 accepts only .ts and .a)", r.JavaScriptSkipped)
	}
	fmt.Fprintf(stdout, "; %d entries reached lowering\n", r.FilesReachingLowering)
	return 0
}

func measure(root string, adapt bool) (*report, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}
	result := &report{Root: filepath.ToSlash(absolute)}
	byReason := make(map[string]*finding)
	var paths []string
	err = filepath.WalkDir(absolute, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		extension := strings.ToLower(filepath.Ext(path))
		if extension == ".js" || extension == ".jsx" {
			result.JavaScriptSkipped++
			return nil
		}
		if extension != ".ts" && extension != ".a" {
			return nil
		}
		result.FilesExamined++
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	add := func(path string, observations []observation) {
		for _, observation := range observations {
			key := observation.Kind + "\x00" + observation.Reason
			existing := byReason[key]
			if existing == nil {
				existing = &finding{Kind: observation.Kind, Reason: observation.Reason, Example: relative(absolute, observation.Example), fileSeen: make(map[string]bool)}
				byReason[key] = existing
			}
			existing.Count++
			file := relative(absolute, path)
			if !existing.fileSeen[file] {
				existing.fileSeen[file] = true
				existing.Files = append(existing.Files, file)
			}
		}
	}
	// Checking all roots together is both truer to a codebase and much faster than rebuilding the
	// TypeScript program once per file. If it checks, lowering still runs once per possible entry,
	// because stage 0 deliberately accepts exactly one entry root.
	if len(paths) != 0 {
		loader := load.Load
		var overlay map[string]string
		baseline, baselineErr := load.Load(paths)
		_ = baseline
		if adapt {
			overlay, result.Adaptations, err = adaptations(paths, baselineErr)
			if err != nil {
				return nil, err
			}
			loader = func(paths []string) (*load.Program, error) { return load.LoadOverlay(paths, overlay) }
		}
		if _, loadErr := loader(paths); loadErr != nil {
			var checkError *load.CheckError
			if errors.As(loadErr, &checkError) {
				diagnosed := make(map[string]bool)
				for _, diagnostic := range checkError.Diagnostics {
					match := checkerDiagnostic.FindStringSubmatch(diagnostic)
					if match == nil {
						add(paths[0], []observation{{"type", "TypeScript diagnostic", diagnostic}})
						continue
					}
					file := diagnosticFile(match[1])
					diagnosed[filepath.Clean(file)] = true
					add(file, []observation{{typescriptCategory(match[2]), "TypeScript TS" + match[2], match[1]}})
				}
				if adapt {
					for _, path := range paths {
						if diagnosed[filepath.Clean(path)] {
							continue
						}
						observations, reached := inspect(path, overlay)
						if reached {
							result.FilesReachingLowering++
						}
						for _, found := range observations {
							if found.Kind == "Refused" || found.Kind == "NotYet" {
								add(path, []observation{found})
							}
						}
					}
				}
			} else {
				add(paths[0], []observation{{"load", normalize(loadErr.Error()), paths[0]}})
			}
		} else {
			for _, path := range paths {
				observations, reached := inspect(path, overlay)
				if reached {
					result.FilesReachingLowering++
				}
				add(path, observations)
			}
		}
	}
	for _, item := range byReason {
		sort.Strings(item.Files)
		item.fileSeen = nil
		result.Reasons = append(result.Reasons, item)
	}
	sort.Slice(result.Reasons, func(i, j int) bool {
		if result.Reasons[i].Count != result.Reasons[j].Count {
			return result.Reasons[i].Count > result.Reasons[j].Count
		}
		if result.Reasons[i].Kind != result.Reasons[j].Kind {
			return result.Reasons[i].Kind < result.Reasons[j].Kind
		}
		return result.Reasons[i].Reason < result.Reasons[j].Reason
	})
	return result, nil
}

func diagnosticFile(where string) string {
	if match := location.FindString(where); match != "" {
		return match[:strings.LastIndex(match, ":")][:strings.LastIndex(match[:strings.LastIndex(match, ":")], ":")]
	}
	return where
}

type observation struct{ Kind, Reason, Example string }

var checkerDiagnostic = regexp.MustCompile(`^([^\n]*?): error TS([0-9]+):`)
var typeOnlyDiagnostic = regexp.MustCompile(`^([^\n]+):([0-9]+):([0-9]+): error TS1484: '([^']+)'`)

type sourceEdit struct {
	line, column int
}

func adaptations(paths []string, baselineErr error) (map[string]string, []adaptation, error) {
	overlay, rewrites, err := importAdaptations(paths, baselineErr)
	if err != nil {
		return nil, nil, err
	}
	_, before := load.LoadOverlay(paths, overlay)
	initial := optionalDiagnosticCount(before)
	changed := 0
	for {
		var count int
		overlay, count, err = optionalAdaptations(paths, overlay, before)
		if err != nil {
			return nil, nil, err
		}
		if count == 0 {
			break
		}
		changed += count
		_, before = load.LoadOverlay(paths, overlay)
	}
	if changed != 0 {
		rewrites = append(rewrites, adaptation{Rewrite: optionalRewrite, Removed: initial - optionalDiagnosticCount(before), DeclarationsChanged: changed})
	}
	initialReturns := diagnosticCount(before, "7030")
	overlay, changed, err = returnAdaptations(paths, overlay, before)
	if err != nil {
		return nil, nil, err
	}
	if changed != 0 {
		_, after := load.LoadOverlay(paths, overlay)
		rewrites = append(rewrites, adaptation{Rewrite: returnRewrite, Removed: initialReturns - diagnosticCount(after, "7030"), FunctionsChanged: changed})
	}
	return overlay, rewrites, nil
}

func optionalDiagnosticCount(err error) int {
	return diagnosticCount(err, "2412") + diagnosticCount(err, "2375") + diagnosticCount(err, "2379")
}

func importAdaptations(paths []string, baselineErr error) (map[string]string, []adaptation, error) {
	var checkError *load.CheckError
	if !errors.As(baselineErr, &checkError) {
		return nil, nil, nil
	}
	edits := make(map[string][]sourceEdit)
	for _, diagnostic := range checkError.Diagnostics {
		match := typeOnlyDiagnostic.FindStringSubmatch(diagnostic)
		if match == nil {
			continue
		}
		line, _ := strconv.Atoi(match[2])
		column, _ := strconv.Atoi(match[3])
		edits[match[1]] = append(edits[match[1]], sourceEdit{line, column})
	}
	overlay := make(map[string]string, len(edits))
	for path, fileEdits := range edits {
		source, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("reading %s to adapt it: %w", path, err)
		}
		lines := strings.Split(string(source), "\n")
		sort.Slice(fileEdits, func(i, j int) bool {
			if fileEdits[i].line != fileEdits[j].line {
				return fileEdits[i].line > fileEdits[j].line
			}
			return fileEdits[i].column > fileEdits[j].column
		})
		for _, edit := range fileEdits {
			if edit.line < 1 || edit.line > len(lines) || edit.column < 1 || edit.column > len(lines[edit.line-1])+1 {
				return nil, nil, fmt.Errorf("TS1484 gave an invalid position in %s:%d:%d", path, edit.line, edit.column)
			}
			line := lines[edit.line-1]
			at := edit.column - 1
			line = line[:at] + "type " + line[at:]
			lines[edit.line-1] = line
		}
		overlay[path] = strings.Join(lines, "\n")
	}
	if len(edits) == 0 {
		return overlay, nil, nil
	}
	_, adaptedErr := load.LoadOverlay(paths, overlay)
	remaining := diagnosticCount(adaptedErr, "1484")
	return overlay, []adaptation{{Rewrite: "type-only imports (TS1484)", Removed: diagnosticCount(baselineErr, "1484") - remaining}}, nil
}

func diagnosticCount(err error, code string) int {
	var checkError *load.CheckError
	if !errors.As(err, &checkError) {
		return 0
	}
	count := 0
	needle := "error TS" + code + ":"
	for _, diagnostic := range checkError.Diagnostics {
		if strings.Contains(diagnostic, needle) {
			count++
		}
	}
	return count
}

func inspect(path string, overlay map[string]string) ([]observation, bool) {
	program, err := load.LoadOverlay([]string{path}, overlay)
	if err != nil {
		var checkError *load.CheckError
		if !errors.As(err, &checkError) {
			return []observation{{"load", normalize(err.Error()), path}}, false
		}
		observations := make([]observation, 0, len(checkError.Diagnostics))
		for _, diagnostic := range checkError.Diagnostics {
			match := checkerDiagnostic.FindStringSubmatch(diagnostic)
			if match == nil {
				observations = append(observations, observation{"type", "TypeScript diagnostic", diagnostic})
			} else {
				observations = append(observations, observation{typescriptCategory(match[2]), "TypeScript TS" + match[2], match[1]})
			}
		}
		return observations, false
	}
	_, err = lower.Lower(context.Background(), program)
	if err == nil {
		return nil, true
	}
	var refused *lower.Refused
	if errors.As(err, &refused) {
		return []observation{{"Refused", normalize(refused.What), refused.Where}}, true
	}
	var notYet *lower.NotYet
	if errors.As(err, &notYet) {
		return []observation{{"NotYet", normalize(notYet.What), notYet.Where}}, true
	}
	return []observation{{"lower", normalize(err.Error()), path}}, true
}

// TypeScript assigns its basic parser diagnostics to the low 1xxx range (and JSX parser diagnostics
// to 17xxx). Later 1xxx codes include option-driven type diagnostics such as TS1484.
// Keeping the code in the reason makes the grouping exact even as TypeScript changes its wording.
func typescriptCategory(text string) string {
	if text == "1484" {
		return "mechanical"
	}
	code, _ := strconv.Atoi(text)
	if code >= 1000 && code < 1200 || code >= 17000 && code < 18000 {
		return "parse"
	}
	return "type"
}

var (
	location = regexp.MustCompile(`(?:[A-Za-z]:)?[^:\n]+:[0-9]+:[0-9]+`)
	quoted   = regexp.MustCompile("(?:'[^']*'|\"[^\"]*\"|`[^`]*`)")
	typeText = regexp.MustCompile(`\b(?:of type|typed|as) [^,;()]+`)
)

func normalize(message string) string {
	message = location.ReplaceAllString(message, "<location>")
	message = quoted.ReplaceAllString(message, "<name>")
	message = typeText.ReplaceAllString(message, "of type <type>")
	if index := strings.Index(message, "("); index >= 0 {
		message = message[:index]
	}
	return strings.Join(strings.Fields(message), " ")
}

func relative(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(path)
}
