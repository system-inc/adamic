/*
 * Replay ESLint's own test rows for every core rule cohere registers, through ESLint's Linter, and
 * write each row with ESLint's verdict, so cohere's rules can be measured against the people who
 * defined them (#jjfa7qb).
 *
 *     node internal/lint/tools/eslint_corpus/extract.mjs --eslint-home <directory> \
 *         --out internal/lint/registry/testdata/eslint-corpus/eslint-<version>
 *
 * `--eslint-home` is any directory whose node_modules resolves `eslint` and `@typescript-eslint/parser`.
 * The tool reads the installed ESLint's version, fetches that tag's tests from GitHub once into
 * `--cache` (default: the system's temporary directory), and writes `<out>/<rule>.json` per rule.
 *
 * This is a regeneration tool, and nothing runs it. The rows it writes are committed, pinned to the
 * tag, and `TestCohereAgreesWithESLintsCoreCorpus` replays the committed rows: a measurement that had
 * to be re-fetched to re-run would be one whose source can vanish. Re-extracting for a new ESLint is
 * a deliberate act, and the test's row count says whether anything was lost on the way.
 *
 * # Why the authority's tests and not oxc's
 *
 * Most of the catalog was ported while oxlint was the gate, and the brief then was "where oxc and
 * another implementation disagree, oxc wins". oxc is a reimplementation, and where it differs from the
 * definition it may simply be wrong: a hook inside a `try` that provably cannot throw is reported by
 * oxc and not by ESLint, and ESLint is right. ESLint's corpus encodes what the rule's authors believe
 * it does, so replaying it finds every place a port inherited a reimplementation's answer.
 *
 * # How a row is run
 *
 * Every row runs as a module named Case.ts under the TypeScript parser, in both engines, so both see
 * the same text in the same mode. That costs the rows that only parse as a script or under another
 * parser, and each of those is written out as skipped with its reason rather than dropped: a corpus
 * that silently shrank would read as a rule that agreed.
 *
 * A test file is run with `RuleTester` replaced by a recorder, which is how its rows are read without
 * re-implementing the file's own construction of them: several build rows in loops or through
 * `unIndent`. Rows that configure ESLint globals are flagged, because cohere asks the checker what a
 * name is and has no counterpart for that configuration.
 */

