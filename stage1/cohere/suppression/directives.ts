// A port of cohere's internal/lint/ecmascript/directives (directives.go and subject.go) to Adamic 0.1:
// the grammar of suppression comments, and the short list of rules whose findings are about those
// comments. Each piece names the Go it reads as.
//
// Where the port differs from the Go, and why:
//
//   - Go's multiple results are objects or undefined, since stage 0 doesn't lower a tuple as a value
//     yet (gitignore's GAPS.md, gap 8): ParseDisable's (Disable, bool) is a Disable or undefined,
//     ParseEnable's (rules, found) the rules or undefined, splitScope's three results a ScopeRead or
//     undefined, and splitReason's two a RuleList.
//   - Recognize's (scopeWord, honored) is the scope word, or '' where the comment isn't one. It would
//     be undefined, but stage 0 writes `return undefined` from a function returning string | undefined
//     as C that clang refuses (mediaquery's GAPS.md, gap 2). For the same reason, and gap 2 here for an
//     array, splitDirective returns a DirectiveRest and ParseEnable an Enable, objects, which may be
//     undefined, where the natural results are a string and a list.
//   - Go's Scope is an int enum; here it is the union of its names.
//   - Go's strings.TrimSpace trims what Go's unicode.IsSpace calls space, which is not what
//     JavaScript's trim trims: Go takes U+0085 and leaves U+FEFF, JavaScript the other way round. The
//     port writes Go's (goTrimSpace), since the port answers as Go cohere does.

// directives.go: the honored spellings. `cohere-disable` is what new code writes; `verify-disable` is
// what this tool's directives were called before the rename to Cohere; `eslint-disable` is the
// spelling the existing corpus uses, permanent, not transitional; `oxlint-disable` is the spelling of
// the gate Cohere replaces.
export const DirectiveCohere = 'cohere-disable';
export const DirectiveVerify = 'verify-disable';
export const DirectiveEslint = 'eslint-disable';
export const DirectiveOxlint = 'oxlint-disable';

// directives.go: disableDirectives, every honored disable spelling.
const disableDirectives: readonly string[] = [DirectiveCohere, DirectiveVerify, DirectiveEslint, DirectiveOxlint];

// directives.go: enableDirectives, which close a block, one per disable spelling.
const enableDirectives: readonly string[] = ['cohere-enable', 'verify-enable', 'eslint-enable', 'oxlint-enable'];

// directives.go: Scope, the span a disable directive covers: the line after the comment, the line the
// comment sits on, or from the comment's own line to a matching enable, or to the end of the file.
export type Scope = 'next-line' | 'same-line' | 'file';

// directives.go: Disable, one parsed disable directive. rules empty means blanket; reason '' means
// the author did not say why.
export interface Disable {
	readonly scope: Scope;
	readonly rules: readonly string[];
	readonly reason: string;
}

// directives.go: the scope words, in ESLint's vocabulary.
export const ScopeWordDisable = 'eslint-disable';
export const ScopeWordDisableLine = 'eslint-disable-line';
export const ScopeWordDisableNextLine = 'eslint-disable-next-line';
export const ScopeWordEnable = 'eslint-enable';

// isGoSpace is Go's unicode.IsSpace: the Latin-1 spaces '\t', '\n', '\v', '\f', '\r', ' ', U+0085 and
// U+00A0, and Unicode's White_Space above them. Every one is a single UTF-16 unit.
function isGoSpace(unit: number): boolean {
	switch (unit) {
		case 0x09:
		case 0x0a:
		case 0x0b:
		case 0x0c:
		case 0x0d:
		case 0x20:
		case 0x85:
		case 0xa0:
		case 0x1680:
		case 0x2028:
		case 0x2029:
		case 0x202f:
		case 0x205f:
		case 0x3000:
			return true;
	}
	return unit >= 0x2000 && unit <= 0x200a;
}

// goTrimSpace is Go's strings.TrimSpace.
export function goTrimSpace(text: string): string {
	let start = 0;
	while (start < text.length && isGoSpace(text.charCodeAt(start))) {
		start++;
	}
	let end = text.length;
	while (end > start && isGoSpace(text.charCodeAt(end - 1))) {
		end--;
	}
	return text.slice(start, end);
}

// directives.go: isLineComment, whether a comment is the `//` form. A file-scope disable and an enable
// are directives only in a block comment, which is ESLint's grammar.
function isLineComment(commentText: string): boolean {
	return commentText.startsWith('//');
}

// directives.go: stripCommentMarkers removes `//`, `/*`, `*/`, and any `*` continuation markers, so a
// reason that wraps across a block comment's lines parses as one reason.
function stripCommentMarkers(commentText: string): string {
	let text = commentText;

	if (text.startsWith('//')) {
		return text.slice(2);
	}

	// strings.TrimSuffix(strings.TrimPrefix(text, "/*"), "*/")
	if (text.startsWith('/*')) {
		text = text.slice(2);
	}
	if (text.endsWith('*/')) {
		text = text.slice(0, text.length - 2);
	}
	if (!text.includes('\n')) {
		return text;
	}

	const lines = text.split('\n');
	for (let index = 0; index < lines.length; index++) {
		const line = goTrimSpace(lines[index] ?? '');
		lines[index] = line.startsWith('*') ? line.slice(1) : line;
	}
	return lines.join(' ');
}

// directives.go: splitDirective's results: what follows an honored directive word.
interface DirectiveRest {
	readonly rest: string;
}

