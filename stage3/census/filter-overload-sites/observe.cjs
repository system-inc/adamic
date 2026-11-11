'use strict';
// Loaded only in private counting processes. The original filter runs unchanged.
const fs = require('node:fs'), path = require('node:path');
const directory = process.env.FILTER_SITE_COUNTS;
if (directory) {
  fs.mkdirSync(directory, { recursive: true });
  const sites = new Map();
  const unit = process.env.STAGE3_CACHE ? path.basename(process.env.STAGE3_CACHE) : '<control-or-producer>';
  const cwd = process.cwd();
  globalThis.__adamicFilterSite = (id, original) => (argument, predicate) => {
    let row = sites.get(id);
    if (!row) sites.set(id, row = { id, calls: 0, completed: 0, identity_equal: 0, returned_argument_array: 0, undefined_argument: 0 });
    row.calls++;
    if (argument === undefined) row.undefined_argument++;
    const result = original(argument, predicate);
    row.completed++;
    if (result === argument) {
      row.identity_equal++;
      if (Array.isArray(argument)) row.returned_argument_array++;
    }
    return result;
  };
  function save(final, reason = final ? "exit" : "periodic") {
    const name = path.join(directory, process.pid + '.json');
    fs.writeFileSync(name + '.tmp', JSON.stringify({ format: 2, pid: process.pid, unit, cwd, final, reason, sites: [...sites.values()] }) + '\n');
    fs.renameSync(name + '.tmp', name);
  }
  save(false, 'start');
  for (const signal of ['SIGTERM', 'SIGINT', 'SIGHUP']) {
    function observeSignal() {
      const otherwiseDefault = process.listenerCount(signal) === 1;
      save(otherwiseDefault, signal);
      if (otherwiseDefault) {
        process.removeListener(signal, observeSignal);
        process.kill(process.pid, signal);
      }
    }
    process.on(signal, observeSignal);
  }
  const timer = setInterval(() => save(false), 5000); timer.unref();
  process.on('exit', () => save(true));
}