import { execFileSync } from 'node:child_process';
import { existsSync, mkdirSync, readFileSync, realpathSync, writeFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import * as NodeOperatingSystem from 'node:os';
import * as NodePath from 'node:path';
import { parseArgs } from 'node:util';

const { values: argumentValues } = parseArgs({
    options: {
        'eslint-home': { type: 'string' },
        'out': { type: 'string' },
        'cache': { type: 'string', default: NodeOperatingSystem.tmpdir() },
        'rule': { type: 'string' },
        'cohere': { type: 'string', default: 'cohere' },
    },
});
if(!argumentValues['eslint-home'] || !argumentValues.out) {
    throw new Error('name --eslint-home (a directory that resolves eslint) and --out (where the rows go)');
}

// Resolved from ESLint's real location, so its own dependencies resolve the way they do for ESLint under pnpm
const homeRequire = createRequire(NodePath.join(NodePath.resolve(argumentValues['eslint-home']), 'package.json'));
const eslintDirectory = realpathSync(NodePath.dirname(homeRequire.resolve('eslint/package.json')));
const eslintRequire = createRequire(NodePath.join(eslintDirectory, 'package.json'));
const { Linter } = eslintRequire('eslint');
const eslintVersion = eslintRequire('./package.json').version;
const typeScriptParser = homeRequire('@typescript-eslint/parser');

// ESLint's tests are not in the npm package, so the tag's source is fetched once and kept
const sourceDirectory = NodePath.join(argumentValues.cache, `eslint-${eslintVersion}`);
const rulesDirectory = NodePath.join(sourceDirectory, 'tests', 'lib', 'rules');
if(!existsSync(rulesDirectory)) {
    mkdirSync(argumentValues.cache, { recursive: true });
    const archive = NodePath.join(argumentValues.cache, `eslint-${eslintVersion}.tar.gz`);
    execFileSync('curl', ['-sSfL', '-o', archive, `https://github.com/eslint/eslint/archive/refs/tags/v${eslintVersion}.tar.gz`]);
    execFileSync('tar', ['xzf', archive, '-C', argumentValues.cache, `eslint-${eslintVersion}/tests`]);
}

// The core rules cohere registers are the ones with no plugin prefix
const ruleNames = execFileSync(argumentValues.cohere, ['--rules'], { encoding: 'utf8' })
    .split('\n')
    .filter((name) => name !== '' && !name.includes('/'));

// unIndent from ESLint's tests/_utils, which the test files build rows with
function unIndent(strings, ...values) {
    const text = strings.map((part, index) => (index === 0 ? part : values[index - 1] + part)).join('');
    const lines = text.replace(/^\n/u, '').replace(/\n\s*$/u, '').split('\n');
    const indents = lines.filter((line) => line.trim()).map((line) => line.match(/ */u)[0].length);
    const minimumIndent = Math.min(...indents);
    return lines.map((line) => line.slice(minimumIndent)).join('\n');
}

// Runs one test file with RuleTester recording rather than testing, and returns what it recorded
function recordedRuns(ruleName) {
    const testPath = NodePath.join(rulesDirectory, `${ruleName}.js`);
    const runs = [];
    class RuleTester {
        constructor(configuration) {
            this.configuration = configuration ?? {};
        }
        run(_name, _rule, tests) {
            runs.push({ base: this.configuration, tests });
        }
        static setDefaultConfig() {}
        static only(item) {
            return item;
        }
    }
    const globalsStandIn = new Proxy({}, { get: () => ({}) });
    const shimmedRequire = function(name) {
        if(typeof name !== 'string') return {};
        // Required both as the module itself and as its named export, so the class carries itself
        if(name.includes('rule-tester')) return Object.assign(RuleTester, { RuleTester });
        if(name.includes('lib/rules/')) return {};
        if(name.endsWith('_utils')) return { unIndent };
        // A fixture parser stands in for the parser it names, so its rows are skipped as a custom parser
        if(name.includes('fixtures/parsers') || name.includes('fixture-parser')) return () => ({ fixtureParser: true });
        if(name === '@typescript-eslint/parser') return typeScriptParser;
        if(name === 'globals') return globalsStandIn;
        if(name === 'eslint-scope') return eslintRequire('eslint-scope');
        if(name.startsWith('node:') || ['assert', 'fs', 'path'].includes(name)) return eslintRequire(name);
        throw new Error(`${ruleName}.js requires ${name}, which this tool does not provide`);
    };
    new Function('require', '__dirname', '__filename', 'module', 'exports', readFileSync(testPath, 'utf8'))(
        shimmedRequire,
        rulesDirectory,
        testPath,
        { exports: {} },
        {},
    );
    return runs;
}

// ESLint's verdict on one row, as each finding's message id and source range
function eslintVerdict(ruleName, row, base) {
    const languageOptions = { ...(base.languageOptions ?? {}), ...(row.languageOptions ?? {}) };
    if(languageOptions.parser && languageOptions.parser !== typeScriptParser) {
        return { skipped: 'a custom parser' };
    }
    let messages;
    try {
        messages = new Linter({ configType: 'flat' }).verify(
            row.code,
            [
                {
                    files: ['**/*.ts'],
                    languageOptions: {
                        ecmaVersion: 'latest',
                        globals: languageOptions.globals ?? {},
                        parser: typeScriptParser,
                        parserOptions: languageOptions.parserOptions ?? {},
                        sourceType: 'module',
                    },
                    rules: { [ruleName]: ['error', ...(row.options ?? [])] },
                },
            ],
            { filename: 'Case.ts' },
        );
    }
    catch(error) {
        return { skipped: `ESLint threw: ${String(error.message).split('\n')[0]}` };
    }
    const fatal = messages.find((message) => !message.ruleId);
    if(fatal) return { skipped: `no parse under the TypeScript parser: ${fatal.message}` };

    // Lines as ESLint counts them, which ends one at \r, \u2028 and \u2029 as well as \n
    const lineStarts = [0];
    for(const lineBreak of row.code.matchAll(/\r\n|[\r\n\u2028\u2029]/gu)) lineStarts.push(lineBreak.index + lineBreak[0].length);
    const offsetOf = (line, column) => lineStarts[line - 1] + column - 1;
    // Only the row's own rule. A row may turn another rule on with an inline `/*eslint ...*/` comment,
    // or name one ESLint cannot find, and either puts that rule's messages beside this one's: three of
    // prefer-const's rows recorded no-undef-init's finding and a missing-definition error as its own.
    return {
        findings: messages.filter((message) => message.ruleId === ruleName).map(function(message) {
            const start = offsetOf(message.line, message.column);
            const end = message.endLine ? offsetOf(message.endLine, message.endColumn) : start;
            return { messageId: message.messageId ?? null, start, end, text: row.code.slice(start, end) };
        }),
    };
}

mkdirSync(argumentValues.out, { recursive: true });
const summary = { eslintVersion, rules: 0, rows: 0, replayed: 0, skipped: 0, unreadable: [] };
for(const ruleName of ruleNames) {
    if(argumentValues.rule && ruleName !== argumentValues.rule) continue;
    if(!existsSync(NodePath.join(rulesDirectory, `${ruleName}.js`))) {
        summary.unreadable.push(`${ruleName}: ESLint ${eslintVersion} has no test file for it`);
        continue;
    }
    let runs;
    try {
        runs = recordedRuns(ruleName);
    }
    catch(error) {
        summary.unreadable.push(`${ruleName}: ${error.message}`);
        continue;
    }
    // One row per line, so a re-extract reads as a diff of the rows that changed
    const lines = [];
    for(const { base, tests } of runs) {
        for(const kind of ['valid', 'invalid']) {
            for(const item of tests[kind] ?? []) {
                const row = typeof item === 'string' ? { code: item } : item;
                const verdict = eslintVerdict(ruleName, row, base);
                const vendored = { code: row.code };
                if(row.options?.length) vendored.options = row.options;
                if(row.languageOptions?.globals || base.languageOptions?.globals) vendored.usesGlobals = true;
                if(verdict.skipped) {
                    vendored.skipped = verdict.skipped;
                    summary.skipped += 1;
                }
                else {
                    vendored.eslint = verdict.findings.map((finding) => [finding.messageId, finding.start, finding.end]);
                    summary.replayed += 1;
                }
                // Escaped to ASCII, because several unicode rules test lone surrogates on purpose
                lines.push(JSON.stringify(vendored).replace(/[\u007f-\uffff]/gu, (character) => '\\u' + character.charCodeAt(0).toString(16).padStart(4, '0')));
            }
        }
    }
    writeFileSync(NodePath.join(argumentValues.out, `${ruleName}.json`), '[\n' + lines.join(',\n') + '\n]\n');
    summary.rules += 1;
    summary.rows += lines.length;
}
console.log(JSON.stringify(summary, null, 1));
if(summary.unreadable.length > 0) process.exitCode = 1;
