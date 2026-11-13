'use strict';
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const tree = path.resolve(process.argv[2]);
function read(file) { return fs.readFileSync(path.join(tree, file), 'utf8').replace(/\r\n/g, '\n'); }
function node(text, predicate) {
    const source = ts.createSourceFile('probe.ts', text, ts.ScriptTarget.Latest, true);
    const found = [];
    function visit(n) { if (predicate(n)) found.push(n); ts.forEachChild(n, visit); }
    visit(source);
    assert.equal(found.length, 1);
    return found[0].getText(source);
}
const sys = read('src/compiler/sys.ts');
const harness = read('src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts');
const poller = node(sys, n => ts.isFunctionDeclaration(n) && n.name?.text === 'createDynamicPriorityPollingWatchFile');
const callbacks = node(harness, n => ts.isClassDeclaration(n) && n.name?.text === 'Callbacks');
const timer = node(harness, n => ts.isMethodDeclaration(n) && n.name?.getText() === 'setTimeout');
const before = 'pollingIntervalQueue(pollingInterval).pollScheduled = host.setTimeout(pollingInterval === PollingInterval.Low ? pollLowPollingIntervalQueue : pollPollingIntervalQueue, pollingInterval, pollingInterval === PollingInterval.Low ? "pollLowPollingIntervalQueue" : "pollPollingIntervalQueue", pollingIntervalQueue(pollingInterval));';
assert.equal(poller.split(before).length, 2);
// Preserve assignment receiver, callee, callback, ms, and tails in that order.
const after = `const scheduledQueue = pollingIntervalQueue(pollingInterval);
        const timer = host.setTimeout;
        const callback = pollingInterval === PollingInterval.Low ? pollLowPollingIntervalQueue : pollPollingIntervalQueue;
        const ms = pollingInterval;
        const timeoutType = pollingInterval === PollingInterval.Low ? "pollLowPollingIntervalQueue" : "pollPollingIntervalQueue";
        const queue = pollingIntervalQueue(pollingInterval);
        scheduledQueue.pollScheduled = timer.call(host, () => callback(timeoutType, queue), ms);`;
function run(owner) {
    const source = `${callbacks}\n${owner}\nclass ProbeHost {
        timeoutCallbacks = new Callbacks(this, "Timeout");
        ${timer}
        getTime() { return 0; }
        getModifiedTime() { return new Date(0); }
    }
    const host = new ProbeHost();
    const watch = createDynamicPriorityPollingWatchFile(host);
    watch("/probe.ts", () => {}, PollingInterval.Low);
    result = host.timeoutCallbacks.log();`;
    const js = ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ES2024}}).outputText;
    const context = {
        PollingInterval: {Low:250, Medium:500, High:2000},
        getModifiedTime: (host, file) => host.getModifiedTime(file),
        result: undefined,
    };
    vm.runInNewContext(js, context);
    return context.result;
}
const original = run(poller);
const adapted = run(poller.replace(before, after));
assert.equal(original, 'Timeout callback:: count: 1\r\n1: pollLowPollingIntervalQueue');
assert.equal(adapted, 'Timeout callback:: count: 1\r\n1: undefined');
const result = {
    case: 'timer-label-baseline',
    sourceTree: tree,
    extraction: 'Complete real poller and Callbacks class, real TestServerHost.setTimeout method; minimal host clock and file stat only. No callback firing is needed to observe the difference.',
    ownerHashes: Object.fromEntries(Object.entries({poller, callbacks, timer}).map(([name,text]) => [name,crypto.createHash('sha256').update(text).digest('hex')])),
    beforeCall: before,
    afterCall: after,
    original,
    adapted,
    byteIdentical: original === adapted,
};
console.log(JSON.stringify(result,null,2));
// This command is the requested byte-identity assertion, so failure is exit 1.
if (original !== adapted) process.exitCode = 1;
