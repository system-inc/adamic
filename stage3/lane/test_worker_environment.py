"""Match the unmodified upstream test entry's development assertion environment."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

DRIVER = Path(__file__).resolve().parents[1] / 'oracle/run-tasks.cjs'


class WorkerEnvironmentTests(unittest.TestCase):
    def test_original_development_assertions_are_enabled(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            tasks = root / 'tasks.json'; tasks.write_text('[{"runner":"compiler","file":"case.ts"}]')
            script = r'''
const assert = require('node:assert/strict');
const cp = require('node:child_process');
const {EventEmitter} = require('node:events');
const driver = process.argv[2], root = process.argv[3];
cp.fork = (script, args, options) => {
    assert.equal(options.env?.NODE_ENV, 'development', 'original normal assertions must be enabled');
    const child = new EventEmitter();
    child.kill = () => {};
    child.send = message => {
        if (message.type === 'test') process.nextTick(() => child.emit('message', {
            type:'result',payload:{task:message.payload,passes:[],errors:[],passing:0,duration:1},
        }));
        else process.nextTick(() => child.emit('exit',0,null));
    };
    return child;
};
process.argv = [process.execPath,driver,root,root+'/tasks.json',root+'/output','1'];
require(driver);
'''
            probe = root / 'probe.cjs'; probe.write_text(script)
            result = subprocess.run(['node', str(probe), str(DRIVER), str(root)],
                                    capture_output=True, text=True, env=dict(os.environ, NODE_ENV='production'))
        self.assertEqual(result.returncode, 0, result.stderr)


if __name__ == '__main__':
    unittest.main()
