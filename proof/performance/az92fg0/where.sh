# #az92fg0: where a warm `adamic build` of hello world spends the time no build-log line names. A CPU profile of the
# binary (an overlay adds a profile around run, exiting through it), and, where strace is on the box, every child
# process with its start and end. Run after environment.sh and sandboxed-job.sh. Never merged.
set -u
W=$(pwd); T=$(mktemp -d "${TMPDIR:-/tmp}/where.XXXXXX")
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
python3 - "$W/cmd/adamic/main.go" "$T/main.go" <<'PYTHON'
import sys
src, out = sys.argv[1:3]; s = open(src).read()
old = '\tos.Exit(run(os.Args[1:]))'
assert s.count(old) == 1
s = s.replace(old, '\tos.Exit(profiledRun(os.Args[1:]))')
s += '''
func profiledRun(arguments []string) int {
	if path := os.Getenv("ADAMIC_CPUPROFILE"); path != "" {
		file, err := os.Create(path)
		if err == nil {
			pprof.StartCPUProfile(file)
			defer func() { pprof.StopCPUProfile(); file.Close() }()
		}
	}
	return run(arguments)
}
'''
s = s.replace('import (', 'import (\n\t"runtime/pprof"', 1)
open(out, 'w').write(s)
PYTHON
gofmt -w "$T/main.go"
printf '{"Replace":{"%s":"%s"}}\n' "$W/cmd/adamic/main.go" "$T/main.go" > "$T/overlay.json"
go build -overlay="$T/overlay.json" -o "$T/adamic" ./cmd/adamic || { echo "VERDICT FAIL build"; exit 1; }
mkdir -p "$T/hello"; printf 'console.log("hello, world");\n' > "$T/hello/hello.a"
export XDG_CACHE_HOME="$T/cache" ADAMIC_BUILD_CACHE_DIR="$T/buildcache"; mkdir -p "$XDG_CACHE_HOME" "$ADAMIC_BUILD_CACHE_DIR"
cd "$T/hello"
"$T/adamic" build hello.a -o "$T/out" > /dev/null 2>&1 && echo "cold build done"
for i in 1 2; do "$T/adamic" build hello.a -o "$T/out" > /dev/null 2>&1; done
s=$(date +%s%N); ADAMIC_CPUPROFILE="$T/cpu.pprof" ADAMIC_BUILD_LOG="$T/build.log" "$T/adamic" build hello.a -o "$T/out" > /dev/null 2>&1; e=$(date +%s%N)
echo "profiled warm build: $(( (e - s) / 1000000 )) ms wall"; cat "$T/build.log"
cd "$W"
go tool pprof -top -nodecount=35 "$T/adamic" "$T/cpu.pprof" 2>/dev/null | head -45
go tool pprof -top -cum -nodecount=40 "$T/adamic" "$T/cpu.pprof" 2>/dev/null | sed -n 1,50p
if command -v strace > /dev/null; then
  cd "$T/hello"; strace -f -tt -e trace=execve,exit_group -o "$T/strace.log" "$T/adamic" build hello.a -o "$T/out" > /dev/null 2>&1; cd "$W"
  echo "== child processes (execve), with time"; grep -E 'execve\(' "$T/strace.log" | grep -v ENOENT | awk '{print $2, $3, substr($0, index($0,"execve"), 200)}' | head -40
  echo "== first and last line"; head -1 "$T/strace.log" | cut -c1-120; tail -1 "$T/strace.log" | cut -c1-120
else echo "no strace on this box"; fi
echo "VERDICT OK"
