// Independent ledger coverage, fixture provenance, golden receipts and negative controls.
const fs = require('node:fs'), path = require('node:path'), cp = require('node:child_process'), crypto = require('node:crypto');
const here = __dirname;
const {contract} = require('./runtime-contract.cjs');
const load = name => JSON.parse(fs.readFileSync(path.join(here, name)));
const classifications = load('classification.json'), fixtures = load('fixtures.json'), observations = load('observations.json');
function assert(ok, message) {if (!ok) throw Error(message);}
function coverage(data) {
 const ledger = JSON.parse(cp.execFileSync('git', ['show', '855bcfaa:stage3/step09-ledger/ledger.json'], {maxBuffer: 32 * 1024 * 1024}));
 const expected = ledger.filter(r => r.kind === 'as_cast' && r.disposition === 'open');
 const ids = new Set(data.rows.map(r => r.id));
 assert(ids.size === expected.length && data.rows.length === expected.length && expected.every(r => ids.has(r.id)), 'cast ledger coverage');
 for (const category of ['tagged','untagged','literal-enum','generic','other']) assert(data.counts[category] === data.rows.filter(r => r.category === category).length, 'category count');
 assert(Object.values(data.counts).reduce((a,b) => a+b,0) === 3957 && data.stock_non_cast_obligations === 271, 'stock population');
 for (const row of fixtures) assert(data.rows.find(r => r.id === row.site)?.category === row.category, `known-answer classification: ${row.site}`);
}
coverage(classifications);
assert(fixtures.length === 20 && new Set(fixtures.map(r => r.file)).size === 20, 'fixture coverage');
for (const category of Object.keys(classifications.counts)) assert(fixtures.filter(r => r.category === category && !r.failing).length === 2 && fixtures.filter(r => r.category === category && r.failing).length === 2, 'two pairs per class');
assert(observations.results.length === fixtures.length, 'observation coverage');
for (const fixture of fixtures) {
 const observed = observations.results.find(r => r.file === fixture.file);
 assert(observed, `missing observation: ${fixture.file}`);
 const text = fs.readFileSync(path.join(here, fixture.file), 'utf8');
 assert(crypto.createHash('sha256').update(text).digest('hex') === observed.sha256, 'stale fixture observation');
 assert(observed.node.exit === 0 && observed.node.stdout === fixture.node_stdout && observed.node.stderr === '', 'Node golden');
 assert(observed.mutants.length > 0, 'missing mutant control');
 if (observed.contract === 'passed') {
  for (const backend of ['native','javascript']) {
   const run = observed[backend];
   assert(contract(fixture, run), `${backend} runtime contract: ${fixture.file}`);
  }
 } else {
  assert(observed.contract === 'blocked' && observed.compiler_exit === 1 && observed.diagnostic.includes('adamic/no-unchecked-cast') && text.startsWith('// a-check: refused adamic/no-unchecked-cast\n'), 'exact compiler blocker and header');
 }
}
function killed(name, mutant) {let caught = false; try {coverage(mutant);} catch {caught = true;} assert(caught, `surviving ${name}`); console.log(`caught ${name}`);}
const dropped = structuredClone(classifications); dropped.rows.shift(); killed('drop one cast', dropped);
const wrongClass = structuredClone(classifications);
wrongClass.rows.find(r => r.id === fixtures[0].site).category = 'other';
wrongClass.counts.tagged--; wrongClass.counts.other++;
killed('move tagged example into other with consistent counts', wrongClass);
const viewFixture = fixtures.find(row => row.category === 'untagged' && row.failing);
for (const backend of ['native', 'javascript']) {
 const receipt = observations.results.find(row => row.file === viewFixture.file)[backend];
 const missingSite = {...receipt, stderr: receipt.stderr.replaceAll(viewFixture.file + ':', '<cast-site-removed>:')};
 assert(receipt.stderr !== missingSite.stderr, 'missing cast-site mutant seam');
 assert(!contract(viewFixture, missingSite), `${backend} missing cast-site mutant survived`);
 console.log(`caught ${backend} field-read stop without cast site`);
}
const blocked = observations.results.filter(r => r.contract === 'blocked').map(r => r.file);
console.log(JSON.stringify({ledger_casts: classifications.rows.length, node_goldens: fixtures.length, runtime_passed: fixtures.length - blocked.length, runtime_blocked: blocked.length, observed_mutant_controls: observations.results.reduce((n,r) => n+r.mutants.length,0)}));
if (process.argv.includes('--require-runtime')) assert(blocked.length === 0, `checked-cast feature missing for ${blocked.join(', ')}`);
