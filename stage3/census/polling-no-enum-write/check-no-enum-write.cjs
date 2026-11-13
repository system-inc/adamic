'use strict';
// Keep census.cjs unchanged from c7f2bc3d. Preserve its child status.
const cp = require('node:child_process');
const path = require('node:path');
const fs = require('node:fs');
const result = cp.spawnSync(process.execPath,[path.join(__dirname,'census.cjs'),process.argv[2]], {encoding:'utf8',maxBuffer:64*1024*1024});
if (result.stderr) process.stderr.write(result.stderr);
if (result.error) { console.error(result.error); process.exit(1); }
if (result.status !== 0) process.exit(result.status ?? 1);
const census = JSON.parse(result.stdout);
fs.writeFileSync(process.argv[3],result.stdout);
console.log(JSON.stringify({counts:census.counts,converged:census.converged,writes:census.writes}));
process.exit(census.converged && census.counts.writes === 0 ? 0 : 1);
