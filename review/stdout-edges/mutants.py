"""Run each old-behavior mutant alone, restore the source even on failure, log the catching test."""
import difflib
import os
import sys
import signal
from pathlib import Path
import subprocess

runtime = Path('internal/native/runtime/adamic.c')
input_c = Path('internal/native/runtime/input.c')
lower = Path('internal/lower/input.go')
originals = {path: path.read_text() for path in [runtime, input_c, lower]}
base = originals[runtime]
logs = Path('/tmp/stdout-mutants')
logs.mkdir(exist_ok=True)


def without(text, start, end):
    left = text.index(start)
    right = text.index(end, left)
    return text[:left] + text[right:]


mutants = [
    ('sanitizer-segv', runtime, base.replace('SIGQUIT, SIGUSR1, SIGUSR2,', 'SIGQUIT, SIGSEGV, SIGUSR1, SIGUSR2,'), './internal/oracle', 'TestStartupPreservesSanitizerSegvReport$'),
    ('1-sigxfsz', runtime, base.replace('\tsignal(SIGXFSZ, SIG_IGN);\n', ''), './internal/oracle', 'TestOutputEdges/(fsize|fsize_out)$'),
    ('2-eagain', runtime, without(base, '\t\t\tif (errno == EAGAIN', '\t\t\treturn -1;'), './internal/oracle', 'TestOutputEdges/nonblock$'),
    ('3-closed', runtime, without(base, '\t// Node opens /dev/null', '\tadamic_arguments_save'), './internal/oracle', 'TestOutputEdges/closed$'),
    ('4-inherited', runtime, base.replace('(current.sa_handler == SIG_IGN && signal_number != SIGINT && signal_number != SIGHUP && signal_number != SIGTERM)', 'current.sa_handler == SIG_IGN'), './internal/oracle', 'TestOutputEdges/ignored$'),
    ('5-fatal', runtime, without(base, '\t// Signals used to stop a process', '\n}\n\nvoid adamic_write_line'), './internal/oracle', 'TestOutputEdges/signals$'),
    ('6-surrogate', runtime, base.replace('write_text(adamic_stderr, message, length);', '(void)write_all(adamic_stderr, message, length);'), './internal/oracle', 'TestOutputEdges/panic$'),
    ('7-spread', lower, originals[lower].replace('''		for _, argument := range arguments {
			if argument.Kind == ast.KindSpreadElement {
				return nil, true, l.notYet(argument, describe(argument))
			}
		}
''', ''), './internal/lower', 'TestInputTupleSpreadsAreNotYet$'),
    ('8-mode0644', input_c, originals[input_c].replace('O_CLOEXEC, 0666)', 'O_CLOEXEC, 0644)'), './internal/oracle', 'TestInputAgreesWithNode/internal/oracle/testdata/write_files.a$'),
    ('terminal', runtime, base.replace('isatty(adamic_stdout) || regular_file ? 2 : 1', 'regular_file ? 2 : 1'), './internal/oracle', 'TestOutputLinesArriveWhileRunning/terminal$'),
    ('realtime', runtime, without(base, '#if defined(SIGRTMIN) && defined(SIGRTMAX)', '\n}\n\nvoid adamic_write_line'), './internal/oracle', f'TestRealtimeSignalsLeaveWhatWasPrinted/{signal.SIGRTMIN}$'),
    ('file', runtime, base.replace('isatty(adamic_stdout) || regular_file ? 2 : 1', 'isatty(adamic_stdout) ? 2 : 1'), './internal/oracle', 'TestOutputLinesArriveWhileRunning/file$'),
]
try:
    for name, path, changed, package, test in mutants:
        if len(sys.argv) > 1 and name not in sys.argv[1:]:
            continue
        assert changed != originals[path], name
        path.write_text(changed)
        patch = ''.join(difflib.unified_diff(originals[path].splitlines(keepends=True), changed.splitlines(keepends=True), fromfile=str(path), tofile=str(path))).encode()
        (logs / (name + '.patch')).write_bytes(patch)
        command = ['go', 'test', '-count=1', '-timeout', '3m', package, '-run', test, '-v']
        with (logs / (name + '.log')).open('w') as output:
            result = subprocess.run(command, stdout=output, stderr=subprocess.STDOUT,
                                    env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), timeout=200)
        log = (logs / (name + '.log')).read_text()
        path.write_text(originals[path])
        assert result.returncode != 0 and '--- FAIL:' in log and 'build failed' not in log and 'native: clang:' not in log, (name, result.returncode, log[-1000:])
        print(name, 'caught by', test, 'exit', result.returncode, flush=True)
finally:
    for path, contents in originals.items():
        path.write_text(contents)
