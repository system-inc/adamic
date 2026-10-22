#!/bin/bash
# Run the stdout-edge programs against Node, the native binary, and the JavaScript
# backend. The shell owns every scenario: ulimit, closed descriptors, ignored INT and
# HUP, signals, and a file read while the program runs. Python is only the piece bash
# cannot do, a non-blocking pipe and a pseudoterminal.
set -u

root=$(cd "$(dirname "$0")/../.." && pwd)
cd "$root"
ulimit -c 0

if [ -n "${ADAMIC_STDOUT_COMPILER:-}" ]; then
 compiler=("$ADAMIC_STDOUT_COMPILER")
else
 compiler=(go run ./cmd/adamic)
fi

work=$(mktemp -d "${TMPDIR:-/tmp}/adamic-stdout-edges.XXXXXX")
mkdir -p "$work/bin" "$work/runs"
report=$work/results.txt
: > "$report"

node_runner=$root/oracle/node.mjs
finite="little_streams lots_streams buffer_bounds stderr_only panic_edges read_after_print"
spinning="spin_little spin_lots"
spreads="spread_pair spread_text spread_path"

log() { printf '%s\n' "$*" | tee -a "$report"; }

build_one() {
	local name=$1
	local source=$root/notes/stdout-edges/programs/$name.a
	local build_log=$work/bin/$name.build
	if ! timeout 60 "${compiler[@]}" build "$source" -o "$work/bin/$name" > "$build_log" 2>&1; then
		return 1
	fi
	timeout 60 "${compiler[@]}" js "$source" > "$work/bin/$name.js" 2>> "$build_log"
}

# run_capture writes stdout, stderr and the exit code for one command.
# Usage: run_capture <dir> <mode> <command...>
run_capture() {
	local dir=$1
	local mode=$2
	shift 2
	mkdir -p "$dir"
	case $mode in
	plain)
		timeout 30 "$@" < /dev/null > "$dir/stdout" 2> "$dir/stderr"
		echo $? > "$dir/exit"
		;;
	ulimit-out)
		(ulimit -f 100; ulimit -c 0; "$@" < /dev/null > "$dir/stdout" 2> "$dir/stderr"; echo $? > "$dir/exit")
		;;
	ulimit-err)
		(ulimit -f 100; ulimit -c 0; "$@" < /dev/null > "$dir/out-tmp" 2> "$dir/stderr"; echo $? > "$dir/exit")
		mv "$dir/out-tmp" "$dir/stdout"
		;;
	ulimit-both)
		(ulimit -f 100; ulimit -c 0; "$@" < /dev/null > "$dir/stdout" 2> "$dir/stderr"; echo $? > "$dir/exit")
		;;
	closed-out)
		timeout 30 "$@" < /dev/null > "$dir/stdout" 2> "$dir/stderr" >&-
		# The >&- above closes stdout after the redirect. Open /dev/null into the
		# capture by redirecting stdout first, then closing it for the child only.
		;;
	esac
}

# closed_* redirections cannot go through the case above: closing has to be the
# child's redirection, and a capture file has to stay open in the shell. Each
# function starts the child with the descriptor closed and captures the other.
run_closed() {
	local dir=$1
	local which=$2
	shift 2
	mkdir -p "$dir"
	case $which in
	stdout)
		timeout 30 "$@" < /dev/null 2> "$dir/stderr" >&-
		echo $? > "$dir/exit"
		: > "$dir/stdout"
		;;
	stderr)
		timeout 30 "$@" < /dev/null > "$dir/stdout" 2>&-
		echo $? > "$dir/exit"
		: > "$dir/stderr"
		;;
	both)
		timeout 30 "$@" < /dev/null >&- 2>&-
		echo $? > "$dir/exit"
		: > "$dir/stdout"
		: > "$dir/stderr"
		;;
	stdin)
		timeout 30 "$@" <&- > "$dir/stdout" 2> "$dir/stderr"
		echo $? > "$dir/exit"
		;;
	esac
}

# Ignored INT and HUP are inherited across exec. The child is the program itself.
run_trapped() {
	local dir=$1
	local trapped=$2
	shift 2
	mkdir -p "$dir"
	bash -c 'ulimit -c 0; trap "" INT HUP; exec "$0" "$@"' "$@" < /dev/null > "$dir/stdout" 2> "$dir/stderr"
	echo $? > "$dir/exit"
}

# Stdout is a regular file. Finite programs are compared when they exit. Spinning
# programs are read while they are still running, then stopped with TERM.
run_file() {
	local dir=$1
	local spinning=$2
	shift 2
	mkdir -p "$dir"
	timeout 30 "$@" < /dev/null > "$dir/stdout" 2> "$dir/stderr" &
	local pid=$!

	wait "$pid"
	echo $? > "$dir/exit"
	cp "$dir/stdout" "$dir/while-running"
	echo finished > "$dir/note"
}

