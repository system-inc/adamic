// Diagnostic for the original stdin barrier, run with the pre-fix Python driver.
// Report the exact fs errno caught by readTextFile, without changing that result.
import fs from 'node:fs';
import { syncBuiltinESMExports } from 'node:module';
import { readTextFile } from '../../../oracle/adamic.mjs';
const original = fs.readFileSync;
fs.readFileSync = function(path, ...options) {
    try { return original.call(this, path, ...options); }
    catch (error) {
        if (path === '/dev/stdin') {
            process.stderr.write(JSON.stringify({ code: error.code, message: error.message }) + '\n');
        }
        throw error;
    }
};
syncBuiltinESMExports();
const first = process.cwd();
process.stdout.write('ready\n');
const signal = readTextFile('/dev/stdin');
process.stderr.write(JSON.stringify({ kind: signal.kind }) + '\n');
console.log('cached ' + (process.cwd() === first));
