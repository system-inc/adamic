#!/usr/bin/env node
// A PreToolUse hook for Bash, registered in .claude/settings.json.
//
// Shell state does not carry between Bash calls, so a variable set in one call is empty in the next,
// and `rm -f $S/build.go.bak` becomes `rm -f /build.go.bak`. Claude Code's own check then holds the
// call for a human's Allow or Deny, and a cloud stream waits on it for as long as nobody looks. This
// hook refuses those commands first, with the reason and the fix, so the model rewrites them and
// keeps going.
//
// It reads the command the way a shell would (quotes, escapes, expansions, ; && || | and newlines,
// subshells, heredocs) and looks at every rm, rmdir, unlink, shred, mv, and find that deletes. Each
// path those touch must stay inside the repository or /tmp, and must not start with a variable that
// can be empty. ${VAR:?} cannot be empty (the shell stops the command instead), so it is allowed; a
// bare $VAR, ${VAR} or $(...) is not. Every unguarded expansion is taken as empty, the worst case,
// so `rm -rf /tmp/$X` is refused too: with X empty it removes /tmp itself.
//
// When the hook cannot read the input or the command, it says nothing and Claude Code's normal
// permission flow decides, which is what happened before this hook existed.

import { readFileSync, realpathSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const removalCommands = new Set(['rm', 'rmdir', 'unlink', 'shred']);
const shells = new Set(['bash', 'sh', 'dash', 'zsh', 'ksh']);
const wrapperCommands = new Set(['sudo', 'command', 'builtin', 'exec', 'nohup', 'nice', 'time', 'env', 'timeout', 'stdbuf', 'ionice', 'xargs']);
const reservedWords = new Set(['!', '{', '}', 'if', 'then', 'elif', 'else', 'fi', 'do', 'done', 'while', 'until']);
const assignmentPattern = /^([A-Za-z_][A-Za-z0-9_]*)=/;

// The lexer. It turns a command line into simple commands, each a list of words, and each word a list
// of parts: literal text, a variable, or a command substitution. Commands nested in $(...), `...`,
// bash -c and eval come back as `nested` source text, read again by the caller.

function lex(source) {
	const commands = [];
	let words = [];
	let parts = [];
	let wordStarted = false;
	let skipNextWord = false;
	const pendingHeredocs = [];
	let index = 0;

	function pushLiteral(text, quoted) {
		const last = parts[parts.length - 1];
		if (last && last.kind === 'Literal' && last.quoted === quoted) {
			last.text += text;
		} else {
			parts.push({ kind: 'Literal', text, quoted });
		}
		wordStarted = true;
	}

	function endWord() {
		if (wordStarted) {
			if (skipNextWord) {
				skipNextWord = false;
				if (pendingHeredocs.length > 0 && pendingHeredocs[pendingHeredocs.length - 1].delimiter === null) {
					pendingHeredocs[pendingHeredocs.length - 1].delimiter = parts.map((part) => (part.kind === 'Literal' ? part.text : part.raw)).join('');
				}
			} else {
				words.push(parts);
			}
		}
		parts = [];
		wordStarted = false;
	}

	function endCommand() {
		endWord();
		if (words.length > 0) {
			commands.push(words);
		}
		words = [];
	}

	// Reads to the matching close of $( or `, honoring quotes, and returns the inner text.
	function readBalanced(open, close) {
		let depth = 1;
		const start = index;
		while (index < source.length) {
			const character = source[index];
			if (character === '\\') {
				index += 2;
				continue;
			}
			if (character === "'" && open !== '`') {
				const end = source.indexOf("'", index + 1);
				if (end === -1) {
					throw new Error('unterminated quote');
				}
				index = end + 1;
				continue;
			}
			if (open !== '`' && character === open) {
				depth++;
			} else if (character === close) {
				depth--;
				if (depth === 0) {
					index++;
					return source.slice(start, index - 1);
				}
			}
			index++;
		}
		throw new Error('unterminated substitution');
	}

	// Reads one expansion starting at a `$` and returns its part.
	function readDollar() {
		index++;
		const next = source[index];
		if (next === '(' && source[index + 1] === '(') {
			index += 2;
			const inner = readBalanced('(', ')');
			if (source[index] === ')') {
				index++;
			}
			return { kind: 'Variable', name: '((arithmetic))', guarded: true, raw: '$((' + inner + '))' };
		}
		if (next === '(') {
			index++;
			const inner = readBalanced('(', ')');
			return { kind: 'Substitution', nested: inner, raw: '$(' + inner + ')' };
		}
		if (next === '{') {
			index++;
			const inner = readBalanced('{', '}');
			const match = /^([A-Za-z_][A-Za-z0-9_]*|[0-9]+|[@*#?$!-])([\s\S]*)$/.exec(inner);
			if (!match) {
				return { kind: 'Variable', name: inner, guarded: false, raw: '${' + inner + '}' };
			}
			const operator = match[2];
			// ${VAR:?} and ${VAR?} stop the command when VAR is unset; ${VAR:-word} and ${VAR:=word}
			// fall back to word. Either way the expansion is never empty. Everything else can be.
			const guarded = /^:?\?/.test(operator) || (/^:?[-=]./.test(operator) && !/^:?[-=]$/.test(operator));
			return { kind: 'Variable', name: match[1], guarded, raw: '${' + inner + '}' };
		}
		const nameMatch = /^([A-Za-z_][A-Za-z0-9_]*|[0-9@*#?$!-])/.exec(source.slice(index));
		if (!nameMatch) {
			return { kind: 'Literal', text: '$', quoted: false };
		}
		index += nameMatch[1].length;
		// $$, $? and $# always expand to a number.
		const guarded = nameMatch[1] === '$' || nameMatch[1] === '?' || nameMatch[1] === '#';
		return { kind: 'Variable', name: nameMatch[1], guarded, raw: '$' + nameMatch[1] };
	}

	function pushPart(part) {
		if (part.kind === 'Literal') {
			pushLiteral(part.text, part.quoted);
		} else {
			parts.push(part);
			wordStarted = true;
		}
	}

	// After a newline, the bodies of any heredocs opened on that line follow; they are data, not
	// commands, so they are skipped.
	function skipHeredocBodies() {
		while (pendingHeredocs.length > 0) {
			const heredoc = pendingHeredocs.shift();
			const delimiter = (heredoc.delimiter || '').replace(/["'\\]/g, '');
			while (index < source.length) {
				let lineEnd = source.indexOf('\n', index);
				if (lineEnd === -1) {
					lineEnd = source.length;
				}
				let line = source.slice(index, lineEnd);
				index = lineEnd + 1;
				if (heredoc.stripTabs) {
					line = line.replace(/^\t+/, '');
				}
				if (line === delimiter) {
					break;
				}
			}
		}
	}

	while (index < source.length) {
		const character = source[index];

		if (character === '\\') {
			if (source[index + 1] === '\n') {
				index += 2;
				continue;
			}
			pushLiteral(source[index + 1] || '', true);
			index += 2;
			continue;
		}
		if (character === "'") {
			const end = source.indexOf("'", index + 1);
			if (end === -1) {
				throw new Error('unterminated quote');
			}
			pushLiteral(source.slice(index + 1, end), true);
			index = end + 1;
			continue;
		}
		if (character === '$' && source[index + 1] === "'") {
			// ANSI-C quoting. Its escapes are not decoded; the text stays as written.
			const end = source.indexOf("'", index + 2);
			if (end === -1) {
				throw new Error('unterminated quote');
			}
			pushLiteral(source.slice(index + 2, end), true);
			index = end + 1;
			continue;
		}
		if (character === '"') {
			index++;
			wordStarted = true;
			while (index < source.length && source[index] !== '"') {
				const inner = source[index];
				if (inner === '\\' && '"\\$`\n'.includes(source[index + 1])) {
					if (source[index + 1] !== '\n') {
						pushLiteral(source[index + 1], true);
					}
					index += 2;
				} else if (inner === '$') {
					const part = readDollar();
					pushPart(part.kind === 'Literal' ? { ...part, quoted: true } : part);
				} else if (inner === '`') {
					index++;
					const nested = readBalanced('`', '`');
					pushPart({ kind: 'Substitution', nested, raw: '`' + nested + '`' });
				} else {
					pushLiteral(inner, true);
					index++;
				}
			}
			if (index >= source.length) {
				throw new Error('unterminated quote');
			}
			index++;
			continue;
		}
		if (character === '$') {
			pushPart(readDollar());
			continue;
		}
		if (character === '`') {
			index++;
			const nested = readBalanced('`', '`');
			pushPart({ kind: 'Substitution', nested, raw: '`' + nested + '`' });
			continue;
		}
		if (character === '#' && !wordStarted) {
			while (index < source.length && source[index] !== '\n') {
				index++;
			}
			continue;
		}
		if (character === '\n') {
			endCommand();
			index++;
			skipHeredocBodies();
			continue;
		}
		if (character === ' ' || character === '\t') {
			endWord();
			index++;
			continue;
		}
		if (character === ';' || character === '&' || character === '|' || character === '(' || character === ')') {
			if (character === '&' && source[index + 1] === '>') {
				// &> and &>> redirect both streams to the next word.
				endWord();
				index += source[index + 2] === '>' ? 3 : 2;
				skipNextWord = true;
				continue;
			}
			endCommand();
			index++;
			if ((character === ';' || character === '&' || character === '|') && source[index] === character) {
				index++;
			}
			continue;
		}
		if (character === '<' || character === '>') {
			// A word of digits right before the operator is a file descriptor, not an argument.
			if (wordStarted && parts.length === 1 && parts[0].kind === 'Literal' && !parts[0].quoted && /^[0-9]+$/.test(parts[0].text)) {
				parts = [];
				wordStarted = false;
			}
			endWord();
			if (source.startsWith('<<<', index)) {
				index += 3;
				skipNextWord = true;
				continue;
			}
			if (source.startsWith('<<', index)) {
				const stripTabs = source[index + 2] === '-';
				index += stripTabs ? 3 : 2;
				pendingHeredocs.push({ delimiter: null, stripTabs });
				skipNextWord = true;
				continue;
			}
			index++;
			if (source[index] === '>' || source[index] === '|' || source[index] === '&' || (character === '<' && source[index] === '>')) {
				index++;
			}
			// >&2 and similar name a descriptor, not a file.
			if (/[0-9-]/.test(source[index] || '') && source[index - 1] === '&') {
				while (/[0-9-]/.test(source[index] || '')) {
					index++;
				}
				continue;
			}
			skipNextWord = true;
			continue;
		}
		pushLiteral(character, false);
		index++;
	}
	endCommand();
	return commands;
}

// Evaluation. A word becomes what the shell would make of it, in the worst case: unguarded
// expansions are empty, guarded ones take the value this command gave them (or the hook's own
// environment's) and are otherwise unknown.

const unknownComponent = '\u2060unknown\u2060';

function evaluateWord(word, variables) {
	let text = '';
	let unguarded = null;
	let startsUnguarded = false;
	let startsUnknown = false;
	let hasUnknown = false;
	word.forEach((part, partIndex) => {
		// Unguarded expansions add nothing, so `$A$B/x` begins with both.
		const atStart = text === '';
		if (part.kind === 'Literal') {
			if (partIndex === 0 && !part.quoted && (part.text === '~' || part.text.startsWith('~/'))) {
				const home = process.env.HOME;
				if (home) {
					text += home + part.text.slice(1);
				} else {
					startsUnknown = true;
					hasUnknown = true;
					text += unknownComponent + part.text.slice(1);
				}
				return;
			}
			text += part.text;
			return;
		}
		if (part.kind === 'Substitution' || !part.guarded) {
			unguarded = unguarded || part.raw;
			if (atStart) {
				startsUnguarded = true;
			}
			return;
		}
		const value = variables.has(part.name) ? variables.get(part.name) : process.env[part.name];
		if (typeof value === 'string' && value !== '') {
			text += value;
		} else {
			hasUnknown = true;
			if (atStart) {
				startsUnknown = true;
			}
			text += unknownComponent;
		}
	});
	return { text, unguarded, startsUnguarded, startsUnknown, hasUnknown };
}

// A word as written, quotes removed and expansions shown as they were typed, for messages.
function sourceTextOfWord(word) {
	return word.map((part) => (part.kind === 'Literal' ? part.text : part.raw)).join('');
}

// The text a word hands to a nested shell (bash -c, eval). The outer shell expands first, so an
// unguarded expansion arrives empty: `sh -c "rm -rf '$S'/"` runs `rm -rf ''/`. Guarded ones are kept
// as written for the nested read to treat as guarded.
function nestedTextOfWord(word) {
	return word.map((part) => (part.kind === 'Literal' ? part.text : part.kind === 'Variable' && part.guarded ? part.raw : '')).join('');
}

function literalOf(word) {
	return word.every((part) => part.kind === 'Literal') ? word.map((part) => part.text).join('') : null;
}

// A command name, also when it comes through a variable this command set: `R=rm; $R -rf /`.
function commandNameOf(word, variables) {
	let name = '';
	for (const part of word) {
		if (part.kind === 'Literal') {
			name += part.text;
		} else if (part.kind === 'Variable' && variables.get(part.name)) {
			name += variables.get(part.name);
		} else {
			return null;
		}
	}
	return name;
}

function isInside(target, root) {
	return target === root || target.startsWith(root.endsWith('/') ? root : root + '/');
}

// The repository, /tmp, and the system's temporary directory (TMPDIR, which on macOS is not /tmp).
function allowedRoots(repository) {
	const written = [repository, '/tmp', path.resolve(tmpdir())];
	const roots = new Set(written);
	for (const root of written) {
		try {
			roots.add(realpathSync(root));
		} catch {
			// A root that does not exist here keeps only its written spelling.
		}
	}
	return [...roots];
}

const fixAdvice =
	'Shell state does not carry between Bash calls, so set the variable in this same command, and write it as "${S:?}/name": ' +
	'an unset variable then stops the command instead of collapsing the path to /. Remove only named files under the repository or /tmp.';

// find and mv may name the repository or /tmp itself (find . -name '*.log' -delete, mv x /tmp), as long
// as no empty expansion is what made it so; rm, rmdir, unlink and shred may not.
function checkTarget(commandName, word, state, problems, rootItselfAllowed = false) {
	const evaluated = evaluateWord(word, state.variables);
	const written = sourceTextOfWord(word);
	if (evaluated.startsUnguarded) {
		problems.push(
			`${commandName} ${written}: the path begins with ${evaluated.unguarded}, which is empty when unset, so the path would become "${evaluated.text || '(nothing)'}".`,
		);
		return;
	}
	if (evaluated.startsUnknown) {
		// ${VAR:?} at the start with a value this hook cannot see: the shell guarantees it is set.
		return;
	}
	if (evaluated.text === '') {
		return;
	}
	let resolved;
	if (path.isAbsolute(evaluated.text)) {
		resolved = path.resolve(evaluated.text);
	} else if (state.base === null) {
		problems.push(`${commandName} ${written}: a relative path after a cd whose target this hook cannot tell, so it cannot say where the path lands.`);
		return;
	} else {
		resolved = path.resolve(state.base, evaluated.text);
	}
	const shown = resolved.split(unknownComponent).join('…');
	const inside = state.roots.some((root) => isInside(resolved, root));
	const isRoot = state.roots.some((root) => resolved === root);
	if (!inside) {
		const because = evaluated.unguarded ? ` (with ${evaluated.unguarded} empty)` : '';
		problems.push(`${commandName} ${written}: resolves to ${shown}${because}, outside the repository and /tmp.`);
	} else if (isRoot && (!rootItselfAllowed || evaluated.unguarded)) {
		const because = evaluated.unguarded ? ` when ${evaluated.unguarded} is empty` : '';
		problems.push(`${commandName} ${written}: resolves to ${shown} itself${because}.`);
	}
}

function optionsEnd(argumentsList, index) {
	return argumentsList[index] && literalOf(argumentsList[index]) === '--';
}

function analyze(source, state, problems, depth) {
	if (depth > 8) {
		return;
	}
	const commands = lex(source);
	for (const command of commands) {
		// Substitutions run in a subshell: their own commands are checked, their assignments and cds stay inside.
		for (const word of command) {
			for (const part of word) {
				if (part.kind === 'Substitution') {
					analyze(part.nested, { ...state, variables: new Map(state.variables) }, problems, depth + 1);
				}
			}
		}

		let words = command;
		while (words.length > 0 && reservedWords.has(literalOf(words[0]) || '')) {
			words = words.slice(1);
		}
		const first = words.length > 0 ? literalOf(words[0]) : null;
		if (first === 'for' || first === 'case' || first === 'select') {
			continue;
		}

		// Leading NAME=value words. Alone they set shell variables for what follows; before a command
		// they set its environment only, and the command's own arguments still see the old values.
		const assignments = [];
		while (words.length > 0) {
			const head = words[0][0];
			if (head && head.kind === 'Literal' && !head.quoted && assignmentPattern.test(head.text)) {
				assignments.push(words[0]);
				words = words.slice(1);
			} else {
				break;
			}
		}
		if (words.length === 0) {
			for (const assignment of assignments) {
				applyAssignment(assignment, state);
			}
			continue;
		}

		// Peel wrappers (sudo rm, env X=1 rm, xargs rm) down to the command they run.
		let name = commandNameOf(words[0], state.variables);
		let argumentsList = words.slice(1);
		let guard = 0;
		while (name !== null && wrapperCommands.has(path.basename(name)) && guard++ < 6) {
			let next = 0;
			while (next < argumentsList.length) {
				const text = literalOf(argumentsList[next]);
				if (text === null) {
					break;
				}
				if (text.startsWith('-') || assignmentPattern.test(text) || (path.basename(name) === 'timeout' && /^[0-9.]+[smhd]?$/.test(text))) {
					next++;
					continue;
				}
				if (path.basename(name) === 'nice' && /^-?[0-9]+$/.test(text)) {
					next++;
					continue;
				}
				break;
			}
			if (next >= argumentsList.length) {
				name = null;
				break;
			}
			name = commandNameOf(argumentsList[next], state.variables);
			argumentsList = argumentsList.slice(next + 1);
		}
		if (name === null) {
			continue;
		}
		const base = path.basename(name);

		if (base === 'export' || base === 'readonly' || base === 'declare' || base === 'typeset' || base === 'local') {
			for (const word of argumentsList) {
				const head = word[0];
				if (head && head.kind === 'Literal' && assignmentPattern.test(head.text)) {
					applyAssignment(word, state);
				}
			}
			continue;
		}
		if (base === 'unset') {
			for (const word of argumentsList) {
				const text = literalOf(word);
				if (text && !text.startsWith('-')) {
					state.variables.set(text, '');
				}
			}
			continue;
		}
		if (base === 'cd' || base === 'pushd' || base === 'popd') {
			state.base = resolveDirectory(base, argumentsList, state);
			continue;
		}
		if (shells.has(base)) {
			const flagIndex = argumentsList.findIndex((word) => {
				const text = literalOf(word);
				return text !== null && /^-[a-z]*c[a-z]*$/.test(text);
			});
			if (flagIndex !== -1 && argumentsList[flagIndex + 1]) {
				analyze(nestedTextOfWord(argumentsList[flagIndex + 1]), { ...state, variables: new Map() }, problems, depth + 1);
			}
			continue;
		}
		if (base === 'eval') {
			analyze(argumentsList.map(nestedTextOfWord).join(' '), state, problems, depth + 1);
			continue;
		}

		if (removalCommands.has(base)) {
			let optionsDone = false;
			for (const word of argumentsList) {
				const text = literalOf(word);
				if (!optionsDone && text === '--') {
					optionsDone = true;
					continue;
				}
				if (!optionsDone && text !== null && text.startsWith('-') && text !== '-') {
					continue;
				}
				checkTarget(base, word, state, problems);
			}
			continue;
		}

		if (base === 'mv') {
			let optionsDone = false;
			for (let wordIndex = 0; wordIndex < argumentsList.length; wordIndex++) {
				const word = argumentsList[wordIndex];
				const text = literalOf(word);
				if (!optionsDone && optionsEnd(argumentsList, wordIndex)) {
					optionsDone = true;
					continue;
				}
				if (!optionsDone && text !== null && text.startsWith('-')) {
					if ((text === '-t' || text === '-S') && argumentsList[wordIndex + 1]) {
						if (text === '-t') {
							checkTarget('mv -t', argumentsList[wordIndex + 1], state, problems, true);
						}
						wordIndex++;
					} else if (text.startsWith('--target-directory=')) {
						checkTarget('mv', [{ kind: 'Literal', text: text.slice('--target-directory='.length), quoted: true }], state, problems, true);
					}
					continue;
				}
				checkTarget('mv', word, state, problems, true);
			}
			continue;
		}

		if (base === 'find') {
			const literals = argumentsList.map(literalOf);
			const deleting = literals.some(
				(text, textIndex) =>
					text === '-delete' ||
					((text === '-exec' || text === '-execdir' || text === '-ok' || text === '-okdir') && removalCommands.has(path.basename(literals[textIndex + 1] || ''))),
			);
			if (!deleting) {
				continue;
			}
			const roots = [];
			for (const word of argumentsList) {
				const text = literalOf(word);
				if (text !== null && (text.startsWith('-') || text === '(' || text === '!' || text === ')')) {
					break;
				}
				roots.push(word);
			}
			if (roots.length === 0) {
				roots.push([{ kind: 'Literal', text: '.', quoted: false }]);
			}
			for (const word of roots) {
				checkTarget('find', word, state, problems, true);
			}
		}
	}
}

function applyAssignment(word, state) {
	const head = word[0];
	const match = assignmentPattern.exec(head.text);
	const name = match[1];
	const valueWord = [{ ...head, text: head.text.slice(match[0].length) }, ...word.slice(1)];
	const evaluated = evaluateWord(valueWord, state.variables);
	// A value built from anything this hook cannot see is unknown, recorded as missing.
	if (evaluated.unguarded || evaluated.hasUnknown) {
		state.variables.delete(name);
	} else {
		state.variables.set(name, evaluated.text);
	}
}

function resolveDirectory(commandName, argumentsList, state) {
	if (commandName === 'popd') {
		return null;
	}
	const targets = argumentsList.filter((word) => {
		const text = literalOf(word);
		return !(text !== null && text.startsWith('-') && text !== '-');
	});
	if (targets.length === 0) {
		return process.env.HOME || null;
	}
	const evaluated = evaluateWord(targets[0], state.variables);
	if (evaluated.text === '-' || evaluated.unguarded || evaluated.hasUnknown) {
		return null;
	}
	if (path.isAbsolute(evaluated.text)) {
		return path.resolve(evaluated.text);
	}
	return state.base === null ? null : path.resolve(state.base, evaluated.text);
}

// Exported for the tests that prove each refusal; the hook itself runs below.
export function review(command, workingDirectory, repository) {
	const problems = [];
	const state = {
		variables: new Map(),
		base: workingDirectory,
		roots: allowedRoots(repository),
	};
	analyze(command, state, problems, 0);
	return problems;
}

function main() {
	let input;
	try {
		input = JSON.parse(readFileSync(0, 'utf8'));
	} catch {
		return;
	}
	const command = input && input.tool_input && input.tool_input.command;
	if (typeof command !== 'string') {
		return;
	}
	const workingDirectory = typeof input.cwd === 'string' ? input.cwd : process.cwd();
	const repository = process.env.CLAUDE_PROJECT_DIR || workingDirectory;
	let problems;
	try {
		problems = review(command, workingDirectory, repository);
	} catch {
		// A command this hook cannot read goes to Claude Code's own permission flow, unchanged.
		return;
	}
	if (problems.length === 0) {
		return;
	}
	const reason = ['Refused before running, so nothing was removed:', ...problems.slice(0, 5).map((problem) => '- ' + problem), fixAdvice].join('\n');
	process.stdout.write(
		JSON.stringify({
			hookSpecificOutput: {
				hookEventName: 'PreToolUse',
				permissionDecision: 'deny',
				permissionDecisionReason: reason,
			},
		}),
	);
}

// Run as the hook, not when the tests import it.
if (process.argv[1] && realpathSync(process.argv[1]) === realpathSync(fileURLToPath(import.meta.url))) {
	main();
}