run_merged() {
	local dir=$1
	shift
	mkdir -p "$dir"
	timeout 30 "$@" < /dev/null > "$dir/stdout" 2>&1
	echo $? > "$dir/exit"
	: > "$dir/stderr"
}

# command_for <runtime> <name> prints the argv of one runtime.
command_for() {
	local runtime=$1
	local name=$2
	local source=$root/notes/stdout-edges/programs/$name.a
	case $runtime in
	node) printf '%s\0' node --disable-warning=ExperimentalWarning "$node_runner" "$source" ;;
	native) printf '%s\0' "$work/bin/$name" ;;
	js) printf '%s\0' node --disable-warning=ExperimentalWarning "$node_runner" "$work/bin/$name.js" ;;
	esac
}

launch() {
	local runtime=$1
	local name=$2
	shift 2
	local args=()
	while IFS= read -r -d '' part; do
		args+=("$part")
	done < <(command_for "$runtime" "$name")
	"$@" "${args[@]}"
}

summarize() {
	python3 - "$report" "$@" << 'PY'
import pathlib, sys
report, scenario, name = sys.argv[1], sys.argv[2], sys.argv[3]
roots = {label: pathlib.Path(path) for label, path in zip(sys.argv[4::2], sys.argv[5::2])}

def load(path):
    exit_code = (path / "exit").read_text().strip() if (path / "exit").exists() else "missing"
    stdout = (path / "stdout").read_bytes() if (path / "stdout").exists() else b""
    stderr = (path / "stderr").read_bytes() if (path / "stderr").exists() else b""
    note = (path / "note").read_text().strip() if (path / "note").exists() else ""
    running = (path / "while-running").read_bytes() if (path / "while-running").exists() else None
    return exit_code, stdout, stderr, note, running

def window(blob, index):
    start = max(0, index - 20)
    piece = blob[start:index + 40]
    return piece.decode("utf-8", "backslashreplace")

loaded = {label: load(path) for label, path in roots.items()}
node = loaded["node"]
lines = [f"== {scenario} {name}"]
differed = False
for label, item in loaded.items():
    if label == "node":
        continue
    reasons = []
    if item[0] != node[0]:
        reasons.append(f"exit {item[0]} != {node[0]}")
    if item[1] != node[1]:
        reasons.append(f"stdout {len(item[1])} != {len(node[1])}")
    if item[2] != node[2]:
        reasons.append(f"stderr {len(item[2])} != {len(node[2])}")
    if item[4] is not None and node[4] is not None and item[4] != node[4]:
        reasons.append(f"while-running {len(item[4])} != {len(node[4])}")
    note = f" note={item[3]}" if item[3] else ""
    if reasons:
        differed = True
        lines.append(f"  {label} DIFFERS: {', '.join(reasons)}{note}")
        for title, got, want in (("stdout", item[1], node[1]), ("stderr", item[2], node[2]), ("while-running", item[4], node[4])):
            if got is None or want is None or got == want:
                continue
            limit = min(len(got), len(want))
            index = next((i for i in range(limit) if got[i] != want[i]), limit)
            lines.append(f"  {title} first difference at {index}")
            lines.append(f"  node[{index}:] {window(want, index)!r}")
            lines.append(f"  {label}[{index}:] {window(got, index)!r}")
    else:
        lines.append(f"  {label} agrees exit={item[0]} stdout={len(item[1])} stderr={len(item[2])}{note}")
lines.append(f"  node exit={node[0]} stdout={len(node[1])} stderr={len(node[2])} note={node[3]}")
text = "\n".join(lines)
print(text)
with open(report, "a", encoding="utf-8") as handle:
    handle.write(text + "\n")
sys.exit(1 if differed else 0)
PY
}

