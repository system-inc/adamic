import { realpathSync } from 'node:fs';
import { mkdir, copyFile, writeFile } from 'node:fs/promises';
import { resolve, join } from 'node:path';
import { fileURLToPath } from 'node:url';

export async function generate(directory) {
  const output = resolve(directory);
  const source = fileURLToPath(new URL('.', import.meta.url));
  if (output === resolve(source)) throw new Error('Output must differ from the template directory');
  await mkdir(output, { recursive: true });
  for (const name of ['worker.mjs', 'bridge.mjs', 'host.mjs', 'wasi.mjs']) {
    await copyFile(join(source, name), join(output, name));
  }
  await writeFile(join(output, 'wrangler.toml'), `name = "adamic-wasm-worker"
main = "worker.mjs"
compatibility_date = "2026-10-07"

[[rules]]
type = "CompiledWasm"
globs = ["**/*.wasm"]
fallthrough = true
`);
}

if (process.argv[1] && realpathSync(process.argv[1]) === fileURLToPath(import.meta.url)) {
  if (process.argv.length !== 3) throw new Error('Usage: generate.mjs <out-directory>');
  await generate(process.argv[2]);
}
