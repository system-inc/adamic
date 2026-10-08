"""Run the host fixture on a file, pipe or a TTY with a known width."""
import base64
import fcntl
import json
import os
import pty
import struct
import subprocess
import sys
import tempfile
import threading
import time
import termios

kind = sys.argv[1]
master = None
if kind == 'tty':
    master, slave = pty.openpty()
    fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack('HHHH', 24, 93, 0, 0))
    attributes = termios.tcgetattr(slave)
    attributes[1] &= ~termios.ONLCR
    termios.tcsetattr(slave, termios.TCSANOW, attributes)
    destination = os.fdopen(slave, 'wb')
elif kind == 'backpressure':
    reader, writer = os.pipe()
    fcntl.fcntl(writer, fcntl.F_SETPIPE_SZ, 4096)
    fcntl.fcntl(writer, fcntl.F_SETFL, fcntl.fcntl(writer, fcntl.F_GETFL) | os.O_NONBLOCK)
    destination = os.fdopen(writer, 'wb')
    pieces = []
    def drain():
        time.sleep(0.2)
        while True:
            piece = os.read(reader, 4096)
            if not piece:
                break
            pieces.append(piece)
        os.close(reader)
    reading = threading.Thread(target=drain)
    reading.start()
elif kind == 'pipe':
    destination = subprocess.PIPE
else:
    destination = tempfile.TemporaryFile()
environment = os.environ.copy()
environment['ASAN_OPTIONS'] = 'detect_leaks=0'
child = subprocess.Popen(sys.argv[2:], stdout=destination, stderr=subprocess.PIPE, env=environment)
if kind == 'tty':
    # Read the terminal while the child runs: macOS drops what a terminal holds once its
    # last writer closes, where Linux keeps it readable until EIO.
    destination.close()
    pieces = []
    def drain_terminal():
        while True:
            try:
                piece = os.read(master, 4096)
            except OSError:
                break
            if not piece:
                break
            pieces.append(piece)
    reading = threading.Thread(target=drain_terminal)
    reading.start()
output, error = child.communicate(timeout=15)
if kind == 'backpressure':
    destination.close()
    reading.join(timeout=15)
    if reading.is_alive():
        raise RuntimeError('pipe reader did not finish')
    output = b''.join(pieces)
elif kind == 'tty':
    reading.join(timeout=15)
    if reading.is_alive():
        raise RuntimeError('terminal reader did not finish')
    output = b''.join(pieces)
    os.close(master)
elif kind == 'file':
    destination.seek(0)
    output = destination.read()
    destination.close()
print(json.dumps({'stdout': base64.b64encode(output).decode('ascii'), 'stderr': base64.b64encode(error).decode('ascii'), 'exitCode': child.returncode}))
