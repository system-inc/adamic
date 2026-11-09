import { statSync, type Stats } from 'node:fs';

function kind(entry: Stats | undefined): string {
	return `${entry?.isFile()}`;
}
console.log(kind(undefined));
console.log(kind(statSync('/')));
