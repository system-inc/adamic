import { readdirSync, type Dirent } from 'node:fs';
function kind(entry: Dirent | undefined): string { return `${entry?.isFile?.()}`; }
console.log(kind(undefined));
let seen = 0;
for (const entry of readdirSync('/', { withFileTypes: true })) { if (kind(entry) !== 'undefined') { seen += 1; } }
console.log(`${seen > 0}`);
