// Loaded through the repository's oracle/node.mjs so checksum.a uses the same hooks.
import { readFileSync } from 'node:fs';
import { pathToFileURL } from 'node:url';
import { ResponseChecksum, readCorpus } from './checksum.a';
const [modulePath, corpusPath, iterationsText, hostPath] = process.argv.slice(2);
const { createHost } = await import(pathToFileURL(hostPath));
const module = new WebAssembly.Module(readFileSync(modulePath));
const host = createHost(module);
const requests = readCorpus(corpusPath);
const iterations = Number(iterationsText);
const checksum = new ResponseChecksum();
for (let round = 0; round < iterations; round++) {
  for (const request of requests) checksum.add(host.call(request));
}
console.log(checksum.text());
