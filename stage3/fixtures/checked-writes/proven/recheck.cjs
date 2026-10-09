'use strict';
// Recompute every original receiver graph on the pinned source, without inventing compiler results.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const { ts, options, createAudit } = require('./receiving-audit.cjs');
const tree = path.resolve(process.argv[2]);
const ledger = JSON.parse(fs.readFileSync(path.join(__dirname, 'sites.json'), 'utf8'));
const program = ts.createProgram(ts.sys.readDirectory(path.join(tree, 'src'), ['.ts'], ['**/lib/**']), options);
const audit = createAudit(program, tree);
const records = ledger.records.map(site => {
    const found = audit.locate(site);
    const actual = found.node ? audit.inspect(found.node, found.original) : { classification: 'd', reason: found.failure };
    return { id: site.id, where: site.where, expected: site.classification, unchanged: actual.classification === site.classification, actual };
});
const report = { evidence_kind: 'stock TypeScript receiver audit, not Adamic compiler proof', records, unchanged: records.filter(row => row.unchanged).length };
fs.writeFileSync(path.join(__dirname, 'recheck.json'), JSON.stringify(report, null, 2) + '\n');
console.log(`${report.unchanged} of ${records.length} pinned receiver classifications unchanged`);
assert.equal(report.unchanged, 98);
