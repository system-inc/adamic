// Object.entries of an Error the runtime made, from a failing host call, with no new Error in the source.
import { readFileSync } from 'node:fs';
try {
	readFileSync('/no/such', 'utf8');
} catch (e) {
	if (e instanceof Error) {
		const o: { message: string } = e;
		console.log(`${Object.entries(o).length}`);
	}
}
