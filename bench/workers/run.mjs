import { defaults, run, markdown } from './harness.mjs';
const names = {
  '--runner': 'runner', '--workerd': 'workerd', '--suite': 'suite', '--output': 'output',
  '--requests': 'requests', '--warmup': 'warmup', '--rounds': 'rounds', '--cold-spawns': 'coldSpawns',
  '--concurrency': 'concurrency', '--compatibility-date': 'compatibilityDate',
  '--timeout-ms': 'timeoutMs', '--idle-settle-ms': 'idleSettleMs', '--memory-poll-ms': 'memoryPollMs',
};
try {
  const options = {};
  for (let i = 2; i < process.argv.length; i += 2) {
    const name = names[process.argv[i]];
    const value = process.argv[i + 1];
    if (!name || value === undefined) throw new Error(`expected flag and value; flags: ${Object.keys(names).join(', ')}`);
    options[name] = name === 'concurrency' ? value.split(',').map(Number) : typeof defaults[name] === 'number' ? Number(value) : value;
  }
  console.log(markdown(await run(options)));
} catch (error) {
  if (error.report) console.log(markdown(error.report));
  console.error(error.message);
  process.exitCode = 1;
}
