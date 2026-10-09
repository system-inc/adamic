package skipcensus

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
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
		if event.Action == "output" {
			messages[k] += event.Output
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
