'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('/tmp/watch-functions-1532/main-tree/built/local/typescript.js');
const target = '/tmp/watch-functions-1532/control-target.ts';
fs.writeFileSync(target, 'export const x = 1;\n');
const results = [];
for (const interval of [250, 500, 2000, undefined, 10]) {
  const before = fs.existsSync(process.env.ADAMIC_WATCH_COUNT) ? fs.readFileSync(process.env.ADAMIC_WATCH_COUNT, 'utf8').trim().split('\n').filter(Boolean).length : 0;
  let result;
  try {
    const watcher = ts.sys.watchFile(target, () => {}, interval, { watchFile: ts.WatchFileKind.PriorityPollingInterval });
    watcher.close();
    result = 'registered and closed';
  } catch (error) { result = error.stack; }
  const rows = fs.readFileSync(process.env.ADAMIC_WATCH_COUNT, 'utf8').trim().split('\n').map(JSON.parse);
  assert.equal(rows.length, before + 1);
  assert.deepEqual(rows.at(-1), { interval: String(interval), type: typeof interval });
  results.push({ interval: String(interval), result });
}
console.log(JSON.stringify({counterFires: true, results}, null, 2));
