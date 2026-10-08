// Run specified whole-file tasks through upstream's unchanged worker protocol.
const fs = require('node:fs');
const path = require('node:path');
const { fork } = require('node:child_process');
const [tree, tasksFile, output, workerCount = '1'] = process.argv.slice(2);
if (!output) throw new Error('usage: run-tasks.cjs TREE TASKS.json NEW-OUTPUT');
fs.mkdirSync(output, { recursive: false });
const started = process.hrtime.bigint();
const tasks = JSON.parse(fs.readFileSync(tasksFile));
const config = path.resolve(output, 'worker.json');
fs.writeFileSync(config, JSON.stringify({ listenForWork: true, runUnitTests: true }));
function startWorker() {
    const child = fork(path.resolve(tree, 'built/local/run.js'), [`--config=${config}`], {
        cwd: path.resolve(tree), stdio: ['ignore', 'inherit', 'inherit', 'ipc'],
    });
    let current;
    const timer = setTimeout(() => { failed = true; child.kill('SIGKILL'); }, 240000);
    child.on('message', message => {
        if (['result', 'progress'].includes(message.type)) {
            fs.appendFileSync(path.join(output, 'tasks.jsonl'), JSON.stringify(message.payload) + '\n');
            failed ||= message.payload.errors.length > 0;
            passing += message.payload.passing;
            failing += message.payload.errors.length;
            completed++;
            dispatch();
        } else if (message.type === 'error') {
            failed = true;
            fs.appendFileSync(path.join(output, 'errors.jsonl'), JSON.stringify(message) + '\n');
            child.kill('SIGKILL');
        }
    });
    child.on('exit', (code, signal) => {
        clearTimeout(timer);
        exited++;
        failed ||= code !== 0;
        if (exited !== Number(workerCount)) return;
        const report = { seconds: Number(process.hrtime.bigint() - started) / 1e9,
            counts: { passing: passing, failing: failing, pending: 0 },
            completed, requested: tasks.length, exit: code, signal, failed };
        const records = fs.existsSync(path.join(output, 'tasks.jsonl')) ?
            fs.readFileSync(path.join(output, 'tasks.jsonl'), 'utf8').trim().split('\n').filter(Boolean).map(JSON.parse) : [];
        const failures = records.flatMap(row => row.errors);
        fs.writeFileSync(path.join(output, 'mocha-errors.txt'), failures.map(error =>
            error.name.filter(Boolean).join(' ') + '\n' + error.error + '\n' + error.stack).join('\n\n'));
        fs.writeFileSync(path.join(output, 'report.json'), JSON.stringify(report, null, 2) + '\n');
        process.exitCode = failed || code !== 0 || completed !== tasks.length ? 1 : 0;
    });
    function dispatch() {
        current = tasks[next++];
        child.send(current ? { type: 'test', payload: current } : { type: 'close' });
    }
    dispatch();
}
let completed = 0, next = 0, exited = 0, passing = 0, failing = 0, failed = false;
if (!Number.isInteger(Number(workerCount)) || Number(workerCount) < 1) throw new Error('invalid workers');
for (let i = 0; i < Number(workerCount); i++) startWorker();