// directives.go: splitDirective finds an honored directive word at the start of the comment body. A
// directive has to be the first thing in its comment to count, or every sentence that mentions one
// becomes a suppression.
function splitDirective(body: string): DirectiveRest | undefined {
	for (const candidate of disableDirectives) {
		if (body.startsWith(candidate)) {
			return { rest: body.slice(candidate.length) };
		}
	}
	return undefined;
}

// directives.go: endsWord, whether the scope word just read is a whole word: the comment ends, or a
// space or tab follows.
function endsWord(rest: string): boolean {
	return rest === '' || rest.startsWith(' ') || rest.startsWith('\t');
}

// directives.go: splitScope's results, the scope and what follows it, when the scope word is whole.
interface ScopeRead {
	readonly scope: Scope;
	readonly remainder: string;
}

// directives.go: splitScope reads the scope suffix and returns what follows it. `-next-line` must be
// tested before the bare form, or every next-line directive reads as a file-level disable; the cases
// are written longest first so the ordering is visible.
function splitScope(rest: string): ScopeRead | undefined {
	if (rest.startsWith('-next-line')) {
		const remainder = rest.slice('-next-line'.length);
		return endsWord(remainder) ? { scope: 'next-line', remainder } : undefined;
	}
	if (rest.startsWith('-line')) {
		const remainder = rest.slice('-line'.length);
		return endsWord(remainder) ? { scope: 'same-line', remainder } : undefined;
	}
	if (endsWord(rest)) {
		// The bare, file-level form.
		return { scope: 'file', remainder: rest };
	}
	// The directive is a prefix of a longer word, `eslint-disable-nonsense` or `eslint-disabled`.
	return undefined;
}

// directives.go: splitReason's results, the rule list and the reason after the first `--`.
interface RuleList {
	readonly rules: string;
	readonly reason: string;
}

// directives.go: splitReason separates the rule list from the ` -- reason` that may follow it. A rule
// name cannot contain `--`, so the first occurrence wins.
function splitReason(rest: string): RuleList {
	const index = rest.indexOf('--');
	if (index >= 0) {
		return { rules: goTrimSpace(rest.slice(0, index)), reason: goTrimSpace(rest.slice(index + 2)) };
	}
	return { rules: goTrimSpace(rest), reason: '' };
}

// directives.go: parseRuleNames splits a comma-separated rule list, dropping empties. The Go returns
// nil for an empty list; the port, an empty one.
function parseRuleNames(list: string): readonly string[] {
	const names: string[] = [];
	if (list === '') {
		return names;
	}
	for (const part of list.split(',')) {
		const trimmed = goTrimSpace(part);
		if (trimmed !== '') {
			names.push(trimmed);
		}
	}
	return names;
}

// directives.go: ParseDisable reads one comment's full source text, delimiters included, as a disable
// directive:
//
//	<tool>-disable[-next-line|-line] [rule[, rule...]] [-- reason]
//
// Everything after the directive word is optional. No rules means blanket. No reason means the author
// did not say why, which is recorded rather than rejected.
export function parseDisable(commentText: string): Disable | undefined {
	const body = goTrimSpace(stripCommentMarkers(commentText));

	const directive = splitDirective(body);
	if (directive === undefined) {
		return undefined;
	}

	const scopeRead = splitScope(directive.rest);
	if (scopeRead === undefined) {
		return undefined;
	}
	if (scopeRead.scope === 'file' && isLineComment(commentText)) {
		return undefined;
	}

	const split = splitReason(scopeRead.remainder);

	return { scope: scopeRead.scope, rules: parseRuleNames(split.rules), reason: split.reason };
}

// directives.go: ParseEnable's results: the rule names an enable closes. No names closes every open
// block.
export interface Enable {
	readonly rules: readonly string[];
}

// directives.go: ParseEnable reads one comment's full source text as an enable directive, or gives
// undefined for a comment that isn't one.
export function parseEnable(commentText: string): Enable | undefined {
	if (isLineComment(commentText)) {
		return undefined;
	}
	const body = goTrimSpace(stripCommentMarkers(commentText));

	for (const directive of enableDirectives) {
		if (!body.startsWith(directive)) {
			continue;
		}
		const rest = body.slice(directive.length);
		// Same anchoring rule as a disable: the directive has to be the whole word.
		if (rest !== '' && !rest.startsWith(' ') && !rest.startsWith('\t')) {
			continue;
		}
		return { rules: parseRuleNames(splitReason(rest).rules) };
	}
	return undefined;
}

// directives.go: Recognize reports which directive a comment is, as its scope word, or '' when it is
// none the suppression index acts on.
export function recognize(commentText: string): string {
	const parsed = parseDisable(commentText);
	if (parsed !== undefined) {
		switch (parsed.scope) {
			case 'next-line':
				return ScopeWordDisableNextLine;
			case 'same-line':
				return ScopeWordDisableLine;
			case 'file':
				return ScopeWordDisable;
		}
	}
	if (parseEnable(commentText) !== undefined) {
		return ScopeWordEnable;
	}
	return '';
}

// subject.go: subjectRules, the rules whose findings are about a directive comment itself. The Go
// writes it only from init, before the first file is walked; the driver registers them before it
// builds an index, the same way.
const subjectRules = new Map<string, boolean>();

// subject.go: RegisterSubject declares that a rule reports on suppression comments themselves, so the
// suppression index must not let a directive silence the finding about itself.
export function registerSubject(ruleName: string): void {
	subjectRules.set(ruleName, true);
}

// subject.go: IsSubject is whether a rule registered itself through RegisterSubject.
export function isSubject(ruleName: string): boolean {
	return subjectRules.get(ruleName) === true;
}
