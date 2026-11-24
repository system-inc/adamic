"""Linux compiler supervisor. Return status only after all descendants exit."""
import ctypes
import errno
import json
import os
from pathlib import Path
import signal
from select import poll, POLLIN
import subprocess
import sys
import time
from tempfile import NamedTemporaryFile


PR_SET_CHILD_SUBREAPER = 36
CLEANUP_SECONDS = 5
OWNER_DEADLINE_EXIT = 3
OWNER_DEADLINE_MESSAGE = 'cleanup owner refused ownership because its deadline passed'


def process_identity(process):
    stat = (process / 'stat').read_text(errors='replace')
    fields = stat.rsplit(')', 1)[1].split()
    return int(fields[1]), int(fields[19])


def process_start_time(pid):
    return process_identity(Path(f'/proc/{pid}'))[1]


def verified_process_start_time(pid, *, allow_thread_reuse=False):
    """Bind the numeric procfs lookup to a task in the caller's namespace."""
    try:
        fd = os.pidfd_open(pid)
    except ProcessLookupError:
        return None
    except OSError as error:
        if allow_thread_reuse and error.errno in (errno.EINVAL, errno.ENOENT):
            prove_non_leader_thread(pid)
            return None
        if error.errno == errno.ENOSYS:
            raise ValueError('the host has no pidfds (ENOSYS)') from error
        if error.errno == errno.EMFILE:
            raise ValueError('the process reached its descriptor limit (EMFILE)') from error
        raise ValueError(f'cannot open a process handle for PID {pid}: {error}') from error
    try:
        with os.fdopen(fd, 'rb', buffering=0):
            info = Path(f'/proc/self/fdinfo/{fd}').read_text()
            if f'Pid:\t{pid}' not in info.splitlines():
                raise ValueError(f'procfs PID view does not match process handle for PID {pid}')
            start_time = process_start_time(pid)
            watcher = poll()
            watcher.register(fd, POLLIN)
            if watcher.poll(0):
                raise ValueError(f'process handle for PID {pid} exited during stat verification')
            return start_time
    except (OSError, IndexError) as error:
        raise ValueError(f'owner PID exists but its procfs identity cannot be read: {error}') from error


def prove_non_leader_thread(pid):
    # The receipt's PID was a thread group ID. Bind this procfs view to
    # the reader before trusting the group ID of a task at that number.
    if verified_process_start_time(os.getpid()) is None:
        raise ValueError('cannot prove the reader procfs PID view')
    try:
        status = Path(f'/proc/{pid}/status').read_text(errors='replace')
    except OSError as error:
        raise ValueError(f'cannot read status for PID {pid}: {error}') from error
    groups = [line.removeprefix('Tgid:').strip() for line in status.splitlines()
              if line.startswith('Tgid:')]
    if len(groups) != 1 or not groups[0].isascii() or not groups[0].isdecimal():
        raise ValueError(f'status for PID {pid} has no valid Tgid')
    group = int(groups[0])
    if group <= 0:
        raise ValueError(f'status for PID {pid} has no valid Tgid')
    if group == pid:
        raise ValueError(f'status Tgid equals PID {pid}; process identity cannot be proven')


def read_boot_id():
    try:
        boot_id = Path('/proc/sys/kernel/random/boot_id').read_text().strip()
    except OSError as error:
        raise ValueError(f'cannot read boot ID: {error}') from error
    if not boot_id:
        raise ValueError('cannot read boot ID: empty value')
    return boot_id


def read_pid_namespace():
    try:
        return os.stat('/proc/self/ns/pid').st_ino
    except OSError as error:
        raise ValueError(f'cannot read PID namespace: {error}') from error


def read_owner_receipt(receipt):
    owner = json.loads(receipt.read_text())
    if not isinstance(owner, dict):
        raise ValueError('receipt is not an object')
    if owner.get('state') == 'expired':
        expired_at = owner.get('expired_at')
        if type(expired_at) is not int or expired_at <= 0:
            raise ValueError('expired receipt lacks an expiry time')
        return owner, 'expired'
    pid = owner.get('pid')
    start_time = owner.get('start_time')
    acknowledged_at = owner.get('acknowledged_at')
    boot_id = owner.get('boot_id')
    pid_namespace = owner.get('pid_namespace')
    if (type(pid) is not int or pid <= 0 or type(start_time) is not int or start_time < 0
            or type(acknowledged_at) is not int or acknowledged_at <= 0
            or not isinstance(boot_id, str) or not boot_id.strip()
            or type(pid_namespace) is not int or pid_namespace <= 0
            or owner.get('state') not in ('owned', 'uncertain', 'released')):
        raise ValueError('receipt lacks a recognized owner identity or state')
    if read_boot_id() != boot_id:
        return owner, 'released' if owner['state'] == 'released' else 'dead'
    if read_pid_namespace() != pid_namespace:
        raise ValueError('receipt PID namespace differs from the reader PID namespace')
    if owner['state'] == 'released':
        return owner, 'released'
    current_start_time = verified_process_start_time(pid, allow_thread_reuse=True)
    if current_start_time is None:
        return owner, 'dead'
    return owner, 'alive' if current_start_time == start_time else 'dead'


def write_owner_receipt(receipt, owner, deadline=None):
    # Stage the bytes before checking the deadline at the publication boundary.
    with NamedTemporaryFile(mode='w', dir=receipt.parent, delete=False) as stream:
        temporary = Path(stream.name)
        try:
            json.dump(owner, stream)
            stream.write('\n')
            stream.close()
            if deadline is None:
                temporary.replace(receipt)
            else:
                if time.monotonic() >= deadline:
                    return False
                try:
                    os.link(temporary, receipt)
                except FileExistsError:
                    return False
            return True
        finally:
            temporary.unlink(missing_ok=True)