# Redirect-and-close needs the command built outside run_closed's naive quoting,
# so finite and spinning scenarios call the runtime functions directly.
run_runtime() {
	local dir=$1
	local scenario=$2
	local runtime=$3
	local name=$4
	local source=$root/notes/stdout-edges/programs/$name.a
	local -a args=()
	case $runtime in
	node) args=(node --disable-warning=ExperimentalWarning "$node_runner" "$source") ;;
	native) args=("$work/bin/$name") ;;
	js) args=(node --disable-warning=ExperimentalWarning "$node_runner" "$work/bin/$name.js") ;;
	*)
		echo "unknown runtime $runtime" >&2
		return 1
		;;
	esac
	case $scenario in
	plain) run_capture "$dir" plain "${args[@]}" ;;
	closed-stdout) run_closed "$dir" stdout "${args[@]}" ;;
	closed-stderr) run_closed "$dir" stderr "${args[@]}" ;;
	closed-both) run_closed "$dir" both "${args[@]}" ;;
	closed-stdin) run_closed "$dir" stdin "${args[@]}" ;;
	trap-quiet) run_trapped "$dir" trap "${args[@]}" ;;
	merged) run_merged "$dir" "${args[@]}" ;;
	file-finite) run_file "$dir" no "${args[@]}" ;;
	ulimit-out | ulimit-err | ulimit-both | nonblock | nonblock-spin | nonblock-timed | pty | file-spin | sig-* | trap-INT | trap-HUP)
		timeout 35 python3 "$root/notes/stdout-edges/posix_run.py" "$scenario" "$dir" "${args[@]}"
		;;
	*)
		echo "unknown scenario $scenario" >&2
		return 1
		;;
	esac
}

compare_scenario() {
	local scenario=$1
	local name=$2
	shift 2
	local pairs=()
	local runtime dir
	for runtime in "$@"; do
		dir=$work/runs/$scenario/$name/$runtime
		run_runtime "$dir" "$scenario" "$runtime" "$name" || true
		pairs+=("$runtime" "$dir")
	done
	set +e
	summarize "$scenario" "$name" "${pairs[@]}"
	local status=$?
	set -e
	if [ "$status" -ne 0 ]; then
		echo "$scenario $name" >> "$work/differed"
	else
		echo "$scenario $name" >> "$work/agreed"
	fi
}

: > "$work/differed"
: > "$work/agreed"
set -e

# One case is one bounded invocation. With no arguments the diagnostic matrix
# includes timing-sensitive streams and Node's inspector signal; differences stay
# visible and the harness exits nonzero. Results always remain in scratch.
if [ "$#" -eq 2 ]; then
 name=$1
 scenario=$2
 case " $finite $spinning forever_mid $spreads " in
  *" $name "*) ;;
  *) echo "unknown program $name" >&2; exit 2 ;;
 esac
 build_one "$name" || { cat "$work/bin/$name.build"; exit 1; }
 (cd "$work"; compare_scenario "$scenario" "$name" node native js)
 log "results $report"
 test ! -s "$work/differed"
 exit
fi

log "building"
for name in $finite $spinning forever_mid $spreads; do
	if build_one "$name"; then
		log "build $name ok"
	else
		log "build $name FAILED"
		sed 's/^/  /' "$work/bin/$name.build" | tee -a "$report"
	fi
done

cd "$work"
runtimes_built="node native js"
for name in $finite; do
	if [ ! -x "$work/bin/$name" ]; then
		continue
	fi
	for scenario in plain ulimit-out ulimit-err ulimit-both closed-stdout closed-stderr closed-both closed-stdin trap-quiet merged file-finite nonblock pty; do
		compare_scenario "$scenario" "$name" $runtimes_built
	done
done

for name in $spinning; do
	if [ ! -x "$work/bin/$name" ]; then
		continue
	fi
	for scenario in sig-TERM sig-INT sig-HUP sig-USR1 sig-USR2 sig-ALRM trap-INT trap-HUP file-spin nonblock-spin; do
		compare_scenario "$scenario" "$name" $runtimes_built
	done
done

if [ -x "$work/bin/forever_mid" ]; then
	for scenario in sig-TERM sig-INT sig-HUP sig-USR1 sig-USR2 sig-ALRM nonblock-timed file-spin; do
		compare_scenario "$scenario" forever_mid $runtimes_built
	done
fi

for name in $spreads; do
	log "== spread $name"
	node_dir=$work/runs/spread/$name/node
	mkdir -p "$node_dir"
	set +e
	node --disable-warning=ExperimentalWarning "$node_runner" "$root/notes/stdout-edges/programs/$name.a" < /dev/null > "$node_dir/stdout" 2> "$node_dir/stderr"
	echo $? > "$node_dir/exit"
	set -e
	log "  node exit=$(cat "$node_dir/exit") stdout=$(cat "$node_dir/stdout") stderr=$(cat "$node_dir/stderr")"
	log "  native build:"
	sed 's/^/    /' "$work/bin/$name.build" | tee -a "$report"
done

log "---- agreed $(wc -l < "$work/agreed") ----"
sort "$work/agreed" | tee -a "$report"
log "---- differed $(wc -l < "$work/differed") ----"
sort "$work/differed" | tee -a "$report"

log "results $report"
test ! -s "$work/differed"
