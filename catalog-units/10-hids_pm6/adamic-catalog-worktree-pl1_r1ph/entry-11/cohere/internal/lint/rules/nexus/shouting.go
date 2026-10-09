package nexus

import (
	"regexp"
	"strings"

	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
)

// What counts as shouting, in one place.
//
// All-caps reads as aggression, and it spreads: a reader mirrors the register of what they are
// answering, so one shouted line ratchets a whole file's comments into block capitals. In a codebase
// written largely by minds that read their own prose back as context, the volume compounds.
//
// The discriminator is shape, and it was bought with a corpus survey rather than guessed. A naive
// "is it uppercase" test cannot separate shouted English from an acronym, and real prose is full of
// both: the measured corpus put NOT, THE, IS, AND, ONE right beside SDK, API, DB, and RPC. An
// acronym is usually a consonant cluster, or carries digits or underscores; a shouted word almost
// always has a vowel and is a real word. Vowel-plus-allowlist was right across the whole corpus.
//
// Backticks are the escape hatch, and they are what keeps the allowlist small: anything that is code
// wears backticks, and inside them nothing is read as shouting.

// allowedUppercaseTokens are acronyms, identifiers, and protocol names that are legitimately
// uppercase and must never read as shouting. A token here is always allowed, whatever its shape.
//
// Ported from the TypeScript original, which earned each entry against a live corpus. The list is
// long on purpose: every entry is a word this codebase actually speaks, and a missing one shows up
// as a false positive that teaches people to disable the rule.
var allowedUppercaseTokens = map[string]bool{
	"AAB": true, "AAC": true, "ACH": true, "ACID": true, "AES": true, "AI": true, "AKA": true, "ANN": true,
	"ANSI": true, "AOV": true, "API": true, "APK": true, "APNS": true, "ARIA": true, "ARN": true,
	"ARP": true, "ARR": true, "ASAP": true, "ASCII": true, "ASIN": true, "ASN": true,
	"AST": true, "AVI": true,
	"AVIF": true, "AWS": true, "B2B": true, "BCP": true, "BEL": true, "BEM": true, "BOM": true, "BTC": true,
	// The six filename and environment tokens below arrived from the TypeScript original after this
	// port was written. They are separated out because they are the one group here that is not an
	// acronym: every one is a pronounceable English word, which is exactly why the vowel heuristic
	// reads them as shouting and why they need naming rather than inferring.
	//
	// Backticking them at the call site is the wrong repair. A filename in the middle of a sentence
	// is the subject being discussed rather than a literal to match character for character, and
	// code font would say the opposite. `PATH` earns its place by the same noun test the operating
	// system entries already pass; the home directory, shell and terminal variables are deliberately
	// absent because those are read for their value instead of named as a thing.
	"AGENTS": true, "CHANGELOG": true, "CLAUDE": true, "LICENSE": true, "PATH": true, "README": true,
	"CAC": true, "CAF": true, "CAPI": true, "CCPA": true, "CD": true, "CDATA": true, "CDN": true,
	"CDT": true, "CFA": true, "CFO": true,
	"CGI": true, "CHIPS": true, "CI": true, "CJS": true, "CLI": true, "CMYK": true, "COA": true,
	"COGS": true, "CORS": true, "CPA": true, "CPM": true, "CPU": true, "CRUD": true, "CSS": true,
	"CST": true, "CSV": true, "CTA": true, "CTE": true, "CTR": true, "CUSIP": true, "DB": true, "DEL": true,
	"DER": true, "DKIM": true, "DM": true, "DNA": true, "DNS": true, "DOM": true, "DOS": true, "DRY": true,
	"DSHEA": true, "DTO": true, "EBML": true, "EBU": true, "ECDH": true, "ECDSA": true, "ECE": true,
	"ECMA": true, "EDT": true, "EIN": true, "EKG": true, "ENV": true, "EOD": true, "EOF": true, "EOL": true,
	"ESC": true, "ESM": true, "EST": true, "ETA": true, "ETH": true, "EXIF": true, "FAQ": true,
	"FBAN": true, "FCM": true, "FDA": true, "FFE": true, "FHIR": true, "FICA": true, "FIFO": true,
	"FIXME": true, "FK": true,
	"FSA": true, "FSI": true, "FTC": true, "FTP": true, "FYI": true, "GA4": true, "GAQL": true,
	"GCLID": true, "GDPR": true, "GIF": true, "GIT": true, "GPU": true, "GUI": true, "GUID": true,
	"HAST": true, "HEIC": true, "HEIF": true, "HEVC": true, "HEX": true, "HIPAA": true, "HMAC": true,
	"HSA": true, "HSLA": true, "HTML": true, "HTTP": true, "HTTPS": true, "IAM": true, "IANA": true,
	"IAP": true, "ICO": true, "ICU": true, "IDE": true, "IDFA": true, "IDK": true, "IEC": true,
	"IEEE": true, "IETF": true, "IFD": true, "IIFE": true, "IP": true, "IPA": true, "IPC": true,
	"IRA": true, "ISIN": true, "ISO": true, "JFIF": true, "JIT": true, "JPEG": true, "JPG": true,
	"JS": true, "JSON": true, "JSONC": true, "JSONL": true, "JSX": true, "JWT": true, "KPI": true,
	"LIFO": true, "LLM": true, "LRE": true, "LRI": true, "LRM": true, "LRO": true, "LTV": true,
	"LUFS": true, "MDAST": true, "MDT": true, "MIME": true, "MIT": true, "MKV": true, "ML": true,
	"MOV": true, "MPEG": true, "MRR": true, "MST": true, "MTD": true, "MVP": true, "NA": true,
	"NAICS": true, "NANP": true, "NDJSON": true, "NET": true, "NOTE": true, "NPM": true, "NUL": true,
	"OAUTH": true, "OCR": true, "OCSP": true, "OFX": true, "OGG": true, "OIDC": true, "OK": true,
	"OKLCH": true, "ONNX": true, "OOM": true, "ORM": true, "OS": true, "OTA": true, "OTP": true,
	"PDF": true, "PDI": true, "PDT": true, "PEM": true, "PHP": true, "PI": true, "PID": true, "PII": true,
	"PK": true, "PKCE": true, "PNG": true, "PNPM": true, "POSIX": true, "PR": true, "PST": true,
	"PTY": true, "QA": true, "QTD": true, "R2": true, "RAM": true, "REPL": true, "REST": true, "RGBA": true,
	"RIFF": true, "RLE": true, "RLI": true, "RLM": true, "RLO": true, "ROAS": true, "ROI": true,
	"RPC": true, "RSA": true, "S3": true, "SAAS": true, "SAFE": true, "SDK": true, "SEO": true, "SES": true,
	"SGR": true, "SHA": true, "SID": true, "SKU": true, "SLA": true, "SNS": true, "SOI": true, "SPA": true,
	"SPKI": true, "SQL": true, "SRE": true, "SSD": true, "SSE": true, "SSH": true, "SSL": true, "SSN": true,
	"SSO": true, "STS": true, "SVG": true, "TIFF": true, "TL": true, "TLS": true, "TODO": true,
	"TOML": true, "TS": true, "TSX": true, "TTL": true, "TTY": true, "TUI": true, "UDID": true,
	"UGC": true, "UI": true,
	"UPC": true, "UPS": true, "URI": true, "URL": true, "URN": true, "US": true, "USPS": true, "UTC": true,
	"UTF": true, "UTM": true, "UUID": true, "UX": true, "VINT": true, "WAL": true, "WASM": true,
	"WCAG": true, "WEBM": true, "WEBP": true, "WEBVTT": true, "WIP": true, "XML": true, "YAML": true,
	"YTD": true, "YUV": true, "YYYY": true,
}

