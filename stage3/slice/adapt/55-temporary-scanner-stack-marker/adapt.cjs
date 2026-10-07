const fs = require('node:fs');
const path = require('node:path');
const tree = path.resolve(process.argv[2]);
if (!fs.existsSync(path.join(tree, 'slice.json'))) throw new Error('adaptation 55 requires a declaration slice');
const file = path.join(tree, 'src/compiler/debug.ts');
const before = fs.readFileSync(file, 'utf8');
// Preserve upstream's callable marker. Restore slices made by the former
// opaque-marker implementation; a newly gathered slice already has AnyFunction.
const after = before.replaceAll('stackCrawlMark?: {}', 'stackCrawlMark?: AnyFunction');
const restored = (before.match(/stackCrawlMark\?: \{\}/g) || []).length;
const count = (after.match(/stackCrawlMark\?: AnyFunction/g) || []).length;
if (count !== 2 && count !== 3) throw new Error(`unexpected callable marker count: ${count}`);
if (restored) fs.writeFileSync(file, after);
console.log(JSON.stringify({ callableMarkers: count, restoredOpaqueMarkers: restored }));
