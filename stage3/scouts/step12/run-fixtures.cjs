// Source Node is the oracle. This runner never uses Adamic JS as its reference.
const fs = require('node:fs');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const ts = require(process.env.SCOUT_TYPESCRIPT);
if (ts.version !== '6.0.3') throw new Error(`wrong TypeScript: ${ts.version}`);
const here = __dirname;
const out = path.resolve(process.argv[2]);
fs.mkdirSync(out); // Never overwrite evidence.
const compiler = process.argv[3];
const rows = JSON.parse(fs.readFileSync(path.join(here, 'fixtures.json'), 'utf8'));
const results = [];
const expected = JSON.parse(fs.readFileSync(path.join(here, 'expected.json'), 'utf8'));
function node(source, stem) {
 const js = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText;
 const file = path.join(out, stem + '.cjs'); fs.writeFileSync(file, js);
 const r = spawnSync(process.execPath, [file]);
 if (r.error) throw r.error;
 fs.writeFileSync(path.join(out, stem + '.stdout'), r.stdout);
 fs.writeFileSync(path.join(out, stem + '.stderr'), r.stderr);
 return r;
}
for (const row of rows) {
 const file = path.join(here, 'fixtures', row.name + '.a');
 const source = fs.readFileSync(file, 'utf8');
 if (source.split(row.old).length !== 2) throw new Error(`ambiguous mutation: ${row.name}`);
 const control = node(source, row.name + '.node');
 const e = expected.find(e => e.name === row.name);
 if (control.status !== 0 || !e || control.stdout.toString() !== e.node_stdout || control.stderr.length !== 0) throw new Error(`Node golden failed: ${row.name}`);
 const mutant = node(source.replace(row.old, row.new), row.name + '.mutant');
 if (control.status === mutant.status && control.stdout.equals(mutant.stdout) && control.stderr.equals(mutant.stderr)) throw new Error(`mutant survived: ${row.name}`);
 const r = spawnSync(compiler, ['build', file, '-o', path.join(out, row.name + '.native')]);
 if (r.error) throw r.error;
 if (!e || r.status !== e.build_exit || !r.stderr.toString().includes(e.diagnostic_contains)) throw new Error(`unexpected Adamic outcome: ${row.name}: ${r.stderr}`);
 if (!source.startsWith('// a-check: ')) throw new Error(`missing a-check header: ${row.name}`);
 fs.writeFileSync(path.join(out, row.name + '.build.stdout'), r.stdout);
 fs.writeFileSync(path.join(out, row.name + '.build.stderr'), r.stderr);
 let nativeEqual = null;
 if (r.status === 0) {
  const n = spawnSync(path.join(out, row.name + '.native'));
  fs.writeFileSync(path.join(out, row.name + '.native.stdout'), n.stdout);
  fs.writeFileSync(path.join(out, row.name + '.native.stderr'), n.stderr);
  nativeEqual = n.status === control.status && n.stdout.equals(control.stdout) && n.stderr.equals(control.stderr);
  if (!nativeEqual) throw new Error(`native mismatch: ${row.name}`);
 }
 results.push({ name: row.name, node_exit: control.status, node_stdout: control.stdout.toString(), mutant_exit: mutant.status, caught: true, build_exit: r.status, diagnostic: r.stderr.toString(), native_equal: nativeEqual });
}
fs.writeFileSync(path.join(out, 'results.json'), JSON.stringify(results, null, 2) + '\n');
console.log(`${results.length} Node controls and source mutants passed`);