// currencyCodes are ISO 4217 codes, which a sentence names the way it names a protocol: "amounts in
// KRW have no minor unit" is prose about a currency, not a raised voice.
//
// In the original these are read from the live currency table rather than retyped, because two lists
// of the same facts drift. That drift is not hypothetical: a parenthetical once read
// "(BIF, DJF, GNF, KMF, RWF, VUV, XPF)" with only some codes backticked, because only the
// vowel-bearing ones tripped the heuristic. Here they are a generated copy of that table, which is
// the same bet with a different failure mode: it cannot drift within a sentence, but it can go stale
// against the source. Regenerate rather than hand-edit.
var currencyCodes = map[string]bool{
	"AED": true, "ALL": true, "AMD": true, "AOA": true, "ARS": true, "AUD": true, "AZN": true, "BAM": true,
	"BDT": true, "BIF": true, "BOB": true, "BRL": true, "CAD": true, "CHF": true, "CLP": true, "CNY": true,
	"COP": true, "CRC": true, "CZK": true, "DJF": true, "DKK": true, "DOP": true, "DZD": true, "EGP": true,
	"ETB": true, "EUR": true, "GBP": true, "GEL": true, "GNF": true, "GTQ": true, "HKD": true, "HUF": true,
	"IDR": true, "ILS": true, "INR": true, "JPY": true, "KES": true, "KHR": true, "KMF": true, "KRW": true,
	"KZT": true, "LBP": true, "LKR": true, "MAD": true, "MDL": true, "MGA": true, "MKD": true, "MMK": true,
	"MXN": true, "MYR": true, "NGN": true, "NOK": true, "NPR": true, "NZD": true, "PEN": true, "PHP": true,
	"PKR": true, "PLN": true, "PYG": true, "QAR": true, "RON": true, "RSD": true, "RUB": true, "RWF": true,
	"SAR": true, "SEK": true, "SGD": true, "THB": true, "TRY": true, "TTD": true, "TWD": true, "TZS": true,
	"UAH": true, "UGX": true, "USD": true, "UYU": true, "UZS": true, "VND": true, "VUV": true, "XAF": true,
	"XOF": true, "XPF": true, "YER": true, "ZAR": true,
}

