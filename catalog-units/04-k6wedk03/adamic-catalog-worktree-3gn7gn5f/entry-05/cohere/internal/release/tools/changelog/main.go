// Command changelog prints a release's section of CHANGELOG.md from what changed in the rule sets, and
// refuses a version smaller than that change allows.
//
//	go run ./internal/release/tools/changelog --version 1.1.0 --previous 1.0.0
//	go run ./internal/release/tools/changelog --version 1.0.0
//
// The sets are compared between the previous release's tag, v<previous>, or the commit --from names, and
// HEAD. With no previous version it writes the first release's section, which lists the sets. The section
// goes to stdout for a person to paste above the last one and to add, under the same heading, what
// changed in cohere itself, which no diff of the sets can say. The policy is the changelog package's.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/system-inc/cohere/internal/release/changelog"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "changelog:", err)
		os.Exit(1)
	}
}

func run() error {
	version := flag.String("version", "", "the version being released, like 1.1.0")
	previous := flag.String("previous", "", "the last released version, like 1.0.0; empty for the first release")
	from := flag.String("from", "", "the commit the last release was made from; v<previous> when empty")
	to := flag.String("to", "HEAD", "the commit being released")
	date := flag.String("date", time.Now().Format("2006-01-02"), "the release date")
	module := flag.String("module", ".", "the cohere module")
	flag.Parse()

	if *version == "" {
		return fmt.Errorf("--version is required")
	}
	after, err := changelog.SetsAt(*module, *to)
	if err != nil {
		return err
	}
	if *previous == "" {
		fmt.Print(changelog.Render(*version, *date, nil, after))
		return nil
	}

	if *from == "" {
		*from = "v" + *previous
	}
	before, err := changelog.SetsAt(*module, *from)
	if err != nil {
		return fmt.Errorf("reading the sets at the last release's tag: %w", err)
	}
	changed, bump, err := changelog.Diff(before, after)
	if err != nil {
		return err
	}
	if err := changelog.RequireBump(*previous, *version, bump); err != nil {
		return err
	}
	fmt.Print(changelog.Render(*version, *date, changed, nil))
	return nil
}
