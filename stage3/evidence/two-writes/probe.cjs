// Node preload: counters do not write to compiler stdout or stderr.
const fs = require('node:fs');
const path = require('node:path');
const domains = new WeakMap();
const rows = new Map();
const examples = [];
let identifyingInput = false;
const mutant = process.env.TWO_WRITES_MUTANT;
function fileOf(node) {
    const seen = new Set();
    while (node && !seen.has(node)) {
        if (node.fileName) return node.fileName;
        seen.add(node);
        node = node.original || node.parent;
    }
    return process.env.TWO_WRITES_INPUT || '<unknown>';
}
function rowFor(site, input) {
    const key = JSON.stringify([site, input]);
    if (!rows.has(key)) rows.set(key, { site, input, assignments: 0, violations: 0, reads: 0, outOfTypeReads: 0 });
    return rows.get(key);
}
function stack() { return new Error().stack.split('\n').slice(3, 8); }
const probe = globalThis.__twoWrites = {
    // T is erased. A legal caller with a literal flags domain must register it.
    // The one stock internal call has Expression.flags: NodeFlags (a bit mask).
    flagsDomain(node, values) { domains.set(node, values); },
    write(site, target, value, original) {
        let input;
        identifyingInput = true;
        try { input = fileOf(original); }
        finally { identifyingInput = false; }
        const row = rowFor(site, input);
        row.assignments++;
        const allowed = domains.get(original);
        const outside = site === 'parent' ? value === undefined :
            allowed ? !allowed.includes(value) : !Number.isInteger(value) || value < 0 || value > 0x7fffffff;
        const counted = mutant === site ? true : outside;
        if (counted) row.violations++;
        let example;
        if (examples.filter(e => e.site === site).length < 12) {
            example = { site, input, kind: original.kind, value: value === undefined ? 'undefined' : typeof value === 'object' ? '<Node>' : value,
                domain: site === 'parent' ? 'Node (undefined excluded)' : allowed || 'NodeFlags bit mask',
                outside, counted, writeStack: stack(), readStack: null };
            examples.push(example);
        }
        // Install after the real store. No observation helper reads the field.
        // The property remains enumerable/configurable and accepts every later store.
        let current = value;
        let active = true;
        Object.defineProperty(target, site, {
            enumerable: true, configurable: true,
            get() {
                if (active && !identifyingInput && mutant !== 'reads-' + site) {
                    row.reads++;
                    if (outside) row.outOfTypeReads++;
                    if (example && !example.readStack) example.readStack = stack();
                }
                return current;
            },
            set(next) { current = next; active = false; },
        });
    },
    snapshot() { return { rows: [...rows.values()], examples }; },
};
if (process.env.TWO_WRITES_RESULTS) {
    process.on('exit', () => {
        const folder = process.env.TWO_WRITES_RESULTS;
        fs.mkdirSync(folder, { recursive: true });
        fs.writeFileSync(path.join(folder, `${process.pid}.json`), JSON.stringify(probe.snapshot(), null, 2) + '\n');
    });
}
module.exports = probe;
