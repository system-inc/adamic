// ESLint's reading of eslint-disable comments with whitespace Go's strings.TrimSpace and JavaScript's
// trim disagree on: U+0085, which Go trims and JavaScript doesn't, and U+FEFF, the other way round.
// GAPS.md ("Go cohere and ESLint disagree on two whitespace characters") has what it printed on ESLint
// 10.12.0 beside what Go cohere answers. Run it where `npm install eslint` ran:
//
//	NODE_PATH=<that directory>/node_modules node eslint_whitespace.cjs

const { Linter } = require('eslint');
const linter = new Linter({ configType: 'flat' });
const config = { rules: { 'no-debugger': 'error' } };
const cases = {
	'plain': '/* eslint-disable no-debugger */\ndebugger;',
	'U+0085 before the word': '/*\u0085eslint-disable no-debugger */\ndebugger;',
	'U+FEFF before the word': '/*\ufeffeslint-disable no-debugger */\ndebugger;',
	'U+0085 after the rule': '/* eslint-disable no-debugger\u0085*/\ndebugger;',
	'U+FEFF after the rule': '/* eslint-disable no-debugger\ufeff*/\ndebugger;',
	'next-line, U+0085 before': '//\u0085eslint-disable-next-line no-debugger\ndebugger;',
	'next-line, U+FEFF before': '//\ufeffeslint-disable-next-line no-debugger\ndebugger;',
	'U+00A0 before the word': '/*\u00a0eslint-disable no-debugger */\ndebugger;',
};
for (const [name, code] of Object.entries(cases)) {
	const messages = linter.verify(code, config);
	console.log(`${name}: ${messages.filter((m) => m.ruleId === 'no-debugger').length === 0 ? 'suppressed' : 'reported'}`);
}
