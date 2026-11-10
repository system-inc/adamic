'use strict';
// Preloaded only into temporary oracle processes. Every executed split is compared.
const fs = require('node:fs');
const path = require('node:path');
const shared = /\s+/;
const directory = process.env.SEMVER_SPLIT_RECORDS;
if (!directory) throw Error('SEMVER_SPLIT_RECORDS is required');
fs.mkdirSync(directory, {recursive:true});
globalThis.__semverSplit = range => {
    const before = range.split(shared);
    const after = range.split(/\s+/);
    const same = before.length === after.length && before.every((value, index) => value === after[index]);
    fs.appendFileSync(path.join(directory, process.pid + '.jsonl'), JSON.stringify({range, before, after, same}) + '\n');
    if (!same) throw Error('parseRange split difference');
    return after;
};
