// Observe Mocha's selected cases; pass every event and argument through unchanged.
const fs = require('node:fs');
const path = require('node:path');
const { createRequire } = require('node:module');
if (process.env.STAGE3_TEST_EVENTS && ['run.js', '_mocha', 'mocha.js'].includes(path.basename(process.argv[1] || ''))) {
    const Mocha = createRequire(path.join(process.cwd(), 'package.json'))('mocha');
    const original = Mocha.Runner.prototype.emit;
    Mocha.Runner.prototype.emit = function (...args) {
        const [event, test, error] = args;
        if (['pass', 'fail', 'pending'].includes(event)) {
            try {
                fs.appendFileSync(process.env.STAGE3_TEST_EVENTS, JSON.stringify({
                    status: event, name: test.titlePath(), milliseconds: test.duration,
                    ...(event === 'fail' ? { error: String(error.message), stack: String(error.stack) } : {}),
                }) + '\n');
            } catch (failure) {
                console.error('Stage 3 test observation failed:', failure.message);
            }
        }
        return Reflect.apply(original, this, args);
    };
}
