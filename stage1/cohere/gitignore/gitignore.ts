// A port of cohere's internal/gitignore/gitignore.go to Adamic 0.1. The Go is beside it in the cohere
// submodule (cohere/internal/gitignore/gitignore.go); each piece here names the Go it reads as.
//
// It decides which paths a repository's ignore files exclude, the way git decides it, without running
// git.
//
// It reads what gitignore(5) names for a working tree: every `.gitignore` from the repository root down,
// each scoped to its own directory, and `.git/info/exclude` when `.git` is a directory. The user's global
// core.excludesFile is not read: it is configuration on one machine, not the repository's, so a check that
// read it would pass on one machine and fail on another. Matching is case sensitive, as git's is unless
// core.ignorecase is set, which is likewise a property of a clone rather than of the repository.
//
// The semantics are gitignore(5)'s:
//
//   - Precedence, highest first: a `.gitignore` in the path's own directory, then each parent's up to the
//     root, then info/exclude. Within one file the last matching line decides, and a `!` line re-includes.
//   - A pattern with a `/` at its start or middle is anchored to its file's directory. One without matches
//     a base name at any depth below it. A trailing `/` matches directories only.
//   - Nothing below an excluded directory can be re-included: git never looks inside one, so the rule
//     that excluded the directory decides every path under it.
//   - Trailing spaces are dropped unless escaped, `#` starts a comment, and a backslash escapes `#`, `!`
//     and any other byte. A file's byte order mark and each line's carriage return are not part of a
//     pattern.
//
// What git would read differently from its documentation is refused by name rather than followed: a
// `.gitignore` that is a symbolic link (git does not follow one inside a working tree), one larger than
// 100 MiB (git skips it with a warning), and a line holding a NUL byte. Each is an error naming the file,
// because skipping one would leave its patterns unapplied without a word.
//
// The matcher only answers. Walking, and pruning what it excludes, belong to the caller: it enters each
// directory it descends into with enter, and asks ignored about each entry it finds there.
//
// Where the port differs from the Go, and why:
//
//   - 0.1 has no input (docs/0.1.md, decision 7), so the Go's reads through os are reads of a
//     WorkingTree, which holds the entries the matcher looks at: ignore files, `.git` entries and
//     symbolic links, keyed by their slash-separated path below the root. When Adamic can read files,
//     a WorkingTree read from disk takes its place and nothing else changes.
//   - 0.1 has no exceptions, so the Go's (value, error) is a Result. Its other two results are tuples:
//     (bool, Source) is a Verdict, and strings.CutPrefix's (after, found) is [after, found].
//   - The Go's unexported fields are #private fields, and a Matcher's are written only while enter
//     builds a new one, as the Go writes only its copy.

import { panic, utf8Length } from 'adamic';
import { compileGlob, Glob } from './glob.ts';
import { base, clean, dir } from './path.ts';

// gitignore.go: IgnoreFileName, the per-directory ignore file.
export const IgnoreFileName = '.gitignore';

// gitignore.go: ExcludeFile, the repository's own exclude file, relative to its root.
export const ExcludeFile = '.git/info/exclude';

// gitignore.go: maximumFileSize, the largest ignore file read. Git skips one past 100 MiB with a warning;
// here it is refused, since skipping it would leave every pattern in it unapplied without a word.
const maximumFileSize = 100 << 20;

// gitignore.go: ErrNestedRepository's text. An error carrying it is a refusal to enter a directory that
// is a repository of its own: its ignore files govern it, through its own newMatcher, and the parent's
// patterns do not.
const nestedRepositoryText = 'is a repository of its own';

// An error, as the Go's error values are: its text, and whether it is ErrNestedRepository, which the Go
// tells by errors.Is.
export interface GitignoreError {
	readonly message: string;
	readonly nestedRepository: boolean;
}

// A value or the error that stopped it: the Go's (value, error).
export type Result<Value> = { readonly kind: 'Ok'; readonly value: Value } | { readonly kind: 'Error'; readonly error: GitignoreError };

function failure<Value>(message: string): Result<Value> {
	return { kind: 'Error', error: { message, nestedRepository: false } };
}

// An entry of a WorkingTree: what os.Lstat would report, and a file's contents, which os.ReadFile would.
export type Entry =
	| { readonly kind: 'File'; readonly contents: string }
	| { readonly kind: 'Directory' }
	| { readonly kind: 'SymbolicLink'; readonly target: string }
	| { readonly kind: 'Other' };

