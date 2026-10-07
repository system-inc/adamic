const fs = require('node:fs'), path = require('node:path');
const tree = path.resolve(process.argv[2]);
if (!fs.existsSync(path.join(tree, 'slice.json'))) throw new Error('adaptation 87 requires a declaration slice');
const file = path.join(tree, 'src/compiler/debug.ts'), before = fs.readFileSync(file, 'utf8');
const original = 'stackCrawlMark || fail', revised = 'stackCrawlMark || (fail as Function)';
const count = before.split(original).length - 1;
if (count !== 1 && !before.includes(revised)) throw new Error('unexpected Debug.fail capture fallback');
if (count) fs.writeFileSync(file, before.replace(original, revised));
console.log(JSON.stringify({functionMetadataAnnotations: count}));
