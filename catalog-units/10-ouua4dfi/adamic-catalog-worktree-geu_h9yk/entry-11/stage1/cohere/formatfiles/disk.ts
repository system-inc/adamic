// The file system as the formatfiles Go sees it through os.Lstat, os.Stat and os.ReadFile, read with
// 'adamic''s fileStatus, readDirectory and readTextFile; and a gitignore WorkingTree read from disk, which
// the gitignore slice's GAPS.md waited for ("a WorkingTree read from disk can take the in-memory one's
// place, and nothing else changes").
//
// Two things the Go has that 'adamic' doesn't, and how they are stood in for:
//
//   - os.Lstat sees a symbolic link that can't be followed, to nothing or round a loop; fileStatus
//     follows links and calls the first no such file and the second failed (gap 1 in GAPS.md). A name
//     its directory lists that fileStatus can't follow is such a link, which is what Lstat would have
//     said.
//   - os.Readlink. The gitignore port follows a link by its target, within its tree; there is no target
//     to read, so a link's entry names a stand-in sibling, and that sibling's entry is what fileStatus and
//     readTextFile find through the link, which is what os.Stat and os.ReadFile find. A link to nothing
//     names a stand-in with no entry.

import { fileStatus, readDirectory, readTextFile } from 'adamic';
import type { Entry, WorkingTree } from '../gitignore/gitignore.ts';
import { base, dir } from '../gitignore/path.ts';

// What os.Lstat says of a path: whether something is there, whether it is a directory (a link never is),
// and whether it is a symbolic link.
export interface LinkStatus {
	readonly exists: boolean;
	readonly directory: boolean;
	readonly symbolicLink: boolean;
}

// lstat is os.Lstat.
export function lstat(path: string): LinkStatus {
	const status = fileStatus(path);
	if (status.kind === 'Ok') {
		return { exists: true, directory: status.type === 'directory' && !status.symbolicLink, symbolicLink: status.symbolicLink };
	}
	// Listed, but nothing there to follow to: a symbolic link to nothing, or round a loop of links
	// (fileStatus says no such file for the first and failed for the second, ELOOP). A permission error
	// is the directory refusing a look at all, which refuses os.Lstat too.
	if (!status.message.endsWith(': permission denied')) {
		const listing = readDirectory(dir(path));
		if (listing.kind === 'Ok' && listing.names.includes(base(path))) {
			return { exists: true, directory: false, symbolicLink: true };
		}
	}
	return { exists: false, directory: false, symbolicLink: false };
}

// statExists is whether os.Stat finds something at path, following symbolic links.
export function statExists(path: string): boolean {
	return fileStatus(path).kind === 'Ok';
}

// statIsDirectory is whether os.Stat finds a directory at path, following symbolic links.
export function statIsDirectory(path: string): boolean {
	const status = fileStatus(path);
	return status.kind === 'Ok' && status.type === 'directory';
}

// followedEntry is what a path names once its links are followed, as a gitignore Entry, or undefined
// when it names nothing.
function followedEntry(path: string): Entry | undefined {
	const status = fileStatus(path);
	if (status.kind === 'Error') {
		return undefined;
	}
	switch (status.type) {
		case 'directory':
			return { kind: 'Directory' };
		case 'other':
			return { kind: 'Other' };
		case 'file': {
			const read = readTextFile(path);
			if (read.kind === 'Error') {
				// An ignore file that can't be read: the Go's os.ReadFile error, which no case here reaches
				// (GAPS.md, "Not covered").
				return { kind: 'Other' };
			}
			return { kind: 'File', contents: read.text };
		}
	}
}

// A gitignore WorkingTree read from disk as the matcher asks for it. The matcher reads its entries when
// it is made (the root's .gitignore, .git and .git/info/exclude) and when it enters a directory (that
// directory's .git and .gitignore), so the walk loads those paths just before, and a pruned directory is
// never read, as the Go never reads it.
export class DiskTree {
	readonly #root: string;
	readonly #entries = new Map<string, Entry>();
	readonly tree: WorkingTree;

	constructor(root: string) {
		this.#root = root;
		this.tree = { root, entries: this.#entries };
	}

	// load reads what a path, relative to the root, is on disk into the tree, if it is anything.
	load(relative: string): void {
		const path = relative === '' ? this.#root : `${this.#root}/${relative}`;
		const status = lstat(path);
		if (!status.exists) {
			return;
		}
		if (!status.symbolicLink) {
			const entry = followedEntry(path);
			if (entry !== undefined) {
				this.#entries.set(relative, entry);
			}
			return;
		}
		// The stand-in sibling the link names, which the gitignore port's stat resolves from the link's
		// directory.
		const target = `\u0000${base(relative)}`;
		this.#entries.set(relative, { kind: 'SymbolicLink', target });
		const parent = dir(relative);
		const followed = followedEntry(path);
		if (followed !== undefined) {
			this.#entries.set(parent === '.' ? target : `${parent}/${target}`, followed);
		}
	}
}
