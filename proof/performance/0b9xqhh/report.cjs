"use strict";
// Usage: node report.cjs <instrumented.json> <counts.json> [plainSeconds]
// Prints each selected member's counted accesses on the instrumented run, most first, the totals, and the price of
// checking them at (6.8 - 1.9) ns a read, which is a bound: a receiver the lowering proves plain pays nothing.
const fs = require("node:fs");

const [instrumentedPath, countsPath, plainSecondsText] = process.argv.slice(2);
if (!instrumentedPath || !countsPath) {
    console.error("usage: node report.cjs <instrumented.json> <counts.json> [plainSeconds]");
    process.exit(2);
}
const { members } = JSON.parse(fs.readFileSync(instrumentedPath, "utf8"));
const counts = JSON.parse(fs.readFileSync(countsPath, "utf8"));
for (const key of ["r", "w", "c", "i"]) {
    if (!Array.isArray(counts[key]) || counts[key].length !== members.length) {
        console.error(`counts.${key} has ${counts[key]?.length} entries, not the ${members.length} members`);
        process.exit(1);
    }
}
const checkedNanoseconds = 6.8 - 1.9;
const rows = members.map((m, i) => ({
    name: m.name,
    id: m.id,
    reads: counts.r[i],
    writes: counts.w[i],
    compound: counts.c[i],
    in: counts.i[i],
    sites: m.static.read + m.static.write + m.static.compound + m.static.in + m.static.destructure,
    uncounted: m.static.uncounted,
}));
const total = r => r.reads + r.writes + r.compound + r.in;
rows.sort((a, b) => total(b) - total(a) || a.id.localeCompare(b.id));
const pad = (s, n) => String(s).padEnd(n);
const num = (s, n) => String(s).padStart(n);
console.log(`${pad("member", 22)} ${num("reads", 13)} ${num("writes", 11)} ${num("compound", 10)} ${num("in", 9)} ${num("sites", 6)} ${num("uncounted", 9)}  declaration`);
for (const r of rows) console.log(`${pad(r.name, 22)} ${num(r.reads, 13)} ${num(r.writes, 11)} ${num(r.compound, 10)} ${num(r.in, 9)} ${num(r.sites, 6)} ${num(r.uncounted, 9)}  ${r.id}`);
const sum = key => rows.reduce((a, r) => a + r[key], 0);
const totals = { reads: sum("reads"), writes: sum("writes"), compound: sum("compound"), in: sum("in") };
const all = totals.reads + totals.writes + totals.compound + totals.in;
console.log(`${pad("total", 22)} ${num(totals.reads, 13)} ${num(totals.writes, 11)} ${num(totals.compound, 10)} ${num(totals.in, 9)} ${num(sum("sites"), 6)} ${num(sum("uncounted"), 9)}`);
// The checker reads TypeParameter's members on every run, so counted sites and no counts mean the counters never ran.
if (all === 0 && sum("sites") > 0) {
    console.error(`${sum("sites")} counted sites and no counts: the instrumented build's counters never ran`);
    process.exit(1);
}
// A compound access reads before it writes, so it is a checked read; writes and `in` checks are priced apart, as the
// ruling registers checked reads and says nothing yet of a view's writes.
const checkedReads = totals.reads + totals.compound;
const plainSeconds = Number(plainSecondsText);
const share = ns => (plainSeconds > 0 ? ` (${(ns / 1e9 / plainSeconds * 100).toFixed(3)}% of the plain run's ${plainSeconds.toFixed(3)} s median)` : "");
const readsNs = checkedReads * checkedNanoseconds;
const allNs = all * checkedNanoseconds;
console.log(`price: checked reads (reads + compound): ${checkedReads}; bound ${(readsNs / 1e6).toFixed(3)} ms at ${checkedNanoseconds.toFixed(1)} ns each${share(readsNs)}`);
console.log(`price: every access (reads, writes, compound, in): ${all}; bound ${(allNs / 1e6).toFixed(3)} ms${share(allNs)}`);
console.log("price: both are upper bounds: a receiver the lowering proves plain pays no check at all.");
