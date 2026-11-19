'use strict';
// #0b9xqhh's census, stage 2: node unit72.cjs MODE TREE SITES.json LATENT.jsonl
// For each rewritten site, the innermost latent unit holding it in TREE (MODE A, the plain tree, at the site's own
// offset; MODE B, the rewritten tree, at its offset moved by the casts spelled before it), with the unit's status
// and its findings. One row a site: `unitMODE<TAB>site<TAB>unit where<TAB>status<TAB>findings`.
const fs = require('node:fs');
const path = require('node:path');
const [mode, tree, sitesFile, latentFile] = process.argv.slice(2);
const sites = JSON.parse(fs.readFileSync(sitesFile, 'utf8')).filter(s => s.rewritten);
const records = fs.readFileSync(latentFile, 'utf8').split('\n').filter(Boolean).map(line => { try { return JSON.parse(line); } catch { return null; } }).filter(r => r && r.file && Array.isArray(r.units));
const byFile = new Map();
for (const r of records) {
    const relative = path.relative(tree, r.file);
    const entry = byFile.get(relative) || {units: [], findings: []};
    entry.units.push(...r.units);
    entry.findings.push(...(r.findings || []));
    byFile.set(relative, entry);
}
const inserted = s => 1 + (' as (' + s.slot + '))').length;
let rows = 0;
for (const site of sites) {
    const entry = byFile.get(site.relative);
    if (!entry) continue;
    let at = site.start;
    if (mode === 'B') for (const other of sites) if (other.relative === site.relative && other.start < site.start) at += inserted(other);
    const holding = entry.units.filter(u => u.body_start !== undefined && u.body_end !== undefined && u.body_start <= at && at < u.body_end).sort((x, y) => (x.body_end - x.body_start) - (y.body_end - y.body_start));
    const unit = holding[0];
    if (!unit) { console.log(`unit${mode}\t${site.key}\t-\tno unit holds it\t`); rows++; continue; }
    const findings = entry.findings.filter(f => f.unit === unit.where).map(f => `${f.kind} ${f.where.split('/').pop()}: ${(f.reason || f.text || '').replace(/\s+/g, ' ').slice(0, 170)}`);
    console.log(`unit${mode}\t${site.key}\t${unit.where.split('/src/').pop()}\t${unit.status}\t${findings.slice(0, 4).join(' || ')}`);
    rows++;
}
console.log(`unit72 ${mode}: ${rows} sites placed in units over ${records.length} file records`);