// The working tree the matcher reads, in place of the Go's file system: root is the repository root as
// the Go's filepath.Abs would give it, which only error messages show, and entries holds what is there,
// keyed by slash-separated paths relative to the root.
export interface WorkingTree {
	readonly root: string;
	readonly entries: ReadonlyMap<string, Entry>;
}

// os.Lstat: the entry itself, a symbolic link included, or undefined when nothing is there.
function lstat(tree: WorkingTree, relative: string): Entry | undefined {
	return tree.entries.get(relative);
}

// os.Stat: the entry a path names once its symbolic links are followed, or undefined when nothing is
// there. A link is followed within the tree, from the directory that holds it; one that leads out of the
// tree, or round more links than the kernel follows, names nothing that can be read here.
function stat(tree: WorkingTree, relative: string): Entry | undefined {
	let current = relative;
	for (let followed = 0; followed < 40; followed++) {
		const entry = lstat(tree, current);
		if (entry === undefined || entry.kind !== 'SymbolicLink') {
			return entry;
		}
		if (entry.target.startsWith('/')) {
			return undefined;
		}
		const parent = dir(current);
		const resolved = clean(parent === '.' ? entry.target : parent + '/' + entry.target);
		if (resolved === '..' || resolved.startsWith('../')) {
			return undefined;
		}
		current = resolved;
	}
	return undefined;
}

// filepath.Join(root, relative), for a message.
function absolute(tree: WorkingTree, relative: string): string {
	return relative === '' ? tree.root : tree.root + '/' + relative;
}

// gitignore.go: Source. It is the rule that decided a path: the ignore file, relative to the repository
// root, the line in it, and the pattern as git prints it, `!` and a trailing `/` included. The zero Source
// means no rule matched.
export class Source {
	readonly file: string;
	readonly line: number;
	readonly pattern: string;

	constructor(file: string, line: number, pattern: string) {
		this.file = file;
		this.line = line;
		this.pattern = pattern;
	}

	// gitignore.go: Source.IsZero. It reports whether no rule decided.
	isZero(): boolean {
		return this.file === '';
	}

	// gitignore.go: Source.String. It is git check-ignore's verbose form, `file:line:pattern`.
	toString(): string {
		return `${this.file}:${this.line}:${this.pattern}`;
	}
}

// A path's answer: whether it is ignored, and the rule that decided, which the Go returns as
// (bool, Source). cohere's differential test names the pair verdict.
export type Verdict = readonly [boolean, Source];

// gitignore.go: Source{}.
function zeroSource(): Source {
	return new Source('', 0, '');
}

// gitignore.go: rule, one compiled pattern line.
//
// anchored is a pattern matched against the path below base rather than against a base name. base is
// the directory of the file the rule came from, relative to the root, "" at the root.
class Rule {
	readonly source: Source;
	readonly negated: boolean;
	readonly directoryOnly: boolean;
	readonly anchored: boolean;
	readonly base: string;
	readonly glob: Glob;

	constructor(source: Source, negated: boolean, directoryOnly: boolean, anchored: boolean, base: string, glob: Glob) {
		this.source = source;
		this.negated = negated;
		this.directoryOnly = directoryOnly;
		this.anchored = anchored;
		this.base = base;
		this.glob = glob;
	}

	// gitignore.go: (*rule).matches.
	matches(relativePath: string, baseName: string): boolean {
		if (!this.anchored) {
			return this.glob.matches(baseName, false);
		}
		let below = relativePath;
		if (this.base !== '') {
			const [after, ok] = cutPrefix(relativePath, this.base + '/');
			if (!ok) {
				return false;
			}
			below = after;
		}
		return this.glob.matches(below, true);
	}
}

// gitignore.go: Matcher. It is one directory's view of a repository's ignore rules: those of every
// ignore file from the root down to that directory. It is never modified once made, so a walk may keep
// one per directory and share them freely.
//
// files are the rule lists of each `.gitignore` from the root down, root first. exclude is
// info/exclude's. excludedBy is the rule that excluded this directory or one above it. It decides every
// path below.
export class Matcher {
	readonly #tree: WorkingTree;
	#directory: string;
	#files: (readonly Rule[])[];
	#exclude: readonly Rule[];
	#excludedBy: Rule | undefined;

