// A port of cohere's internal/lint/suppression (suppression.go, parse.go and directive_subject.go) to
// Adamic 0.1: the comments that tell a rule "not here", and what they did. Each piece names the Go it
// reads as.
//
// Two properties matter more than the parsing, and both are about not lying. A suppression is never
// silent: every finding withheld is counted, so a run that suppressed forty findings cannot print the
// same line as a run that found none. And a suppression that never fires is visible too: an unused one
// is a rule silently scoped off a file that no longer needs it.
//
// The index works on source text and offsets rather than on a parsed file. A suppression is a fact
// about a line, not about a node.
//
// Where the port differs from the Go, and why:
//
//   - Offsets are UTF-16 indexes where the Go's are byte offsets (scan.ts says why both decide alike).
//   - Go's Kind is an int enum with a String method; here it is the union of those strings.
//   - The Go's methods check for a nil *Index, which a caller with no index passes. The port's Index is
//     always one Build made, so it has no such check; a caller with no index has no Index to ask.

import { isSubject, parseDisable, parseEnable } from './directives.ts';
import { buildLineIndex, scanComments, scanEnables } from './scan.ts';

// suppression.go: Kind, the scope a directive covers. 'next-line' silences the line after the comment
// (372 of 387 directives in cohere's corpus, so the form to get exactly right), 'same-line' the line
// the comment sits on, and 'file' from the comment's own line to a matching enable, or to the end of
// the file.
export type Kind = 'next-line' | 'same-line' | 'file';

// suppression.go: matchesRuleName resolves a comment's rule name against a registry rule name: exact
// first, then plugin-qualified. The suffix has to fall on a `/` boundary, or `no-enum` would match
// `consistency-no-enum` and a directive would silence rules its author never named.
function matchesRuleName(named: string, ruleName: string): boolean {
	if (named === ruleName) {
		return true;
	}
	// strings.TrimSuffix(named, ruleName), and whether it trimmed anything.
	if (named.endsWith(ruleName) && ruleName !== '') {
		const prefix = named.slice(0, named.length - ruleName.length);
		return prefix.endsWith('/');
	}
	return false;
}

// suppression.go: Directive, one suppression comment, resolved to the lines and rules it covers.
export class Directive {
	readonly kind: Kind;

	// The rule names the comment named. Empty means it named none, which silences everything in its
	// scope: a blanket.
	readonly rules: readonly string[];

	// The text after ` -- `, trimmed. '' means the author did not say why.
	readonly reason: string;

	// The zero-based line the comment starts on.
	line = 0;

	// The inclusive zero-based line span this directive covers.
	startLine = 0;
	endLine = 0;

	// The comment's own range, so an unused directive can be reported where it sits.
	pos = 0;
	end = 0;

	// The findings this directive actually withheld.
	applied = 0;

	constructor(kind: Kind, rules: readonly string[], reason: string) {
		this.kind = kind;
		this.rules = rules;
		this.reason = reason;
	}

	// suppression.go: HasReason is whether the author said why.
	hasReason(): boolean {
		return this.reason !== '';
	}

	// suppression.go: Covers is whether this directive applies to ruleName on a zero-based line. One
	// naming no rules covers every rule; one naming rules covers only those, matched by plugin-qualified
	// suffix.
	covers(ruleName: string, line: number): boolean {
		if (line < this.startLine || line > this.endLine) {
			return false;
		}
		if (this.rules.length === 0) {
			return true;
		}
		for (const named of this.rules) {
			if (matchesRuleName(named, ruleName)) {
				return true;
			}
		}
		return false;
	}
}

// directive_subject.go: coversAFindingAboutADirective is ESLint's coverage of a column -1 finding, in
// line terms, for rules that report on a directive comment itself: only a file-scope directive opened
// on an earlier line covers it, through its closing enable's own line inclusive.
function coversAFindingAboutADirective(candidate: Directive, line: number): boolean {
	return candidate.kind === 'file' && candidate.line < line;
}

// parse.go: parseDirective reads one comment's text and returns the directive it declares, or
// undefined.
function parseDirective(commentText: string): Directive | undefined {
	const parsed = parseDisable(commentText);
	if (parsed === undefined) {
		return undefined;
	}
	return new Directive(parsed.scope, parsed.rules, parsed.reason);
}

// suppression.go: RuleReference, one rule name a disable or enable comment names, with the comment's
// range. Enables are included because ESLint checks them too.
export interface RuleReference {
	readonly name: string;
	readonly pos: number;
	readonly end: number;
}

// suppression.go: namesIntersect. A block naming no rules is blanket, so any enable closes it.
function namesIntersect(enableRules: readonly string[], blockRules: readonly string[]): boolean {
	if (blockRules.length === 0) {
		return true;
	}
	for (const left of enableRules) {
		for (const right of blockRules) {
			if (left === right) {
				return true;
			}
		}
	}
	return false;
}

