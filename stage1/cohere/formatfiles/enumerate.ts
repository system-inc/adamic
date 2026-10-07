// A port of cohere's internal/format/formatfiles/enumerate.go to Adamic 0.1: which files a tree offers
// the formatter, the walk, its ignore layers, and an account of every file it did not offer. Each piece
// names the Go it reads as.
//
// Where the port differs from the Go, and why:
//
//   - The format options. The Go's Enumerate starts by calling formatoptions.Resolve(root), which walks
//     up from the root reading settings files as JSON, and 0.1 refuses JSON.parse. The port takes Resolve's
//     answer as given (Resolution): the house ignore list, its declaration and lint globs. The test's Go
//     side calls Resolve on each tree and writes its answer into the cases, so the port is held to the
//     Go's walk with the Go's own resolution. The ignorePatterns layer uses the lint glob port in ../config/glob.ts,
//     relative to the settings directory, with the Go's restriction on directory pruning.
//   - Git's ignore rules are the gitignore slice's port, reading a WorkingTree from disk (disk.ts).
//   - Go's filepath.Walk and filepath.WalkDir are written out here, as Go defines them, over fileStatus
//     and readDirectory (walkTree and walkDirectories). Their callbacks' errors are the WalkStep union.
//   - Go's maps (IgnoredByLayer, DeclinedExtensions) are Maps; their printing order is the driver's.

import { panic, readDirectory } from 'adamic';
import { LintGlob } from '../config/glob.ts';
import { compilePatterns, ExcludeFile, IgnoreFileName, newMatcher, type Matcher, type Patterns } from '../gitignore/gitignore.ts';
import { base, clean, dir } from '../gitignore/path.ts';
import { DiskTree, lstat, statExists, statIsDirectory, type LinkStatus } from './disk.ts';
import { ext, goToLower, join, rel } from './golang.ts';

// formatoptions: what Resolve gives Enumerate. houseIgnore is the format block's `ignore` list, and
// houseIgnoreDeclared whether any tier declared one; source is the settings file that governs root, ''
// for none, and ignorePatterns its globs.
export interface Resolution {
	readonly houseIgnore: readonly string[];
	readonly houseIgnoreDeclared: boolean;
	readonly source: string;
	readonly ignorePatterns: readonly string[];
}

// formatoptions.ErrPrettierConfigRemains and NexusTierFileName, which the refusal of a .prettierignore
// names.
const errPrettierConfigRemains = 'Prettier config remains where cohere no longer reads it';
const nexusTierFileName = 'NexusCohereSettings.json';

// enumerate.go: the names the walk's layers are counted under. GitignoreLayer counts every git ignore
// source together.
export const GitignoreLayer = '.gitignore';
export const HouseIgnoreLayer = 'format.ignore';
export const IgnorePatternsLayer = 'ignorePatterns';

// enumerate.go: Enumeration, what a walk found, and why each file did not survive it. Every number is a
// subtraction someone can check: walked minus the ignore counts minus unhandled minus symbolicLinks is
// the number of files plus Adamic files held back.
export class Enumeration {
	readonly root: string;
	walked = 0;
	readonly ignoredByLayer = new Map<string, number>();
	unhandled = 0;
	symbolicLinks = 0;
	readonly nestedRepositories: string[] = [];
	readonly declinedExtensions = new Map<string, number>();
	readonly directories: string[] = [];
	readonly ignoreFiles: string[] = [];
	readonly files: string[] = [];
	readonly adamic: string[] = [];

	constructor(root: string) {
		this.root = root;
	}
}

// What Enumerate gives: the enumeration, or the Go's error.
export type Enumerated = { readonly kind: 'Ok'; readonly enumeration: Enumeration } | { readonly kind: 'Error'; readonly message: string };

// enumerate.go: ignoreLayer, the house lines or the lint globs after git's.
type IgnoreLayer =
	| { readonly kind: 'Lines'; readonly name: string; readonly lines: Patterns }
	| { readonly kind: 'Globs'; readonly name: string; readonly patterns: readonly string[]; readonly globs: readonly LintGlob[]; readonly from: string };

