// TypeScript 6.0.3 sys.ts shapes: cwd memoize (1501), executing path (1497),
// env (getEnvironmentVariable), newLine/os.EOL, write/process.exit (fixtures 17-20).
// Clocks mirror performanceCore.ts' timestamp and sys.ts' memoryUsage().heapUsed.
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import {performance} from 'node:perf_hooks';
const mode = process.argv[2];
const log = value => console.log(String(value));
if (mode === 'output') {
    process.stdout.write('Version 6.0.3' + os.EOL);
    const write = process.stdout.write.bind(process.stdout);
    write('héllo 🌍\u0000\uD800' + os.EOL);
    process.stderr.write('diagnostic' + os.EOL);
} else if (mode === 'exit') {
    process.exitCode = 7;
    console.log('buffered');
    fs.writeSync(1, 'stdio');
    fs.writeSync(2, 'stderr');
    try { process.exit(Number(process.argv[3])); }
    finally { console.log('finally'); }
} else if (mode === 'status') {
    log(process.exitCode === undefined);
    process.exitCode = 2;
    log(process.exitCode);
    process.exitCode = undefined;
    log(process.exitCode === undefined);
    console.log('continues');
} else if (mode === 'natural') {
    process.exitCode = 2;
    console.log('natural cleanup');
} else if (mode === 'clocks') {
    const origin = performance.timeOrigin;
    const first = performance.now();
    let fractional = !Number.isInteger(first);
    let ordered = true, previous = first;
    for (let i = 0; i < 10000; i++) {
        const now = performance.now();
        ordered &&= now >= previous;
        fractional ||= !Number.isInteger(now);
        previous = now;
    }
    log(ordered); log(fractional);
    log(previous > first);
    log(origin === performance.timeOrigin);
    log(Math.abs(origin + performance.now() - Date.now()) < 1000);
    log(Number.isInteger(Date.now()));
} else if (mode === 'memory') {
    const first = process.memoryUsage();
    const bytes = new Uint8Array(4 * 1024 * 1024);
    bytes.fill(42);
    const view = bytes.subarray(1);
    const now = process.memoryUsage();
    log(Object.keys(now).join(','));
    log(Object.values(now).every(n => Number.isFinite(n) && n >= 0));
    log(now.rss > 0 && now.heapTotal >= now.heapUsed);
    log(now.arrayBuffers - first.arrayBuffers >= bytes.byteLength);
    log(now.external >= now.arrayBuffers);
    log(view[0] === 42);
} else if (mode === 'identity') {
    log(path.isAbsolute(process.execPath));
    log(process.execPath === process.execPath);
    log(process.argv[0] === process.execPath);
    log(process.argv.slice(3).join('|'));
    log(process.execArgv.length === 0);
    // Bundle identity is the script on Node, the image natively; lib files share its directory.
    log(fs.existsSync(path.join(path.dirname(process.argv[1]), 'lib.d.ts')));
    log(process.cwd() === process.env.HOST_EXPECT_CWD);
} else if (mode === 'cwd') {
    let value, callback = () => process.cwd();
    const getCurrentDirectory = () => {
        if (callback) { value = callback(); callback = undefined; }
        return value;
    };
    log(getCurrentDirectory() === process.cwd());
    process.chdir(process.env.HOST_CHILD);
    log(getCurrentDirectory() !== process.cwd());
    log(process.cwd() === process.env.HOST_CHILD);
} else if (mode === 'environment') {
    const getEnvironmentVariable = name => process.env[name] || '';
    log(process.env.HOST_MISSING === undefined);
    log(getEnvironmentVariable('HOST_MISSING') === '');
    log(process.env.HOST_EMPTY === '');
    log(getEnvironmentVariable('HOST_VALUE'));
    const before = process.env.HOST_VALUE;
    process.env.HOST_VALUE = 'changed'; // External host mutation, not admitted Adamic code.
    log(getEnvironmentVariable('HOST_VALUE'));
    log(before);
} else if (mode === 'eol') {
    log(JSON.stringify(os.EOL));
    process.stdout.write('Version 6.0.3' + os.EOL + 'diagnostic' + os.EOL);
} else { throw Error('unknown mode'); }
