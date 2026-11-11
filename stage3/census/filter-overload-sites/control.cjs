'use strict';
const fs = require('node:fs'), path = require('node:path'), cp = require('node:child_process'), assert = require('node:assert/strict');
const root = process.argv[2], dest = process.argv[3];
if (['child', 'signal-child', 'application-signal-child', 'kill-child'].includes(process.argv[4])) {
  const ts = require('../../api/node_modules/typescript');
  const source = ts.createSourceFile('core.ts', fs.readFileSync(path.join(root, 'src/compiler/core.ts'), 'utf8'), ts.ScriptTarget.Latest, true);
  const declaration = source.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === 'filter' && n.body);
  const js = ts.transpileModule(declaration.getText(source).replace('export function', 'function'), { compilerOptions: { target: ts.ScriptTarget.ES2024 } }).outputText;
  const original = new Function(js + ';return filter;')();
  process.env.FILTER_SITE_COUNTS = dest;
  require('./observe.cjs');
  const run = globalThis.__adamicFilterSite('control', original);
  const a = [1, 2]; assert.equal(run(a, () => true), a);
  assert.deepEqual(run(a, x => x > 1), [2]);
  const empty = []; assert.equal(run(empty, () => false), empty);
  assert.equal(run(undefined, () => true), undefined);
  assert.throws(() => run(a, () => { throw new Error('control'); }));
  if (process.argv[4] !== 'child') {
    if (process.argv[4] === 'application-signal-child') process.on('SIGTERM', () => process.exit(7));
    setInterval(() => {}, 1000);
    process.kill(process.pid, process.argv[4] === 'kill-child' ? 'SIGKILL' : 'SIGTERM');
  }
} else {
  fs.mkdirSync(dest, { recursive: true });
  const child = cp.spawnSync(process.execPath, [__filename, root, dest, 'child'], { encoding: 'utf8' });
  console.log('control child exit code: ' + child.status);
  assert.equal(child.status, 0, child.stderr);
  const record = JSON.parse(fs.readFileSync(path.join(dest, child.pid + '.json')));
  assert.equal(record.final, true);
  assert.deepEqual(record.sites, [{ id: 'control', calls: 5, completed: 4, identity_equal: 3, returned_argument_array: 2, undefined_argument: 1 }]);
  console.log(JSON.stringify(record));
  for (const mode of ['signal-child', 'application-signal-child', 'kill-child']) {
    const directory = dest + '-' + mode; fs.mkdirSync(directory, { recursive: true });
    const signalChild = cp.spawnSync(process.execPath, [__filename, root, directory, mode], { encoding: 'utf8' });
    console.log(JSON.stringify({ mode, child_exit_code: signalChild.status, child_signal: signalChild.signal }));
    const observed = JSON.parse(fs.readFileSync(path.join(directory, signalChild.pid + '.json')));
    assert.equal(observed.format, 2);
    if (mode === 'application-signal-child') { assert.equal(signalChild.status, 7); assert.equal(observed.final, true); assert.equal(observed.sites[0].calls, 5); }
    else if (mode === 'signal-child') { assert.equal(signalChild.signal, 'SIGTERM'); assert.equal(observed.final, true); assert.equal(observed.reason, 'SIGTERM'); assert.equal(observed.sites[0].calls, 5); }
    else { assert.equal(signalChild.signal, 'SIGKILL'); assert.equal(observed.final, false); assert.equal(observed.reason, 'start'); }
    console.log(JSON.stringify(observed));
  }
}