// suppression.go: resolveFileScopeEnds closes each file-scope directive at the enable comment that
// matches it, by rule name, because blocks overlap in real code. An enable naming no rules closes every
// open block.
function resolveFileScopeEnds(found: readonly Directive[], sourceText: string, lineOf: (offset: number) => number): void {
	let hasFileScope = false;
	for (const candidate of found) {
		if (candidate.kind === 'file') {
			hasFileScope = true;
			break;
		}
	}
	if (!hasFileScope) {
		return;
	}

	const enables = scanEnables(sourceText, lineOf);
	for (const block of found) {
		if (block.kind !== 'file') {
			continue;
		}
		for (const enable of enables) {
			if (enable.line <= block.line) {
				continue;
			}
			if (enable.rules.length === 0 || namesIntersect(enable.rules, block.rules)) {
				block.endLine = enable.line;
				break;
			}
		}
	}
}

// suppression.go: Index, every directive in one file, ready to answer whether a finding is suppressed.
export class Index {
	readonly #directives: readonly Directive[];
	readonly #ruleReferences: readonly RuleReference[];
	readonly #lineOf: (offset: number) => number;

	constructor(directives: readonly Directive[], ruleReferences: readonly RuleReference[], lineOf: (offset: number) => number) {
		this.#directives = directives;
		this.#ruleReferences = ruleReferences;
		this.#lineOf = lineOf;
	}

	// suppression.go: RuleReferences, every rule name a disable or enable comment named, in source
	// order.
	ruleReferences(): readonly RuleReference[] {
		return this.#ruleReferences;
	}

	// suppression.go: Suppresses is whether a finding for ruleName at an offset is silenced, and records
	// that the directive responsible was used. Recording on the query is what makes an unused directive
	// detectable at all.
	suppresses(ruleName: string, offset: number): boolean {
		if (this.#directives.length === 0) {
			return false;
		}

		const line = this.#lineOf(offset);
		const reportsOnDirectives = isSubject(ruleName);
		for (const candidate of this.#directives) {
			if (reportsOnDirectives && !coversAFindingAboutADirective(candidate, line)) {
				continue;
			}
			if (candidate.covers(ruleName, line)) {
				candidate.applied++;
				return true;
			}
		}
		return false;
	}

	// suppression.go: Directives, every directive found, in source order.
	directives(): readonly Directive[] {
		return this.#directives;
	}

	// suppression.go: AppliedCount, how many findings the directive at an index withheld.
	appliedCount(index: number): number {
		const candidate = this.#directives[index];
		return candidate === undefined ? 0 : candidate.applied;
	}

	// suppression.go: TotalApplied, how many findings this file's directives withheld.
	totalApplied(): number {
		let total = 0;
		for (const candidate of this.#directives) {
			total += candidate.applied;
		}
		return total;
	}

	// suppression.go: Unused, the directives that never withheld anything.
	unused(): readonly Directive[] {
		return this.#directives.filter((candidate) => candidate.applied === 0);
	}

	// suppression.go: WithoutReason, the directives that never said why.
	withoutReason(): readonly Directive[] {
		return this.#directives.filter((candidate) => !candidate.hasReason());
	}

	// suppression.go: LineOf, the zero-based line an offset falls on.
	lineOf(offset: number): number {
		return this.#lineOf(offset);
	}
}

// suppression.go: Build scans source text once and resolves every directive in it.
export function build(sourceText: string): Index {
	const lineOf = buildLineIndex(sourceText);
	const found: Directive[] = [];
	const references: RuleReference[] = [];

	for (const comment of scanComments(sourceText)) {
		const commentText = sourceText.slice(comment.pos, comment.end);
		const parsed = parseDirective(commentText);
		if (parsed === undefined) {
			const enable = parseEnable(commentText);
			if (enable !== undefined) {
				for (const name of enable.rules) {
					references.push({ name, pos: comment.pos, end: comment.end });
				}
			}
			continue;
		}
		for (const name of parsed.rules) {
			references.push({ name, pos: comment.pos, end: comment.end });
		}

		parsed.pos = comment.pos;
		parsed.end = comment.end;
		parsed.line = lineOf(comment.pos);

		switch (parsed.kind) {
			case 'next-line':
				parsed.startLine = parsed.line + 1;
				parsed.endLine = parsed.line + 1;
				break;
			case 'same-line':
				parsed.startLine = parsed.line;
				parsed.endLine = parsed.line;
				break;
			case 'file':
				// Runs from its own line to a matching enable, or to the end of the file. It must not
				// reach backward.
				parsed.startLine = parsed.line;
				parsed.endLine = lineOf(sourceText.length);
				break;
		}

		found.push(parsed);
	}

	resolveFileScopeEnds(found, sourceText, lineOf);

	return new Index(found, references, lineOf);
}
