'use strict';
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const assert = require('node:assert/strict');
const ts = require('typescript');
const tree = path.resolve(process.argv[2]);
const mutant = process.argv[3];
function read(file) { return fs.readFileSync(path.join(tree, 'src', file), 'utf8').replace(/\r\n/g, '\n'); }
function extract(text, predicate) {
    const source = ts.createSourceFile('witness.ts', text, ts.ScriptTarget.Latest, true), found = [];
    function visit(node) { if (predicate(node)) found.push(node.getText(source)); ts.forEachChild(node, visit); }
    visit(source); assert.equal(found.length, 1); return found[0];
}
const fn = (text, name) => extract(text, n => ts.isFunctionDeclaration(n) && n.name?.text === name);
const method = (text, name) => extract(text, n => ts.isMethodDeclaration(n) && n.name?.getText() === name);
const harness = read('testRunner/unittests/helpers/virtualFileSystemWithWatch.ts');
const callbacks = extract(harness, n => ts.isClassDeclaration(n) && n.name?.text === 'Callbacks');
const timer = method(harness, 'setTimeout');
const fixture = `${callbacks}\nclass ProbeHost {
    timeoutCallbacks = new Callbacks(this, "Timeout");
    ${timer}
    time = 0;
    getTime() { return this.time; }
    setTime(time) { this.time = time; }
}`;
function run(source, context = {}) {
    const js = ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS}}).outputText;
    const sandbox = {exports: {}, result: undefined, arrayFrom: Array.from, Debug: {assert}, ...context};
    vm.runInNewContext(js, sandbox, {timeout: 10000});
    return JSON.parse(JSON.stringify(sandbox.result));
}
function replace(text, before, after, name) { assert.equal(text.split(before).length, 2, name + ': mutant target'); return text.replace(before, after); }
const sys = read('compiler/sys.ts');
let poller = fn(sys, 'createDynamicPriorityPollingWatchFile');
if (mutant === 'poller') poller = replace(poller, 'callback(t1, t2)', 'callback(t1, pollingIntervalQueue(pollingInterval))', 'poller-schedule-capture');
if (mutant === 'label') poller = replace(poller, 'timeoutDelay, t1);', 'timeoutDelay);', 'timer-label-baseline');
const pollingDeclarations = sys.slice(sys.indexOf('export enum PollingInterval'), sys.indexOf('function setCustomPollingValues'));
// Only the enum, level helper and level variables are needed from this region.
const pollingSetup = [
    extract(sys, n => ts.isEnumDeclaration(n) && n.name.text === 'PollingInterval'),
    fn(sys, 'createPollingIntervalBasedLevels'),
    'const defaultChunkLevels = {Low:32, Medium:64, High:256};',
    'let pollingChunkSize = createPollingIntervalBasedLevels(defaultChunkLevels);',
    'let unchangedPollThresholds = createPollingIntervalBasedLevels(defaultChunkLevels);',
    fn(sys, 'setCustomPollingValues'),
].join('\n');
const pollResult = run(`${fixture}\n${pollingSetup}\n${poller}
    const seen = [];
    const host = new ProbeHost();
    host.getModifiedTime = () => new Date(0);
    const watch = createDynamicPriorityPollingWatchFile(host);
    watch("/probe.ts", () => {}, PollingInterval.Low);
    const pending = host.timeoutCallbacks.log();
    setCustomPollingValues({getEnvironmentVariable: key => ({
        TSC_WATCH_POLLINGINTERVAL_LOW:"500", TSC_WATCH_POLLINGINTERVAL_MEDIUM:"250"
    })[key] || ""});
    host.timeoutCallbacks.invoke(1);
    result = {pending, visited:seen};`, {
    getModifiedTime: (host, name) => host.getModifiedTime(name),
    // Record the real poller's choice of queue; the stat loop is outside this
    // witness's capture contract and is replaced by an observable queue read.
    pollWatchedFileQueue: (host, queue) => {
        for (const file of queue) if (file) host.visited?.push(file.fileName);
        queue.length = 0; return 0;
    },
});
// Repeat with the observation connected to the host; this keeps vm state local.
// The first result is the unchanged blocker probe: the actual registry must
// retain the exact original label before the timer fires.
assert.equal(pollResult.pending, 'Timeout callback:: count: 1\r\n1: pollLowPollingIntervalQueue', 'timer-label-baseline');
const chosen = run(`${fixture}\n${pollingSetup}\n${poller}
    const host = new ProbeHost(); host.visited=[];
    host.getModifiedTime = () => new Date(0);
    const watch = createDynamicPriorityPollingWatchFile(host);
    watch("/probe.ts", () => {}, PollingInterval.Low);
    setCustomPollingValues({getEnvironmentVariable: key => ({
        TSC_WATCH_POLLINGINTERVAL_LOW:"500", TSC_WATCH_POLLINGINTERVAL_MEDIUM:"250"
    })[key] || ""});
    host.timeoutCallbacks.invoke(1);
    result = host.visited;`, {
    getModifiedTime: (host, name) => host.getModifiedTime(name),
    pollWatchedFileQueue: (host, queue) => { for (const file of queue) if (file) host.visited.push(file.fileName); queue.length = 0; return 0; },
});
assert.deepEqual(chosen, ['/probe.ts'], 'poller-schedule-capture');
let requestMethod = method(read('server/typingInstallerAdapter.ts'), 'scheduleRequest');
if (mutant === 'request') requestMethod = replace(requestMethod, 'callback(t1)', 'callback(`${request.projectName}::${request.kind}`)', 'request-schedule-capture');
const requestResult = run(`${fixture}\nclass Adapter {
    static requestDelayMillis = 100;
    host = new ProbeHost(); activeRequestCount=0;
    logger={hasLevel:()=>false};
    installer={send:request=>this.sent.push(JSON.stringify(request))}; sent=[];
    ${requestMethod}
}
const TypingsInstallerAdapter = Adapter;
const reads=[]; let name="before";
const request={get projectName(){reads.push(name);return name;},set projectName(value){name=value;},kind:"discover"};
const adapter=new Adapter(); adapter.scheduleRequest(request);
const pending=adapter.host.timeoutCallbacks.log();
request.projectName="after";
adapter.host.timeoutCallbacks.invoke(1);
result={pending,reads,sent:adapter.sent};`, {LogLevel:{verbose:3}});
assert.deepEqual(requestResult, {pending:'Timeout callback:: count: 1\r\n1: before::discover', reads:['before','after'], sent:['{"projectName":"after","kind":"discover"}']}, 'request-schedule-capture');
const watchSource = read('compiler/watchPublic.ts');
const watchResult = run(`${fixture}
let timerToUpdateProgram, reportFileChangeDetectedOnCreateProgram=false;
const events=[]; const host=new ProbeHost(); host.clearTimeout=id=>host.timeoutCallbacks.unregister(id);
const writeLog=message=>events.push(message); const updateProgram=()=>events.push("updated");
${fn(watchSource,'scheduleProgramUpdate')}
${fn(watchSource,'updateProgramWithWatchStatus')}
scheduleProgramUpdate(); const pending=host.timeoutCallbacks.log();
host.timeoutCallbacks.invoke(1);
result={pending,events,reportFileChangeDetectedOnCreateProgram};`);
assert.deepEqual(watchResult,{pending:'Timeout callback:: count: 1\r\n1: timerToUpdateProgram',events:['Scheduling update','updated'],reportFileChangeDetectedOnCreateProgram:true},'watch-host-label-and-tail');
console.log(JSON.stringify({poller:{pending:pollResult.pending,visited:chosen},request:requestResult,watchHost:watchResult},null,2));