// shoutedTwoLetterWords are two-letter words that are shouting when capitalized.
//
// At two characters the vowel test cannot help, since IS and IP have the same shape. So the short
// words are named, and every other two-letter token is assumed to be an abbreviation.
var shoutedTwoLetterWords = map[string]bool{
	"AM": true, "AN": true, "AS": true, "AT": true, "BE": true, "BY": true, "DO": true, "GO": true,
	"HE": true, "IF": true, "IN": true, "IS": true, "IT": true, "ME": true, "MY": true, "NO": true,
	"OF": true, "ON": true, "OR": true, "SO": true, "TO": true, "UP": true, "US": true, "WE": true,
}

var (
	digitOrUnderscore = regexp.MustCompile(`[0-9_]`)
	vowel             = regexp.MustCompile(`[AEIOUY]`)
	uppercaseToken    = regexp.MustCompile(`\b[A-Z][A-Z0-9_]*\b`)
)

// isShoutedToken reports whether an uppercase token is shouting rather than an acronym.
func isShoutedToken(token string) bool {
	if allowedUppercaseTokens[token] || currencyCodes[token] {
		return false
	}
	// Digits or underscores mean an identifier, never a shout: C1, H3, DATABASE_URL.
	if digitOrUnderscore.MatchString(token) {
		return false
	}
	if len(token) < 2 {
		return false
	}
	if len(token) == 2 {
		return shoutedTwoLetterWords[token]
	}
	// A vowel is the tell of a pronounceable word; a consonant cluster is an acronym.
	return vowel.MatchString(token)
}

