// Build the independent source oracle once using the same transform as oracle/node.mjs.
import { readFileSync, writeFileSync } from 'node:fs';
import { stripTypeScriptTypes } from 'node:module';
import { join } from 'node:path';
const [source, output] = process.argv.slice(2);
for (const name of ['main', 'inline', 'classes']) {
 const text = stripTypeScriptTypes(readFileSync(join(source, name + '.ts'), 'utf8'), { mode: 'transform' });
 writeFileSync(join(output, name + '.mjs'), text.replaceAll('./inline.ts', './inline.mjs').replaceAll('./classes.ts', './classes.mjs'));
}
