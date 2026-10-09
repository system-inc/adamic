// Observe original Mocha events, without replacing reporters or changing results.
const fs = require('node:fs');
const path = require('node:path');
const { createRequire } = require('node:module');
if (process.env.STAGE3_ERROR_DIR && ['run.js', '_mocha', 'mocha.js'].includes(path.basename(process.argv[1] || ''))) {
    const Mocha = createRequire(path.join(process.cwd(), 'package.json'))('mocha');
    const starts = new WeakMap();
    const originalRun = Mocha.Runnable.prototype.run;
    Mocha.Runnable.prototype.run = function (...args) {
        starts.set(this, process.hrtime.bigint());
        return Reflect.apply(originalRun, this, args);
    };
    const originalEmit = Mocha.Runner.prototype.emit;
    Mocha.Runner.prototype.emit = function (...args) {
        const [event, runnable, error] = args;
        // The parallel host replays failures on synthetic tests; record only real runs.
        if (event === 'fail' && starts.has(runnable)) {
            try {
                const stack = String(error.stack || '');
                const row = {
                    title: runnable.titlePath().filter(Boolean).join(' '),
                    kind: runnable.type,
                    message: String(error.message || error),
                    stack,
                    stack_top: stack.split('\n').find(line => /^\s*at /.test(line)) || null,
                    timeout_ms: runnable.timeout(),
                    elapsed_ms: Number(process.hrtime.bigint() - starts.get(runnable)) / 1e6,
                    pid: process.pid,
                };
                fs.appendFileSync(path.join(process.env.STAGE3_ERROR_DIR, `${process.pid}.jsonl`), JSON.stringify(row) + '\n');
            } catch (failure) {
                // Metadata failure cannot alter the upstream verdict.
                console.error('Stage 3 error observation failed:', failure.message);
            }
        }
        return Reflect.apply(originalEmit, this, args);
    };
}
