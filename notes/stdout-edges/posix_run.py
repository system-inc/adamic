# posix_run.py is the part of the shell harness that needs a pipe, a non-blocking
# descriptor, a pseudoterminal, or a file sampled while the process is still running.
# The shell starts it. Exit status, stdout and stderr are written into the directory
# named by the first argument after the mode.
import fcntl
import os
import pty
import select
import signal
import subprocess
import sys
import termios
import time
import threading

mode, directory, *command = sys.argv[1:]
os.makedirs(directory, exist_ok=True)


def cpu_seconds(pid):
 if sys.platform.startswith("linux"):
  try:
   data = open(f"/proc/{pid}/stat", encoding="utf-8").read()
   fields = data[data.rfind(")") + 2:].split()
   return (int(fields[11]) + int(fields[12])) / os.sysconf("SC_CLK_TCK")
  except (OSError, ValueError, IndexError):
   return 0
 try:
  text = subprocess.check_output(["ps", "-o", "time=", "-p", str(pid)], timeout=2).decode().strip()
  minutes, seconds = text.split(":")
  return int(minutes) * 60 + float(seconds)
 except (OSError, ValueError, subprocess.SubprocessError):
  return 0


def ready(process):
 # Only the spin after the fixed output can spend this CPU budget. Starvation
 # delays readiness; it cannot imitate execution. Drain pipes concurrently.
 deadline = time.monotonic() + 15
 while process.poll() is None and cpu_seconds(process.pid) < 0.5:
  if time.monotonic() >= deadline:
   process.kill()
   process.wait(timeout=2)
   raise TimeoutError("program never reached its spin")
  time.sleep(0.01)
 return process.poll() is None


def collector(process):
 result = []
 thread = threading.Thread(target=lambda: result.append(process.communicate()), daemon=True)
 thread.start()
 return thread, result


def write_result(exit_code, stdout, stderr, note="", while_running=None):
	with open(os.path.join(directory, "stdout"), "wb") as handle:
		handle.write(stdout)
	with open(os.path.join(directory, "stderr"), "wb") as handle:
		handle.write(stderr)
	with open(os.path.join(directory, "exit"), "w", encoding="utf-8") as handle:
		handle.write(str(exit_code))
	if note:
		with open(os.path.join(directory, "note"), "w", encoding="utf-8") as handle:
			handle.write(note)
	if while_running is not None:
		with open(os.path.join(directory, "while-running"), "wb") as handle:
			handle.write(while_running)


def finish(process, stdout, stderr, note=""):
	code = process.returncode if process.returncode is not None else 999
	write_result(code, stdout, stderr, note)


def nonblock(kind):
	reader, writer = os.pipe()
	err_read, err_write = os.pipe()
	flags = fcntl.fcntl(writer, fcntl.F_GETFL)
	fcntl.fcntl(writer, fcntl.F_SETFL, flags | os.O_NONBLOCK)
	process = subprocess.Popen(command, stdout=writer, stderr=err_write, stdin=subprocess.DEVNULL)
	os.close(writer)
	os.close(err_write)
	chunks = []
	errors = []
	started = time.monotonic()
	idle_since = None
	limit = 8 if kind == "nonblock-timed" else 20
	watch = [reader, err_read]
	try:
		while watch and time.monotonic() - started < limit:
			ready, _, _ = select.select(watch, [], [], 0.2)
			if not ready:
				if process.poll() is not None:
					break
				if kind == "nonblock-spin":
					idle_since = time.monotonic() if idle_since is None else idle_since
					if cpu_seconds(process.pid) >= 0.5:
						break
				if kind == "nonblock-timed" and cpu_seconds(process.pid) >= 0.5:
					break
				continue
			idle_since = None
			for descriptor in ready:
				size = 4096 if kind == "nonblock" else 1 << 16
				try:
					piece = os.read(descriptor, size)
				except OSError:
					watch.remove(descriptor)
					continue
				if not piece:
					watch.remove(descriptor)
					continue
				(chunks if descriptor == reader else errors).append(piece)
				if kind == "nonblock":
					time.sleep(0.0005)
		if process.poll() is None:
			process.send_signal(signal.SIGTERM)
			note = "stopped"
		else:
			note = ""
		try:
			process.wait(timeout=2)
		except subprocess.TimeoutExpired:
			process.kill()
			process.wait(timeout=2)
			note = "killed"
		# The writer is gone. Drain whatever the slow reader has not taken.
		deadline = time.monotonic() + 2
		while watch and time.monotonic() < deadline:
			ready, _, _ = select.select(watch, [], [], 0.05)
			if not ready:
				break
			for descriptor in ready:
				piece = os.read(descriptor, 1 << 16)
				if not piece:
					watch.remove(descriptor)
					continue
				(chunks if descriptor == reader else errors).append(piece)
		finish(process, b"".join(chunks), b"".join(errors), note)
	finally:
		os.close(reader)
		os.close(err_read)
		if process.poll() is None:
			process.kill()
			process.wait()


