// The record cost probe's two kept shapes in plain JavaScript (zz_record_cost_probe_test.go), its
// Node rows: node records.mjs counted-kept|graph-kept N. A node {next: null} is held by a record {}
// under "k", and in the graph shape the node holds the record back, a cycle per record; every record
// is kept in one array to the end. It prints the harness's line: make and free seconds and the peak
// resident set in KiB. Free is dropping the array and, when the test runs node with --expose-gc, a
// full collection, since V8 frees nothing at the moment the last reference goes.
const [mode, text] = process.argv.slice(2);
const count = Number(text);
if ((mode !== 'counted-kept' && mode !== 'graph-kept') || !Number.isInteger(count) || count < 0) {
  console.error('usage: records.mjs counted-kept|graph-kept N');
  process.exit(2);
}
const graph = mode === 'graph-kept';
let records = [];
const start = process.hrtime.bigint();
for (let index = 0; index < count; index++) {
  const node = { next: null };
  const record = {};
  record.k = node;
  if (graph) node.next = record;
  records.push(record);
}
const made = process.hrtime.bigint();
records = null;
if (typeof globalThis.gc === 'function') globalThis.gc();
const freed = process.hrtime.bigint();
const seconds = (from, to) => (Number(to - from) / 1e9).toFixed(6);
console.log(`mode=${mode} n=${count} make=${seconds(start, made)} free=${seconds(made, freed)} peak_kb=${process.resourceUsage().maxRSS}`);
