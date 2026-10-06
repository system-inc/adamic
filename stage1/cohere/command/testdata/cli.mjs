// Bind the typed module's result to the host status. The shared oracle runner strips source types
// and loads Adamic's runtime; no flag parsing or output decisions live in this adapter.
import { pathToFileURL } from 'node:url';
const source = process.argv[2];
const runner = process.argv[3];
process.argv.splice(3, 1);
await import(pathToFileURL(runner).href);
const result = await import(pathToFileURL(source).href);
process.exitCode = result.commandExitCode;
