// The port's driver: it reads the cases file named by its one argument (case.ts says the format), decides
// every query of every tree and every pattern list in it, and matches every glob, in the words git
// check-ignore --verbose --non-matching uses, so its output can be held byte for byte to Go cohere's and
// to git's. Each answer line is `<ignored> <file>:<line>:<pattern><TAB><path>`, with `::` for a path no
// rule decided, and `<ignored>` 1 or 0.
//
//	node oracle/node.mjs stage1/cohere/gitignore/main.ts stage1/cohere/gitignore/sample-cases.txt

import { panic, programArguments, readTextFile } from 'adamic';
import { parseCases } from './case.ts';
import { compilePatterns, newMatcher, type Matcher, type Result, type Verdict } from './gitignore.ts';
import { dir } from './path.ts';
import { compileGlob } from './glob.ts';

// verdictLine is a verdict in check-ignore's verbose words, without the path.
function verdictLine(verdict: Verdict): string {
	const [ignored, source] = verdict;
	return `${ignored ? '1' : '0'} ${source.isZero() ? '::' : source.toString()}`;
}

// scopeFor is the matcher for directory, as a walk has it: entered from the nearest directory above it
// that was entered before, and kept for the next path there. The root's is always kept.
function scopeFor(scopes: Map<string, Matcher>, directory: string): Result<Matcher> {
	let ancestor = directory;
	let entered = scopes.get(ancestor);
	while (entered === undefined) {
		const parent = dir(ancestor);
		ancestor = parent === '.' ? '' : parent;
		entered = scopes.get(ancestor);
	}
	if (ancestor === directory) {
		return { kind: 'Ok', value: entered };
	}
	const scope = entered.enter(directory);
	if (scope.kind === 'Ok') {
		scopes.set(directory, scope.value);
	}
	return scope;
}

const casesPath = programArguments()[0] ?? panic('usage: main.ts <cases file>');
const read = readTextFile(casesPath);
if (read.kind === 'Error') {
	panic(read.message);
}
const asked = parseCases(read.text);

for (const tree of asked.trees) {
	console.log(`tree ${tree.name}`);
	const created = newMatcher({ root: tree.root, entries: tree.entries });
	if (created.kind === 'Error') {
		console.log(`error ${created.error.message}`);
		continue;
	}
	for (const query of tree.queries) {
		const answer = created.value.ignoredPath(query.path, query.isDirectory);
		if (answer.kind === 'Error') {
			console.log(`error ${answer.error.message}${answer.error.nestedRepository ? ' (nested repository)' : ''}`);
			continue;
		}
		console.log(`${verdictLine(answer.value)}\t${query.path}`);
	}

	// The same paths again, as cohere's walk asks them: each through the matcher of its own directory,
	// entered from the nearest one entered before, with whether that directory is itself excluded.
	console.log(`walk ${tree.name}`);
	const scopes = new Map<string, Matcher>([['', created.value]]);
	for (const query of tree.queries) {
		const parent = dir(query.path);
		const scope = scopeFor(scopes, parent === '.' ? '' : parent);
		if (scope.kind === 'Error') {
			console.log(`error ${scope.error.message}${scope.error.nestedRepository ? ' (nested repository)' : ''}`);
			continue;
		}
		const excluded = scope.value.excluded();
		console.log(`${verdictLine(scope.value.ignored(query.path, query.isDirectory))}\t${query.path}\t${verdictLine(excluded)}`);
	}
}

for (const patternsCase of asked.patterns) {
	console.log(`patterns ${patternsCase.name}`);
	const compiled = compilePatterns(patternsCase.lines, patternsCase.name);
	if (compiled.kind === 'Error') {
		console.log(`error ${compiled.error.message}`);
		continue;
	}
	for (const query of patternsCase.queries) {
		console.log(`${verdictLine(compiled.value.ignored(query.path, query.isDirectory))}\t${query.path}`);
	}
}

for (let index = 0; index < asked.globs.length; index++) {
	const globCase = asked.globs[index];
	if (globCase !== undefined) {
		const matched = compileGlob(globCase.pattern, globCase.path).matches(globCase.text, globCase.path);
		console.log(`glob ${index} ${matched ? '1' : '0'}`);
	}
}
