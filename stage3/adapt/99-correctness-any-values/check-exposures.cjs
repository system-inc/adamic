'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const selection = require('./exposure-selection.json');

const [evidence, mode = 'check', readme = path.join(__dirname, 'README.md')] = process.argv.slice(2);
assert(evidence, 'usage: node check-exposures.cjs <replay-directory> [check|render] [README.md]');

const input = file => fs.readFileSync(path.join(evidence, file), 'utf8');
const location = value => value.replace(/^.*\/src\//, '');
const any = 'a value of type any: any has no proven type';
const predicate = 'a checked predicate target whose base has no required field to check (MapLike<string[]>)';
const sourcePin = require('../../source.json').commit;

function findings(side) {
    assert.equal(input(side + '.source').trim(), sourcePin, 'upstream source pin');
    return selection.map((pair, index) => {
        const anchor = pair[side === 'before' ? 'beforeAnchor' : 'headAnchor'];
        const unit = pair[side === 'before' ? 'baseline' : 'unit'];
        const record = JSON.parse(input(`${side}-${index}.json`));
        assert.equal(record.module_format, 'CommonJSBundle', 'unexpected replay module format');
        assert(record.units?.some(candidate => location(candidate.where) === unit), `missing selected unit ${unit}`);
        const findings = record.findings.map(row => ({ ...row, where: location(row.where), unit: location(row.unit) }));
        assert(findings.some(row => row.unit === unit && row.where === anchor[0] &&
            row.kind === anchor[1] && row.reason === anchor[2]), `missing exact replay anchor for ${unit}`);
        return findings;
    });
}

const baseline = findings('before');
const candidate = findings('head');
const selected = [];
for (const [index, { baseline: priorUnit, unit, rows }] of selection.entries()) {
    const before = baseline[index];
    const head = candidate[index];
    const earlier = before.filter(row => row.unit === priorUnit && row.kind === 'Refused' && row.reason === any);
    assert(earlier.length, `missing earlier any stop in ${priorUnit}`);
    const after = head.filter(row => row.unit === unit);
    for (const [where, kind, reason, meter] of rows) {
        assert(after.some(row => row.where === where && row.kind === kind && row.reason === reason),
            `missing pinned replay row ${unit} → ${where} (${reason})`);
        selected.push([where, unit, kind, reason, meter]);
    }
    // Select the dependency finding from the actual unit replay; the reviewed scopes
    // explicitly include both callers, even when either caller has other retained stops.
    if (unit === 'compiler/commandLineParser.ts:3623:5' || unit === 'compiler/commandLineParser.ts:3543:1') {
        const dependencies = after.filter(row => row.reason === predicate);
        assert.equal(dependencies.length, 1, `expected one MapLike predicate dependency in ${unit}`);
        assert(!before.some(row => row.unit === priorUnit && row.reason === predicate), `predicate already observed in ${priorUnit}`);
        const [row] = dependencies;
        selected.push([row.where, unit, row.kind, predicate, 'Exact reason absent · @system_adamic']);
    }
}
assert.equal(selected.filter(row => row[3] === predicate).length, 2, 'both MapLike predicate callers must be selected');
const expected = selected.map(([site, unit, verdict, reason, meter]) =>
    `| \`${site}\` | \`${unit}\` | ${verdict} | ${reason.replaceAll('|', '&#124;')} | ${meter} |`);
if (mode === 'render') {
    console.log(expected.join('\n'));
} else {
    assert.equal(mode, 'check');
    const markdown = fs.readFileSync(readme, 'utf8');
    const section = markdown.split('### Paired full-mode exposure inventory\n');
    assert.equal(section.length, 2, 'paired exposure section');
    const actual = section[1].split('\n### ')[0].split('\n').filter(line => line.startsWith('| `'));
    assert.deepEqual(actual.sort(), expected.sort(), 'README must match the pinned paired exposure inventory');
    console.log(`${actual.length} pinned paired exposure rows match, including both MapLike predicate callers`);
}
