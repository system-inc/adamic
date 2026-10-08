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
	return CheckLogAwaiting(input, output, rows, nil)
}

// Landed says whether a branch is on main. An error means it can't be told, which fails the check closed.
type Landed func(branch string) (bool, error)

var awaitedBranch = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)

// checkPending holds a pending row to its reason: it names the branch it awaits, and its skip message says so,
// as "awaits <branch>" or "dependency: <branch>" (@system_adamic, Oct 8: internal/oracle's TestEntriesAcceptance
// skips with "acceptance dependency: codex/step12-entries-fixtures 8d864f9c", the same pending reason).
func checkPending(row Row) error {
	if !awaitedBranch.MatchString(row.Awaits) {
		return fmt.Errorf("a pending skip names the branch it awaits")
	}
	literal, err := strconv.Unquote(row.Message)
	if err != nil {
		literal = row.Message
	}
	if !strings.Contains(literal, "awaits "+row.Awaits) && !strings.Contains(literal, "dependency: "+row.Awaits) {
		return fmt.Errorf("a pending skip's message says %q", "awaits "+row.Awaits)
	}
	return nil
}

// CheckLogAwaiting is CheckLog with pending skips (@system_adamic, Oct 8): a skip that waits on another
// branch passes while that branch is off main and fails the moment it lands and the test still skips, so a
// skip can't outlive its reason. Every pending skip is listed by name. Without landed, a pending skip can't
// be checked and counts as unclassified.
func CheckLogAwaiting(input io.Reader, output io.Writer, rows []Row, landed Landed) error {
	decoder := json.NewDecoder(input)
	messages := map[string]string{}
	required, unknown, total, pending, overdue := 0, 0, 0, 0, 0
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
		if row.Class == "pending" {
			if landed == nil || checkPending(row) != nil {
				unknown++
				fmt.Fprintf(output, "unknown\t%s\t%s\n", event.Package, event.Test)
				continue
			}
			on, err := landed(row.Awaits)
			if err != nil {
				overdue++
				fmt.Fprintf(output, "pending-unknown\t%s\t%s\t%s\tawaits %s: %v\n", event.Package, event.Test, row.ID, row.Awaits, err)
				continue
			}
			if on {
				overdue++
				fmt.Fprintf(output, "pending-landed\t%s\t%s\t%s\tawaits %s, which is on main, and the test still skips\n", event.Package, event.Test, row.ID, row.Awaits)
				continue
			}
			pending++
			fmt.Fprintf(output, "pending\t%s\t%s\t%s\tawaits %s\n", event.Package, event.Test, row.ID, row.Awaits)
			continue
		}
		switch row.Class {
		case "required-input", "measurement", "not-applicable", "opt-in-lane":
		default:
			unknown++
			fmt.Fprintf(output, "unknown\t%s\t%s\n", event.Package, event.Test)
			continue
		}
		fmt.Fprintf(output, "%s\t%s\t%s\t%s\t%s\n", row.Class, event.Package, event.Test, row.ID, row.Provides)
		if row.Class == "required-input" {
			required++
		}
	}
	fmt.Fprintf(output, "skips=%d required-input=%d unknown=%d pending=%d\n", total, required, unknown, pending)
	if required+unknown+overdue > 0 {
		return fmt.Errorf("gate skipped %d required-input tests; %d unclassified skips; %d pending skips past or unsure of their reason", required, unknown, overdue)
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
