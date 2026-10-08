// Generated .ts files are temporary CLI inputs, never repository programs.
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const cp = require('node:child_process');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw Error(`expected 6.0.3, got ${ts.version}`);
const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'ts2488-'));
const source = fs.readFileSync(path.join(__dirname, 'repeated-destructure.a'), 'utf8');
const cases = [
 ['repeated', source, [0, 1, 1]],
 ['unsuppressed', source.replace('// @ts-expect-error The first destructuring diagnoses the undefined tuple.\n', ''), [1, 2, 2]],
 ['single', source.split('\n').filter((_, index) => index !== 2 && index !== 3).join('\n'), [1, 1, 1]],
 ['iterable', source.replace(' | undefined', '').replace('// @ts-expect-error The first destructuring diagnoses the undefined tuple.\n', ''), [0, 0, 0]],
 ['wrong-assignment', source.replace('// @ts-expect-error The first destructuring diagnoses the undefined tuple.\n', '').replace(' | undefined', '') + '\nconst wrong: number = "wrong";\n', [1, 1, 1]],
];
const flags = ['--pretty', 'false', '--ignoreConfig', '--strict', '--noUncheckedIndexedAccess', '--exactOptionalPropertyTypes', '--noEmit', '--target', 'es2024', '--lib', 'es2024', '--module', 'esnext', '--moduleResolution', 'bundler', '--moduleDetection', 'force', '--verbatimModuleSyntax', '--allowImportingTsExtensions'];
for (const [name, text, expected] of cases) {
 const file = path.join(directory, name + '.ts');
 fs.writeFileSync(file, text);
 const programs = [
  ['tsc', process.execPath, [require.resolve('typescript/bin/tsc'), ...flags, file]],
  ['typescript-go', process.env.TS2488_TSGO, [...flags, file]],
  ['adamic', process.env.TS2488_LOAD, [file]],
 ];
 programs.forEach(([checker, command, args], index) => {
  if (!command) throw Error(`missing command for ${checker}`);
  const result = cp.spawnSync(command, args, { encoding: 'utf8' });
  if (result.error) throw result.error;
  const output = result.stdout + result.stderr;
  const codes = [...output.matchAll(/error TS(\d+):/g)].map(match => Number(match[1]));
  console.log(JSON.stringify({ name, checker, exit: result.status, output, codes }));
  const code = name === 'wrong-assignment' ? 2322 : 2488;
  if (codes.length !== expected[index] || codes.some(value => value !== code) || result.status !== (expected[index] ? 1 : 0) && !(checker !== 'adamic' && expected[index] && result.status === 2)) throw Error(`unexpected ${checker} result for ${name}`);
 });
}
console.log('PASS: 15 checker observations; repeated, single, iterable and assignment controls');
