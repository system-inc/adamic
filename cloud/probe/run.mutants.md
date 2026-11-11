# Mutant for the probe runner (#y1p4qzt)

`cloud/probe/run.sh`, the two comparisons that decide agreement. Old:

```sh
			if cmp -s "$out/$build_name.js.stdout" "$out/node.stdout" && [ "$js_exit" = "$node_exit" ]; then
			if cmp -s "$out/$build_name.native.stdout" "$out/node.stdout" && [ "$native_exit" = "$node_exit" ]; then
```

New, every backend reads as agreeing with Node:

```sh
			if true; then
			if true; then
```

Run: `cloud/probe/run.sh --check-control`

Red: control/differs.a reads as agrees on both backends, and the run exits 1 with "the control didn't read as agrees.ts agreeing and differs.a differing".