// enumerate.go: (ignoreLayer).covers, whether the layer excludes a path relative to the walk root.
function covers(layer: IgnoreLayer, root: string, relative: string, directory: boolean): boolean {
	if (layer.kind === 'Lines') { return layer.lines.ignored(relative, directory)[0]; }
	const fromSettings = rel(layer.from, join(root, relative));
	if (fromSettings === '' || fromSettings === '..' || fromSettings.startsWith('../')) { return false; }
	for (let index = 0; index < layer.globs.length; index++) {
		const pattern = layer.patterns[index] ?? panic('formatfiles: no pattern for a compiled glob');
		if (directory && pattern !== '**' && !pattern.endsWith('/**')) { continue; }
		const glob = layer.globs[index] ?? panic('formatfiles: no compiled glob');
		if (glob.matches(fromSettings)) { return true; }
	}
	return false;
}

// What a walk's callback returns: go on, skip this directory (filepath.SkipDir), or stop with an error.
type WalkStep = { readonly kind: 'Continue' } | { readonly kind: 'SkipDirectory' } | { readonly kind: 'Error'; readonly message: string };

const proceed: WalkStep = { kind: 'Continue' };
const skipDirectory: WalkStep = { kind: 'SkipDirectory' };

// A walk's callback: the path, what Lstat said of it, and whether reading it failed (the Go's err).
type Visit = (path: string, status: LinkStatus, failed: boolean) => WalkStep;

// path/filepath's walk: a directory is visited after it is read, with the error reading it, and then
// each of its entries in name order; SkipDir from a file skips the rest of its directory.
function walkTree(path: string, status: LinkStatus, visit: Visit): WalkStep {
	if (!status.directory) {
		return visit(path, status, false);
	}
	const listing = readDirectory(path);
	const step = visit(path, status, listing.kind === 'Error');
	if (listing.kind === 'Error' || step.kind !== 'Continue') {
		return step;
	}
	for (const name of listing.names) {
		const child = join(path, name);
		const childStatus = lstat(child);
		if (!childStatus.exists) {
			const failed = visit(child, childStatus, true);
			if (failed.kind === 'Error') {
				return failed;
			}
			continue;
		}
		const walked = walkTree(child, childStatus, visit);
		if (walked.kind === 'Error' || (walked.kind === 'SkipDirectory' && !childStatus.directory)) {
			return walked;
		}
	}
	return proceed;
}

// filepath.Walk.
function walk(root: string, visit: Visit): WalkStep {
	const status = lstat(root);
	const step = status.exists ? walkTree(root, status, visit) : visit(root, status, true);
	return step.kind === 'SkipDirectory' ? proceed : step;
}

// path/filepath's walkDir: a directory is visited before it is read, and again with the error if
// reading it fails; SkipDir from a file skips the rest of its directory.
function walkDirectories(path: string, status: LinkStatus, visit: Visit): WalkStep {
	const first = visit(path, status, false);
	if (first.kind !== 'Continue' || !status.directory) {
		return first.kind === 'SkipDirectory' && status.directory ? proceed : first;
	}
	const listing = readDirectory(path);
	if (listing.kind === 'Error') {
		const failed = visit(path, status, true);
		return failed.kind === 'SkipDirectory' ? proceed : failed;
	}
	for (const name of listing.names) {
		const child = join(path, name);
		const walked = walkDirectories(child, lstat(child), visit);
		if (walked.kind === 'SkipDirectory') {
			break;
		}
		if (walked.kind === 'Error') {
			return walked;
		}
	}
	return proceed;
}

// filepath.WalkDir.
function walkDir(root: string, visit: Visit): WalkStep {
	const status = lstat(root);
	const step = status.exists ? walkDirectories(root, status, visit) : visit(root, status, true);
	return step.kind === 'SkipDirectory' ? proceed : step;
}

// enumerate.go: parentOf, a slash path's directory, '' for an entry of the root, as the matchers key
// directories.
function parentOf(relative: string): string {
	const parent = dir(relative);
	return parent === '.' ? '' : parent;
}

// enumerate.go: HasOwnRepository reports whether a directory is the root of a git repository of its
// own: it holds a `.git`, a directory for a clone or a file for a submodule's gitlink.
export function hasOwnRepository(directory: string): boolean {
	return statExists(join(directory, '.git'));
}

