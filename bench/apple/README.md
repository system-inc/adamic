# What a call into AppKit costs

`go run ./bench/apple` (macOS) times three loops written three ways: Adamic (`testdata/crossing.a`), Swift (`crossing.swift`, `-O`) and Objective-C with ARC (`crossing.m`, `-O2`). Each cell is the wall time of the process at N iterations less its time at 0, so starting up and making the window cancel out; runs are interleaved, best of five.

- **scalar**: a `BOOL` getter, `-[NSView isHidden]`. One message, nothing allocated.
- **object**: an object getter, `-[NSWindow contentView]`, held for the iteration.
- **string**: `-[NSWindow setTitle:]` then `-[NSWindow title]`, a string crossing each way.

## October 6, 2026, M-series Mac, macOS 27.0, Apple clang 21

Not a quiet machine: load averages were 37 to 71 (other work was running), and the Objective-C scalar loop moved from 1.3 to 4.0 ns between the two runs. Read these as orders of magnitude; a quiet-machine run is owed before anything is optimized against them.

| ns per iteration | Adamic | Swift | Objective-C |
| --- | --- | --- | --- |
| scalar, run 1 | 4.8 | 3.0 | 1.3 |
| scalar, run 2 | 4.6 | 2.2 | 4.0 |
| object, run 1 | 15.8 | 7.3 | 9.5 |
| object, run 2 | 13.3 | 3.4 | 7.8 |
| string, run 1 | 201 | 131 | 142 |
| string, run 2 | 194 | 116 | 128 |

What Adamic pays that Swift doesn't, from reading the generated C, not yet measured one by one:

- **object**: a box allocated and freed for every object result (`runtime/foreign.c`). Keeping one box per object would also make `===` true for the same object.
- **scalar**: `adamic_apple_unbox` is a call into another translation unit that checks the value's kind, plus the stack check every Adamic function makes.
- **string**: each crossing goes by UTF-16 units: `adamic_apple_string_from` copies the units out, widens each to a double, and builds the string from those. An ASCII or UTF-8 fast path is the obvious first step.
