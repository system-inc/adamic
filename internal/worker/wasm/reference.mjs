// Oracle process only. Keep the Workers harness free of Node's WASI implementation.
// Use the driver's own crossing function, replacing only its in-function assertions.
import assert from 'node:assert/strict';
import { readFileSync, writeFileSync } from 'node:fs';
import { pathToFileURL } from 'node:url';

const [binary, driverHost, requestsPath, outputPath] = process.argv.slice(2);
const source = readFileSync(driverHost, 'utf8');
const begin = source.indexOf('\tconst call = bytes => {');
const end = source.indexOf('\n\tfor (let index = 0; index < 1000;', begin);
assert.ok(begin >= 0 && end > begin, 'driver crossing must be found');
// This generated oracle is a standalone Node script, never shipped to a Worker.
const crossing = source.slice(begin, end)
  .replace("assert.equal(actual, plain.handleRequest(Buffer.from(bytes).toString('utf8')), 'response differs from source on Node');", 'results.push(actual);');
const prelude = `import assert from 'node:assert/strict';
import {readFileSync, writeFileSync, openSync} from 'node:fs';
import {WASI} from 'node:wasi';
const results = [];
const wasi = new WASI({version:'preview1', args:[], env:{}, preopens:{}, stdout:openSync(${JSON.stringify(outputPath + '.stdout')},'w'), stderr:openSync(${JSON.stringify(outputPath + '.stderr')},'w')});
const module = new WebAssembly.Module(readFileSync(${JSON.stringify(binary)}));
const instance = new WebAssembly.Instance(module, wasi.getImportObject());
wasi.initialize(instance);
const api = instance.exports;
const baselineLive = api.adamic_live();
const decoder = new TextDecoder('utf-8', {ignoreBOM:true});
${crossing}
for (const text of JSON.parse(readFileSync(${JSON.stringify(requestsPath)},'utf8'))) call(new TextEncoder().encode(text));
writeFileSync(${JSON.stringify(outputPath)}, JSON.stringify(results));
`;
writeFileSync(outputPath + '.mjs', prelude);
// No eval: import the generated oracle from a real file.
await import(pathToFileURL(outputPath + '.mjs'));
