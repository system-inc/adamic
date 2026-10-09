// cohere/command/cohere/discovery.go: project markers, not TypeScript source enumeration.
// Directory reads are sequential here; Go sorts the parallel walk's result, so order is identical.
import { fileStatus, readDirectory } from 'adamic';
import { DiskTree, lstat } from '../formatfiles/disk.ts';
import { compareCodePoints, join, rel } from '../formatfiles/golang.ts';
import { clean, dir } from '../gitignore/path.ts';
import { ExcludeFile, IgnoreFileName, newMatcher, type Matcher } from '../gitignore/gitignore.ts';
import { ignorePatternsOf } from './settings.ts';
import { LintGlob, matchAny } from './glob.ts';

export interface Project {
	readonly directory: string;
	readonly engine: string;
}

export class Discovery {
	readonly projects: Project[] = [];
	readonly refused: string[] = [];
	readonly nestedRepositories: string[] = [];
	submodules = 0;
	ignored = 0;
	neverDescended = 0;
	listed = 0;
}

export type Discovered = { readonly kind: 'Ok'; readonly found: Discovery } | { readonly kind: 'Error'; readonly message: string };

function nearestRepository(start: string): string {
	for (let directory = clean(start); ; directory = dir(directory)) {
		if (lstat(join(directory, '.git')).exists) { return directory; }
		if (dir(directory) === directory) { return start; }
	}
}

function pathBefore(left: string, right: string): number {
	if (left === '.' || right === '.') { return left === right ? 0 : left === '.' ? -1 : 1; }
	const leftSegments = left.split('/');
	const rightSegments = right.split('/');
	for (let index = 0; index < leftSegments.length && index < rightSegments.length; index++) {
		const compared = compareCodePoints(leftSegments[index] ?? '', rightSegments[index] ?? '');
		if (compared !== 0) { return compared; }
	}
	return leftSegments.length - rightSegments.length;
}

function fromRepository(repository: string, root: string, relative: string): string {
	const path = rel(repository, join(root, relative));
	return path === '.' ? '' : path;
}

// DiskTree loads ignore inputs before Enter, including the ancestors when root is inside a repository.
function loadScope(disk: DiskTree, relative: string): void {
	let prefix = '';
	for (const segment of relative.split('/')) {
		if (segment === '') { continue; }
		prefix = prefix === '' ? segment : prefix + '/' + segment;
		disk.load(prefix + '/.git');
		disk.load(prefix + '/' + IgnoreFileName);
	}
}

function visit(root: string, repository: string, relative: string, scopeGiven: Matcher, disk: DiskTree, patterns: readonly LintGlob[], found: Discovery): string {
	const absolute = join(root, relative);
	const listing = readDirectory(absolute);
	if (listing.kind === 'Error') {
		// Runtime errors are coarser than errno. Known Linux errors are translated; unknown ones are
		// named as a gap instead of claimed as Go's exact error.
		const reason = listing.message.endsWith(': no such directory') ? 'no such file or directory'
			: listing.message.endsWith(': not a directory') ? 'not a directory'
			: listing.message.endsWith(': permission denied') ? 'permission denied' : 'unrepresented filesystem error';
		return `reading ${absolute}: open ${absolute}: ${reason}`;
	}
	found.listed++;
	const names: string[] = [];
	const directories: string[] = [];
	for (const name of listing.names) {
		const status = lstat(join(absolute, name));
		if (name === '.git' && relative !== '.') {
			if (status.directory) { found.nestedRepositories.push(relative); } else { found.submodules++; }
			return '';
		}
		if (status.directory) { directories.push(name); } else { names.push(name); }
	}
	let scope = scopeGiven;
	if (relative !== '.') {
		const inside = fromRepository(repository, root, relative);
		loadScope(disk, inside);
		const entered = scope.enter(inside);
		if (entered.kind === 'Error') { return entered.error.message; }
		scope = entered.value;
	}
	for (const marker of ['tsconfig.json', 'Package.swift']) {
		if (!names.includes(marker)) { continue; }
		const markerPath = relative === '.' ? marker : relative + '/' + marker;
		if (matchAny(patterns, markerPath)) { found.refused.push(markerPath); }
		else { found.projects.push({ directory: relative, engine: marker === 'tsconfig.json' ? 'TypeScript' : 'Swift' }); }
	}
	for (const name of directories) {
		const child = relative === '.' ? name : relative + '/' + name;
		if (['node_modules', '.build', '.cache', '.git', 'testdata'].includes(name)) {
			if (name !== '.git') { found.neverDescended++; }
			continue;
		}
		if (scope.ignored(fromRepository(repository, root, child), true)[0]) { found.ignored++; continue; }
		const error = visit(root, repository, child, scope, disk, patterns, found);
		if (error !== '') { return error; }
	}
	return '';
}

export function discoverProjects(root: string, ignorePatterns: readonly string[]): Discovered {
	const repository = nearestRepository(root);
	const disk = new DiskTree(repository);
	disk.load(IgnoreFileName);
	disk.load('.git');
	disk.load(ExcludeFile);
	const made = newMatcher(disk.tree);
	if (made.kind === 'Error') { return { kind: 'Error', message: made.error.message }; }
	let scope = made.value;
	const prefix = fromRepository(repository, root, '.');
	if (prefix !== '') {
		loadScope(disk, prefix);
		const entered = scope.enter(prefix);
		if (entered.kind === 'Error') { return { kind: 'Error', message: entered.error.message }; }
		scope = entered.value;
	}
	const found = new Discovery();
	const patterns = ignorePatterns.map((pattern) => new LintGlob(pattern));
	const error = visit(root, repository, '.', scope, disk, patterns, found);
	if (error !== '') { return { kind: 'Error', message: error }; }
	found.projects.sort((left, right) => {
		const compared = pathBefore(left.directory, right.directory);
		return compared !== 0 ? compared : left.engine === right.engine ? 0 : left.engine === 'TypeScript' ? -1 : 1;
	});
	found.refused.sort(compareCodePoints);
	found.nestedRepositories.sort(compareCodePoints);
	return { kind: 'Ok', found };
}

// command/cohere/rootIgnorePatterns: a missing/default non-file settings path contributes no globs.
export function discoverConfiguredProjects(root: string): Discovered {
    const path = join(root, 'CohereSettings.json');
    const status = fileStatus(path);
    if (status.kind !== 'Ok' || status.type !== 'file') { return discoverProjects(root, []); }
    const patterns = ignorePatternsOf(path);
    if (patterns.kind === 'Error') { return patterns; }
    return discoverProjects(root, patterns.values);
}
