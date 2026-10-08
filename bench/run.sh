#!/usr/bin/env bash
# Source cloud/setup.sh's env file first. Requires Python 3 and the repository's Go/clang tools.
set -euo pipefail
cd "$(dirname "$0")/.."
exec python3 - "$@" <<'PY'
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import platform
import resource
import shutil
import subprocess
import sys
import tempfile
import time
import zipfile

PROGRAMS = ["primes.a", "number_sum_sort.a", "string_build.a", "word_count.ts",
            "map_200k.a", "json_encode.a", "trees.ts", "nbody.ts"]
BUN_VERSION = "1.3.14"


def command(args, **kwargs):
    return subprocess.run(args, check=True, stdout=subprocess.PIPE,
                          stderr=subprocess.PIPE, **kwargs)


def version(args):
    return command(args).stdout.decode().strip().splitlines()[0]


def read_optional(path):
    try:
        return Path(path).read_text().strip()
    except OSError:
        return "unavailable"


def locate_bun(work):
    supplied = os.environ.get("BUN")
    found = shutil.which(supplied or "bun")
    if supplied and not found:
        raise RuntimeError(f"BUN executable not found: {supplied}")
    if found:
        return found
    candidate = Path.home() / ".bun/bin/bun"
    if candidate.is_file() and os.access(candidate, os.X_OK):
        return str(candidate)
    machine = platform.machine().lower()
    arch = {"x86_64": "x64", "amd64": "x64", "aarch64": "aarch64", "arm64": "aarch64"}.get(machine)
    system = {"Linux": "linux", "Darwin": "darwin"}.get(platform.system())
    if not arch or not system:
        raise RuntimeError("Cannot install Bun on this platform; set BUN to its executable")
    distribution = f"bun-{system}-{arch}"
    cache = Path.home() / ".cache/adamic-bench" / f"bun-{BUN_VERSION}-{system}-{arch}"
    binary = cache / "bun"
    if not binary.is_file():
        url = f"https://github.com/oven-sh/bun/releases/download/bun-v{BUN_VERSION}/{distribution}.zip"
        archive = work / "bun.zip"
        print(f"Installing Bun {BUN_VERSION} from {url}", file=sys.stderr, flush=True)
        command(["curl", "--fail", "--location", "--retry", "2", "--max-time", "180", url, "-o", str(archive)])
        with zipfile.ZipFile(archive) as bundle:
            contents = bundle.read(f"{distribution}/bun")
        cache.mkdir(parents=True, exist_ok=True)
        temporary = cache / f"bun.{os.getpid()}.tmp"
        temporary.write_bytes(contents)
        temporary.chmod(0o755)
        temporary.replace(binary)
    if version([str(binary), "--version"]) != BUN_VERSION:
        raise RuntimeError("Cached Bun version differs from the pinned installer version")
    return str(binary)


def run(args, expected, timeout, environment):
    before = resource.getrusage(resource.RUSAGE_CHILDREN).ru_utime
    start = time.perf_counter()
    result = command(args, timeout=timeout, env=environment)
    wall = time.perf_counter() - start
    user = resource.getrusage(resource.RUSAGE_CHILDREN).ru_utime - before
    if not result.stdout:
        raise RuntimeError(f"Empty checksum output from {args}")
    if expected is not None and result.stdout != expected:
        raise RuntimeError(f"CHECKSUM MISMATCH: {args}\nexpected {expected!r}\nactual {result.stdout!r}")
    return result.stdout, wall, user


