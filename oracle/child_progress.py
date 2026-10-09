"""Report child output activity without changing a driver's JSON observation."""
import os
import threading
import time


def progress(data):
    descriptor = os.environ.get('ADAMIC_CHILDGUARD_PROGRESS_FD')
    if data and descriptor is not None:
        os.write(int(descriptor), b'.')


def watch_files(child, streams):
    # Regular files keep their original descriptor semantics. Only size changes
    # count as progress; polling a silent child never sends a keepalive.
    def watch():
        sizes = [0 for _ in streams]
        while child.poll() is None:
            for index, stream in enumerate(streams):
                try:
                    size = os.fstat(stream.fileno()).st_size
                except (OSError, ValueError, AttributeError):
                    continue
                if size != sizes[index]:
                    sizes[index] = size
                    progress(b'.')
            time.sleep(0.1)
    thread = threading.Thread(target=watch, daemon=True)
    thread.start()
    return thread


def communicate(child):
    pieces = [[], []]
    readers = []
    def drain(stream, output):
        while True:
            data = os.read(stream.fileno(), 65536)
            if not data:
                return
            output.append(data)
            progress(data)
    for stream, output in zip((child.stdout, child.stderr), pieces):
        if stream is not None:
            thread = threading.Thread(target=drain, args=(stream, output))
            thread.start()
            readers.append(thread)
    child.wait()
    for reader in readers:
        reader.join()
    return tuple(b''.join(output) for output in pieces)
