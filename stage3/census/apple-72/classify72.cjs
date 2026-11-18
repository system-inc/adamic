'use strict';
// #0b9xqhh's census: node classify72.cjs TREE_B SITES.json OFF.json ON.json B.json OUT.json
// OFF is main's lowering over the plain tree, ON the prototype over the plain tree (2-o at existing casts
// alone), B the prototype over the rewritten tree (:72 at every family 5 relation it could spell). A stop
// compares by rule, file and words, counted, since lines don't move and columns do.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const [treeB, sitesFile, offFile, onFile, bFile, outFile] = process.argv.slice(2);
const load = file => {
    const raw = JSON.parse(fs.readFileSync(file, 'utf8'));
    return {boundary: raw.Boundary, stops: raw.Stops.map(s => {
        const m = /^(.*):(\d+):(\d+)$/.exec(s.Site);
        const file = m ? m[1].split('/tree-').pop().replace(/^[^/]*\//, '') : s.Site;
        return {file, line: m ? +m[2] : 0, rule: s.Rule.split('/').pop(), words: s.Words};
    })};
};
const off = load(offFile), on = load(onFile), b = load(bFile);
const family5 = s => s.words.includes('absent from structural source');
const key = s => `${s.rule}|${s.file}|${s.words}`;
function minus(x, y) {
    const pool = new Map();
    for (const s of y) pool.set(key(s), (pool.get(key(s)) || 0) + 1);
    return x.filter(s => { const n = pool.get(key(s)) || 0; if (n) { pool.set(key(s), n - 1); return false; } return true; });
}
// The function a line sits in, in tree B: its name, or the file for top level.
const sources = new Map();
function enclosing(file, line) {
    let source = sources.get(file);
    if (!source) {
        const full = path.join(treeB, file);
        if (!fs.existsSync(full)) return `${file}:?`;
        source = ts.createSourceFile(full, fs.readFileSync(full, 'utf8'), ts.ScriptTarget.Latest, true);
        sources.set(file, source);
    }
    let position;
    try { position = source.getPositionOfLineAndCharacter(line - 1, 0); } catch { return `${file}:?`; }
    let name = '<top>';
    (function visit(node) {
        if (node.pos > position || node.end < position) return;
        if ((ts.isFunctionLike(node) || ts.isClassLike(node)) && node.name && ts.isIdentifier(node.name)) name = node.name.text;
        ts.forEachChild(node, visit);
    })(source);
    return `${file}#${name}`;
}
const sites = JSON.parse(fs.readFileSync(sitesFile, 'utf8'));
const rewritten = sites.filter(s => s.rewritten);
const gone = minus(on.stops, b.stops), added = minus(b.stops, on.stops);
const byUnit = new Map();
for (const s of added) { const u = enclosing(s.file, s.line); byUnit.set(u, [...(byUnit.get(u) || []), s]); }
const shape = s => `${s.rule}: ${s.words.replace(/\b[A-Z][A-Za-z0-9_<>,.\[\]| ]*\b/g, 'T').replace(/\d+/g, 'N').slice(0, 110)}`;
const tally = (list, f) => { const m = new Map(); for (const x of list) m.set(f(x), (m.get(f(x)) || 0) + 1); return [...m].sort((p, q) => q[1] - p[1]); };
const result = {boundaries: {off: off.boundary, on: on.boundary, b: b.boundary}, counts: {}, sites: []};
result.counts.family5 = {off: off.stops.filter(family5).length, on: on.stops.filter(family5).length, b: b.stops.filter(family5).length};
result.counts.events = {off: off.stops.length, on: on.stops.length, b: b.stops.length};
result.counts.twoO = {gone: minus(off.stops, on.stops).length, added: minus(on.stops, off.stops).length};
result.counts.rewrite = {gone: gone.length, goneFamily5: gone.filter(family5).length, added: added.length, addedFamily5: added.filter(family5).length};
result.goneOtherFamilies = tally(gone.filter(s => !family5(s)), s => s.rule);
result.addedShapes = tally(added, shape);
for (const site of rewritten) {
    const unit = enclosing(site.relative, site.line);
    const next = (byUnit.get(unit) || []).map(s => `${s.rule} ${s.file}:${s.line}: ${s.words.slice(0, 160)}`);
    result.sites.push({key: site.key, unit, expression: site.expression, slot: site.slot.slice(0, 120), next});
}
result.left = tally(sites.filter(s => s.left), s => s.left.replace(/: .*/, ''));
fs.writeFileSync(outFile, JSON.stringify(result, null, 1));
console.log(`classify72: events off ${off.stops.length}, on ${on.stops.length}, rewritten ${b.stops.length}; boundaries ${JSON.stringify(result.boundaries)}`);
console.log(`classify72: family 5 stops off ${result.counts.family5.off}, on ${result.counts.family5.on}, rewritten ${result.counts.family5.b}`);
console.log(`classify72: 2-o at existing casts: ${result.counts.twoO.gone} gone, ${result.counts.twoO.added} added`);
console.log(`classify72: the rewrite: ${gone.length} gone (${result.counts.rewrite.goneFamily5} family 5), ${added.length} added (${result.counts.rewrite.addedFamily5} family 5)`);
for (const [rule, n] of result.goneOtherFamilies) console.log(`classify72:   gone, not family 5: ${n} ${rule}`);
for (const [s, n] of result.addedShapes.slice(0, 40)) console.log(`classify72:   added ${n} ${s}`);
for (const [why, n] of result.left) console.log(`classify72:   left alone ${n}: ${why}`);
