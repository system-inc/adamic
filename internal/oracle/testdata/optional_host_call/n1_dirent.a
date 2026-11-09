import { readdirSync, type Dirent } from 'node:fs';

function kind(entry: Dirent | undefined): string {
	return `${entry?.isFile()}`;
}
console.log(kind(undefined));
const entries = readdirSync('/', { withFileTypes: true });
let read = 0;
for (const entry of entries) {
	if (kind(entry) !== 'undefined') {
		read += 1;
	}
}
console.log(`${read === entries.length}`);
