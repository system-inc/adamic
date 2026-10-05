// A sample of what main.ts decides: the tree of cohere's TestTheMatcherFollowsGitignoreSemantics
// (cohere/internal/gitignore/gitignore_test.go) with its cases, and a few of glob_test.go's. The test
// (gitignore_test.go) writes a cases.ts of its own, with everything cohere's tests and git's corpus ask.

import type { GlobCase, TreeCase } from './case.ts';
import type { Entry } from './gitignore.ts';

export const trees: readonly TreeCase[] = [
	{
		name: 'semantics',
		root: '/repository',
		entries: new Map<string, Entry>([
			['.git', { kind: 'Directory' }],
			['.git/info/exclude', { kind: 'File', contents: '# a comment\nper-repo\nkept-by-gitignore\n' }],
			[
				'.gitignore',
				{
					kind: 'File',
					contents:
						'*.log\n!keep.log\n/anchored\nbuild/\ndocs/generated\n\\#literal\n\\!bang\ntrailing   \nescaped\\ \n**/deep/x\nvendor/\n!vendor/kept\n!kept-by-gitignore\n',
				},
			],
			['sub/.gitignore', { kind: 'File', contents: '!*.log\nlocal\n/rooted\n' }],
			['sub/inner/.gitignore', { kind: 'File', contents: '*.tmp\n' }],
		]),
		queries: [
			{ path: 'app.log', isDirectory: false },
			{ path: 'keep.log', isDirectory: false },
			{ path: 'sub/app.log', isDirectory: false },
			{ path: 'anchored', isDirectory: false },
			{ path: 'sub/anchored', isDirectory: false },
			{ path: 'build', isDirectory: true },
			{ path: 'build', isDirectory: false },
			{ path: 'sub/build', isDirectory: true },
			{ path: 'docs/generated', isDirectory: false },
			{ path: 'sub/docs/generated', isDirectory: false },
			{ path: '#literal', isDirectory: false },
			{ path: '!bang', isDirectory: false },
			{ path: 'trailing', isDirectory: false },
			{ path: 'escaped ', isDirectory: false },
			{ path: 'deep/x', isDirectory: false },
			{ path: 'a/b/deep/x', isDirectory: false },
			{ path: 'vendor/kept', isDirectory: false },
			{ path: 'sub/local', isDirectory: false },
			{ path: 'sub/inner/local', isDirectory: false },
			{ path: 'sub/rooted', isDirectory: false },
			{ path: 'sub/inner/rooted', isDirectory: false },
			{ path: 'sub/inner/a.tmp', isDirectory: false },
			{ path: 'sub/a.tmp', isDirectory: false },
			{ path: 'per-repo', isDirectory: false },
			{ path: 'sub/per-repo', isDirectory: false },
			{ path: 'kept-by-gitignore', isDirectory: false },
		],
	},
	{
		name: 'byte order mark',
		root: '/repository',
		entries: new Map<string, Entry>([['.gitignore', { kind: 'File', contents: '\uFEFFfirst\r\nsecond\r\n' }]]),
		queries: [
			{ path: 'first', isDirectory: false },
			{ path: 'second', isDirectory: false },
		],
	},
];

export const globCases: readonly GlobCase[] = [
	{ pattern: 'a/**/b', text: 'a/x/y/b', path: true },
	{ pattern: 'foo**/bar', text: 'foo/x/bar', path: true },
	{ pattern: '[[:digit:]]', text: '7', path: true },
	{ pattern: '?', text: 'é', path: false },
	{ pattern: '??', text: 'é', path: false },
];
