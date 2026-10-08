// Test harness only: never assembled into a Worker.
import { readFileSync } from 'node:fs';
import { pathToFileURL } from 'node:url';
const [handler, corpus, expectedPath] = process.argv.slice(2);
if (!corpus) {
	await import(pathToFileURL(handler));
} else {
	const { evaluate } = await import(pathToFileURL(handler));
	const texts = JSON.parse(readFileSync(corpus, 'utf8'));
	const expected = expectedPath ? readFileSync(expectedPath, 'utf8').trimEnd().split('\n') : undefined;
	for (let index = 0; index < texts.length; index++) {
		const actual = evaluate(Math.floor(index / 10000), texts[index]);
		if (expected) {
			if (actual !== expected[index]) throw new Error(`native disagreement at corpus ${index}: ${JSON.stringify(texts[index])}\nactual: ${actual}\nexpected: ${expected[index]}`);
		} else console.log(actual);
	}
	if (expected && expected.length !== texts.length) throw new Error('native corpus observation count differs');
	if (expected) console.log(`compared ${texts.length} texts byte for byte with native`);
}
