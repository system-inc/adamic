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
// These pinned terminal-check witnesses retain main's exit-70 oracle convention.
// Other programs keep Node's ordinary exception behavior. A changed source hash
// cannot borrow an old fixture's terminal convention.
const terminalFixtures = new Map([
	["docs/step-18/fixtures/runtime-chain-parentheses-write.a", "471acb50ef53b7d4735ed81759f5ad46b759512e3fba578044ea59d695ca9198"],
	["stage3/parser-next/stack/pathological.a", "9bb31080be12884f64e706471d0aeb729ee0194812fce8884202da06ce83ca53"],
	["internal/oracle/testdata/concat_too_long.a", "0accc878d2c2a8ed30641a131c3606e9932ca7326ba7e64d14d9a7e9dc69ccce"],
	["internal/oracle/testdata/dead_zone.a", "fad9a712c540773834ed86ff70b7bbf77eee136d4f41f6b2ca83ced94101dd1e"],
	["internal/oracle/testdata/from_code_point_fails.a", "9c194ad095ae355d222c25bfbe1b08a2437b8ec7c6e9ce4210b780ce15ce8dd3"],
	["internal/oracle/testdata/import_cycles/classes/c.a", "90eeb9fc11ad95eaeabdf179c3f36e141861388a028eb3160f7d157fcbea1371"],
	["internal/oracle/testdata/import_cycles/indirect/b.a", "a086b304b88183100951fa73c7c60d32ddff1655e28d865ff470f28d0c11535c"],
	["internal/oracle/testdata/import_cycles/read/a.a", "fe84c25233f0184fe20aa879969acce913df15d5c9985f4747071df8e0ba54f1"],
	["internal/oracle/testdata/library_method_values_dead_zone.a", "4c3028458a44d355238b7772d1324bf7d393887c3002549f3110257355ec1fcd"],
	["internal/oracle/testdata/method_coverage_empty_alias_tdz.a", "41b465cdaff3767058189471e6ba7d370775034a2d08237d950c6a635cc193a2"],
	["internal/oracle/testdata/narrowed_writes.a", "4e8450502bf85e23b563f115090a66b58acc5450984bc32a890b2b5c244c2057"],
	["internal/oracle/testdata/nested_destructured_tdz.a", "ea3747a0715235f89bbfcaf9d1a48297543ab1c73e497ef2d5d327c5596b02a6"],
	["internal/oracle/testdata/nested_tdz.a", "d45a1ac448116e7316b275435f72f8908630acd30e39cec34c492b22ee6e0602"],
	["internal/oracle/testdata/nested_tdz_write.a", "1a2a8f23b16b520489ca2efbeef51a0c57a9fe1aa465b9c1273291054876931c"],
	["internal/oracle/testdata/normalize_coverage_limit.a", "8ee1bfd50ca3a69d9ac986228b936fc94d06e15ac7074d7c62ab59b0b327711b"],
	["internal/oracle/testdata/pad_too_long.a", "5d63c594506879cd0dccef2e4955786078f0c2a9247b41c47ff572c7bc97354e"],
	["internal/oracle/testdata/stack_forever.a", "14e06789596f7b3f083cc4a9d1ee06ba978bbae01aaf2fa42e80e06d969a0132"],
	["internal/oracle/testdata/stack_over.a", "88f7d5db0b76ae545c7b8e5193085b20046debac8d19c1db5f4ba0398ec818c1"],
	["internal/oracle/testdata/stack_overflow.a", "4bfff1246fb5b82a39134640757d1c293c40cccf4c4f4a596c028486d9dd0fff"],
	["internal/oracle/testdata/stack_tail_call.a", "b1d7a3b550252b1bd4a07670a449453a7e578b08c18a6c2cbbfe838e6390c92a"],
	["internal/oracle/testdata/switch_case_declarations/dead_zone.a", "31fdad1d767b03db9f35de8a37b7e2cad07380ea891414dcf161aed292ade846"],
	["internal/oracle/testdata/switch_case_declarations/dead_zone_direct.a", "5374d0f60734227442ae4f1b74959be241ea1cb48ed0d1e68ccec0b7be30648b"],
	["internal/oracle/testdata/switch_case_declarations/dead_zone_initializer.a", "e443cd47c87aa87c55ae62296c41f64acaa83949c7e044742e50651174f23348"],
	["internal/oracle/testdata/switch_case_declarations/dead_zone_write.a", "7e2725bd001bc715ff58718d757e98b8489831ef2b8537f134fe47e5b875690f"],
	["internal/oracle/testdata/user_iterators_rest_tdz.a", "45b67f43e19aaca5198329e6ab119dbc7b6ba772173cf831006176e8fa9e9eea"],
	["stage3/drivers/scanner/probes/cyclic-premature-value/main.a", "cedfe65a05ae7762b0abcda757f81754d3bd01d02c9fb11af3abf89f5b7c3716"],
	["stage3/fixtures/cycles/06_import_order_mutant/main.a", "b6ba97a6d97e9bb9b6edbafc67979d088067bfcd7efb0c0e9d1b900ed8104e25"],
	["stage3/fixtures/records/15_inherited_read.a", "d52f251873b7acc0c8972818c02bad0663aeb5f2f5dd258dbb58d6d501687c26"],
]);
// Module-cycle witnesses also pin every source in their static import graph.
const terminalDependencies = new Map([
	['internal/oracle/testdata/import_cycles/classes/c.a', [
		['internal/oracle/testdata/import_cycles/classes/a.a', 'ca84122d173d4dcce912fce946c17f759f3962447e4108189e0f3ff269642e45'],
		['internal/oracle/testdata/import_cycles/classes/b.a', 'abbd6bba6f30486ae882b01eda15da56746f5f7db2b3dadb018f074e6415a121'],
	]],
	['internal/oracle/testdata/import_cycles/indirect/b.a', [
		['internal/oracle/testdata/import_cycles/indirect/a.a', '4d7a295787258563654d497394d33a5d5bb29fe76b4213a599ae4f6171bf66c4'],
	]],
	['internal/oracle/testdata/import_cycles/read/a.a', [
		['internal/oracle/testdata/import_cycles/read/b.a', 'dc8cba54a484fb0b0ad857c436e76e72da2cc0dc13fb798b059555c3705451be'],
	]],
	['stage3/drivers/scanner/probes/cyclic-premature-value/main.a', [
		['stage3/drivers/scanner/probes/cyclic-premature-value/barrel.a', '5a9d36265363f3955ee06f00db3fbbb193e03860129db8499e38e5922d9dfe84'],
		['stage3/drivers/scanner/probes/cyclic-premature-value/reader.a', '632b67e47160f12d2485665fe2206313012f11dc76e8f67a08e57ea9e826b180'],
		['stage3/drivers/scanner/probes/cyclic-premature-value/types.a', '4a5bcfaa2210d096dc08b2f001529abf3a4f274e40ac9c3703009c1be794dc5f'],
	]],
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
