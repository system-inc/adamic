# View-site-roads probe report

Run date: 2026-11-12 (UTC).

Adamic HEAD: `daf2e578390c0db7603a7fc027d11d31581457f9` (detached for both runs). The requested branch was fetched and `git submodule update --init --recursive` completed. Submodules: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.

Loaded `/workspace/adamic-tools/env.sh` before setup and each run. Created its external work directory `/tmp/adamic-gate`. Fetched upstream TypeScript into `/tmp/vsr-typescript` and verified commit `050880ce59e30b356b686bd3144efe24f875ebc8`; exported `ADAMIC_TYPESCRIPT_SOURCE=/tmp/vsr-typescript`. Copied stage3/api package.json and package-lock.json into `/tmp/vsr-api` and ran `npm ci --prefix /tmp/vsr-api --ignore-scripts --no-audit --no-fund`. Prepended `/tmp/vsr-api/node_modules/.bin` to PATH: `tsc --version` reports `Version 6.0.3`. No dependency installation was made in the repository. Heavy commands ran one at a time.

Tools: Node v24.19.0, Go go1.27.1 linux/amd64, clang 20.1.8. Native probes use the runner's `--sanitize` build (ASan and UBSan).

## Control verdict

**PASS**, exit 0, for `cloud/probe/run.sh --check-control`: agrees.ts agrees on both backends; differs.a differs on both with the expected array-bounds panic and exit 70.

```text
Probes in cloud/probe/control, on v24.19.0 and clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261).
- head: daf2e578390c0db7603a7fc027d11d31581457f9

| probe | Node | build | JavaScript | native |
|---|---|---|---|---|
| differs.a | 1,2,3,9 (exit 0) | head | differs:  (exit 70), adamic: panic: index 3 is outside an array of length 3 | differs:  (exit 70), adamic: panic: index 3 is outside an array of length 3 |
| agrees.ts | 1,2,3 (exit 0) | head | agrees | agrees |

Full output: /tmp/adamic-gate/adamic-probe.Qzd2IB/logs
```

## Requested probe table

Command: `cloud/probe/run.sh probe/view-site-roads --logs /tmp/vsr-logs` (exit 0). This exit indicates runner completion; the table reports the semantic outcomes.

Four probes stop on both backends. Three logical array-assignment probes differ on both backends; two additional nullish-assignment probes differ natively while JavaScript agrees. The readonly native probe produces an AddressSanitizer SEGV; the hidden optional-field native probe panics.

```text
Probes in probe/view-site-roads, on v24.19.0 and clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261).
- head: daf2e578390c0db7603a7fc027d11d31581457f9

| probe | Node | build | JavaScript | native |
|---|---|---|---|---|
| v01_assign_control.ts | rex woof (exit 70) | head | stops: adamic: v01_assign_control.ts:6:11: Adamic 0.1 refuses a value of type Dog[] seen as Animal[] \| undefined, which can wr | stops: adamic: v01_assign_control.ts:6:11: Adamic 0.1 refuses a value of type Dog[] seen as Animal[] \| undefined, which can wr |
| v02_nullish_assign.ts | rex woof (exit 70) | head | differs: rex woof tom meow (exit 0) | differs: rex woof tom woof (exit 0) |
| v03_or_assign.ts | rex woof (exit 70) | head | differs: rex woof tom meow (exit 0) | differs: rex woof tom woof (exit 0) |
| v04_and_assign.ts | rex woof (exit 70) | head | differs: rex woof tom meow (exit 0) | differs: rex woof tom woof (exit 0) |
| v05_shorthand.ts | rex woof (exit 70) | head | stops: adamic: v05_shorthand.ts:6:38: Adamic 0.1 refuses a value of type Dog[] seen as Animal[], which can write Animal where D | stops: adamic: v05_shorthand.ts:6:38: Adamic 0.1 refuses a value of type Dog[] seen as Animal[], which can write Animal where D |
| v06_readonly_nullish.ts | s1 (exit 0) | head | agrees | differs:  (exit 1), ==2827==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x5615c7aecac2 bp 0 |
| v07_hiding_nullish.ts | str1 (exit 0) | head | agrees | differs:  (exit 70), adamic: panic: compiler bug: a numeric field holds a reference |
| v08_hiding_shorthand.ts | str1 (exit 0) | head | stops: adamic: v08_hiding_shorthand.ts:4:55: Adamic 0.1 refuses optional property b in { a: number; b?: number; } absent from s | stops: adamic: v08_hiding_shorthand.ts:4:55: Adamic 0.1 refuses optional property b in { a: number; b?: number; } absent from s |
| v09_hiding_assign_control.ts | str1 (exit 0) | head | stops: adamic: v09_hiding_assign_control.ts:4:7: Adamic 0.1 refuses optional property b in { a: number; b?: number; } absent fr | stops: adamic: v09_hiding_assign_control.ts:4:7: Adamic 0.1 refuses optional property b in { a: number; b?: number; } absent fr |

Full output: /tmp/vsr-logs
```

## Runtime differences

The stdout and stderr blocks below are copied from the logs. Empty streams are explicitly marked. Exits are the statuses recorded in the runner table. Both backends are included for each probe with a runtime difference, including an agreeing backend. Sanitizer stderr is limited to its first 30 lines; the observed report has 21 lines, so it is included in full.

### v02_nullish_assign.ts

Node stdout:

```text
rex woof
```

Node exit: 70.

Node stderr (additional context):

```text
adamic: panic: TypeError: dog.bark is not a function
```

JavaScript stdout:

```text
rex woof
tom meow
```

JavaScript stderr:

(empty)

JavaScript exit: 0.

Native stdout:

```text
rex woof
tom woof
```

Native stderr:

(empty)

Native exit: 0.

### v03_or_assign.ts

Node stdout:

```text
rex woof
```

Node exit: 70.

Node stderr (additional context):

```text
adamic: panic: TypeError: dog.bark is not a function
```

JavaScript stdout:

```text
rex woof
tom meow
```

JavaScript stderr:

(empty)

JavaScript exit: 0.

Native stdout:

```text
rex woof
tom woof
```

Native stderr:

(empty)

Native exit: 0.

### v04_and_assign.ts

Node stdout:

```text
rex woof
```

Node exit: 70.

Node stderr (additional context):

```text
adamic: panic: TypeError: dog.bark is not a function
```

JavaScript stdout:

```text
rex woof
tom meow
```

JavaScript stderr:

(empty)

JavaScript exit: 0.

Native stdout:

```text
rex woof
tom woof
```

Native stderr:

(empty)

Native exit: 0.

### v06_readonly_nullish.ts

Node stdout:

```text
s1
```

Node exit: 0.

Node stderr (additional context):

(empty)

JavaScript stdout:

```text
s1
```

JavaScript stderr:

(empty)

JavaScript exit: 0.

Native stdout:

(empty)

Native stderr:

```text
AddressSanitizer:DEADLYSIGNAL
=================================================================
==2827==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x5615c7aecac2 bp 0x000000000001 sp 0x7fffeee77f80 T0)
==2827==The signal is caused by a READ memory access.
==2827==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.
    #0 0x5615c7aecac2 in adamic_release /adamic-build/runtime/adamic.h:197:13
    #1 0x5615c7aecdb7 in adamic_initialize_v06_readonly_nullish_ts_c4c3a8ec44ae6fa6e14233e594db3ac9 /adamic-build/main.c:59:99
    #2 0x5615c7aeca9d in adamic_initialize_v06_readonly_nullish_ts_28602762e73c1d93591d6a93b3b9bdd2 /adamic-build/main.c:72:2
    #3 0x5615c7aeca52 in main /adamic-build/main.c:80:2
    #4 0x7f4671613ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)
    #5 0x7f4671613d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)
    #6 0x5615c7a06590 in _start (/tmp/vsr-logs/v06_readonly_nullish.ts/head.bin+0x81590)

==2827==Register values:
rax = 0x07fe000000000001  rbx = 0x00007bb6707e00c0  rcx = 0x0000000000000000  rdx = 0x00000ac2b8f865b0  
rdi = 0x3ff000000000000c  rsi = 0x3ff0000000000000  rbp = 0x0000000000000001  rsp = 0x00007fffeee77f80  
 r8 = 0x00000f76ce0fc019   r9 = 0x0000000000000000  r10 = 0x00000f76ce0fc019  r11 = 0x0000000000000000  
r12 = 0x00000f76ce0fc018  r13 = 0x00007fffeee780f8  r14 = 0x00007bb6707e0090  r15 = 0x3ff0000000000000  
AddressSanitizer can not provide additional info.
SUMMARY: AddressSanitizer: SEGV /adamic-build/runtime/adamic.h:197:13 in adamic_release
==2827==ABORTING
```

Native exit: 1.

### v07_hiding_nullish.ts

Node stdout:

```text
str1
```

Node exit: 0.

Node stderr (additional context):

(empty)

JavaScript stdout:

```text
str1
```

JavaScript stderr:

(empty)

JavaScript exit: 0.

Native stdout:

(empty)

Native stderr:

```text
adamic: panic: compiler bug: a numeric field holds a reference
```

Native exit: 70.

## Compile-time stops (verbatim)

For each probe below, the JavaScript and native compiler stderr files are byte-for-byte identical; the shared complete stop text is pasted once and applies to both backends. No backend executable ran for these stops.

### v01_assign_control.ts — JavaScript and native

```text
adamic: /workspace/adamic/probe/view-site-roads/v01_assign_control.ts:6:11: Adamic 0.1 refuses a value of type Dog[] seen as Animal[] | undefined, which can write Animal where Dog is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

### v05_shorthand.ts — JavaScript and native

```text
adamic: /workspace/adamic/probe/view-site-roads/v05_shorthand.ts:6:38: Adamic 0.1 refuses a value of type Dog[] seen as Animal[], which can write Animal where Dog is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

### v08_hiding_shorthand.ts — JavaScript and native

```text
adamic: /workspace/adamic/probe/view-site-roads/v08_hiding_shorthand.ts:4:55: Adamic 0.1 refuses optional property b in { a: number; b?: number; } absent from structural source { a: number; }, which can hide fields; declare b on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### v09_hiding_assign_control.ts — JavaScript and native

```text
adamic: /workspace/adamic/probe/view-site-roads/v09_hiding_assign_control.ts:4:7: Adamic 0.1 refuses optional property b in { a: number; b?: number; } absent from structural source { a: number; }, which can hide fields; declare b on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

## Evidence and scope

Full probe logs: `/tmp/vsr-logs`. Control logs: `/tmp/adamic-gate/adamic-probe.Qzd2IB/logs`. Captured runner outputs: `/tmp/vsr-control.txt` and `/tmp/vsr-probe.txt`; exit files both contain `0`. These /tmp artifacts remain local and are not part of the commit.

This is a probe report, not a compiler fix. The report branch is based directly on the requested HEAD and contains only `proof/lower-vsp/report.md`.
