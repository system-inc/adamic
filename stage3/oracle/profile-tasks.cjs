// Observe upstream IPC task results without changing scheduling or verdicts.
const fs = require('node:fs');
const childProcess = require('node:child_process');
const originalFork = childProcess.fork;
childProcess.fork = function (...args) {
    if (process.env.STAGE3_DISCOVERY_ONLY === '1') process.exit(0);
    const child = Reflect.apply(originalFork, this, args);
    if (process.env.STAGE3_TASK_PROFILE) {
        child.on('message', message => {
            if (['result', 'progress'].includes(message.type)) {
                try {
                    fs.appendFileSync(process.env.STAGE3_TASK_PROFILE, JSON.stringify(message.payload) + '\n');
                } catch (error) {
                    console.error('Stage 3 task observation failed:', error.message);
                }
            }
        });
    }
    return child;
};
// The host discovers only top-level unit suites. Source maps identify their files.
if (process.env.STAGE3_TASK_PROFILE && !process.send) {
    for (const name of ['describe', 'it']) {
        let registered;
        Object.defineProperty(global, name, {
            configurable: true,
            get() { return registered; },
            set(original) {
                if (typeof original !== 'function') { registered = original; return; }
                registered = function (...args) {
                    const savedLimit = Error.stackTraceLimit;
                    Error.stackTraceLimit = Infinity;
                    const stack = String(new Error().stack);
                    Error.stackTraceLimit = savedLimit;
                    const files = [...stack.matchAll(/(src\/testRunner\/unittests\/[^():]+\.ts):\d+:\d+/g)];
                    if (files.length) {
                        try {
                            fs.appendFileSync(process.env.STAGE3_TASK_PROFILE + '.sources', JSON.stringify({
                                title: args[0], file: files.at(-1)[1],
                            }) + '\n');
                        } catch (error) {
                            console.error('Stage 3 source observation failed:', error.message);
                        }
                    }
                    return Reflect.apply(original, this, args);
                };
                Object.assign(registered, original);
            },
        });
    }
}
