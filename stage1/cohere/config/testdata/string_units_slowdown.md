# House-config string length reproduction

`house_config_string_units.a` calls the real house loader once with a missing
local config. It loads the embedded tiers and all four variants, then prints
the ordinary settings result. The fixture measures nothing.

`TestHouseConfigStringUnits` compares the JavaScript backend and sanitized
native result with source-on-Node output. Only native execution has a two-second
deadline; checking, compilation and reference execution are outside it. The
native process group is killed when the deadline expires. The test intentionally
fails on area/library `48123187`.

The source regression is runtime commit `aef95985d`: immortal literals no longer
cache UTF-16 lengths. Repeated length and indexed reads in the JSON parser call
`adamic_string_units`, which rescans the large embedded strings.

Run from the repository root, with setup's environment sourced and Node 24.19.0:

```sh
go test ./stage1/cohere/config -run '^TestHouseConfigStringUnits$' \
  -count=1 -timeout=3m -v > /workspace/string-units-current.log 2>&1
```

## Previous-function control

This substitutes only the previous `adamic_string_units` body through a Go
build overlay. It does not edit runtime code. The old body mutates literal
metadata, so this is a single-threaded diagnostic control, not a thread-safe fix.
All temporary source and output stay outside the checkout.

```sh
export STRING_UNITS_CONTROL_DIR=$(mktemp -d /workspace/string-units-control.XXXXXX)
python3 - <<'PY_CONTROL'
import json
import os
from pathlib import Path
import subprocess

root = Path.cwd()
output = Path(os.environ['STRING_UNITS_CONTROL_DIR'])
source = root / 'internal/native/runtime/string_index.c'
current = source.read_text()
previous = subprocess.check_output(
    ['git', 'show', 'aef95985d^:internal/native/runtime/string_index.c'], text=True)

def function(text):
    begin = text.index('size_t adamic_string_units(const adamic_string *string) {')
    position = text.index('{', begin) + 1
    depth = 1
    while depth:
        if text[position] == '{':
            depth += 1
        if text[position] == '}':
            depth -= 1
        position += 1
    return begin, position

begin, end = function(current)
old_begin, old_end = function(previous)
replacement = output / 'string_index.c'
replacement.write_text(current[:begin] + previous[old_begin:old_end] + current[end:])
(output / 'overlay.json').write_text(json.dumps(
    {'Replace': {str(source): str(replacement)}}))
PY_CONTROL
go test -overlay="$STRING_UNITS_CONTROL_DIR/overlay.json" \
  ./stage1/cohere/config -run '^TestHouseConfigStringUnits$' \
  -count=1 -timeout=3m -v > "$STRING_UNITS_CONTROL_DIR/test.log" 2>&1
```

## Native profile

This adds `-pg` to an external copy of native compilation flags. Both the emitted
program and runtime are instrumented. `ADAMIC_STRING_UNITS_ARTIFACTS` retains the
emitted C and binary in the requested external directory. The regression test
still has its two-second deadline and is expected to fail before the runtime fix.
Check that it reached that deadline, rather than a build or reference failure,
and that `house` exists before proceeding.

```sh
export STRING_UNITS_PROFILE_DIR=$(mktemp -d /workspace/string-units-profile.XXXXXX)
python3 - <<'PY_PROFILE'
import json
import os
from pathlib import Path

root = Path.cwd()
output = Path(os.environ['STRING_UNITS_PROFILE_DIR'])
source = root / 'internal/native/native.go'
text = source.read_text()
needle = 'flags = append(flags, "-ffp-contract=off")'
assert text.count(needle) == 1
replacement = output / 'native.go'
replacement.write_text(text.replace(
    needle, 'flags = append(flags, "-ffp-contract=off", "-pg")', 1))
(output / 'overlay.json').write_text(json.dumps(
    {'Replace': {str(source): str(replacement)}}))
PY_PROFILE
ADAMIC_STRING_UNITS_ARTIFACTS="$STRING_UNITS_PROFILE_DIR" \
  go test -overlay="$STRING_UNITS_PROFILE_DIR/overlay.json" \
  ./stage1/cohere/config -run '^TestHouseConfigStringUnits$' \
  -count=1 -timeout=3m -v > "$STRING_UNITS_PROFILE_DIR/test.log" 2>&1
```

Run the retained binary separately to completion so gprof can write `gmon.out`.
The config path below remains absent: the fixture needs the embedded house tiers,
not an input file. `timeout` bounds this separate profiling run at 30 seconds.
Do not pipe either execution into a reader that can terminate it early.

```sh
(
  cd "$STRING_UNITS_PROFILE_DIR"
  ASAN_OPTIONS=detect_leaks=0 timeout 30s ./house \
    "$STRING_UNITS_PROFILE_DIR/CohereSettings.json" \
    > native.stdout 2> native.stderr
  gprof ./house ./gmon.out > gprof.txt 2> gprof.stderr
)
```

Read `gprof.txt` for the self time and call count of `adamic_string_units`, and
its callers in the JSON parser, string length, indexing and slicing paths.
