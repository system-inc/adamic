package tailwind

import (
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
)

func messageDeprecatedClassReplaceable(className string, replacement string) rule.Message {
	return rule.Message{
		Id: "deprecatedClassReplaceable",
		Description: "The class \"" + className + "\" was renamed in a Tailwind release and \"" +
			replacement + "\" is the current spelling. Both work today, which is the problem: the old " +
			"name keeps compiling until the release that removes it, and by then it is spread across " +
			"files nobody is looking at. Rename it while the two are still equivalent.",
	}
}

func messageDeprecatedClassIrreplaceable(className string) rule.Message {
	return rule.Message{
		Id: "deprecatedClassIrreplaceable",
		Description: "The class \"" + className + "\" is deprecated and has no direct replacement. " +
			"The opacity utilities were removed in favour of writing the opacity into the color " +
			"itself, such as `bg-black/50` instead of `bg-black bg-opacity-50`, so this needs a small " +
			"rewrite rather than a rename.",
	}
}

// NoDeprecatedClassesOptions lets a project name the surfaces that carry class strings.
type NoDeprecatedClassesOptions struct {
	TailwindLocationOptions
	TailwindClassLiteralOptions
}

// deprecation is one renamed or removed utility.
//
// Pattern matches the class's base, which is the name with its variants and importance marker
// stripped. Replacement is the current spelling, with `$1` standing for whatever the pattern
// captured; an empty Replacement means the utility was removed outright and no rename exists.
type deprecation struct {
	Pattern     *regexp.Regexp
	Replacement string
	// SinceMajor and SinceMinor are the Tailwind release that deprecated it. A project on an older
	// Tailwind is not wrong to use the old name, so the rule stays quiet about renames that have
	// not happened yet for them.
	SinceMajor int
	SinceMinor int
}

// deprecations is the table upstream keeps, ported directly.
//
// This is a real table rather than a computation, which is why it ports cleanly: unlike collapsing,
// nothing here is derived from the theme, so there is no engine to ask. It does go stale with each
// Tailwind release that renames something, and the staleness is visible the same way the collapse
// table's is: a class this does not know is simply not reported, so the cost of falling behind is
// silence rather than a wrong answer.
var deprecations = []deprecation{
	// Tailwind 4.0 renamed the unsuffixed scale steps, so `shadow` became `shadow-sm` and the old
	// `shadow-sm` became `shadow-xs`. That shift is why these read as surprising: the old name still
	// resolves, to a different size than it used to.
	{Pattern: regexp.MustCompile(`^shadow$`), Replacement: "shadow-sm", SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^inset-shadow$`), Replacement: "inset-shadow-sm", SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^drop-shadow$`), Replacement: "drop-shadow-sm", SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^blur$`), Replacement: "blur-sm", SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^backdrop-blur$`), Replacement: "backdrop-blur-sm", SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^rounded$`), Replacement: "rounded-sm", SinceMajor: 4, SinceMinor: 0},

	// The opacity utilities were removed rather than renamed: opacity moved into the color itself,
	// so `bg-black bg-opacity-50` is now `bg-black/50`. No mechanical replacement exists.
	{Pattern: regexp.MustCompile(`^bg-opacity-(.*)$`), SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^text-opacity-(.*)$`), SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^border-opacity-(.*)$`), SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^divide-opacity-(.*)$`), SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^ring-opacity-(.*)$`), SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^placeholder-opacity-(.*)$`), SinceMajor: 4, SinceMinor: 0},

	// The bare forms come before the captured ones, matching upstream's order: `flex-shrink` must
	// match its own pattern rather than being read as `flex-shrink-` with an empty capture.
	{Pattern: regexp.MustCompile(`^flex-shrink$`), Replacement: "shrink", SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^flex-shrink-(.*)$`), Replacement: "shrink-$1", SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^flex-grow$`), Replacement: "grow", SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^flex-grow-(.*)$`), Replacement: "grow-$1", SinceMajor: 4, SinceMinor: 0},

	{Pattern: regexp.MustCompile(`^overflow-ellipsis$`), Replacement: "text-ellipsis", SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^decoration-slice$`), Replacement: "box-decoration-slice", SinceMajor: 4, SinceMinor: 0},
	{Pattern: regexp.MustCompile(`^decoration-clone$`), Replacement: "box-decoration-clone", SinceMajor: 4, SinceMinor: 0},

	// Tailwind 4.1 reordered the two-axis position names so the block axis comes first.
	{Pattern: regexp.MustCompile(`^bg-left-top$`), Replacement: "bg-top-left", SinceMajor: 4, SinceMinor: 1},
	{Pattern: regexp.MustCompile(`^bg-left-bottom$`), Replacement: "bg-bottom-left", SinceMajor: 4, SinceMinor: 1},
	{Pattern: regexp.MustCompile(`^bg-right-top$`), Replacement: "bg-top-right", SinceMajor: 4, SinceMinor: 1},
	{Pattern: regexp.MustCompile(`^bg-right-bottom$`), Replacement: "bg-bottom-right", SinceMajor: 4, SinceMinor: 1},
	{Pattern: regexp.MustCompile(`^object-left-top$`), Replacement: "object-top-left", SinceMajor: 4, SinceMinor: 1},
	{Pattern: regexp.MustCompile(`^object-left-bottom$`), Replacement: "object-bottom-left", SinceMajor: 4, SinceMinor: 1},
	{Pattern: regexp.MustCompile(`^object-right-top$`), Replacement: "object-top-right", SinceMajor: 4, SinceMinor: 1},
	{Pattern: regexp.MustCompile(`^object-right-bottom$`), Replacement: "object-bottom-right", SinceMajor: 4, SinceMinor: 1},
}

