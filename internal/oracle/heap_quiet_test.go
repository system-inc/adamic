package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"strconv"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// heap.c gives a thread's spare chunks back only at a quiet point (#vea9pph, Adamic's ruling): a wait
// that would block at least ADAMIC_HEAP_QUIET_MS first waits that long, and if it ends with nothing
// to do, the spares past the cap go back and the wait goes on to its own deadline. These run whole
// programs through both quiet points, the event loop's wait and a pool worker's park, unsanitized,
// since under AddressSanitizer every value is malloc's and there are no chunks to give back.
//   - "burst then wait": a million values held at module level, dropped, then one 300 ms timer.
//     Read in the timer's callback, the process holds far less than when they were dropped, and the
//     counted build says chunks went back. Keeping every spare reads the two the same.
//   - "fast timer churn": a 1 ms timer churns values every tick, so no wait ever reaches the quiet
//     span, and nothing goes back (Adamic's condition 2).
//   - "deadline": a 100 ms timer armed after a burst fires as late as the same timer with no burst,
//     within 5 ms, since the wait's remainder runs to the loop's own deadline.
//   - "parked workers": parallelMap's workers build and drop values, park, and give theirs back while
//     the main thread spins, so no quiet point of the main thread's can be what gave them back.
//
// Mutants, each red here: no quiet span in host_default_wait (burst then wait); the remainder waited
// as the full timeout again after the span (deadline, 10 ms late); a release as a chunk empties, at
// the drop (fast timer churn releases); park waiting without its span (parked workers).
const heapQuietResident = `import { readFileSync } from 'node:fs';

// resident is VmRSS from /proc/self/status, or -1 where there's none.
function resident(): number {
	try {
		const status = readFileSync('/proc/self/status', 'utf8');
		for (const line of status.split('\n')) {
			if (line.startsWith('VmRSS:')) {
				const value = line.slice(6).trim();
				const space = value.indexOf(' ');
				return Number(space < 0 ? value : value.slice(0, space));
			}
		}
	} catch (error) {
		return -1;
	}
	return -1;
}
`

const heapQuietBurst = `interface Item {
	readonly value: number;
	readonly label: string;
}
let held: Item[] = [];
for (let index = 0; index < 1000000; index += 1) {
	held.push({ value: index, label: ` + "`k${index % 997}`" + ` });
}
let total = 0;
for (const item of held) {
	total = (total + item.value + item.label.length) % 1000003;
}
held = [];
`

var heapQuietSources = map[string]string{
	"burst then wait": heapQuietResident + heapQuietBurst + `const dropped = resident();
setTimeout(() => {
	const idle = resident();
	console.error(` + "`rss dropped ${dropped} idle ${idle}`" + `);
	console.log(` + "`total ${total}`" + `);
}, 300);
`,
	"fast timer churn": `import 'node:timers';
interface Item {
	readonly value: number;
	readonly label: string;
}
let held: Item[] = [];
let ticks = 0;
let total = 0;
function tick(): void {
	held = [];
	for (let index = 0; index < 50000; index += 1) {
		held.push({ value: index, label: ` + "`t${index % 97}`" + ` });
	}
	total = (total + held.length) % 1000003;
	ticks += 1;
	if (ticks < 100) {
		setTimeout(tick, 1);
	} else {
		held = [];
		console.log(` + "`ticks ${ticks} total ${total}`" + `);
	}
}
setTimeout(tick, 1);
`,
	"deadline": `import { performance } from 'node:perf_hooks';
` + heapQuietBurst + `const start = performance.now();
setTimeout(() => {
	console.error(` + "`delay ${Math.round(performance.now() - start)}`" + `);
	console.log(` + "`fired ${total}`" + `);
}, 100);
`,
	"deadline control": `import { performance } from 'node:perf_hooks';
const start = performance.now();
setTimeout(() => {
	console.error(` + "`delay ${Math.round(performance.now() - start)}`" + `);
	console.log('fired');
}, 100);
`,
}

const heapQuietWorkers = `import { performance } from 'node:perf_hooks';
import { parallelMap } from 'adamic';
interface Item {
	readonly value: number;
	readonly label: string;
}
function build(seed: number): number {
	const made: Item[] = [];
	for (let index = 0; index < 300000; index += 1) {
		made.push({ value: index + seed, label: ` + "`w${index % 89}`" + ` });
	}
	let total = 0;
	for (const item of made) {
		total = (total + item.value + item.label.length) % 1000003;
	}
	return total;
}
const seeds: readonly number[] = [0, 1, 2, 3, 4, 5, 6, 7];
const totals = parallelMap(seeds, build);
let sum = 0;
for (const each of totals) {
	sum = (sum + each) % 1000003;
}
// The main thread spins rather than waits, so it never reaches a quiet point of its own: only the
// parked workers can give their chunks back meanwhile.
const start = performance.now();
let spins = 0;
while (performance.now() - start < 300) {
	spins += 1;
}
console.log(` + "`sum ${sum} spun ${spins > 0}`" + `);
`

