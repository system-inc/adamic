# #41vb71c: the cost of a quiet argument demand on a tsc-shaped loop

bench.a is a graph program (a cycle of Nodes) whose interface-method calls go through a parameter, so their targets are Unknown, and each call hands a fresh object and a fresh array the callee doesn't keep: 2 visitors times 5M calls. job.sh builds adamic at the checked-out head and at the parent named in it, compiles bench.a with each, and times 5 alternating runs with /usr/bin/time. Run it with ahra loom proof on one box.

Measured on Home, Nov 9 2026 (job 01M6ZEM4DGN8HRDDEKJBM2GGV2, 64 cores, load 0.8): parent 0f44327f median 0.94 s, head eab085d6 median 1.15 s, 1.223x, peak RSS 1,624 KB on both. home-run.log is that run. Its counted builds failed (exit 2) because --count came before the file; job.sh here has --count after -o, as the CLI wants. For the narrower head, set PARENT in job.sh to the sha it's compared against.