// NoDeprecatedClasses reports class names a Tailwind release renamed or removed.
//
//	valid:   <div className="shrink-0 rounded-sm" />
//	invalid: <div className="flex-shrink-0" />
//	invalid: <div className="hover:flex-shrink-0" />
//	invalid: <div className="bg-opacity-50" />
//
// The reason this is worth a rule rather than an upgrade note is that both spellings work. A
// deprecated class keeps compiling and keeps rendering until the release that removes it, so
// nothing surfaces it: no error, no visual difference, no failing test. It spreads by copy-paste
// into files nobody is reading, and the cost lands all at once on the upgrade that finally drops it.
//
// Some of these are worse than a rename. Tailwind 4 shifted the unsuffixed scale steps, so `shadow`
// still resolves but to a different size than it did before, which is a silent visual change rather
// than a broken class.
//
// Matching is against the class's base, with variants and importance stripped and then restored in
// the fix, so `sm:hover:flex-grow-2!` becomes `sm:hover:grow-2!`. Measured against upstream rather
// than assumed: the plugin dissects the class and matches the base, and a port that matched the raw
// name would miss every prefixed use, which is most of them in real markup.
var NoDeprecatedClasses = rule.Rule{
	Name: "better-tailwindcss/no-deprecated-classes",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		surfaces := DefaultClassLiteralSurfaces()
		if configured, isConfigured := rule.OptionsAs[NoDeprecatedClassesOptions](options); isConfigured {
			surfaces = configured.ClassLiteralSurfaces()
		}

		reader := surfaces.ReaderFor(ctx.FileCache)

		report := func(node *ast.Node) {
			for _, literal := range reader.ClassLiteralsIn(node) {
				reportDeprecations(ctx, literal)
			}
		}

		listeners := rule.Listeners{}
		for _, kind := range ListenerKinds() {
			listeners[kind] = report
		}
		return listeners
	},
}

// reportDeprecations finds renamed classes in one literal and proposes each rename at its own class.
//
// Each finding's fix rewrites its own class and nothing else. The whole-literal rewrite this replaced
// was attached to every finding, so a literal with two deprecated classes proposed the same edit
// twice and the engine refused the second, and the rewrite also claimed every separator, colliding
// with `no-unnecessary-whitespace` in the same pass. See class_tokens.go.
//
// A repeat of an earlier class is reported and not renamed: `no-duplicate-classes` deletes it, and
// a rename of the same bytes would be refused as an overlap. Once it is gone there is nothing left
// to rename.
func reportDeprecations(ctx rule.Context, literal ClassLiteral) {
	classes := SplitClasses(literal.Text)
	if len(classes) == 0 {
		return
	}

	// The class tokens line up with `classes` one for one, or there is no fix to offer.
	var tokens []classToken
	tokenized := false
	if ctx.SourceFile != nil {
		tokens, tokenized = classTokensIn(ctx.SourceFile.Text(), literal.Range, literal.Text)
	}
	classTokens := classesOf(tokens)
	if tokenized && len(classTokens) == len(classes) {
		for index, token := range classTokens {
			if token.Text != classes[index] {
				tokenized = false
				break
			}
		}
	} else {
		tokenized = false
	}

	type finding struct {
		className   string
		replacement string
		index       int
	}
	var findings []finding
	for index, className := range classes {
		if replacement, isDeprecated := deprecationFor(className); isDeprecated {
			findings = append(findings, finding{className: className, replacement: replacement, index: index})
		}
	}

	if len(findings) == 0 {
		return
	}

	// A literal holding a template hole is only partly known, so its classes can be reported but its
	// text must not be rewritten from the fragments the reader could see.
	canFix := !strings.Contains(literal.Text, "${")

	for _, found := range findings {
		if found.replacement == "" {
			// Removed outright: reported and left as written, because this rule has no rewrite to offer
			// and dropping the class would change what renders.
			ctx.ReportRange(literal.Range, messageDeprecatedClassIrreplaceable(found.className))
			continue
		}

		message := messageDeprecatedClassReplaceable(found.className, found.replacement)
		if !canFix {
			ctx.ReportRange(literal.Range, message)
			continue
		}

		// Without tokens the literal's source is not its decoded value, and it is reported without a
		// fix for the reason no-unnecessary-whitespace gives.
		var fixes []rule.Fix
		if tokenized && !containsString(classes[:found.index], found.className) {
			fixes = []rule.Fix{rule.ReplaceRange(classTokens[found.index].Range, found.replacement)}
		}

		ctx.Report(rule.Diagnostic{
			Range:      literal.Range,
			Message:    message,
			SourceFile: ctx.SourceFile,
			Fixes:      fixes,
		})
	}
}

