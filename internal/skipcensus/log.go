package skipcensus

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// CheckLog lists every named skip, fails closed on unknown/ambiguous skips, and
// rejects required inputs. Package-level [no test files] events are not tests.
// The skip reason disambiguates multiple sites in one enclosing test; line numbers
// are diagnostic only, so historical logs remain usable when source lines move.
func CheckLog(input io.Reader, output io.Writer, rows []Row) error {
	decoder := json.NewDecoder(input)
	messages := map[string]string{}
	active := map[string]map[string]bool{}
	packageLines := map[string]string{}
	required, unknown, total := 0, 0, 0
	for {
		var event struct{ Action, Package, Test, Output string }
		err := decoder.Decode(&event)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("test log: %w", err)
		}
		k := event.Package + "/" + event.Test
		if (event.Action == "run" || event.Action == "cont") && event.Test != "" {
			if event.Action == "run" {
				messages[k] = ""
			}
			if active[event.Package] == nil {
				active[event.Package] = map[string]bool{}
			}
			active[event.Package][event.Test] = true
		}
		if event.Action == "output" {
			if event.Test == "" {
				packageLines[event.Package] += event.Output
				for {
					end := strings.IndexByte(packageLines[event.Package], '\n')
					if end < 0 {
						break
					}
					line := packageLines[event.Package][:end+1]
					packageLines[event.Package] = packageLines[event.Package][end+1:]
					if len(MissingInputVariables(line)) > 0 {
						for test := range active[event.Package] {
							messages[event.Package+"/"+test] += line
						}
					}
				}
			}
			messages[k] += event.Output
			continue
		}
		if event.Action == "pause" {
			delete(active[event.Package], event.Test)
		}
		if event.Action == "pass" || event.Action == "fail" || event.Action == "skip" {
			if len(MissingInputVariables(packageLines[event.Package])) > 0 {
				for test := range active[event.Package] {
					messages[event.Package+"/"+test] += packageLines[event.Package]
				}
				packageLines[event.Package] = ""
			}
			delete(active[event.Package], event.Test)
		}
		if event.Action == "pass" && event.Test != "" {
			findings := degradedFindings(event.Package, event.Test, messages[k], rows)
			for _, finding := range findings {
				total++
				if finding.Class == "required-input" {
					required++
				}
				if finding.Class == "unknown" {
					unknown++
				}
				fmt.Fprintf(output, "%s\t%s\t%s\t%s\tdegraded-input\t%s\n", finding.Class, event.Package, event.Test, finding.ID, strings.Join(finding.Variables, ","))
			}
			delete(messages, k)
			continue
		}
		if event.Action != "skip" || event.Test == "" {
			continue
		}
		total++
		directory := strings.TrimPrefix(event.Package, "github.com/system-inc/adamic/")
		test := strings.Split(event.Test, "/")[0]
		var candidates []Row
		for _, row := range rows {
			if row.Kind == "degraded-input" {
				continue
			}
			if strings.TrimSuffix(row.File, "/"+base(row.File)) != directory {
				continue
			}
			for _, caller := range row.Callers {
				if caller == test {
					candidates = append(candidates, row)
					break
				}
			}
		}
		if len(candidates) > 1 {
			var matched []Row
			for _, row := range candidates {
				literal, err := strconv.Unquote(row.Message)
				if err == nil && skipReasonMatches(literal, messages[k]) {
					matched = append(matched, row)
				}
			}
			candidates = matched
		}
		if len(candidates) != 1 {
			unknown++
			fmt.Fprintf(output, "unknown\t%s\t%s\n", event.Package, event.Test)
			continue
		}
		row := candidates[0]
		switch row.Class {
		case "required-input", "measurement", "not-applicable", "opt-in-lane":
		default:
			unknown++
			fmt.Fprintf(output, "unknown\t%s\t%s\n", event.Package, event.Test)
			continue
		}
		fmt.Fprintf(output, "%s\t%s\t%s\t%s\n", row.Class, event.Package, event.Test, row.ID)
		if row.Class == "required-input" {
			required++
		}
	}
	fmt.Fprintf(output, "skips=%d required-input=%d unknown=%d\n", total, required, unknown)
	if required+unknown > 0 {
		return fmt.Errorf("gate skipped %d required-input tests; %d unclassified skips", required, unknown)
	}
	return nil
}
func base(path string) string { parts := strings.Split(path, "/"); return parts[len(parts)-1] }

// Skipf's numeric exit code changes the rendered message, not the declared site.
// Other unsupported format directives still fail closed when sites are ambiguous.
func skipReasonMatches(literal, output string) bool {
	if strings.Contains(output, literal) {
		return true
	}
	if !strings.Contains(literal, "%d") {
		return false
	}
	pattern := strings.ReplaceAll(regexp.QuoteMeta(literal), "%d", "[+-]?[0-9]+")
	matched, _ := regexp.MatchString(pattern, output)
	return matched
}

// Forwarded overlay output belongs to the parent test that emitted it. It can
// match a declared testdata site in that package without knowing overlay names.
func degradedFindings(pkg, test, message string, rows []Row) []Row {
	directory := strings.TrimPrefix(pkg, "github.com/system-inc/adamic/")
	top := strings.Split(test, "/")[0]
	findings := map[string]Row{}
	for _, line := range strings.Split(message, "\n") {
		variables := MissingInputVariables(line)
		if len(variables) == 0 {
			continue
		}
		var matches []Row
		for _, row := range rows {
			if row.Kind != "degraded-input" {
				continue
			}
			own := strings.TrimSuffix(row.File, "/"+base(row.File))
			overlay := strings.HasPrefix(row.File, directory+"/testdata/")
			caller := false
			for _, name := range row.Callers {
				if name == top {
					caller = true
				}
			}
			if !overlay && (own != directory || !caller) {
				continue
			}
			literal, err := strconv.Unquote(row.Message)
			if err == nil && diagnosticReasonMatches(literal, line, row.Variables) {
				matches = append(matches, row)
			}
		}
		if len(matches) != 1 {
			key := strings.Join(variables, ",")
			findings["unknown/"+key] = Row{Class: "unknown", ID: "unclassified diagnostic", Variables: variables}
		} else {
			row := matches[0]
			if row.Class != "required-input" && row.Class != "not-applicable" {
				row.Class = "unknown"
			}
			findings[key(row)] = row
		}
	}
	keys := make([]string, 0, len(findings))
	for key := range findings {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var result []Row
	for _, key := range keys {
		result = append(result, findings[key])
	}
	return result
}
