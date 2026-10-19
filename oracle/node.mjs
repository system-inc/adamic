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
	// Main added these terminal placeholder and async-rejection witnesses.
	["internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_element_plain.a", "1eb94d37614330494f21b0d61e405183d067c9de231941e780b7356f1816c25c"],
	["internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_element.a", "37e4e5c76731a999d9153d9435be8299a0486a45d2c173fe306486085a1d60a8"],
	["internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_element_method.a", "6f32b516bef7bd318bebb855eff6a8186e16f71a8d96acdd421831499006588a"],
	["internal/oracle/testdata/async_coverage_reject_nested.a", "10fa813d2cb2ec049cf3b55944492e3a2db9e890a866486b9b3a5e3cdaea4262"],
	["internal/oracle/testdata/namespaces_unknown_write_before.a", "32d4a539c31220adb40189c4a4e1cb0d0f4e4eff5228539f7567c72e03538a52"],
	["internal/oracle/testdata/element_access_presence.a", "3056e90354f2749fb53b5fb5ea34c701b1bb4f524b8e677bf6f6005cf00f94f3"],
	["internal/oracle/testdata/namespaces_unknown_void_before.a", "79a5b5970dfaf5e59ab89c34066973391d0725625fa61bc12c577b6fe2106d26"],
	["internal/oracle/testdata/async_coverage_reject_eager.a", "bc4f5ce41f6731e5e3ffa96ed16047bf22c7314b3ee5119e4c989067a1a3f111"],
	["internal/oracle/testdata/async_coverage_reject_empty.a", "7f68067134db38d0405c91ee6fcc620d5eccdce69d786efe76aeb3a64054e9ad"],
	["internal/oracle/testdata/namespaces_unknown_enum_before.a", "32388d9f7258d18f4b557bf6ae69fffc4f20a17b8b5c124c7e8dd089666a0892"],
	["internal/oracle/testdata/namespaces_unknown_before.a", "4f1be7b872b6d8c72ab38b60a6241ed370a09e77a02fb486b37e0603b657653a"],
	["internal/oracle/testdata/async_throw.a", "6c3e00e7233b1131d43ca9e746748fc2b9f035daf6560a37555d03f81c8386fd"],
	["internal/oracle/testdata/namespaces_unknown_function_before.a", "78153a5b9e1f38ef774091542042ba449691052c55984e0959ac066e47394b3d"],
	["internal/oracle/testdata/047cb0d_n_arrayindex.a", "b7e22bf1e20ff18644490b57789c40afc99b9ad938e63d276421f529e8c63dac"],
	["internal/oracle/testdata/047cb0d_n_element.a", "37e4e5c76731a999d9153d9435be8299a0486a45d2c173fe306486085a1d60a8"],
	["internal/oracle/testdata/047cb0d_n_element_method.a", "6f32b516bef7bd318bebb855eff6a8186e16f71a8d96acdd421831499006588a"],
	["internal/oracle/testdata/047cb0d_n_element_plain.a", "1eb94d37614330494f21b0d61e405183d067c9de231941e780b7356f1816c25c"],
	["internal/oracle/testdata/concat_too_long.a", "0accc878d2c2a8ed30641a131c3606e9932ca7326ba7e64d14d9a7e9dc69ccce"],
	["internal/oracle/testdata/from_code_point_fails.a", "9c194ad095ae355d222c25bfbe1b08a2437b8ec7c6e9ce4210b780ce15ce8dd3"],

	["internal/oracle/testdata/narrowed_fields.a", "d072173334da8535c54e5afa482bc25680a23729fa4abc52d355b3c66cf80ab1"],
	["internal/oracle/testdata/narrowed_methods.a", "59741aecd02b4fafd8aae1fdb0d329a703b97d520840a8b527d92e7c4e1f127f"],
	["internal/oracle/testdata/narrowed_reads.a", "d689a2ac70535e33bd9f14e943a57577647594e74806d92d9934c3bc7feb802c"],

	["internal/oracle/testdata/narrowed_writes.a", "4e8450502bf85e23b563f115090a66b58acc5450984bc32a890b2b5c244c2057"],
	["internal/oracle/testdata/normalize_coverage_limit.a", "8ee1bfd50ca3a69d9ac986228b936fc94d06e15ac7074d7c62ab59b0b327711b"],
	["internal/oracle/testdata/pad_too_long.a", "5d63c594506879cd0dccef2e4955786078f0c2a9247b41c47ff572c7bc97354e"],
	["internal/oracle/testdata/regexp_null_narrowed.a", "691a4afae5803416ec4bd0b8fb142c5e3fad934f0060d4b49df1035f75c7f846"],
	["internal/oracle/testdata/reuse_narrowed.a", "85c888600818c6fb1c9aa0442ad1a4e68db9bf969941a107558e763b49e687b1"],
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
