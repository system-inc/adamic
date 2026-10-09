// node.mjs: run an Adamic program on Node, the oracle every native binary is compared against.
//
//	node --disable-warning=ExperimentalWarning oracle/node.mjs main.a [arguments...]
//
// It runs the source itself, with its types stripped, and never anything Adamic lowered: an oracle
// built from Adamic's own output would agree with Adamic's bugs. Both .ts and .a load the same way,
// and 'adamic' resolves to the runtime beside this file.
import { readFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { relative, resolve, sep } from 'node:path';
import { registerHooks, stripTypeScriptTypes } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';

const runtimeUrl = new URL('./adamic.mjs', import.meta.url).href;

registerHooks({
	resolve(specifier, context, nextResolve) {
		if (specifier === 'adamic') {
			return { url: runtimeUrl, shortCircuit: true };
		}
		return nextResolve(specifier, context);
	},
	load(url, context, nextLoad) {
		if (url.endsWith('.a') || url.endsWith('.ts')) {
			const source = readFileSync(fileURLToPath(url), 'utf8');
			return { format: 'module', source: stripTypeScriptTypes(source, { mode: 'transform' }), shortCircuit: true };
		}
		return nextLoad(url, context);
	},
});

// The program sees argv as if Node ran it directly, [node, program, ...arguments]: this file takes its
// own place out, so programArguments() is process.argv.slice(2) here too.
// Language dead-zone reads and writes now throw catchable ReferenceError and
// are removed from this terminal list by lowering-chain-fixes. Placeholder and
// remaining terminal-check witnesses retain their exit-70 oracle convention.
// Other programs keep Node's ordinary exception behavior. A changed source hash
// cannot borrow an old fixture's terminal convention.
const terminalFixtures = new Map([
	["docs/step-18/fixtures/runtime-chain-parentheses-write.a", "471acb50ef53b7d4735ed81759f5ad46b759512e3fba578044ea59d695ca9198"],
	["stage3/parser-next/stack/pathological.a", "9bb31080be12884f64e706471d0aeb729ee0194812fce8884202da06ce83ca53"],
	["internal/oracle/testdata/concat_too_long.a", "0accc878d2c2a8ed30641a131c3606e9932ca7326ba7e64d14d9a7e9dc69ccce"],
	["internal/oracle/testdata/from_code_point_fails.a", "9c194ad095ae355d222c25bfbe1b08a2437b8ec7c6e9ce4210b780ce15ce8dd3"],
	["internal/oracle/testdata/narrowed_writes.a", "4e8450502bf85e23b563f115090a66b58acc5450984bc32a890b2b5c244c2057"],
	["internal/oracle/testdata/normalize_coverage_limit.a", "8ee1bfd50ca3a69d9ac986228b936fc94d06e15ac7074d7c62ab59b0b327711b"],
	["internal/oracle/testdata/pad_too_long.a", "5d63c594506879cd0dccef2e4955786078f0c2a9247b41c47ff572c7bc97354e"],
	["internal/oracle/testdata/stack_forever.a", "14e06789596f7b3f083cc4a9d1ee06ba978bbae01aaf2fa42e80e06d969a0132"],
	["internal/oracle/testdata/stack_over.a", "88f7d5db0b76ae545c7b8e5193085b20046debac8d19c1db5f4ba0398ec818c1"],
	["internal/oracle/testdata/stack_overflow.a", "4bfff1246fb5b82a39134640757d1c293c40cccf4c4f4a596c028486d9dd0fff"],
	["internal/oracle/testdata/stack_tail_call.a", "b1d7a3b550252b1bd4a07670a449453a7e578b08c18a6c2cbbfe838e6390c92a"],
	["stage3/fixtures/cycles/06_import_order_mutant/main.a", "b6ba97a6d97e9bb9b6edbafc67979d088067bfcd7efb0c0e9d1b900ed8104e25"],
	["stage3/fixtures/records/15_inherited_read.a", "d52f251873b7acc0c8972818c02bad0663aeb5f2f5dd258dbb58d6d501687c26"],
]);
// Module-cycle witnesses also pin every source in their static import graph.
const terminalDependencies = new Map([
	['stage3/fixtures/cycles/06_import_order_mutant/main.a', [
		['stage3/fixtures/cycles/06_import_order_mutant/_namespaces/ts.a', 'afca00bbafae39f1033b08f1f724dcc6529c1ce9e2405710bc4ad41bf68e769d'],
		['stage3/fixtures/cycles/06_import_order_mutant/core.a', 'beb7f28dfaf869148319009e9843e2e5520d4d297bd9adecd48fe3cb55568c17'],
		['stage3/fixtures/cycles/06_import_order_mutant/utilities.a', 'fd0f58af403606d661bbc466f927854172353b7383575350204e7c8ac13946e3'],
	]],
]);
let terminalSource;
let forceTerminal = false;
if (process.argv[2] === '--terminal-check') {
	forceTerminal = true;
	process.argv.splice(2, 1);
} else if (process.argv[2] === '--terminal-source') {
	terminalSource = process.argv[3];
	process.argv.splice(2, 2);
}
const program = process.argv[2];
terminalSource ??= program;
const repository = fileURLToPath(new URL('../', runtimeUrl));
const key = relative(repository, resolve(terminalSource)).split(sep).join('/');
const expectedHash = terminalFixtures.get(key);
const matchesHash = (source, expectedHash) => createHash('sha256').update(readFileSync(source)).digest('hex') === expectedHash;
const terminalOracle = forceTerminal || expectedHash !== undefined && matchesHash(terminalSource, expectedHash) && (terminalDependencies.get(key) ?? []).every(([source, hash]) => matchesHash(new URL(source, new URL('../', runtimeUrl)), hash));
process.argv.splice(1, 1);

const runtime = await import(runtimeUrl);
if (terminalOracle) runtime.terminalCheckOracle();
await import(pathToFileURL(program).href);
