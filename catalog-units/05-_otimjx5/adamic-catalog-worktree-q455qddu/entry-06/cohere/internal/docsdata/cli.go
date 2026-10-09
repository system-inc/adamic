package docsdata

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Cli is cli.json: the flags the binary's own flag set prints, and each verb's.
type Cli struct {
	Flags []CliFlag `json:"flags"`
	Verbs []CliVerb `json:"verbs,omitempty"`
}

// CliFlag is one flag as the flag package describes it.
type CliFlag struct {
	Name string `json:"name"`
	// Argument is the name the help gives the flag's value: a backquoted word in its usage, or else its
	// type. Empty for a boolean flag, which takes none.
	Argument string `json:"argument,omitempty"`
	Usage    string `json:"usage"`
	// Default is the flag's default, present only when it differs from its type's zero value, which is
	// when the flag package prints it.
	Default string `json:"default,omitempty"`
}

// CliVerb is a subcommand with a flag set of its own.
type CliVerb struct {
	Name string `json:"name"`
	// Usage is the verb's synopsis lines, and Summary the prose its help prints before the flags.
	Usage   []string  `json:"usage"`
	Summary string    `json:"summary,omitempty"`
	Flags   []CliFlag `json:"flags"`
}

// flagHeading is the first line of one flag in flag.PrintDefaults: two spaces, the name, and the
// argument's name when it takes one.
var flagHeading = regexp.MustCompile(`^  -([^\s]+)(?: (\S+))?$`)

// flagDefault is the suffix PrintDefaults appends to a non-zero default: quoted for a string flag.
var flagDefault = regexp.MustCompile(` \(default ("(?:[^"\\]|\\.)*"|[^()"]*)\)$`)

// parseCli reads the top-level help and each verb's.
func parseCli(help []byte, verbHelp map[string][]byte) (Cli, error) {
	_, flags, err := parseFlagHelp(help)
	if err != nil {
		return Cli{}, fmt.Errorf("docsdata: cohere -help: %w", err)
	}
	cli := Cli{Flags: flags}
	verbs := make([]string, 0, len(verbHelp))
	for verb := range verbHelp {
		verbs = append(verbs, verb)
	}
	sort.Strings(verbs)
	for _, verb := range verbs {
		preamble, flags, err := parseFlagHelp(verbHelp[verb])
		if err != nil {
			return Cli{}, fmt.Errorf("docsdata: cohere %s -help: %w", verb, err)
		}
		entry := CliVerb{Name: verb, Flags: flags}
		var summary []string
		for _, line := range preamble {
			trimmed := strings.TrimSpace(line)
			switch {
			case trimmed == "":
			case strings.HasPrefix(trimmed, "usage: "):
				entry.Usage = append(entry.Usage, strings.TrimPrefix(trimmed, "usage: "))
			case len(entry.Usage) > 0 && len(summary) == 0 && strings.HasPrefix(line, "       "):
				entry.Usage = append(entry.Usage, trimmed)
			default:
				summary = append(summary, trimmed)
			}
		}
		if len(entry.Usage) == 0 {
			return Cli{}, fmt.Errorf("docsdata: cohere %s -help printed no usage line", verb)
		}
		entry.Summary = strings.Join(summary, " ")
		cli.Verbs = append(cli.Verbs, entry)
	}
	return cli, nil
}

// parseFlagHelp splits help text into the lines before the first flag and the flags, as
// flag.PrintDefaults writes them: each flag a heading line, then its usage on lines indented four
// spaces and a tab. A flag whose usage is missing refuses the parse, so a change in the format fails
// here rather than publishing flags with no description.
func parseFlagHelp(help []byte) ([]string, []CliFlag, error) {
	var preamble []string
	var flags []CliFlag
	var usage []string
	finish := func() error {
		if len(flags) == 0 {
			return nil
		}
		current := &flags[len(flags)-1]
		if len(usage) == 0 {
			return fmt.Errorf("flag -%s has no usage", current.Name)
		}
		text := strings.Join(usage, "\n")
		if match := flagDefault.FindStringSubmatchIndex(text); match != nil {
			value := text[match[2]:match[3]]
			if strings.HasPrefix(value, `"`) {
				unquoted, err := strconv.Unquote(value)
				if err != nil {
					return fmt.Errorf("flag -%s: default %s: %w", current.Name, value, err)
				}
				value = unquoted
			}
			current.Default = value
			text = text[:match[0]]
		}
		current.Usage = text
		usage = nil
		return nil
	}
	for _, line := range strings.Split(strings.TrimRight(string(help), "\n"), "\n") {
		if match := flagHeading.FindStringSubmatch(line); match != nil {
			if err := finish(); err != nil {
				return nil, nil, err
			}
			flags = append(flags, CliFlag{Name: match[1], Argument: match[2]})
			continue
		}
		if len(flags) == 0 {
			preamble = append(preamble, line)
			continue
		}
		continued, isUsage := strings.CutPrefix(line, "    \t")
		if !isUsage {
			return nil, nil, fmt.Errorf("line %q is neither a flag nor its usage", line)
		}
		usage = append(usage, continued)
	}
	if err := finish(); err != nil {
		return nil, nil, err
	}
	if len(flags) == 0 {
		return nil, nil, fmt.Errorf("no flags")
	}
	return preamble, flags, nil
}