func TestHeapQuietPoints(t *testing.T) {
	t.Parallel()
	if goruntime.GOOS != "linux" {
		t.Skip("reads /proc/self/status, and the release is pinned on Linux")
	}
	build := func(t *testing.T, name, extension, source string) (*ir.Program, string) {
		t.Helper()
		path := filepath.Join(t.TempDir(), "main"+extension)
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		return program, path
	}
	binary := func(t *testing.T, program *ir.Program, count bool) string {
		t.Helper()
		output := filepath.Join(t.TempDir(), "native")
		if err := native.Build(native.C(program), output, native.Options{Count: count}); err != nil {
			t.Fatal(err)
		}
		return output
	}
	// The programs write their own measurements to stderr, which differ from Node's, so they agree with
	// Node by their output and exit code.
	agrees := func(t *testing.T, path string, got run) {
		t.Helper()
		if truth := onNode(t, path); truth.exitCode != got.exitCode || !bytes.Equal(truth.stdout, got.stdout) {
			t.Fatalf("native: exit %d, stdout %q; Node: exit %d, stdout %q; stderr %q", got.exitCode, got.stdout, truth.exitCode, truth.stdout, got.stderr)
		}
	}
	released := func(stderr []byte) int {
		match := regexp.MustCompile(`adamic: heap counts: released (\d+)`).FindSubmatch(stderr)
		if match == nil {
			return 0
		}
		value, _ := strconv.Atoi(string(match[1]))
		return value
	}
	number := func(t *testing.T, pattern string, text []byte) int {
		t.Helper()
		match := regexp.MustCompile(pattern).FindSubmatch(text)
		if match == nil {
			t.Fatalf("no %q in %q", pattern, text)
		}
		value, err := strconv.Atoi(string(match[1]))
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	t.Run("burst then wait", func(t *testing.T) {
		t.Parallel()
		program, path := build(t, "burst then wait", ".ts", heapQuietSources["burst then wait"])
		plain := execute(t, binary(t, program, false))
		agrees(t, path, plain)
		dropped, idle := number(t, `rss dropped (\d+)`, plain.stderr), number(t, `idle (\d+)`, plain.stderr)
		if idle*2 > dropped {
			t.Errorf("read in the timer's callback the process holds %d KB, against %d KB when the values were dropped: the wait gave nothing back", idle, dropped)
		}
		if counted := execute(t, binary(t, program, true)); released(counted.stderr) == 0 {
			t.Errorf("the counted build released no chunks: %s", counted.stderr)
		}
		t.Logf("dropped %d KB, idle %d KB", dropped, idle)
	})
	t.Run("fast timer churn", func(t *testing.T) {
		t.Parallel()
		program, path := build(t, "fast timer churn", ".ts", heapQuietSources["fast timer churn"])
		counted := execute(t, binary(t, program, true))
		agrees(t, path, counted)
		if count := released(counted.stderr); count != 0 {
			t.Errorf("a loop ticking every millisecond gave back %d chunks, want none: no wait of its reaches the quiet span", count)
		}
	})
	t.Run("deadline", func(t *testing.T) {
		t.Parallel()
		burst, burstPath := build(t, "deadline", ".ts", heapQuietSources["deadline"])
		control, controlPath := build(t, "deadline control", ".ts", heapQuietSources["deadline control"])
		burstBinary, controlBinary := binary(t, burst, false), binary(t, control, false)
		least := func(path, name string) int {
			best := -1
			for range 3 {
				got := execute(t, name)
				agrees(t, path, got)
				if delay := number(t, `delay (\d+)`, got.stderr); best < 0 || delay < best {
					best = delay
				}
			}
			return best
		}
		after, alone := least(burstPath, burstBinary), least(controlPath, controlBinary)
		if after > alone+5 {
			t.Errorf("a 100 ms timer armed after a burst fired after %d ms, against %d ms with no burst: the wait after the quiet span overran the loop's deadline", after, alone)
		}
		t.Logf("after a burst %d ms, alone %d ms", after, alone)
	})
	t.Run("parked workers", func(t *testing.T) {
		t.Parallel()
		if goruntime.NumCPU() < 2 {
			t.Skip("with one processor the main thread runs every range, and no worker holds anything to give back")
		}
		program, path := build(t, "parked workers", ".a", heapQuietWorkers)
		counted := execute(t, binary(t, program, true))
		agrees(t, path, counted)
		if released(counted.stderr) == 0 {
			t.Errorf("parked workers gave back no chunks: %s", counted.stderr)
		}
	})
}
