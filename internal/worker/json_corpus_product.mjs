// Prepare immutable inputs, not backend verdicts. Each Worker process crosses all
// four schemas, retaining descriptor-cache witnesses while bounding outer parsing.
import { readFileSync, writeFileSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import { join } from 'node:path';
const [mode, directory] = process.argv.slice(2);
const corpus = join(directory, 'corpus.json'), witness = join(directory, 'stringify.txt');
execFileSync(process.execPath, [`json_${mode}_corpus.mjs`, corpus, witness]);
const texts = JSON.parse(readFileSync(corpus, 'utf8'));
if (texts.length !== 40000) throw new Error(`expected 40000 texts, got ${texts.length}`);
const expected = mode === 'encode' ? readFileSync(witness, 'utf8').trimEnd().split('\n') : [];
for (let start = 0; start < texts.length; start += 250) {
 writeFileSync(join(directory, `part-${String(start).padStart(5, '0')}.json`), JSON.stringify(texts.slice(start, start + 250)));
}
for (let start = 0; start < texts.length; start += 1000) {
 const name = String(start).padStart(5, '0');
 const offsets = [0, 10000, 20000, 30000].map(schema => schema + start / 4);
 writeFileSync(join(directory, `${name}.json`), JSON.stringify(offsets.flatMap(offset => texts.slice(offset, offset + 250))));
 if (mode === 'encode') writeFileSync(join(directory, `${name}.txt`), offsets.flatMap(offset => expected.slice(offset, offset + 250)).join('\n') + '\n');
}