var (
	// A fence spans lines, so its newlines are preserved rather than collapsed.
	fencedBlock = regexp.MustCompile("(?s)```.*?```")

	// Inline backticks, including a pair that wraps across a line. A block comment reflows, so a
	// long snippet often ends up split with the continuation behind an asterisk gutter. Matching a
	// single line would leave that pair unmasked and read the code inside it as shouting, which is
	// the opposite of what backticks are for.
	inlineBackticks = regexp.MustCompile("`[^`]*`")

	// Quoted strings, for the same reason backticks are masked: a quote marks the text inside it as
	// a literal rather than as the writer's voice. Prose about strings needs to show them, and a
	// sentence discussing "ORDER STATUS" is about a value, not shouting it.
	//
	// One line at a time here; a pair that wraps onto the next line is maskWrappedDoubleQuotes's.
	doubleQuoted = regexp.MustCompile(`"[^"\n]*"`)

	// Single-quoted tokens, deliberately narrower than the double-quote rule, because an apostrophe
	// is the same character as an opening single quote and prose is full of them. A naive pair-match
	// would start at a possessive and swallow the rest of the sentence, blinding the rule to
	// whatever followed. So only a pair that looks like a token is masked: no spaces inside.
	singleQuotedToken = regexp.MustCompile(`'[A-Za-z0-9_.\-/]+'`)

	// A single-quoted phrase in capitals, such as 'RENAME COLUMN' or 'ON CONFLICT DO NOTHING'. The
	// token pattern above stops at the first space, so before this a quoted keyword phrase was read
	// as shouting even though the writer had quoted it, which is exactly the repair the rule asks for.
	//
	// No lowercase letter is allowed inside, and that is what keeps it from repeating the apostrophe
	// failure: a possessive or a contraction is followed by lowercase prose before any closing quote,
	// so it can never form a match. The boundary half lives in maskSingleQuotedCapitalPhrases.
	singleQuotedCapitalPhrase = regexp.MustCompile(`'[A-Z0-9_][A-Z0-9_ .,:=\-/]*[A-Z0-9_]'`)

	notNewline  = regexp.MustCompile(`[^\n]`)
	jsDocGutter = regexp.MustCompile(`^\s*\*?\s?`)
	// A block comment's gutter alone, leading whitespace, the asterisk and one space or tab, which
	// unlike jsDocGutter leaves a line with no asterisk untouched so its indentation can be measured.
	blockCommentGutter = regexp.MustCompile(`^[ \t]*\*(?: |\t|$)`)
	exampleTag         = regexp.MustCompile(`^@example\b`)
	anyJsDocTag        = regexp.MustCompile(`^@\w+`)
	commandLine        = regexp.MustCompile(`^\s*(\$|ahra\s|git\s|pnpm\s|npm\s|sqlite3\s|curl\s|s\s+c\b)`)
)

// maskCodeAndCommands blanks out anything that is code, so nothing inside it is ever read as
// shouting: fenced blocks, inline backticks, quoted strings, JSDoc examples, and command lines.
//
// Replacing with spaces rather than deleting preserves every offset, so a caller reporting positions
// keeps true line and column numbers. Newlines are preserved for the same reason: collapsing a
// fence's newlines to spaces keeps character offsets right but destroys the line count, and every
// line after the fence then reports against the wrong source line.
func maskCodeAndCommands(text string) string {
	masked := blankKeepingNewlines(fencedBlock, text)
	masked = maskIndentedCodeBlocks(masked)
	masked = blankKeepingNewlines(inlineBackticks, masked)
	masked = blankAll(doubleQuoted, masked)
	masked = maskWrappedDoubleQuotes(masked)
	masked = blankAll(singleQuotedToken, masked)
	masked = maskSingleQuotedCapitalPhrases(masked)
	masked = maskJsDocExamples(masked)
	masked = maskCommandLines(masked)
	return masked
}

// blankKeepingNewlines replaces every match with spaces, leaving its newlines in place.
func blankKeepingNewlines(pattern *regexp.Regexp, text string) string {
	return pattern.ReplaceAllStringFunc(text, func(match string) string {
		return notNewline.ReplaceAllString(match, " ")
	})
}

