"""Hold observational instrumentation to upstream's original event API."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

PROFILE = Path(__file__).resolve().parents[1] / 'oracle/profile-tasks.cjs'


class ProfileTests(unittest.TestCase):
    def observe(self, source, output):
        script = r'''
const assert = require('node:assert/strict');
const fs = require('node:fs');
const events = require('node:events');
const cp = require('node:child_process');
const child = new events.EventEmitter();
const calls = [];
cp.fork = function(...args) { calls.push([this, args]); return child; };
process.env.STAGE3_TASK_PROFILE = process.argv[2];
require(process.argv[1]);
const receiver = {};
assert.equal(cp.fork.call(receiver, 'runner', ['arg'], {cwd:'tree'}), child);
assert.deepEqual(calls, [[receiver, ['runner', ['arg'], {cwd:'tree'}]]]);
const payload = {task:{runner:'compiler',file:'case'},duration:123,passes:[{name:['case','baseline']}],errors:[],passing:1};
child.emit('message', {type:'timeout',payload:{duration:40000}});
child.emit('message', {type:'result',payload});
assert.ok(fs.existsSync(process.argv[2]), 'upstream task observation was dropped');
assert.deepEqual(fs.readFileSync(process.argv[2], 'utf8'), JSON.stringify(payload)+'\n');
const original = function(...args) { calls.push([this, args]); return 123; };
original.skip = 'original property';
for (const name of ['describe', 'it']) {
    global[name] = original;
    assert.equal(global[name].call(receiver, 'title', 456), 123);
    assert.deepEqual(calls.at(-1), [receiver, ['title',456]]);
    assert.equal(global[name].skip, 'original property');
}
'''
        return subprocess.run(['node', '-e', script, str(source), str(output)],
                              capture_output=True, text=True,
                              env=dict(os.environ, STAGE3_DISCOVERY_ONLY='0'))

    def test_original_arguments_events_and_results_are_preserved(self):
        with tempfile.TemporaryDirectory() as scratch:
            result = self.observe(PROFILE, Path(scratch) / 'profile')
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_dropped_task_observation_mutant(self):
        with tempfile.TemporaryDirectory() as scratch:
            mutant = Path(scratch) / 'profile.cjs'
            source = PROFILE.read_text()
            before = "if (['result', 'progress'].includes(message.type))"
            self.assertEqual(source.count(before), 1)
            mutant.write_text(source.replace(before, 'if (false)'))
            result = self.observe(mutant, Path(scratch) / 'profile')
        self.assertEqual(result.returncode, 1)
        self.assertIn('AssertionError', result.stderr)
        self.assertIn('upstream task observation was dropped', result.stderr)
        print('caught dropped IPC observation by original-task payload assertion')