def reap_descendants(deadline):
    while True:
        try:
            pid, _ = os.waitpid(-1, os.WNOHANG)
        except ChildProcessError:
            pid = None
        if deadline is not None and time.monotonic() >= deadline:
            raise TimeoutError('compiler descendants did not terminate before the deadline')
        if pid is None:
            return
        if pid:
            continue

        # The supervisor adopts orphans, including those in another session.
        # Kill only its current children. As they exit, their children become
        # ours and are handled by the next pass until waitpid reports ECHILD.
        parent = os.getpid()
        if os.readlink('/proc/self') != str(parent):
            raise RuntimeError('procfs does not show the supervisor PID')
        for process in Path('/proc').iterdir():
            if not process.name.isdecimal() or not process.name.isascii():
                continue
            try:
                fd = os.pidfd_open(int(process.name))
                try:
                    # fdinfo names this handle's PID in the procfs view.
                    info = Path(f'/proc/self/fdinfo/{fd}').read_text()
                    if f'Pid:\t{process.name}' not in info.splitlines():
                        continue
                    owner, start_time = process_identity(process)
                    if owner != parent or start_time < 0:
                        continue
                    # A live pidfd after the stat read binds that snapshot
                    # to its original task. An exited task cannot be signalled
                    # through this handle even if its numeric PID is reused.
                    watcher = poll()
                    watcher.register(fd, POLLIN)
                    if not watcher.poll(0):
                        signal.pidfd_send_signal(fd, signal.SIGKILL)
                finally:
                    os.close(fd)
            except (IndexError, ValueError, FileNotFoundError, PermissionError, ProcessLookupError):
                pass
        time.sleep(0.01)


def main():
    deadline, cleanup_deadline, output_fd, *argv = sys.argv[1:]
    deadline, cleanup_deadline = float(deadline), float(cleanup_deadline)
    libc = ctypes.CDLL(None, use_errno=True)
    if libc.prctl(PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0) != 0:
        error = ctypes.get_errno()
        raise OSError(error, os.strerror(error))
    with os.fdopen(int(output_fd), 'wb') as stdout:
        code = None
        if time.monotonic() < deadline:
            child = subprocess.Popen(argv, stdout=stdout)
            try:
                try:
                    code = child.wait(timeout=max(0, deadline - time.monotonic()))
                except subprocess.TimeoutExpired:
                    pass
                if time.monotonic() >= deadline:
                    code = None
            finally:
                try:
                    reap_descendants(min(cleanup_deadline, time.monotonic() + CLEANUP_SECONDS))
                except TimeoutError:
                    # The caller enforces its own deadline. Keep ownership
                    # until the remaining writers have actually stopped.
                    reap_descendants(None)
                    raise
                finally:
                    child.poll()
    print('null' if code is None else code)


def own_supervisor(receipt, deadline, cleanup_deadline, output_fd, argv):
    receipt = Path(receipt)
    admitted = False
    code = None
    failed = False
    verified = False
    try:
        pid = os.getpid()
        start_time = verified_process_start_time(pid)
        if start_time is None:
            raise ValueError('cannot open the cleanup owner process handle')
        owner = {'pid': pid, 'start_time': start_time,
                 'boot_id': read_boot_id(), 'pid_namespace': read_pid_namespace(),
                 'acknowledged_at': time.time_ns(), 'state': 'owned'}
        admitted = write_owner_receipt(receipt, owner, float(deadline))
        if not admitted:
            print(OWNER_DEADLINE_MESSAGE, file=sys.stderr, flush=True)
            return OWNER_DEADLINE_EXIT
        command = [sys.executable, '-B', str(Path(__file__).resolve()),
                   deadline, cleanup_deadline, output_fd, *argv]
        supervisor = subprocess.Popen(command, stdout=subprocess.PIPE,
                                      pass_fds=(int(output_fd),))
        try:
            status, _ = supervisor.communicate(
                timeout=max(0, float(cleanup_deadline) - time.monotonic()))
        except Exception as error:
            failed = True
            read_failed = not isinstance(error, subprocess.TimeoutExpired)
            try:
                supervisor.send_signal(signal.SIGCONT)
            finally:
                try:
                    supervisor.stdout.close()
                finally:
                    # Keep ownership even when the status channel or its
                    # close fails. Only exit 0 attests to subtree cleanup.
                    exit_code = supervisor.wait()
            if read_failed:
                print(f'supervisor status read failed: {error}', file=sys.stderr, flush=True)
            verified = exit_code == 0 and not read_failed
        else:
            if supervisor.returncode == 0:
                code = json.loads(status)
                verified = True
            else:
                failed = True
    except Exception as error:
        print(f'cleanup owner failed: {error}', file=sys.stderr, flush=True)
        failed = True
    finally:
        os.close(int(output_fd))
        os.close(2)
        if admitted:
            write_owner_receipt(receipt, dict(owner, state='released' if verified else 'uncertain'))
    try:
        os.write(sys.stdout.fileno(), (json.dumps(code) + '\n').encode())
    except BrokenPipeError:
        pass  # Cleanup has completed or been marked uncertain.
    return int(failed)


if __name__ == '__main__':
    if sys.argv[1] == '--owner':
        raise SystemExit(own_supervisor(*sys.argv[2:6], sys.argv[6:]))
    main()
