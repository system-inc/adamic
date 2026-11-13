# Custom polling intervals disagree with inlined enum reads in the published bundle

Tested with TypeScript 6.0.3 from npm and Node v24.19.0 on Linux x64.

Save this as repro.cjs in a directory containing the published package:

```sh
npm install typescript@6.0.3
node repro.cjs
```

```js
process.env.TSC_WATCHFILE = 'DynamicPriorityPolling';
process.env.TSC_WATCH_POLLINGINTERVAL_LOW = '17';
process.env.TSC_WATCH_POLLINGINTERVAL_MEDIUM = '41';
process.env.TSC_WATCH_POLLINGINTERVAL_HIGH = '89';
const scheduled = [];
global.setTimeout = (callback, interval) => {
    scheduled.push(interval);
    return { unref() {} };
};
global.clearTimeout = () => {};
const ts = require('typescript');
const watchers = [];
console.log(ts.version);
console.log(ts.PollingInterval.Low, ts.PollingInterval.Medium, ts.PollingInterval.High);
for (const interval of [250, 500, 2000, 17, 41, 89]) {
    try {
        watchers.push(ts.sys.watchFile(__filename, () => {}, interval));
        console.log(interval, 'ok');
    } catch (error) {
        console.log(interval, error.name + ': ' + error.message);
    }
}
console.log(scheduled);
watchers.forEach(watcher => watcher.close());
```

Expected: the configured enum values and dynamic polling queue intervals agree.
Using the exported Low/Medium/High values should register watchers without
throwing, and those queues should schedule at 17/41/89.

Actual:

```text
6.0.3
17 41 89
250 ok
500 ok
2000 ok
17 TypeError: Cannot read properties of undefined (reading 'push')
41 TypeError: Cannot read properties of undefined (reading 'push')
89 TypeError: Cannot read properties of undefined (reading 'push')
[ 250, 500, 2000 ]
```

In src/compiler/sys.ts, setCustomPollingValues passes PollingInterval to
setCustomLevels, which mutates its runtime Low/Medium/High properties.
The published bundle retains that mutation but inlines enum member reads in
queue construction and selection to 250/500/2000. The exported runtime values
therefore disagree with the internal queue identities.

Transpiling the source modules with transpileModule to CommonJS, without enum
inlining, instead schedules 17/41/89 and throws for callers supplying the default
250/500/2000. The published _tsc.js shows the same default scheduling and
configured-caller exceptions as the published typescript.js when its CLI
dispatch is replaced with exports for inspection. No file watching callback
needs to run to reproduce the queue lookup failure.

This draft has not been filed.