def open_pty():
	leader, follower = pty.openpty()
	attributes = termios.tcgetattr(follower)
	attributes[0] = 0
	attributes[1] = 0
	attributes[3] = 0
	termios.tcsetattr(follower, termios.TCSANOW, attributes)
	process = subprocess.Popen(command, stdout=follower, stderr=subprocess.PIPE, stdin=subprocess.DEVNULL)
	os.close(follower)
	fcntl.fcntl(leader, fcntl.F_SETFL, fcntl.fcntl(leader, fcntl.F_GETFL) | os.O_NONBLOCK)
	chunks = []
	errors = []
	err = process.stderr
	watch = [leader]
	if err is not None:
		fcntl.fcntl(err.fileno(), fcntl.F_SETFL, fcntl.fcntl(err.fileno(), fcntl.F_GETFL) | os.O_NONBLOCK)
		watch.append(err.fileno())
	deadline = time.monotonic() + 15
	try:
		while time.monotonic() < deadline:
			ready, _, _ = select.select(watch, [], [], 0.2)
			for descriptor in ready:
				try:
					piece = os.read(descriptor, 1 << 16)
				except OSError:
					continue
				if not piece:
					continue
				(chunks if descriptor == leader else errors).append(piece)
			if process.poll() is not None:
				# Drain what the terminal and the stderr pipe still hold.
				for _ in range(20):
					ready, _, _ = select.select(watch, [], [], 0.05)
					if not ready:
						break
					for descriptor in ready:
						try:
							piece = os.read(descriptor, 1 << 16)
						except OSError:
							piece = b""
						if not piece:
							continue
						(chunks if descriptor == leader else errors).append(piece)
				break
		if process.poll() is None:
			process.kill()
			process.wait(timeout=2)
			note = "killed"
		else:
			process.wait(timeout=2)
			note = "pty"
		stdout = b"".join(chunks).replace(b"\r\n", b"\n").replace(b"\r", b"\n")
		finish(process, stdout, b"".join(errors), note)
	finally:
		os.close(leader)
		if process.poll() is None:
			process.kill()
			process.wait()


def file_while_running():
	stdout_path = os.path.join(directory, "stdout")
	with open(stdout_path, "wb") as stdout_file:
		process = subprocess.Popen(command, stdout=stdout_file, stderr=subprocess.PIPE, stdin=subprocess.DEVNULL)
		thread, collected = collector(process)
		ready(process)
		running = open(stdout_path, "rb").read()
		note = "finished"
		if process.poll() is None:
			note = "running"
			process.send_signal(signal.SIGTERM)
		thread.join(timeout=3)
		if thread.is_alive():
			process.kill()
			thread.join(timeout=2)
			note = "killed"
		_, stderr = collected[0]

		final = open(stdout_path, "rb").read()
		write_result(process.returncode if process.returncode is not None else 999, final, stderr, note, running)


def signaled(signal_name, trapped):
	number = signal_name if signal_name.startswith("SIG") else "SIG" + signal_name
	signum = getattr(signal, number)
	if trapped:
		wrapped = ["bash", "-c", 'ulimit -c 0; trap "" INT HUP; exec "$@"', command[0], *command]
	else:
		wrapped = command

	def drop_core():
		import resource
		resource.setrlimit(resource.RLIMIT_CORE, (0, 0))

	process = subprocess.Popen(
		wrapped,
		stdout=subprocess.PIPE,
		stderr=subprocess.PIPE,
		stdin=subprocess.DEVNULL,
		preexec_fn=None if trapped else drop_core,
	)
	thread, collected = collector(process)
	if not ready(process):
		thread.join(timeout=2)
		stdout, stderr = collected[0]
		write_result(process.returncode, stdout, stderr, "early")
		return
	process.send_signal(signum)
	thread.join(timeout=2)
	if thread.is_alive():
		process.kill()
		thread.join(timeout=2)
		note = "survived"
	else:
		note = "signaled"
	stdout, stderr = collected[0]
	write_result(process.returncode if note == "signaled" else 999, stdout, stderr, note)


def ulimit(which):
	stdout_path = os.path.join(directory, "stdout")
	stderr_path = os.path.join(directory, "stderr")
	shell = 'ulimit -f 100; ulimit -c 0; exec "$@"'
	wrapped = ["bash", "-c", shell, command[0], *command]
	stdout_target = open(stdout_path, "wb") if which in ("ulimit-out", "ulimit-both") else subprocess.PIPE
	stderr_target = open(stderr_path, "wb") if which in ("ulimit-err", "ulimit-both") else subprocess.PIPE
	process = subprocess.Popen(wrapped, stdout=stdout_target, stderr=stderr_target, stdin=subprocess.DEVNULL)
	if hasattr(stdout_target, "close") and which in ("ulimit-out", "ulimit-both"):
		stdout_target.close()
	if hasattr(stderr_target, "close") and which in ("ulimit-err", "ulimit-both"):
		stderr_target.close()
	try:
		out, err = process.communicate(timeout=20)
	except subprocess.TimeoutExpired:
		process.kill()
		out, err = process.communicate()
	stdout = open(stdout_path, "rb").read() if which in ("ulimit-out", "ulimit-both") else out
	stderr = open(stderr_path, "rb").read() if which in ("ulimit-err", "ulimit-both") else err
	write_result(process.returncode if process.returncode is not None else 999, stdout, stderr, which)


if mode == "nonblock":
	nonblock("nonblock")
elif mode == "nonblock-spin":
	nonblock("nonblock-spin")
elif mode == "nonblock-timed":
	nonblock("nonblock-timed")
elif mode == "pty":
	open_pty()
elif mode == "file-spin":
	file_while_running()
elif mode.startswith("sig-"):
	signaled(mode[4:], False)
elif mode.startswith("trap-"):
	signaled(mode[5:], True)
elif mode in ("ulimit-out", "ulimit-err", "ulimit-both"):
	ulimit(mode)
else:
	raise SystemExit(f"unknown mode {mode}")
