// Call upstream's --tests/worker implementation without repeating build dependencies.
import path from 'node:path';
import { pathToFileURL } from 'node:url';
const tree = path.resolve(process.argv[2]);
const { default: options } = await import(pathToFileURL(path.join(tree, 'scripts/build/options.mjs')));
const { runConsoleTests } = await import(pathToFileURL(path.join(tree, 'scripts/build/tests.mjs')));
await runConsoleTests('./built/local/run.js', 'min', !options.tests && options.workers > 1);
