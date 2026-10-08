"""Remove a child's working directory between two explicit cwd calls."""
import base64
import json
import os
import subprocess
import sys
import tempfile

working = tempfile.mkdtemp(prefix='node-process-cached-')
environment = os.environ.copy()
environment['ASAN_OPTIONS'] = 'detect_leaks=0'
child = subprocess.Popen(sys.argv[1:], cwd=working, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE, env=environment)
ready = child.stdout.readline()
os.rmdir(working)
output, error = child.communicate(b'continue', timeout=15)
print(json.dumps({'stdout': base64.b64encode(ready + output).decode('ascii'), 'stderr': base64.b64encode(error).decode('ascii'), 'exitCode': child.returncode}))
