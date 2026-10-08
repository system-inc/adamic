// Independent JavaScript option contract, unchanged raw source with the u flag.
import { readFileSync } from 'node:fs';
for (const [pattern, text] of JSON.parse(readFileSync(process.argv[2], 'utf8'))) {
  try { console.log(`${pattern}\t${new RegExp(pattern, 'u').test(text)}`); }
  catch (error) { if (!(error instanceof SyntaxError)) throw error; console.log(`${pattern}\tSyntaxError`); }
}
