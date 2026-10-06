import assert from 'node:assert/strict';
import { mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawnSync } from 'node:child_process';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

const loader = fileURLToPath(new URL('./register-dot-a.mjs', import.meta.url));
const runner = fileURLToPath(new URL('./node.mjs', import.meta.url));
const directory = mkdtempSync(join(tmpdir(), 'dot-a-loader-'));

function run(entry, wrapper = false, args = []) {
	return spawnSync(process.execPath, ['--disable-warning=ExperimentalWarning',
		...(wrapper ? [runner] : ['--import', loader]), entry, ...args], { encoding: 'utf8' });
}

test('source modules, transform mode, runtime and argv agree through both entry paths', () => {
	writeFileSync(join(directory, 'dep.a'), 'export enum Value { Answer = 42 }\n');
	const entry = join(directory, 'main.a');
	writeFileSync(entry, "import { Value } from './dep.a';\nimport { programArguments } from 'adamic';\nconsole.log(`${Value.Answer}:${programArguments().join(',')}`);\n");
	for (const wrapper of [false, true]) {
		const result = run(entry, wrapper, ['one', 'two']);
		assert.equal(result.status, 0, result.stderr);
		assert.equal(result.stdout, '42:one,two\n');
		assert.equal(result.stderr, '');
	}
});

test('runtime panic and uncaught source errors keep the oracle exit contract', () => {
	for (const source of ["import { panic } from 'adamic'; panic('probe');", "throw new Error('probe');"]) {
		const entry = join(directory, 'panic.a');
		writeFileSync(entry, source);
		for (const wrapper of [false, true]) {
			const result = run(entry, wrapper);
			assert.equal(result.status, 70);
			assert.equal(result.stdout, '');
			assert.match(result.stderr, /^adamic: panic: (Error: )?probe\n$/);
		}
	}
});
