import { statSync } from 'node:fs';
interface Sized {
	readonly size: number;
}
const sized: Sized = statSync('.');
console.log(sized.toString());