// enumerate.go: repositoryBoundaryBetween reports whether a repository of its own begins at root or
// between root and the directory of the settings governing it.
export function repositoryBoundaryBetween(rootGiven: string, settingsDirectoryGiven: string): boolean {
	const root = clean(rootGiven);
	const settingsDirectory = clean(settingsDirectoryGiven);
	for (let directory = root; directory !== settingsDirectory; directory = dir(directory)) {
		if (hasOwnRepository(directory)) {
			return true;
		}
		if (dir(directory) === directory) {
			return false;
		}
	}
	return false;
}

// enumerate.go: NestedRepositoryContaining returns the outermost repository of its own that holds
// fileName below root, relative to root, or '' when the file belongs to root's repository or lies
// outside root.
export function nestedRepositoryContaining(rootGiven: string, fileName: string): string {
	const root = clean(rootGiven);
	let nested = '';
	for (let directory = dir(clean(fileName)); ; directory = dir(directory)) {
		const relative = rel(root, directory);
		if (relative === '' || relative === '.' || relative === '..' || relative.startsWith('../')) {
			return nested;
		}
		if (hasOwnRepository(directory)) {
			nested = relative;
		}
		if (dir(directory) === directory) {
			return nested;
		}
	}
}

// The scope a path is asked through: the matcher of its directory, entered before it.
function scopeOf(scopes: ReadonlyMap<string, Matcher>, relative: string): Matcher {
	return scopes.get(parentOf(relative)) ?? panic(`no matcher for the directory of ${relative}`);
}

// enumerate.go: NestedRepositoriesBelow finds every repository of its own below root, the outermost of
// each, relative to root. It prunes only by git's ignore rules.
export function nestedRepositoriesBelow(root: string): { readonly kind: 'Ok'; readonly nested: readonly string[] } | { readonly kind: 'Error'; readonly message: string } {
	const disk = new DiskTree(root);
	disk.load(IgnoreFileName);
	disk.load('.git');
	disk.load(ExcludeFile);
	const matcher = newMatcher(disk.tree);
	if (matcher.kind === 'Error') {
		return { kind: 'Error', message: matcher.error.message };
	}
	const scopes = new Map<string, Matcher>([['', matcher.value]]);
	const nested: string[] = [];
	const visit = (path: string, status: LinkStatus, failed: boolean): WalkStep => {
		if (failed || !status.directory) {
			return proceed;
		}
		const relative = rel(root, path);
		if (relative === '' || relative === '.') {
			return proceed;
		}
		if (base(path) === '.git') {
			return skipDirectory;
		}
		if (hasOwnRepository(path)) {
			nested.push(relative);
			return skipDirectory;
		}
		const scope = scopeOf(scopes, relative);
		if (scope.ignored(relative, true)[0]) {
			return skipDirectory;
		}
		disk.load(`${relative}/.git`);
		disk.load(`${relative}/${IgnoreFileName}`);
		const entered = scope.enter(relative);
		if (entered.kind === 'Error') {
			return { kind: 'Error', message: entered.error.message };
		}
		scopes.set(relative, entered.value);
		return proceed;
	};
	const walked = walkDir(root, visit);
	if (walked.kind === 'Error') {
		return { kind: 'Error', message: `walking ${root} for nested repositories: ${walked.message}` };
	}
	return { kind: 'Ok', nested };
}