def main():
    parser = argparse.ArgumentParser(description="Release native vs Node vs Bun; five interleaved rounds")
    parser.add_argument("--only", help="comma-separated names, default all eight programs (six workload categories plus JSON encode)")
    parser.add_argument("--check-only", action="store_true", help="build and verify outputs, without timings or a report")
    parser.add_argument("--output", default="bench/RESULTS.md")
    parser.add_argument("--timeout", type=float, default=120, help="seconds per execution")
    options = parser.parse_args()
    if options.timeout <= 0:
        parser.error("--timeout must be positive")
    sources = PROGRAMS
    if options.only is not None:
        wanted = set(options.only.split(","))
        known = {Path(name).stem for name in PROGRAMS}
        if not wanted or wanted - known:
            parser.error(f"unknown program names: {sorted(wanted - known)}")
        sources = [name for name in PROGRAMS if Path(name).stem in wanted]
    node = shutil.which("node")
    if not node or version([node, "--version"]) != "v24.19.0":
        raise RuntimeError("Node v24.19.0 is required; source the setup env file or put that version on PATH")
    environment = dict(os.environ)
    for variable in ("NODE_OPTIONS", "BUN_OPTIONS", "ADAMIC_THREADS"):
        environment.pop(variable, None)
    with tempfile.TemporaryDirectory(prefix="adamic-honest-bench-") as directory:
        work = Path(directory)
        bun = locate_bun(work)
        cpu = next((line.split(":", 1)[1].strip() for line in read_optional("/proc/cpuinfo").splitlines()
                    if line.startswith("model name")), platform.processor() or "unavailable")
        metadata = {
            "UTC": datetime.datetime.now(datetime.timezone.utc).isoformat(),
            "machine": platform.node(), "OS/kernel/architecture": platform.platform(), "CPU": cpu,
            "logical CPUs": os.cpu_count(),
            "available CPUs (affinity)": len(os.sched_getaffinity(0)) if hasattr(os, "sched_getaffinity") else "unavailable",
            "cgroup cpu.max (quota period, microseconds)": read_optional("/sys/fs/cgroup/cpu.max"),
            "Adamic source commit": version(["git", "rev-parse", "HEAD"]),
            "working tree": command(["git", "status", "--short"]).stdout.decode().strip() or "clean",
            "Go": version(["go", "version"]), "clang": version(["clang", "--version"]),
            "Node": version([node, "--version"]), "Bun": version([bun, "--version"]),
            "Python": platform.python_version(), "Node executable": node, "Bun executable": bun,
            "runner SHA256": hashlib.sha256(Path("bench/run.sh").read_bytes()).hexdigest(),
        }
        print(json.dumps(metadata, indent=2), flush=True)
        not_measured = ["JSON decode: intentionally omitted; native refuses `JSON.parse`: its result's type can't be proven from the text (internal/lower/library_json_stringify.go)."]
        programs = {}
        for source in sources:
            path = Path("bench") / source
            name = path.stem
            native = work / name
            args = ["go", "run", "./cmd/adamic", "build", str(path), "-o", str(native)]
            print("Building: " + " ".join(args), file=sys.stderr, flush=True)
            built = subprocess.run(args, stdout=subprocess.PIPE, stderr=subprocess.PIPE, env=environment)
            if built.returncode:
                diagnostic = (built.stdout + built.stderr).decode(errors="replace").strip()
                if name == "json_encode" and "JSON.stringify" in diagnostic and (
                        "refuses" in diagnostic.lower() or "can't lower" in diagnostic.lower()):
                    not_measured.append("JSON encode: native build refused:\n\n```text\n" + diagnostic + "\n```")
                    print(not_measured[-1], file=sys.stderr, flush=True)
                    continue
                raise RuntimeError(f"Build failed: {args}\n{diagnostic}")
            node_args = [node, "--disable-warning=ExperimentalWarning"]
            if path.suffix == ".a":
                node_args += ["oracle/node.mjs"]
            node_args += [str(path)]
            programs[name] = {
                "source": str(path), "hash": hashlib.sha256(path.read_bytes()).hexdigest(),
                "commands": {"native": [str(native)], "Node": node_args, "Bun": [bun, str(path)]},
                "samples": {kind: [] for kind in ("native", "Node", "Bun")},
            }
        if not programs:
            raise RuntimeError("No measurable programs")
        # Separate untimed check, followed by an output check on EVERY measured run.
        for name, program in programs.items():
            expected = None
            for kind, args in program["commands"].items():
                output, _, _ = run(args, expected, options.timeout, environment)
                expected = output
            program["checksum"] = expected
            print(f"Checksum OK {name}: {expected.decode().strip()!r}", flush=True)
        if options.check_only:
            return
        metadata["load before timings (1/5/15 minutes)"] = " / ".join(f"{load:.2f}" for load in os.getloadavg())
        print("Load before: " + metadata["load before timings (1/5/15 minutes)"], flush=True)
        kinds = ["native", "Node", "Bun"]
        for round_index in range(5):
            # Rotate first runtime to spread fixed ordering effects; no concurrent benchmark children.
            order = kinds[round_index % 3:] + kinds[:round_index % 3]
            for name, program in programs.items():
                for kind in order:
                    _, wall, user = run(program["commands"][kind], program["checksum"], options.timeout, environment)
                    program["samples"][kind].append((wall, user))
                    print(f"round {round_index + 1} {name} {kind}: wall {wall:.6f}s user {user:.6f}s", flush=True)
        metadata["load after timings (1/5/15 minutes)"] = " / ".join(f"{load:.2f}" for load in os.getloadavg())
        print("Load after: " + metadata["load after timings (1/5/15 minutes)"], flush=True)
        report = ["# Native, Node and Bun benchmarks", "", "Run with `source /workspace/adamic-tools/env.sh; bash bench/run.sh`.", "",
                  "Six workload categories; string building and splitting use two existing programs. JSON encode is attempted separately. No compiler/runtime optimizations were made.", "",
                  "## Machine and versions", "", "```json", json.dumps(metadata, indent=2), "```", "",
                  "## Method", "",
                  "Each source is built with `go run ./cmd/adamic build SOURCE -o BINARY`: shipped release build, clang -O2, no sanitizers or counters. Build/setup time is excluded. Node runs existing .ts sources directly and .a sources through oracle/node.mjs; Bun runs the same sources directly. No generated JavaScript is used.", "",
                  "One untimed execution per runtime validates exact, nonempty stdout before timing. All 15 timed executions per program must also match byte for byte or the script exits nonzero without writing a report. Each execution is a fresh process, with no in-process warmup. Startup, Node's type stripping, checksum calculation and output are included.", "",
                  "Five rounds interleave runtimes for each program. The first runtime rotates each round: native/Node/Bun, Node/Bun/native, Bun/native/Node, native/Node/Bun, Node/Bun/native. Wall time uses Python perf_counter around process launch and wait; user CPU time uses the difference in getrusage(RUSAGE_CHILDREN). Children run sequentially, so it measures that execution's user CPU, excluding kernel CPU and the Python parent. Best means minimum wall time; reported user time belongs to that same sample, not an independently selected minimum.", "",
                  "This is a shared cloud machine, without CPU pinning or isolation. Load averages and the full sample spread are recorded; small differences are not evidence of a repeatable win. Earlier setup and validation remain in the load averages; they finished before timing. No setup, tests or builds were started alongside timing by this runner. NODE_OPTIONS, BUN_OPTIONS and ADAMIC_THREADS are cleared for benchmark children.", "",
                  "## Workloads", "",
                  "- primes: ten fresh Uint8Array sieves with limits 2,000,000 through 2,000,009; count and sum every prime.",
                  "- number_sum_sort: one million random integers and one million nearly sorted integers; full sums before/after, comparator sorts, order check and an order-sensitive checksum. Adapted from the existing sort.ts generator/cases.",
                  "- string_build: existing 6,000 bounded formatter documents; construction and full UTF-16 checksum.",
                  "- word_count: existing text construction and splitting of a million words, map counts and frequency sort.",
                  "- map_200k: insert 200,000 distinct string keys, look every key up in reverse order, fold insertion order.",
                  "- json_encode: 100,000 fresh encodings of a nested object with runtime scalar values, followed by a full UTF-16 checksum scan. This measures encoding plus checksum work, not isolated encoder throughput.",
                  "- trees: existing recursive binary-tree build/walk, about 68 million allocated nodes, maximum depth 18.",
                  "- nbody: existing five-body floating-point simulation, one million steps, energy checksums.", "",
                  "## Results", "", "Times in seconds. Ratios are native wall / competitor wall; above 1 means native loses.", "",
                  "| Program | Native wall | Native user | Node wall | Node user | Bun wall | Bun user | Native / Node | Native / Bun | Output |",
                  "|---|---:|---:|---:|---:|---:|---:|---:|---:|---|"]
        losses = {"Node": [], "Bun": []}
        wins = {"Node": [], "Bun": []}
        for name, program in programs.items():
            best = {kind: min(program["samples"][kind]) for kind in kinds}
            ratios = {kind: best["native"][0] / best[kind][0] for kind in ("Node", "Bun")}
            row = [name] + [f"{value:.6f}" for kind in kinds for value in best[kind]]
            row += [f"{ratios[kind]:.2f}x" for kind in ("Node", "Bun")] + ["matches"]
            report.append("| " + " | ".join(row) + " |")
            for kind, ratio in ratios.items():
                (losses if ratio > 1 else wins)[kind].append(f"{name} ({ratio:.2f}x)")
        report += ["", "Native losses are measured best-wall comparisons, including startup:", ""]
        for kind in ("Node", "Bun"):
            report.append(f"- Native loses to {kind}: " + (", ".join(losses[kind]) or "none in this run") + ".")
            report.append(f"- Native wins against {kind}: " + (", ".join(wins[kind]) or "none in this run") + ".")
        report += ["", "## Not measured", ""] + not_measured
        report += ["", "## Sources and checksums", "", "Source SHA256 identifies the exact workload; multiline tree/n-body output is retained.", ""]
        for name, program in programs.items():
            report += [f"### {program['source']}", "", f"SHA256: `{program['hash']}`", "", "```text", program["checksum"].decode().rstrip("\n"), "```", ""]
        report += ["## Every timing sample", "", "Samples are in round order. Each cell is wall / user seconds.", "",
                   "| Program | Runtime | Round 1 | Round 2 | Round 3 | Round 4 | Round 5 |", "|---|---|---:|---:|---:|---:|---:|"]
        for name, program in programs.items():
            for kind in kinds:
                report.append("| " + " | ".join([name, kind] + [f"{wall:.6f} / {user:.6f}" for wall, user in program["samples"][kind]]) + " |")
        destination = Path(options.output)
        destination.parent.mkdir(parents=True, exist_ok=True)
        with tempfile.NamedTemporaryFile(mode="w", dir=destination.parent, delete=False) as output:
            output.write("\n".join(report) + "\n")
            temporary = Path(output.name)
        temporary.replace(destination)
        print(f"Report written: {destination}", flush=True)


try:
    main()
except (RuntimeError, OSError, subprocess.SubprocessError, zipfile.BadZipFile, KeyError) as error:
    print(f"bench: {error}", file=sys.stderr)
    sys.exit(1)
PY
