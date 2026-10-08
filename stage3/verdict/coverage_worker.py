"""Persistent Node CLI workers for the one complete coverage measurement."""
import json
import base64
import os
from pathlib import Path
import queue
import selectors
import subprocess

ROOT = Path(__file__).resolve().parent


class Workers:
    def __init__(self, bundle, output):
        self.bundle, self.output = bundle, output
        self.available = queue.Queue()
        self.processes = []
        self.logs = []

    def __enter__(self):
        self.output.mkdir()
        for index in range(int(os.environ.get('TSC_JOBS', '4'))):
            log = (self.output / f'{index}.stderr').open('wb')
            self.logs.append(log)
            process = subprocess.Popen(['node', str(ROOT / 'coverage_worker.cjs'), str(self.bundle)],
                                       stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=log, text=True)
            self.processes.append(process)
            self.available.put(process)
        return self

    def __exit__(self, *error):
        for process in self.processes:
            process.stdin.close()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
            process.stdout.close()
        for log in self.logs:
            log.close()

    def execute(self, argv, cwd, stdout, stderr):
        # Namespaced cases still use the exact ordinary instrumented CLI and
        # mount isolation. Persistent workers cannot cross mount namespaces.
        if Path(argv[0]).name == 'bwrap':
            capture = cwd / '.verdict-capture'
            capture.mkdir()
            position = argv.index('--chdir')
            argv = argv[:position] + ['--bind',str(capture),'/.verdict-capture'] + argv[position:]
            env = dict(os.environ,TSC_COVERAGE_FILE='/.verdict-capture/function-coverage.json')
            return subprocess.run(argv,cwd=cwd,stdout=stdout,stderr=stderr,env=env,
                                  timeout=float(os.environ.get('TSC_TIMEOUT','60')))
        process = self.available.get()
        try:
            process.stdin.write(json.dumps({'cwd':str(cwd),'args':argv[1:]})+'\n')
            process.stdin.flush()
            with selectors.DefaultSelector() as selector:
                selector.register(process.stdout,selectors.EVENT_READ)
                if not selector.select(float(os.environ.get('TSC_TIMEOUT','60'))):
                    process.kill()
                    raise RuntimeError('coverage worker timed out')
            response = json.loads(process.stdout.readline())
            if 'error' in response:
                raise RuntimeError(response['error'])
            stdout.write(base64.b64decode(response['stdout_base64']))
            stderr.write(response['stderr'].encode())
            return subprocess.CompletedProcess(argv,response['exit'])
        finally:
            self.available.put(process)
