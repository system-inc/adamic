// Node exposes the status a CLI must return. Stage 0 has no process binding yet.
declare const process: { exitCode: number };
process.exitCode = 7;
