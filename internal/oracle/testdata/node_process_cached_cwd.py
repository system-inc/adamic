"""Remove a child's working directory between two explicit cwd calls."""
import base64
import json
import os
import subprocess
import sys
sys.dont_write_bytecode = True
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../../../oracle')))
from child_progress import progress, watch_files, communicate
import tempfile
import errno
import threading
import time

working = tempfile.mkdtemp(prefix='node-process-cached-')
environment = os.environ.copy()
environment['ASAN_OPTIONS'] = 'detect_leaks=0'
# A separate named FIFO is opened synchronously by the child. Unlike /dev/stdin,
# it does not share an inherited descriptor's platform-dependent nonblocking flags.
with tempfile.TemporaryDirectory(prefix='node-process-barrier-') as barrier_directory:
    barrier = os.path.join(barrier_directory, 'continue')
    os.mkfifo(barrier)
    child = subprocess.Popen(sys.argv[1:] + [barrier], cwd=working,
                             stdin=subprocess.DEVNULL, stdout=subprocess.PIPE,
                             stderr=subprocess.PIPE, env=environment)
    output, error = [], []
    ready = threading.Event()

    def drain(stream, pieces, announce=False):
        prefix = b''
        while True:
            part = os.read(stream.fileno(), 65536)
            if not part:
                return
            pieces.append(part)
            progress(part)
            if announce and not ready.is_set():
                prefix = (prefix + part)[:6]
                if prefix == b'ready\n':
                    ready.set()

    # One reader owns each pipe from start to EOF. Mixing buffered readline with
    # communicate loses read-ahead bytes; draining live also handles async output.
    readers = [threading.Thread(target=drain, args=(child.stdout, output, True)),
               threading.Thread(target=drain, args=(child.stderr, error))]
    for reader in readers:
        reader.start()
    writer = None
    try:
        if not ready.wait():
            raise RuntimeError('child did not announce ready')
        os.rmdir(working)
        while writer is None:
            try:
                writer = os.open(barrier, os.O_WRONLY | os.O_NONBLOCK)
            except OSError as failure:
                if failure.errno != errno.ENXIO:
                    raise
                if child.poll() is not None:
                    raise RuntimeError('child did not open the barrier FIFO') from failure
                time.sleep(0.001)
        os.write(writer, b'continue')
        os.close(writer)
        writer = None
        child.wait()
    finally:
        if writer is not None:
            os.close(writer)
        if child.poll() is None:
            child.kill()
        child.wait()
        for reader in readers:
            reader.join()
            if reader.is_alive():
                raise RuntimeError('child output did not reach EOF')
        child.stdout.close()
        child.stderr.close()
        if os.path.isdir(working):
            os.rmdir(working)
    print(json.dumps({'stdout': base64.b64encode(b''.join(output)).decode('ascii'),
                      'stderr': base64.b64encode(b''.join(error)).decode('ascii'),
                      'exitCode': child.returncode}))
