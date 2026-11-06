'use strict';
const fs = require('node:fs'), path = require('node:path');
const append = fs.appendFileSync;
function record(name, file) {
    if (typeof file !== 'string' && !Buffer.isBuffer(file) && !(file instanceof URL)) return;
    const resolved = file instanceof URL ? require('node:url').fileURLToPath(file) : path.resolve(String(file));
    append(process.env.RO_REPLAY_TRACE, JSON.stringify({operation:name, file:resolved}) + '\n');
}
for (const name of ['writeFileSync','writeFile','appendFileSync','appendFile']) {
    const original = fs[name];
    fs[name] = function(file, ...args) { record(name, file); return original.call(this, file, ...args); };
}
for (const name of ['writeFile','appendFile']) {
    const original = fs.promises[name];
    fs.promises[name] = function(file, ...args) { record('promises.'+name, file); return original.call(this,file,...args); };
}