	constructor(tree: WorkingTree, directory: string, files: (readonly Rule[])[], exclude: readonly Rule[], excludedBy: Rule | undefined) {
		this.#tree = tree;
		this.#directory = directory;
		this.#files = files;
		this.#exclude = exclude;
		this.#excludedBy = excludedBy;
	}

	// gitignore.go: (*Matcher).Root, the repository root the matcher reads.
	root(): string {
		return this.#tree.root;
	}

	// gitignore.go: (*Matcher).Directory, the directory this matcher answers for, relative to the root, ""
	// for the root itself.
	directory(): string {
		return this.#directory;
	}

	// gitignore.go: (*Matcher).Excluded. It reports whether this directory, or one above it, is excluded,
	// and by which rule. A walk that entered one anyway gets the same answer for everything inside.
	excluded(): Verdict {
		if (this.#excludedBy === undefined) {
			return [false, zeroSource()];
		}
		return [true, this.#excludedBy.source];
	}

	// gitignore.go: (*Matcher).Enter. It returns the matcher for relativeDirectory, which is this
	// matcher's directory or one below it, slash separated and relative to the root.
	//
	// It reads each `.gitignore` between the two once, and stops reading at a directory that is itself
	// excluded, as git does: the rule that excluded it then decides everything below. A directory holding
	// a `.git` is refused as a nested repository, and an ignore file that cannot be read faithfully (a
	// symbolic link, which git does not follow in a tree; one too large) is an error naming it, never
	// skipped.
	enter(relativeDirectoryAsGiven: string): Result<Matcher> {
		const relativeDirectory = cleanRelative(relativeDirectoryAsGiven);
		if (relativeDirectory === this.#directory) {
			return { kind: 'Ok', value: this };
		}
		let [remainder, below] = cutPrefix(relativeDirectory, this.#directory);
		if (this.#directory !== '') {
			[remainder, below] = cutPrefix(remainder, '/');
		}
		if (!below || remainder === '') {
			return failure(`gitignore: ${relativeDirectory} is not below ${displayDirectory(this.#directory)}`);
		}

		// The Go copies the matcher and cuts its files' capacity, so appending copies them: the copy here
		// is a new Matcher over a copy of the list.
		const entered = new Matcher(this.#tree, this.#directory, this.#files.slice(), this.#exclude, this.#excludedBy);
		for (const name of remainder.split('/')) {
			const child = joinRelative(entered.#directory, name);
			if (lstat(this.#tree, joinRelative(child, '.git')) !== undefined) {
				return {
					kind: 'Error',
					error: { message: `gitignore: ${absolute(this.#tree, child)} ${nestedRepositoryText}`, nestedRepository: true },
				};
			}
			entered.#directory = child;
			if (entered.#excludedBy !== undefined) {
				continue;
			}
			const decided = entered.decide(child, base(child), true, entered.#files.length);
			if (decided !== undefined && !decided.negated) {
				entered.#excludedBy = decided;
				continue;
			}
			const rules = readTreeFile(this.#tree, child);
			if (rules.kind === 'Error') {
				return rules;
			}
			entered.#files.push(rules.value);
		}
		return { kind: 'Ok', value: entered };
	}

	// gitignore.go: (*Matcher).Ignored. It reports whether relativePath, an entry of this matcher's
	// directory given relative to the root, is excluded, and the rule that decided. A rule that
	// re-included the path is reported with false.
	//
	// isDirectory says whether the path is a directory, which is what a pattern with a trailing `/` asks;
	// a symbolic link to a directory is not one, as git does not follow it. Asking about a path outside
	// this directory is a programming error and panics, because the answer would silently leave out the
	// ignore files between the two.
	ignored(relativePathAsGiven: string, isDirectory: boolean): Verdict {
		const relativePath = cleanRelative(relativePathAsGiven);
		let parent = dir(relativePath);
		if (parent === '.') {
			parent = '';
		}
		if (parent !== this.#directory) {
			panic(
				`gitignore: asked about ${relativePath} through the matcher for ${displayDirectory(this.#directory)}; Enter its directory first`,
			);
		}
		if (this.#excludedBy !== undefined) {
			return [true, this.#excludedBy.source];
		}
		const decided = this.decide(relativePath, base(relativePath), isDirectory, this.#files.length);
		if (decided === undefined) {
			return [false, zeroSource()];
		}
		return [!decided.negated, decided.source];
	}

	// gitignore.go: (*Matcher).IgnoredPath. It is ignored for a path anywhere below this matcher's
	// directory: it enters the directories in between first, reading their ignore files, so the answer
	// names a nested file's line when one decides. It is for a question about one path, such as
	// explaining why a file was left out; a walk enters each directory itself and asks ignored, which
	// reads every ignore file once.
	ignoredPath(relativePathAsGiven: string, isDirectory: boolean): Result<Verdict> {
		const relativePath = cleanRelative(relativePathAsGiven);
		let parent = dir(relativePath);
		if (parent === '.') {
			parent = '';
		}
		const scope = this.enter(parent);
		if (scope.kind === 'Error') {
			return scope;
		}
		return { kind: 'Ok', value: scope.value.ignored(relativePath, isDirectory) };
	}

	// gitignore.go: (*Matcher).decide. It is the rule that decides relativePath among the first fileCount
	// tree files and info/exclude, or undefined: the deepest file first, the last line of each first,
	// info/exclude last.
	decide(relativePath: string, baseName: string, isDirectory: boolean, fileCount: number): Rule | undefined {
		for (let fileIndex = fileCount - 1; fileIndex >= 0; fileIndex--) {
			const rules = this.#files[fileIndex] ?? panic('gitignore: a file index past the files read');
			const decided = lastMatching(rules, relativePath, baseName, isDirectory);
			if (decided !== undefined) {
				return decided;
			}
		}
		return lastMatching(this.#exclude, relativePath, baseName, isDirectory);
	}

}

// gitignore.go: New. It reads a repository's root `.gitignore` and, when `.git` is a directory, its
// info/exclude.
export function newMatcher(tree: WorkingTree): Result<Matcher> {
	const files: (readonly Rule[])[] = [];
	const rules = readTreeFile(tree, '');
	if (rules.kind === 'Error') {
		return rules;
	}
	files.push(rules.value);

	let exclude: readonly Rule[] = [];
	const git = stat(tree, '.git');
	if (git !== undefined && git.kind === 'Directory') {
		// info/exclude lives outside the tree, so a symbolic link to it is followed, as git follows it.
		const contents = readIgnoreFile(tree, ExcludeFile, true);
		if (contents.kind === 'Error') {
			return contents;
		}
		const parsed = parseRules(contents.value ?? '', ExcludeFile, '');
		if (parsed.kind === 'Error') {
			return parsed;
		}
		exclude = parsed.value;
	}
	return { kind: 'Ok', value: new Matcher(tree, '', files, exclude, undefined) };
}

// gitignore.go: Patterns. It is a list of ignore-file lines that is not a file in the tree, read with the
// same syntax and relative to the root it is applied at: a project setting that says what to skip in the
// words a `.gitignore` would use. Within the list the last matching line decides, as within one file.
export class Patterns {
	readonly #rules: readonly Rule[];

	constructor(rules: readonly Rule[]) {
		this.#rules = rules;
	}

	// gitignore.go: (*Patterns).Ignored. It reports whether the list excludes relativePath, given relative
	// to the root it applies at, and the line that decided. Unlike a Matcher it knows nothing of
	// directories above the path: a walk applying it prunes each directory it excludes, so a path below
	// one is never asked about.
	ignored(relativePathAsGiven: string, isDirectory: boolean): Verdict {
		const relativePath = cleanRelative(relativePathAsGiven);
		const decided = lastMatching(this.#rules, relativePath, base(relativePath), isDirectory);
		if (decided === undefined) {
			return [false, zeroSource()];
		}
		return [!decided.negated, decided.source];
	}
}

// gitignore.go: CompilePatterns. It reads lines as the lines of an ignore file named name, which is what
// each Source reports. A line holding a newline or a NUL byte cannot be an ignore-file line and is
// refused.
export function compilePatterns(lines: readonly string[], name: string): Result<Patterns> {
	const rules: Rule[] = [];
	for (let index = 0; index < lines.length; index++) {
		let line = lines[index] ?? panic('gitignore: a line index past the lines');
		if (line.includes('\n') || line.includes('\u0000')) {
			return failure(
				`gitignore: ${name} entry ${index + 1}, ${quote(line)}, holds a newline or a NUL byte, which no ignore-file line can`,
			);
		}
		line = trimSuffix(line, '\r');
		if (line === '' || line.startsWith('#')) {
			continue;
		}
		rules.push(compileRule(trimTrailingSpaces(line), name, index + 1, ''));
	}
	return { kind: 'Ok', value: new Patterns(rules) };
}

// gitignore.go: lastMatching.
function lastMatching(rules: readonly Rule[], relativePath: string, baseName: string, isDirectory: boolean): Rule | undefined {
	for (let index = rules.length - 1; index >= 0; index--) {
		const candidate = rules[index] ?? panic('gitignore: a rule index past the rules');
		if (candidate.directoryOnly && !isDirectory) {
			continue;
		}
		if (candidate.matches(relativePath, baseName)) {
			return candidate;
		}
	}
	return undefined;
}

// gitignore.go: readTreeFile. It reads the `.gitignore` of directory, relative to the root, returning no
// rules when it has none.
function readTreeFile(tree: WorkingTree, directory: string): Result<readonly Rule[]> {
	const relativeFile = joinRelative(directory, IgnoreFileName);
	const contents = readIgnoreFile(tree, relativeFile, false);
	if (contents.kind === 'Error') {
		return contents;
	}
	if (contents.value === undefined) {
		return { kind: 'Ok', value: [] };
	}
	return parseRules(contents.value, relativeFile, directory);
}

// gitignore.go: readIgnoreFile. It returns a file's contents, or undefined when it does not exist. follow
// says whether a symbolic link is read through: git follows one for info/exclude and not for a
// `.gitignore` in the tree.
function readIgnoreFile(tree: WorkingTree, relativeFile: string, follow: boolean): Result<string | undefined> {
	const filePath = absolute(tree, relativeFile);
	let info = lstat(tree, relativeFile);
	if (info === undefined) {
		return { kind: 'Ok', value: undefined };
	}
	if (info.kind === 'SymbolicLink') {
		if (!follow) {
			return failure(
				`gitignore: ${filePath} is a symbolic link, which git does not follow inside a working tree, so its patterns would not apply; replace it with the file it points to`,
			);
		}
		info = stat(tree, relativeFile);
		if (info === undefined) {
			return { kind: 'Ok', value: undefined };
		}
	}
	if (info.kind !== 'File') {
		return failure(`gitignore: ${filePath} is not a regular file, so it cannot be read as an ignore file`);
	}
	if (utf8Length(info.contents) > maximumFileSize) {
		return failure(
			`gitignore: ${filePath} is larger than 100 MiB, which git skips with a warning, leaving its patterns unapplied; cohere refuses it rather than skip it`,
		);
	}
	return { kind: 'Ok', value: info.contents };
}

// gitignore.go: byteOrderMark, which opens a file written by an editor that marks its encoding. The Go
// sees its three bytes; a string sees one character.
const byteOrderMark = '\uFEFF';

// gitignore.go: parseRules. It compiles an ignore file's lines. file names it in every Source, and base
// is the directory its anchored patterns are relative to.
function parseRules(contentsAsRead: string, file: string, base: string): Result<readonly Rule[]> {
	const contents = contentsAsRead.startsWith(byteOrderMark) ? contentsAsRead.slice(byteOrderMark.length) : contentsAsRead;
	const rules: Rule[] = [];
	const lines = contents.split('\n');
	for (let index = 0; index < lines.length; index++) {
		let line = trimSuffix(lines[index] ?? panic('gitignore: a line index past the lines'), '\r');
		if (line === '' || line.startsWith('#')) {
			continue;
		}
		if (line.includes('\u0000')) {
			return failure(`gitignore: ${file}:${index + 1} holds a NUL byte, which no pattern can contain`);
		}
		line = trimTrailingSpaces(line);
		rules.push(compileRule(line, file, index + 1, base));
	}
	return { kind: 'Ok', value: rules };
}

// gitignore.go: trimTrailingSpaces. It drops the spaces ending a line, except those a backslash escapes.
// Only spaces: a trailing tab is part of the pattern.
function trimTrailingSpaces(line: string): string {
	// spacesStart is where the current run of unescaped spaces began, or -1 outside one.
	let spacesStart = -1;
	for (let index = 0; index < line.length; index++) {
		switch (line[index]) {
			case ' ':
				if (spacesStart < 0) {
					spacesStart = index;
				}
				break;
			case '\\':
				// The escaped byte is kept whatever it is, a space included.
				index++;
				spacesStart = -1;
				break;
			default:
				spacesStart = -1;
				break;
		}
	}
	if (spacesStart >= 0) {
		return line.slice(0, spacesStart);
	}
	return line;
}

// gitignore.go: compileRule. It reads one pattern line: an optional `!`, the pattern, and an optional
// trailing `/`.
function compileRule(line: string, file: string, lineNumber: number, base: string): Rule {
	let pattern = line;
	let negated = false;
	if (pattern.startsWith('!')) {
		negated = true;
		pattern = pattern.slice(1);
	}
	let directoryOnly = false;
	if (pattern.endsWith('/')) {
		directoryOnly = true;
		pattern = pattern.slice(0, pattern.length - 1);
	}

	let sourcePattern = pattern;
	if (negated) {
		sourcePattern = '!' + sourcePattern;
	}
	if (directoryOnly) {
		sourcePattern += '/';
	}

	const anchored = pattern.includes('/');
	const glob = anchored ? compileGlob(trimPrefix(pattern, '/'), true) : compileGlob(pattern, false);
	return new Rule(new Source(file, lineNumber, sourcePattern), negated, directoryOnly, anchored, base, glob);
}

// gitignore.go: cleanRelative. It normalizes a relative path to the slash form the matcher keys on, ""
// for the root. A walk passes paths already in that form, which are returned as they are.
function cleanRelative(relative: string): string {
	if (isCleanRelative(relative)) {
		return relative;
	}
	const cleaned = clean(relative);
	if (cleaned === '.' || cleaned === '/') {
		return '';
	}
	return trimPrefix(cleaned, '/');
}

// gitignore.go: isCleanRelative. It reports whether a path is already what cleanRelative would make it:
// slash separated, with no empty, `.` or `..` segment and no leading or trailing slash. The Go also
// refuses a backslash on Windows, where it separates; Adamic's paths are slash separated everywhere.
function isCleanRelative(relative: string): boolean {
	if (relative === '') {
		return true;
	}
	let start = 0;
	for (let index = 0; index <= relative.length; index++) {
		if (index < relative.length && relative[index] !== '/') {
			continue;
		}
		const segment = relative.slice(start, index);
		if (segment === '' || segment === '.' || segment === '..') {
			return false;
		}
		start = index + 1;
	}
	return true;
}

// gitignore.go: joinRelative.
export function joinRelative(directory: string, name: string): string {
	if (directory === '') {
		return name;
	}
	return directory + '/' + name;
}

// gitignore.go: displayDirectory.
function displayDirectory(directory: string): string {
	if (directory === '') {
		return 'the repository root';
	}
	return directory;
}

// strings.CutPrefix: text without the prefix, and whether it had it.
function cutPrefix(text: string, prefix: string): readonly [string, boolean] {
	if (text.startsWith(prefix)) {
		return [text.slice(prefix.length), true];
	}
	return [text, false];
}

// strings.TrimPrefix.
function trimPrefix(text: string, prefix: string): string {
	return text.startsWith(prefix) ? text.slice(prefix.length) : text;
}

// strings.TrimSuffix.
function trimSuffix(text: string, suffix: string): string {
	return suffix !== '' && text.endsWith(suffix) ? text.slice(0, text.length - suffix.length) : text;
}

// hexDigit is one lowercase hexadecimal digit.
function hexDigit(value: number): string {
	return '0123456789abcdef'[value] ?? panic('gitignore: a hexadecimal digit past f');
}

// fmt's %q, for the text of one refusal: a double-quoted string with Go's escapes. Printable ASCII is
// itself, and the C escapes and \x are used for the rest of ASCII. Other characters are written as they
// are, which is what %q does for every printable one; a non-printable one beyond ASCII, which %q would
// spell as \u, is not distinguished here.
function quote(text: string): string {
	let quoted = '"';
	for (const character of text) {
		const code = character.codePointAt(0) ?? panic('gitignore: a character with no code point');
		switch (character) {
			case '"':
				quoted += '\\"';
				break;
			case '\\':
				quoted += '\\\\';
				break;
			case '\u0007':
				quoted += '\\a';
				break;
			case '\b':
				quoted += '\\b';
				break;
			case '\f':
				quoted += '\\f';
				break;
			case '\n':
				quoted += '\\n';
				break;
			case '\r':
				quoted += '\\r';
				break;
			case '\t':
				quoted += '\\t';
				break;
			case '\v':
				quoted += '\\v';
				break;
			default:
				if (code < 0x20 || code === 0x7f) {
					quoted += '\\x' + hexDigit(code >> 4) + hexDigit(code & 0xf);
				} else {
					quoted += character;
				}
				break;
		}
	}
	return quoted + '"';
}