// blankAll replaces every match with spaces of the same length. Used only for single-line patterns,
// where there is no newline to preserve.
func blankAll(pattern *regexp.Regexp, text string) string {
	return pattern.ReplaceAllStringFunc(text, func(match string) string {
		return strings.Repeat(" ", len(match))
	})
}

// maskWrappedDoubleQuotes blanks a double-quoted phrase that a block comment wrapped onto the next
// line, which the one-line pattern cannot see.
//
// A comment reflows like prose, so `"the caller's job and you must NOT"` ends up as its opening half
// on one line and its closing half behind the next line's gutter (ahra's Shouting.ts did exactly
// this), and the capitals inside a quote the writer closed read as shouting. The backtick mask
// already spans a wrap for the same reason.
//
// It runs after the one-line pass, which pairs every two quotes sharing a line, so each line has at
// most one quote left. Only those leftovers pair, and only across one wrap: the leftover on a line
// with the leftover on the very next line. Pairing quotes freely across lines would let a stray
// inch mark (`a 5" board`) open a span that eats a later quote's opening, leaving the phrase that
// quote protected unmasked; confined to leftovers on adjacent lines, the pass can only blank text,
// never unblank what the one-line pass already masked.
func maskWrappedDoubleQuotes(text string) string {
	if !strings.Contains(text, "\n") || strings.Count(text, `"`) < 2 {
		return text
	}

	lines := strings.Split(text, "\n")
	for index := 0; index+1 < len(lines); index++ {
		opening := strings.IndexByte(lines[index], '"')
		if opening < 0 {
			continue
		}
		closing := strings.IndexByte(lines[index+1], '"')
		if closing < 0 {
			continue
		}
		lines[index] = lines[index][:opening] + strings.Repeat(" ", len(lines[index])-opening)
		// The closing quote is blanked with the rest, so it cannot also open a pair with the line
		// after it.
		lines[index+1] = strings.Repeat(" ", closing+1) + lines[index+1][closing+1:]
	}
	return strings.Join(lines, "\n")
}

// maskSingleQuotedCapitalPhrases blanks a capital phrase the writer set in single quotes.
//
// A quote is only a quote at a word boundary: the opening one must not follow a letter or a digit,
// and the closing one must not be followed by one. Without that, the plural possessive in "the
// users' 'NOT NULL'" could open a pair, and a quote glued to a suffix ('RENAME COLUMN'd) would mask
// a phrase the writer never closed. RE2 has no lookaround, so the boundary is checked on the match
// indices rather than in the pattern.
func maskSingleQuotedCapitalPhrases(text string) string {
	// Two quotes and a space between them are the least a phrase can be, and almost no comment holds
	// a quote at all, so the common case pays one byte scan.
	if strings.Count(text, "'") < 2 {
		return text
	}

	matches := singleQuotedCapitalPhrase.FindAllStringIndex(text, -1)
	if matches == nil {
		return text
	}

	masked := []byte(text)
	for _, match := range matches {
		start, end := match[0], match[1]
		if start > 0 && isAsciiLetterOrDigit(text[start-1]) {
			continue
		}
		if end < len(text) && isAsciiLetterOrDigit(text[end]) {
			continue
		}
		for index := start; index < end; index++ {
			masked[index] = ' '
		}
	}
	return string(masked)
}

