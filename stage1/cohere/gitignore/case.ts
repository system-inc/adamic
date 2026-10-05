// The shape of what main.ts is asked: working trees with the paths to decide in each, and globs with
// the text to match. cases.ts holds them as constants, because 0.1 programs have no input; the test
// (gitignore_test.go) writes its own cases.ts from cohere's tests and git's.

import type { Entry } from './gitignore.ts';

// One path to decide, and whether it is a directory.
export interface Query {
	readonly path: string;
	readonly isDirectory: boolean;
}

// A working tree: its name for the output, the root error messages show, its entries, and its queries.
export interface TreeCase {
	readonly name: string;
	readonly root: string;
	readonly entries: ReadonlyMap<string, Entry>;
	readonly queries: readonly Query[];
}

// A list of ignore-file lines that is not a file in a tree, compiled as compilePatterns does, with the
// paths to decide by it.
export interface PatternsCase {
	readonly name: string;
	readonly lines: readonly string[];
	readonly queries: readonly Query[];
}

// One glob matched against one text, as a path or as a base name.
export interface GlobCase {
	readonly pattern: string;
	readonly text: string;
	readonly path: boolean;
}