// enumerate.go: Enumerate walks a project root and returns the files handles accepts. The layers, in
// order: git's ignore rules, read as git reads them; the house list, read with the same syntax; and the
// project's ignorePatterns. Each is counted separately, so a misconfigured layer
// shows as a suspicious zero rather than as a slightly smaller total.
export function enumerate(root: string, resolution: Resolution, handles: (fileName: string) => boolean): Enumerated {
	const enumeration = new Enumeration(root);

	const leftover = join(root, '.prettierignore');
	if (lstat(leftover).exists) {
		return {
			kind: 'Error',
			message: `${leftover}: ${errPrettierConfigRemains}; delete it, since the format block's "ignore" in the Nexus tier (${nexusTierFileName}) and the ignorePatterns of CohereSettings.json say what the walk skips`,
		};
	}

	const disk = new DiskTree(root);
	disk.load(IgnoreFileName);
	disk.load('.git');
	disk.load(ExcludeFile);
	const matcher = newMatcher(disk.tree);
	if (matcher.kind === 'Error') {
		return { kind: 'Error', message: matcher.error.message };
	}
	enumeration.ignoreFiles.push(join(root, IgnoreFileName));
	if (statIsDirectory(join(root, '.git'))) {
		enumeration.ignoreFiles.push(clean(`${root}/${ExcludeFile}`));
	}
	enumeration.ignoredByLayer.set(GitignoreLayer, 0);
	const layers: IgnoreLayer[] = [];
	if (resolution.houseIgnoreDeclared) {
		const house = compilePatterns(resolution.houseIgnore, HouseIgnoreLayer);
		if (house.kind === 'Error') {
			return { kind: 'Error', message: house.error.message };
		}
		layers.push({ kind: 'Lines', name: HouseIgnoreLayer, lines: house.value });
	}
	// The ignorePatterns layer applies only inside the repository governed by the settings.
	if (resolution.source !== '' && !repositoryBoundaryBetween(root, dir(resolution.source))) {
		layers.push({ kind: 'Globs', name: IgnorePatternsLayer, patterns: resolution.ignorePatterns,
			globs: resolution.ignorePatterns.map((pattern) => new LintGlob(pattern)), from: dir(resolution.source) });
	}
	for (const layer of layers) {
		enumeration.ignoredByLayer.set(layer.name, 0);
	}
	const count = (layer: string): void => {
		enumeration.ignoredByLayer.set(layer, (enumeration.ignoredByLayer.get(layer) ?? 0) + 1);
	};

	// The matcher for each directory entered, by its path relative to root.
	const scopes = new Map<string, Matcher>([['', matcher.value]]);

	const visit = (path: string, status: LinkStatus, failed: boolean): WalkStep => {
		if (failed) {
			return proceed;
		}
		const relative = rel(root, path);
		if (relative === '') {
			return proceed;
		}
		if (status.directory) {
			// .git is never formatted and walking it is pure cost on a large repo.
			if (base(path) === '.git') {
				return skipDirectory;
			}
			// A directory with its own .git is somebody else's tree, and the ignore layers do not
			// exclude it.
			if (relative !== '.' && hasOwnRepository(path)) {
				enumeration.nestedRepositories.push(relative);
				return skipDirectory;
			}
			// Prune an ignored directory rather than descending and filtering its contents.
			if (relative === '.') {
				enumeration.directories.push(path);
				return proceed;
			}
			const scope = scopeOf(scopes, relative);
			if (scope.ignored(relative, true)[0]) {
				count(GitignoreLayer);
				return skipDirectory;
			}
			for (const layer of layers) {
				if (covers(layer, root, relative, true)) {
					count(layer.name);
					return skipDirectory;
				}
			}
			disk.load(`${relative}/.git`);
			disk.load(`${relative}/${IgnoreFileName}`);
			const entered = scope.enter(relative);
			if (entered.kind === 'Error') {
				return { kind: 'Error', message: entered.error.message };
			}
			scopes.set(relative, entered.value);
			enumeration.directories.push(path);
			enumeration.ignoreFiles.push(join(path, IgnoreFileName));
			return proceed;
		}

		// A symbolic link is never a directory to filepath.Walk, which does not follow it, so a link to a
		// directory arrives here as a file, and is skipped after the ignore layers have had their say.
		enumeration.walked++;

		const scope = scopeOf(scopes, relative);
		if (scope.ignored(relative, false)[0]) {
			count(GitignoreLayer);
			return proceed;
		}
		for (const layer of layers) {
			if (covers(layer, root, relative, false)) {
				count(layer.name);
				return proceed;
			}
		}

		if (status.symbolicLink) {
			enumeration.symbolicLinks++;
			return proceed;
		}

		// cohere 483acc0e: .a files wait until a TypeScript program claims them.
		if (path.endsWith('.a') && base(path) !== '.a') {
			enumeration.adamic.push(path);
			return proceed;
		}

		if (!handles(path)) {
			enumeration.unhandled++;
			let extension = goToLower(ext(path));
			if (extension === '') {
				extension = '(none)';
			}
			enumeration.declinedExtensions.set(extension, (enumeration.declinedExtensions.get(extension) ?? 0) + 1);
			return proceed;
		}

		enumeration.files.push(path);
		return proceed;
	};
	const walked = walk(root, visit);
	if (walked.kind === 'Error') {
		return { kind: 'Error', message: `walking ${root}: ${walked.message}` };
	}

	return { kind: 'Ok', enumeration };
}
