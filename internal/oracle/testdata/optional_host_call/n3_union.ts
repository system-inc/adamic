import { readdirSync, statSync, type Dirent, type Stats } from 'node:fs';

function kind(entry: Stats | Dirent | undefined): string {
	return `${entry?.isFile()}`;
}
console.log(kind(undefined));
console.log(kind(statSync('/')));
const entries = readdirSync('/', { withFileTypes: true });
let read = 0;
for (const entry of entries) {
	if (kind(entry) !== 'undefined') {
		read += 1;
	}
}
console.log(`${read === entries.length}`);
