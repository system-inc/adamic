"""Prove Mocha observation preserves original events and records selected cases."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

OBSERVER = Path(__file__).resolve().parents[1] / 'oracle/observe-tests.cjs'


class CaseObserverTests(unittest.TestCase):
    def test_observer_preserves_event_receiver_arguments_and_return(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            (root / 'node_modules/mocha').mkdir(parents=True)
            (root / 'package.json').write_text('{}')
            (root / 'node_modules/mocha/index.js').write_text('class Runner { emit(...args) { this.calls.push(args); return 123; } } module.exports = {Runner};')
            source = r'''
const assert = require('node:assert/strict');
const fs = require('node:fs');
process.argv[1] = '_mocha';
require(process.argv[2]);
const {Runner} = require('mocha');
const runner = new Runner(); runner.calls = [];
const test = {titlePath: () => ['', 'describe', 'it'], duration: 17};
const error = new Error('real error text');
assert.equal(runner.emit('pass', test), 123);
assert.equal(runner.emit('fail', test, error), 123);
assert.equal(runner.emit('pending', test), 123);
assert.equal(runner.emit('other', 42), 123);
assert.deepEqual(runner.calls, [['pass',test],['fail',test,error],['pending',test],['other',42]]);
const rows = fs.readFileSync(process.env.STAGE3_TEST_EVENTS,'utf8').trim().split('\n').map(JSON.parse);
assert.deepEqual(rows.map(row=>row.status), ['pass','fail','pending']);
assert.equal(rows[1].error, 'real error text');
assert.equal(rows[1].stack, error.stack);
assert.deepEqual(rows[0].name, test.titlePath());
'''
            script = root / 'probe.cjs'; script.write_text(source)
            result = subprocess.run(['node', str(script), str(OBSERVER)], cwd=root,
                env=dict(os.environ, STAGE3_TEST_EVENTS=str(root / 'events.jsonl')), capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)


if __name__ == '__main__':
    unittest.main()
