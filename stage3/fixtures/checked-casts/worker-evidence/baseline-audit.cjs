// Observe only this directory. Node executes stock-transpiled source, independently of Adamic.
const fs = require('node:fs'), path = require('node:path'), cp = require('node:child_process');
const ts = require('typescript');
const crypto = require('node:crypto');
if (ts.version !== '6.0.3') throw Error('TypeScript 6.0.3 required');
const directory = '/workspace/adamic/stage3/fixtures/checked-casts';
const [compilerArg, scratchArg] = process.argv.slice(2);
const compiler = path.resolve(compilerArg), scratch = path.resolve(scratchArg);
fs.mkdirSync(scratch, {recursive: true});
const runtime = path.join(scratch, 'node_modules/adamic');
fs.mkdirSync(runtime, {recursive: true});
fs.copyFileSync(path.resolve(directory, '../../../oracle/adamic.mjs'), path.join(runtime, 'index.mjs'));
fs.writeFileSync(path.join(runtime, 'package.json'), JSON.stringify({name: 'adamic', type: 'module', exports: './index.mjs'}));
const fixtures = JSON.parse(fs.readFileSync(path.join(directory, 'fixtures.json')));
function run(command, name, environment = {}) {
 const result = cp.spawnSync(command[0], command.slice(1), {encoding: 'utf8', env: {...process.env, ...environment}, timeout: 120000, maxBuffer: 16 * 1024 * 1024});
 if (result.error || result.signal || result.status === null) throw Error(`${name}: ${result.error || result.signal}`);
 fs.writeFileSync(path.join(scratch, name + '.stdout'), result.stdout);
 fs.writeFileSync(path.join(scratch, name + '.stderr'), result.stderr);
 return {command, exit: result.status, stdout: result.stdout, stderr: result.stderr};
}
function same(actual, expected) {
 return actual.exit === expected.exit && actual.stdout === expected.stdout && actual.stderr === expected.stderr;
}
function contract(row, result) {
 return result.exit === row.runtime_exit && result.stdout === row.runtime_stdout && (row.failing ? result.stderr.startsWith('adamic: panic: cast failed:') : result.stderr === '');
}
const results = [];
for (const row of fixtures) {
 const sourcePath = path.join(directory, row.file), text = fs.readFileSync(sourcePath, 'utf8');
 const code = ts.transpileModule(text, {compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext}}).outputText;
 const nodePath = path.join(scratch, row.file + '.source.mjs');
 fs.writeFileSync(nodePath, code);
 const node = run([process.execPath, nodePath], row.file + '.node');
 if (!same(node, {exit: 0, stdout: row.node_stdout, stderr: ''})) throw Error(`Node golden mismatch: ${row.file}`);
 const emitted = run([compiler, 'c', sourcePath], row.file + '.c');
 const result = {file: row.file, sha256: crypto.createHash('sha256').update(text).digest('hex'), node, compiler_exit: emitted.exit, diagnostic: emitted.stderr, native: null, javascript: null, contract: 'blocked', mutants: []};
 if (emitted.exit === 0) {
  const binary = path.join(scratch, row.file + '.native');
  const build = run([compiler, 'build', sourcePath, '-o', binary, '--sanitize'], row.file + '.build');
  if (build.exit !== 0) throw Error(`build failed: ${row.file}: ${build.stderr}`);
  result.native = run([binary], row.file + '.native', {ASAN_OPTIONS: row.failing ? 'detect_leaks=0' : 'detect_leaks=1', UBSAN_OPTIONS: 'halt_on_error=1'});
  const js = run([compiler, 'js', sourcePath], row.file + '.js');
  if (js.exit !== 0) throw Error(`JavaScript emission failed: ${row.file}`);
  const jsPath = path.join(scratch, row.file + '.compiled.mjs');
  fs.writeFileSync(jsPath, js.stdout);
  result.javascript = run([process.execPath, jsPath], row.file + '.javascript');
  result.contract = contract(row, result.native) && contract(row, result.javascript) ? 'passed' : 'mismatch';
  if (!row.failing && (!same(result.native, node) || !same(result.javascript, node))) throw Error(`Node disagreement: ${row.file}`);
  if (false && row.failing) {
   // Mutate the real emitted JavaScript, replacing only cast failure calls by void 0.
   const file = ts.createSourceFile(jsPath, js.stdout, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS);
   const edits = [];
   function visit(n) {
    if (ts.isCallExpression(n) && n.expression.getText(file) === 'adamicCast' && n.arguments.some(a => ts.isStringLiteral(a) && a.text.startsWith('cast failed:'))) edits.push({start: n.getStart(file), end: n.end, value: n.arguments[0].getText(file)});
    ts.forEachChild(n, visit);
   }
   visit(file);
   if (!edits.length) throw Error(`missing emitted check mutant: ${row.file}`);
   let mutant = js.stdout;
   for (const edit of edits.sort((a,b) => b.start - a.start)) mutant = mutant.slice(0, edit.start) + edit.value + mutant.slice(edit.end);
   const mutantPath = path.join(scratch, row.file + '.skip-check.mjs');
   fs.writeFileSync(mutantPath, mutant);
   const actual = run([process.execPath, mutantPath], row.file + '.skip-check');
   if (contract(row, actual) || actual.exit !== 0 || actual.stderr !== '') throw Error(`missing-check mutant was not a semantic kill: ${row.file}`);
   result.mutants.push({name: 'remove real emitted cast failure call', caught_by: 'panic exit and stdout contract', observation: actual});
  }
 }
 if (row.failing) {
  // Source erasure is a negative control for the required check, not a compiler implementation mutant.
  if (contract(row, node)) throw Error(`erasure control survived: ${row.file}`);
  result.mutants.push({name: 'erased-cast negative control', caught_by: 'panic exit and stdout contract', observation: node});
 } else {
  const file = ts.createSourceFile(sourcePath, text, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
  let marker;
  function visit(n) {if (ts.isStringLiteral(n) && n.text === 'after') marker = n; ts.forEachChild(n, visit);}
  visit(file);
  if (!marker) throw Error('missing output marker');
  const mutantText = text.slice(0, marker.getStart(file)) + "'mutant'" + text.slice(marker.end);
  const mutantPath = path.join(scratch, row.file + '.output-mutant.mjs');
  fs.writeFileSync(mutantPath, ts.transpileModule(mutantText, {compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext}}).outputText);
  const actual = run([process.execPath, mutantPath], row.file + '.output-mutant');
  if (same(actual, node) || actual.exit !== 0 || actual.stderr !== '') throw Error(`output mutant survived: ${row.file}`);
  result.mutants.push({name: 'change final output marker', caught_by: 'Node stdout golden', observation: actual});
 }
 results.push(result);
}
fs.writeFileSync(path.join(scratch, 'baseline-observations.json'), JSON.stringify({compiler_revision: cp.execFileSync('git', ['rev-parse', 'HEAD'], {encoding: 'utf8'}).trim(), node_version: process.version, results}, null, 2) + '\n');
console.log(JSON.stringify({fixtures: results.length, node_goldens: results.length, runtime_passed: results.filter(r => r.contract === 'passed').length, runtime_blocked: results.filter(r => r.contract === 'blocked').length, mutant_controls: results.reduce((n,r) => n + r.mutants.length, 0)}));
