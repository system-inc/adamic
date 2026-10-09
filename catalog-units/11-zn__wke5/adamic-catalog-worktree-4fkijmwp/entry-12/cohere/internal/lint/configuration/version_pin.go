package configuration

import (
	"fmt"
	"strconv"
	"strings"

	release "github.com/system-inc/cohere/internal/release/packaging"
)

// releaseVersion is the version this binary was published as, or "dev" for a local build. A variable
// so a test can stand in a release.
var releaseVersion = func() string { return release.Current().Version }

// checkCoherePin refuses a published release the project's pin does not admit, naming both versions.
//
// A local build reports "dev" and is not held to the pin: it names no release, and every build cohere's
// own launcher makes from a commit is one. The pin is for the published binary a machine installed, which
// is the one that can silently be a different release from the one the project chose.
func checkCoherePin(path string, versionRange VersionRange) error {
	version := releaseVersion()
	if version == "dev" {
		return nil
	}
	admitted, err := versionRange.Admits(version)
	if err != nil {
		return fmt.Errorf("lint config %s pins cohere to %q, and this binary's own version %q does not read as a "+
			"release version: %w", path, versionRange.Text, version, err)
	}
	if !admitted {
		return fmt.Errorf("lint config %s pins cohere to %q, and this binary is cohere %s: install a release "+
			"that %q admits, or change the pin", path, versionRange.Text, version, versionRange.Text)
	}
	return nil
}

// VersionRange is the "cohere" key of a project's settings: the cohere releases that project accepts,
// written the way npm writes a dependency range ("^1.0.0", "~1.4.2", ">=1.2.0 <2.0.0", "1.3.0 || 1.4.0").
//
// A project pins cohere so a machine running a different release refuses rather than reporting a
// verdict the project never agreed to. Rule names, options and sets are part of a release's contract,
// so a binary outside the range can read the same file and mean something else by it.
type VersionRange struct {
	// Text is the range as the file wrote it, for messages.
	Text string

	// alternatives are the `||`-separated comparator sets; a version satisfies the range when it
	// satisfies every comparator of any one set.
	alternatives [][]versionComparator
}

type versionComparator struct {
	operator string
	version  semanticVersion
}

type semanticVersion struct {
	major      int
	minor      int
	patch      int
	prerelease []string
}

// ParseVersionRange reads a range, refusing anything it cannot read exactly. A pin that half-parses
// would admit or refuse releases nobody chose, which is worse than no pin.
//
// It reads the forms npm's ranges use for a single dependency: an exact version, `=`, `<`, `<=`, `>`,
// `>=`, a caret, a tilde, space-separated comparators that must all hold, and `||` between
// alternatives. Wildcards (`1.x`, `*`) and hyphen ranges are refused by name rather than guessed at.
func ParseVersionRange(text string) (VersionRange, error) {
	// Go whitespace: the "cohere" pin in cohere's own settings, which no JavaScript tool reads.
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return VersionRange{}, fmt.Errorf("the \"cohere\" range is empty: write the releases this project accepts, like \"^1.0.0\"")
	}
	parsed := VersionRange{Text: trimmed}
	for _, alternative := range strings.Split(trimmed, "||") {
		// Go whitespace: the "cohere" pin in cohere's own settings, which no JavaScript tool reads.
		fields := strings.Fields(alternative)
		if len(fields) == 0 {
			return VersionRange{}, fmt.Errorf("the \"cohere\" range %q has an empty alternative around `||`", trimmed)
		}
		var comparators []versionComparator
		for _, field := range fields {
			expanded, err := parseComparator(field)
			if err != nil {
				return VersionRange{}, fmt.Errorf("the \"cohere\" range %q: %w", trimmed, err)
			}
			comparators = append(comparators, expanded...)
		}
		parsed.alternatives = append(parsed.alternatives, comparators)
	}
	return parsed, nil
}

// parseComparator reads one comparator, expanding a caret or tilde into the bounds npm gives it.
func parseComparator(field string) ([]versionComparator, error) {
	if field == "-" || strings.ContainsAny(field, "xX*") && !strings.ContainsAny(field, "-+") {
		return nil, fmt.Errorf("%q is a wildcard or hyphen range, which this reader does not take; write the bounds, like \">=1.2.0 <2.0.0\"", field)
	}
	for _, operator := range []string{">=", "<=", ">", "<", "=", "^", "~"} {
		rest, found := strings.CutPrefix(field, operator)
		if !found {
			continue
		}
		version, err := parseSemanticVersion(rest)
		if err != nil {
			return nil, err
		}
		switch operator {
		case "^":
			return []versionComparator{{">=", version}, {"<", caretCeiling(version)}}, nil
		case "~":
			return []versionComparator{{">=", version}, {"<", semanticVersion{version.major, version.minor + 1, 0, []string{"0"}}}}, nil
		default:
			return []versionComparator{{operator, version}}, nil
		}
	}
	version, err := parseSemanticVersion(field)
	if err != nil {
		return nil, err
	}
	return []versionComparator{{"=", version}}, nil
}

