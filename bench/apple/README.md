# What a call into AppKit costs

`go run ./bench/apple` (macOS) times three loops written three ways: Adamic (`testdata/crossing.a`), Swift (`crossing.swift`, `-O`) and Objective-C with ARC (`crossing.m`, `-O2`), and a frame of SwiftUI two ways (`frame.a`, `frame.swift`): the supplements list, its rows a class of their own, rendered again and laid out. Each cell is the wall time of the process at N iterations less its time at 0, so starting up and making the window cancel out; runs are interleaved, best of five. A counted build then gives Adamic's allocations, retains and releases per iteration, which don't depend on the machine.

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

## October 6, 2026, 22:43 MDT: frames, and what each crossing counts

Load averages 89 to 138, worse than the first run: wall times this noisy come out negative where a loop is fast, so only the frame row means anything, and only roughly (best of three).

| per iteration | Adamic | Swift |
| --- | --- | --- |
| frame, microseconds | 1646 | 986 |

Adamic's counts per iteration, from the counted build, are exact. Swift's traffic can't be read from inside its process the same way (it would take Instruments), so there's no column for it.

| per iteration | allocations | retains | releases |
| --- | ---: | ---: | ---: |
| scalar (`isHidden`) | 0 | 1 | 1 |
| object (`contentView`) | 1 | 1 | 2 |
| string (`setTitle:`, `title`) | 1 | 2 | 3 |
| frame (the supplements list) | 81 | 179 | 190 |

The object loop's allocation is the box made for every object result; one box per object would remove it, and make `===` true for the same object. A frame's 81 allocations are the views made fresh each render (a `Shown` per view, the row objects, the arrays of children) and the strings crossing. Most would go with views built once and changed, which is what SwiftUI's own diffing assumes. A quiet-machine run is still owed for the wall times.

## October 7, 2026, 05:42 MDT: one box per object

An Objective-C object now has one box while Adamic holds it, so `===` holds for the same object. It doesn't remove the object loop's allocation: nothing holds the content view between iterations, so its box goes with each one and the next crossing makes another (1 allocation, 1 retain, 2 releases, unchanged). What it costs is the table: a CFDictionary, the first cut, took the object loop from about 15 ns to 51; the open-addressed table that replaced it puts it back at 16.8 (best of five, load 26 to 71, so roughly). Removing that allocation needs a box that outlives Adamic's last reference to it, which this doesn't try.

| ns per iteration | Adamic | Swift | Objective-C |
| --- | --- | --- | --- |
| object, CFDictionary | 50.7 | 1.7 | 8.9 |
| object, open-addressed table | 16.8 | 8.1 | 5.9 |
