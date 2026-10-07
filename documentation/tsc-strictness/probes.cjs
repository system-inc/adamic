// Run with Node 24; TSC_SURVEY_TYPESCRIPT must point to TypeScript 6.0.3's typescript.js.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const ts = require(process.env.TSC_SURVEY_TYPESCRIPT);
assert.equal(ts.version, '6.0.3');
const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'strictness-probes-'));
let checks = 0;
function check(source) {
  const file = path.join(dir, 'probe.ts');
  fs.writeFileSync(file, source);
  const options = { strict: true, exactOptionalPropertyTypes: true,
    noUncheckedIndexedAccess: true, target: ts.ScriptTarget.ES2024,
    module: ts.ModuleKind.ESNext, skipLibCheck: true };
  const program = ts.createProgram([file], options);
  const codes = ts.getPreEmitDiagnostics(program).map(d => d.code);
  let js;
  program.emit(undefined, (name, text) => { if (name.endsWith('.js')) js = text; });
  checks++;
  return {codes, js};
}
function caught(name, fn) {
  assert.throws(fn, undefined, `mutant survived: ${name}`);
  console.log(`caught mutant: ${name}`);
}
try {
  const narrow = check('interface Slot { p?: string }; const s: Slot = {}; s.p = undefined;');
  const honest = check('interface Slot { p?: string | undefined }; const s: Slot = {}; s.p = undefined;');
  assert.deepEqual(narrow.codes, [2412]);
  assert.deepEqual(honest.codes, []);
  assert.equal(narrow.js, honest.js);
  assert.deepEqual(check('interface Slot { p?: string | undefined }; const s: Slot = {}; s.p = 42;').codes, [2322]);
  assert.deepEqual(check('function f(a: {x: number}[]) { if(a.length === 1) return a[0].x; return 0; }').codes, [2532]);
  assert.deepEqual(check('function panic(): never { throw Error("missing element"); } function f(a: {x: number}[]) { if(a.length === 1) { const v = a[0] ?? panic(); return v.x; } return 0; }').codes, []);
  const fingerprint = o => [Object.hasOwn(o, 'p'), 'p' in o, Object.keys(o), {...o}];
  const original = {}; original.p = undefined;
  assert.deepEqual(fingerprint(original), [true, true, ['p'], {p: undefined}]);
  caught('delete optional slot instead of writing undefined', () => {
    const mutant = {}; mutant.p = undefined; delete mutant.p;
    assert.deepEqual(fingerprint(mutant), fingerprint(original));
  });
  function required(v) { if(v === undefined || v === null) throw Error('missing element'); return v; }
  for (const v of [0, '', false, {x: 3}]) assert.equal(required(v), v);
  assert.throws(() => required(undefined), /missing element/);
  caught('erase checked unwrap into an unchecked identity', () => {
    const mutant = v => v;
    assert.throws(() => mutant(undefined), /missing element/);
  });
  // U059's local condition admits an empty, well-typed array.
  const arrowCondition = a => !(a.length > 1 || a.hasTrailingComma || a[0].constraint);
  assert.throws(() => arrowCondition([]), TypeError);
  // U051's local upper-bound test admits a negative, well-typed number.
  const parameterName = (a, pos) => pos < a.length ? a[pos].escapedName : 'fallback';
  assert.throws(() => parameterName([{escapedName: 'x'}], -1), TypeError);
  const sparse = new Array(1);
  assert.equal(sparse.length, 1);
  assert.throws(() => sparse[0].x, TypeError);
  // U083: absence is the existing zero flag, rather than an invariant violation.
  for (const n of [undefined, 0, 1, 2, 3, NaN, -1]) assert.equal(n & 2, (n ?? 0) & 2);
  caught('default every missing value to zero', () => {
    const mutant = v => v ?? 0;
    assert.throws(() => mutant(undefined), /missing element/);
  });
  // Reusing a dynamic lookup may change observable evaluation count/order.
  let reads = 0;
  const table = {get p() { reads++; return {x: reads}; }};
  assert.equal(table.p.x + table.p.x, 3);
  assert.equal(reads, 2);
  // U073: group 1 is mandatory; groups 2 and 3 are alternative, optional captures.
  const re = /(\sx\s*=\s*)(?:(?:'([^']*)')|(?:"([^"]*)"))/im;
  for (const input of [" x='v'", ' x="v"']) {
    const m = re.exec(input); assert.ok(m); assert.equal(typeof m[1], 'string');
    assert.ok(m[2] === undefined || m[3] === undefined);
  }
  console.log(`PASS: ${checks} stock TypeScript checks; identical emitted JS for honest optional type; Node presence, holes, bounds, flags, getters and captures; 3 mutants caught`);
} finally { fs.rmSync(dir, {recursive: true, force: true}); }