// caretCeiling is the exclusive upper bound of `^version`: the next release that changes the leftmost
// nonzero part, as npm defines it, so `^0.3.1` stays below 0.4.0 and `^0.0.3` below 0.0.4.
//
// The ceiling carries the lowest prerelease, so a prerelease of the next major (2.0.0-beta) is outside
// `^1.0.0` rather than inside it by sorting below 2.0.0.
func caretCeiling(version semanticVersion) semanticVersion {
	lowest := []string{"0"}
	switch {
	case version.major > 0:
		return semanticVersion{version.major + 1, 0, 0, lowest}
	case version.minor > 0:
		return semanticVersion{0, version.minor + 1, 0, lowest}
	default:
		return semanticVersion{0, 0, version.patch + 1, lowest}
	}
}

// parseSemanticVersion reads MAJOR.MINOR.PATCH with an optional prerelease and build, as semver.org
// defines them. Build metadata is read and dropped, because it does not order versions.
func parseSemanticVersion(text string) (semanticVersion, error) {
	core, _, _ := strings.Cut(text, "+")
	core, prerelease, hasPrerelease := strings.Cut(core, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return semanticVersion{}, fmt.Errorf("%q is not a version: write all three parts, like 1.4.0", text)
	}
	numbers := make([]int, 3)
	for index, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 || (len(part) > 1 && part[0] == '0') {
			return semanticVersion{}, fmt.Errorf("%q is not a version: %q is not a whole number without leading zeros", text, part)
		}
		numbers[index] = number
	}
	version := semanticVersion{major: numbers[0], minor: numbers[1], patch: numbers[2]}
	if hasPrerelease {
		if prerelease == "" {
			return semanticVersion{}, fmt.Errorf("%q is not a version: the prerelease after `-` is empty", text)
		}
		version.prerelease = strings.Split(prerelease, ".")
	}
	return version, nil
}

// Admits reports whether a release version falls inside the range.
//
// It follows npm's rule for prereleases: a prerelease satisfies a set of comparators only when one of
// them names the same major, minor and patch with a prerelease of its own. So `^1.0.0` does not admit
// 1.1.0-beta.1, and `>=1.1.0-beta.0 <2.0.0` does. A project opts into prereleases by writing one.
func (versionRange VersionRange) Admits(version string) (bool, error) {
	candidate, err := parseSemanticVersion(version)
	if err != nil {
		return false, err
	}
	for _, comparators := range versionRange.alternatives {
		if admitsAll(comparators, candidate) {
			return true, nil
		}
	}
	return false, nil
}

func admitsAll(comparators []versionComparator, candidate semanticVersion) bool {
	prereleaseAllowed := len(candidate.prerelease) == 0
	for _, comparator := range comparators {
		order := compareVersions(candidate, comparator.version)
		var holds bool
		switch comparator.operator {
		case "=":
			holds = order == 0
		case ">":
			holds = order > 0
		case ">=":
			holds = order >= 0
		case "<":
			holds = order < 0
		case "<=":
			holds = order <= 0
		}
		if !holds {
			return false
		}
		// The ceiling a caret or tilde adds carries a prerelease too, but it never opts one in: a
		// prerelease of the ceiling's own major, minor and patch sorts at or above it, so `<` refuses it
		// before this line.
		if len(comparator.version.prerelease) > 0 &&
			comparator.version.major == candidate.major && comparator.version.minor == candidate.minor &&
			comparator.version.patch == candidate.patch {
			prereleaseAllowed = true
		}
	}
	return prereleaseAllowed
}

// compareVersions orders two versions by semver precedence: the numbers, then a release above any of
// its prereleases, then the prerelease identifiers left to right.
func compareVersions(left semanticVersion, right semanticVersion) int {
	for _, pair := range [][2]int{{left.major, right.major}, {left.minor, right.minor}, {left.patch, right.patch}} {
		if pair[0] != pair[1] {
			if pair[0] < pair[1] {
				return -1
			}
			return 1
		}
	}
	switch {
	case len(left.prerelease) == 0 && len(right.prerelease) == 0:
		return 0
	case len(left.prerelease) == 0:
		return 1
	case len(right.prerelease) == 0:
		return -1
	}
	for index := 0; index < len(left.prerelease) && index < len(right.prerelease); index++ {
		if order := compareIdentifiers(left.prerelease[index], right.prerelease[index]); order != 0 {
			return order
		}
	}
	return compareInts(len(left.prerelease), len(right.prerelease))
}

// compareIdentifiers orders prerelease identifiers: numeric ones numerically and below any
// alphanumeric one, alphanumeric ones by ASCII.
func compareIdentifiers(left string, right string) int {
	leftNumber, leftErr := strconv.Atoi(left)
	rightNumber, rightErr := strconv.Atoi(right)
	switch {
	case leftErr == nil && rightErr == nil:
		return compareInts(leftNumber, rightNumber)
	case leftErr == nil:
		return -1
	case rightErr == nil:
		return 1
	default:
		return strings.Compare(left, right)
	}
}

func compareInts(left int, right int) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}
