// Coverage classifier only. Comparisons use the unmodified Adamic runtime.
import { readFileSync } from 'node:fs';
import { registerHooks, stripTypeScriptTypes } from 'node:module';
import { fileURLToPath } from 'node:url';
const throwingPanic = 'data:text/javascript,export function panic(message) { throw new Error(message); }';
registerHooks({
 resolve(specifier, context, next) {
  if(specifier === 'adamic') return {url: throwingPanic, shortCircuit: true};
  return next(specifier, context);
 },
 load(url, context, next) {
  if(url.endsWith('.ts') || url.endsWith('.a')) return {format: 'module', source: stripTypeScriptTypes(readFileSync(fileURLToPath(url), 'utf8')), shortCircuit: true};
  return next(url, context);
 }
});
const { Parser } = await import('../../../../typescript/parser/parser.ts');
const rows = readFileSync(process.argv[2], 'utf8').trim().split('\n');
const failures = [];
for(let index = 0; index < rows.length; index++) {
 const path = rows[index].split('\t')[0];
 try { new Parser(readFileSync(path, 'utf8'), path).file(); }
 catch(error) { failures.push({index, error: String(error)}); }
}
console.log(JSON.stringify(failures));
