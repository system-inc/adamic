import { defaults, run, markdown } from './harness.mjs';
const flags = {
  '--repo': 'repo', '--entry': 'entry', '--corpus': 'corpus', '--name': 'name',
  '--iterations': 'iterations', '--rounds': 'rounds', '--output': 'output',
  '--wasm-opt': 'wasmOpt', '--artifacts': 'artifacts', '--adamic': 'adamic', '--node': 'node',
  '--timeout-ms': 'timeoutMs', '--build-timeout-ms': 'buildTimeoutMs',
};
try {
  const options = {};
  for (let index = 2; index < process.argv.length; index += 2) {
    const name = flags[process.argv[index]];
    const value = process.argv[index + 1];
    if (!name || value === undefined) throw new Error(`expected flag and value; flags: ${Object.keys(flags).join(', ')}`);
    options[name] = typeof defaults[name] === 'number' ? Number(value) : value;
  }
  console.log(markdown(await run(options)));
} catch (error) {
  if (error.report) console.log(markdown(error.report));
  console.error(error.message);
  process.exitCode = 1;
}
