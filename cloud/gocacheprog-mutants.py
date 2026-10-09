#!/usr/bin/env python3
"""Break cloud/go-pin.sh and cloud/gocacheprog.sh one way at a time and require test_gocacheprog.py to catch each."""
from pathlib import Path
import os
import subprocess
import tempfile

source = Path(__file__).resolve().parent
scratch = Path(tempfile.mkdtemp(prefix="gocacheprog-mutants-"))
print("mutant logs:", scratch)
mutants = [
    # The pin removed: any Go that answers counts, as setup.sh accepted before the pin.
    ("pin-removed", "go-pin.sh",
     'pinnedGo() { [ "$(cd / && GOTOOLCHAIN=local "$1" env GOVERSION 2> /dev/null)" = "$goPin" ]; }',
     'pinnedGo() { "$1" version > /dev/null 2>&1; }', "test_split_toolchain_never_gets_the_program"),
    # A failed download that has already deleted the machine's Go.
    ("delete-first", "go-pin.sh", '\tstaging=$(mktemp -d "$1/go-pin.XXXXXX")\n',
     '\tstaging=$(mktemp -d "$1/go-pin.XXXXXX")\n\trm -rf "$1/go"\n', "test_split_toolchain_never_gets_the_program"),
    # env.sh lacks the variable.
    ("variable-missing", "gocacheprog.sh", "\\texport GOCACHEPROG=%q\\nfi\\n", "\\t:\\nfi\\n",
     "test_pinned_go_sets_the_program"),
    # env.sh lacks the flags that make actions shareable across trees.
    ("flags-missing", "gocacheprog.sh", "\techo 'case \" ${GOFLAGS:-} \" in", "\t: echo 'case \" ${GOFLAGS:-} \" in",
     "test_pinned_go_sets_the_program"),
    # The flags line reads $GOFLAGS bare, so a shell under set -u dies sourcing env.sh.
    ("flags-unset-unsafe", "gocacheprog.sh", "\techo 'case \" ${GOFLAGS:-} \" in", "\techo 'case \" $GOFLAGS \" in",
     "test_sourced_under_set_u"),
    # An older setup's block survives beside the new one.
    ("legacy-kept", "gocacheprog.sh", '\t/^unset GOCACHEPROG$/ { next }\n', "", "test_older_env_is_replaced_not_stacked"),
    # The key ignores the program's source, so a changed program is never rebuilt.
    ("key-blind", "gocacheprog.sh", ' "$repository"/cmd/adamic-gocacheprog/*.go', "", "test_changed_source_rebuilds"),
]
for name, file, old, new, catcher in mutants:
    directory = scratch / name
    directory.mkdir()
    for original in ["go-pin.sh", "gocacheprog.sh"]:
        text = (source / original).read_text()
        if original == file:
            assert old in text, f"{name}: the line to break isn't in {file}"
            text = text.replace(old, new)
        (directory / original).write_text(text)
    log = scratch / f"{name}.log"
    with log.open("wb") as output:
        result = subprocess.run(["python3", str(source / "test_gocacheprog.py")],
                                env=dict(os.environ, ADAMIC_GOCACHEPROG_SOURCE=str(directory)),
                                stdout=output, stderr=subprocess.STDOUT)
    if result.returncode != 1 or catcher not in log.read_text():
        raise SystemExit(f"mutant {name} escaped or failed for the wrong reason: {log}")
    print(f"{name}: caught by {catcher} (exit 1)")