// deprecationFor returns the current spelling of a class, and whether it is deprecated at all.
//
// An empty replacement with a true second value means deprecated-and-removed: there is a finding to
// report and no rename to propose.
func deprecationFor(className string) (string, bool) {
	variants, base, important := dissectClass(className)

	for _, entry := range deprecations {
		if !tailwindAtLeast(entry.SinceMajor, entry.SinceMinor) {
			continue
		}
		match := entry.Pattern.FindStringSubmatchIndex(base)
		if match == nil {
			continue
		}
		if entry.Replacement == "" {
			return "", true
		}

		replacedBase := string(entry.Pattern.ExpandString(nil, entry.Replacement, base, match))
		return buildClass(variants, replacedBase, important), true
	}

	return "", false
}

// dissectClass splits a class into its variants, its base, and whether it is important.
//
// The base is what the deprecation table matches, so `sm:hover:flex-grow-2!` has base `flex-grow-2`
// and is reported exactly as the bare form would be. Matching the whole class name instead would
// miss every prefixed use, and prefixed uses are most of real markup.
//
// Splitting on `:` is safe for variants in a way it would not be for finding a class's root,
// because an arbitrary variant's brackets cannot contain an unescaped colon at the top level in
// practice; upstream splits the same way. The importance marker is a trailing `!`.
func dissectClass(className string) (variants string, base string, important bool) {
	base = className

	if index := strings.LastIndex(base, ":"); index >= 0 {
		variants = base[:index+1]
		base = base[index+1:]
	}

	if strings.HasSuffix(base, "!") {
		important = true
		base = strings.TrimSuffix(base, "!")
	}

	return variants, base, important
}

// buildClass reassembles what dissectClass took apart.
func buildClass(variants string, base string, important bool) string {
	rebuilt := variants + base
	if important {
		rebuilt += "!"
	}
	return rebuilt
}

// tailwindAtLeast reports whether the Tailwind this table was generated against is new enough for a
// deprecation to apply.
//
// A project on Tailwind 4.0 is not wrong to write `bg-left-top`, because 4.1 is what renamed it. The
// version comes from the collapse table, which records the engine that produced it, so both tables
// describe the same Tailwind rather than drifting apart.
func tailwindAtLeast(major int, minor int) bool {
	installedMajor, installedMinor := parseTailwindVersion(tailwindVersionForDeprecations())
	if installedMajor != major {
		return installedMajor > major
	}
	return installedMinor >= minor
}

// tailwindVersionForDeprecations is indirected so a test can pin a version.
//
// It reads the version the collapse table was generated against rather than asking the filesystem,
// so both tables describe the same Tailwind. A rule reporting renames from a newer release than the
// one the tables were built from would be reporting an engine nobody here is running.
var tailwindVersionForDeprecations = func() string { return tailwindengine.TailwindVersion }

// parseTailwindVersion reads the leading major and minor out of a semantic version.
//
// An unreadable version reports as 0.0, which disables every deprecation rather than enabling all
// of them. That is the safe direction: this rule reporting nothing is a gap someone notices when
// they upgrade, while reporting renames that have not happened yet is noise on correct code.
//
// Cut rather than Split, because this runs for every deprecation entry against every class the rule
// reads and Split allocates its slice each time: 670K objects on a cold ahra run (#smshtp5). Cutting
// at the first two dots reads the same two fields without allocating, and parsing on every call keeps
// a test's pinned version working with no cache to invalidate.
func parseTailwindVersion(version string) (int, int) {
	majorText, rest, hasMinor := strings.Cut(version, ".")
	if !hasMinor {
		return 0, 0
	}
	minorText, _, _ := strings.Cut(rest, ".")

	major, majorOk := atoi(majorText)
	minor, minorOk := atoi(minorText)
	if !majorOk || !minorOk {
		return 0, 0
	}
	return major, minor
}

func atoi(text string) (int, bool) {
	if text == "" {
		return 0, false
	}
	value := 0
	for _, character := range text {
		if character < '0' || character > '9' {
			return 0, false
		}
		value = value*10 + int(character-'0')
	}
	return value, true
}
