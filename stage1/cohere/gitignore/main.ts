// The port's driver: it decides every query of every tree and every pattern list in cases.ts, and matches
// every glob case, in the words git check-ignore --verbose --non-matching uses, so its output can be held
// byte for byte to Go cohere's and to git's. Each answer line is `<ignored> <file>:<line>:<pattern><TAB><path>`, with
// `::` for a path no rule decided, and `<ignored>` 1 or 0.

import { globCases, patternsCases, trees } from './cases.ts';
import { compilePatterns, newMatcher } from './gitignore.ts';
import { compileGlob } from './glob.ts';

for (const tree of trees) {
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
		const verdict = answer.value;
		console.log(`${verdict.ignored ? '1' : '0'} ${verdict.source.isZero() ? '::' : verdict.source.toString()}\t${query.path}`);
	}
}

for (const patternsCase of patternsCases) {
	console.log(`patterns ${patternsCase.name}`);
	const compiled = compilePatterns(patternsCase.lines, patternsCase.name);
	if (compiled.kind === 'Error') {
		console.log(`error ${compiled.error.message}`);
		continue;
	}
	for (const query of patternsCase.queries) {
		const verdict = compiled.value.ignored(query.path, query.isDirectory);
		console.log(`${verdict.ignored ? '1' : '0'} ${verdict.source.isZero() ? '::' : verdict.source.toString()}\t${query.path}`);
	}
}

for (let index = 0; index < globCases.length; index++) {
	const globCase = globCases[index];
	if (globCase !== undefined) {
		const matched = compileGlob(globCase.pattern, globCase.path).matches(globCase.text, globCase.path);
		console.log(`glob ${index} ${matched ? '1' : '0'}`);
	}
}