// isAsciiLetterOrDigit reports whether a byte is a word character a quote could be glued to.
func isAsciiLetterOrDigit(character byte) bool {
	return (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
		(character >= '0' && character <= '9')
}

// maskJsDocExamples blanks the lines under an @example tag, which are code whatever they look like.
//
// The lines under it are a call and its result, written the way a reader would type them, so
// hour24To12(0) with a comment showing 'AM' is a snippet rather than someone shouting. Backticking
// inside an example is the wrong repair too: the whole block is already code by context. The block
// runs from the tag to the next tag or the end of the comment, which is how JSDoc delimits it.
func maskJsDocExamples(text string) string {
	// An example block opens with an `@example` tag, so a text with no `@` at all cannot hold one.
	// Measured at 12.2ms of the 35.7ms masking cost before this check, from splitting and running
	// two regexes per line on every comment in the tree.
	if !strings.ContainsRune(text, '@') {
		return text
	}

	lines := strings.Split(text, "\n")
	insideExample := false
	for index, line := range lines {
		withoutGutter := jsDocGutter.ReplaceAllString(line, "")

		if exampleTag.MatchString(withoutGutter) {
			insideExample = true
			lines[index] = strings.Repeat(" ", len(line))
			continue
		}
		// Any other JSDoc tag closes the example.
		if insideExample && anyJsDocTag.MatchString(withoutGutter) {
			insideExample = false
		}
		if insideExample {
			lines[index] = strings.Repeat(" ", len(line))
		}
	}
	return strings.Join(lines, "\n")
}

// maskIndentedCodeBlocks blanks an indented code block, the markdown form of a fence: lines indented
// four spaces or a tab, opened after a blank line (or at the start of the text) and running until a
// line that is not indented. Without it, a sandbox command posted as an indented block had its
// $HOME and TMPDIR read as shouting, and the quieted draft lowercased them into a command that no
// longer ran, while the same command in a fence survived (#vy055vj).
//
// An indented line straight after prose is a paragraph continuation in CommonMark, not code, so it
// is still read; that is what the blank-line condition is for. A block comment's gutter is set aside
// before the indentation is measured, so a comment's own indentation is never mistaken for a code
// block, while an indented block written inside a JSDoc comment still is one. Mirrors nexus's
// Shouting.ts, so the quieter and this rule agree on what is code.
func maskIndentedCodeBlocks(masked string) string {
	// Nothing can be indented four deep without four spaces or a tab somewhere.
	if !strings.Contains(masked, "    ") && !strings.ContainsRune(masked, '\t') {
		return masked
	}

	lines := strings.Split(masked, "\n")
	previousIsBlank := true
	insideBlock := false
	for index, line := range lines {
		content := blockCommentGutter.ReplaceAllString(line, "")
		if text.TrimWhitespace(content) == "" {
			previousIsBlank = true
			continue
		}
		isIndented := strings.HasPrefix(content, "    ") || strings.HasPrefix(content, "\t")
		insideBlock = isIndented && (insideBlock || previousIsBlank)
		previousIsBlank = false
		if insideBlock {
			lines[index] = strings.Repeat(" ", len(line))
		}
	}
	return strings.Join(lines, "\n")
}

// maskCommandLines blanks a line that is plainly a command the reader is meant to run.
func maskCommandLines(text string) string {
	// Nothing to mask unless one of the command starters appears somewhere in the text. One scan of
	// the whole string is far cheaper than two regexes per line, and the overwhelming majority of
	// comments hold no command at all.
	//
	// Measured on 300 real files: this stage was 15.2ms of the 35.7ms masking cost, the single most
	// expensive step, because it ran a gutter-stripping ReplaceAllString and a match on every line
	// of every file. The abbreviation gate taught the same lesson one layer down: the cheapest
	// question first, and a per-line regex is where a small constant gets multiplied.
	if !containsAnyCommandStarter(text) {
		return text
	}

	lines := strings.Split(text, "\n")
	for index, line := range lines {
		// The gutter is stripped before matching, which the TypeScript original does not do. Its
		// pattern anchors at the start of the line, so a command inside a block comment sits behind
		// a " * " gutter and never matches, and the rule reads the command as prose. The
		// example-block masker in this same file already strips the gutter for exactly this reason,
		// so this follows that convention rather than inventing one.
		if commandLine.MatchString(jsDocGutter.ReplaceAllString(line, "")) {
			lines[index] = strings.Repeat(" ", len(line))
		}
	}
	return strings.Join(lines, "\n")
}

// shoutedTokensIn returns the distinct shouted tokens in a comment body, in first-seen order.
//
// The body is masked first, so code is invisible. Order is first-seen rather than sorted because a
// reader looking for what to fix scans the comment top to bottom.
func shoutedTokensIn(body string) []string {
	// The cheapest question first, and it lets the whole rule skip 94% of comments.
	//
	// `uppercaseToken` is `\b[A-Z][A-Z0-9_]*\b`, and `isShoutedToken` declines anything under two
	// characters, so nothing this function can report exists in a body without two uppercase letters
	// somewhere. Masking only ever replaces characters with spaces, never adds an uppercase letter,
	// so a body that fails this test cannot pass it after masking either. Checking before masking
	// therefore skips both stages rather than one.
	//
	// Measured over 8,814 real comments: 5.7% hold two adjacent uppercase letters. The byte scan is
	// 579µs against 5,413µs for the regex over the same corpus, and masking is another 6,019µs that
	// this avoids entirely.
	if !hasAdjacentUppercaseLetters(body) {
		return nil
	}

	masked := maskCodeAndCommands(body)

	var tokens []string
	seen := map[string]bool{}
	for _, token := range uppercaseToken.FindAllString(masked, -1) {
		if seen[token] || !isShoutedToken(token) {
			continue
		}
		seen[token] = true
		tokens = append(tokens, token)
	}
	return tokens
}

// commandStarters are the literal prefixes commandLine can match after its optional leading space.
//
// Kept beside the pattern deliberately: if the pattern gains an alternative, this must gain it too,
// or maskCommandLines will stop masking that command and the rule will read it as prose. The gate is
// a superset check, so an extra entry here costs one wasted line scan and a missing one costs a
// false finding.
// The trailing space is dropped from each entry on purpose: the pattern separates the command from
// its argument with `\s`, which matches a tab as well as a space, so requiring a literal space here
// would miss `git\tstatus`. A bare word over-admits slightly, which is the safe direction.
var commandStarters = []string{"$", "ahra", "git", "pnpm", "npm", "sqlite3", "curl"}

// containsAnyCommandStarter reports whether any command prefix appears anywhere in the text.
//
// A superset of what `commandLine` can match, never a subset, for the same reason the abbreviation
// gate is: a gate that admits too much costs one wasted line scan, and a gate that admits too little
// silently stops masking a command, so the rule reads it as prose and shouts at it.
//
// The `s\s+c` arm is the one that cannot be reduced to a literal, since the pattern accepts any run
// of whitespace between the two letters. A fixture caught that: a hand-written `"s c"` entry misses
// `s   c`, and 800 real files did not happen to contain the multi-space form. **Absence from a
// corpus is not absence in general**, which is why this arm keeps its regex rather than a literal.
func containsAnyCommandStarter(text string) bool {
	for _, starter := range commandStarters {
		if strings.Contains(text, starter) {
			return true
		}
	}
	return structureCommandStarter.MatchString(text)
}

// structureCommandStarter is the `s\s+c` arm, which no literal can express.
var structureCommandStarter = regexp.MustCompile(`s\s+c`)

// hasAdjacentUppercaseLetters reports whether two uppercase letters appear in a row.
//
// That is the shortest thing `uppercaseToken` can match that `isShoutedToken` will not immediately
// decline, so it is a superset of what the rule can report. Over-approximating costs one masking
// pass on a comment that turns out to hold only acronyms; under-approximating would silence the rule
// for a real shout, which is the failure that cannot be allowed.
func hasAdjacentUppercaseLetters(text string) bool {
	previousWasUppercase := false
	for index := 0; index < len(text); index++ {
		character := text[index]
		isUppercase := character >= 'A' && character <= 'Z'
		if isUppercase && previousWasUppercase {
			return true
		}
		previousWasUppercase = isUppercase
	}
	return false
}
