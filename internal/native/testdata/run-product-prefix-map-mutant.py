#!/usr/bin/env python3
"""Compare retained product bytes with cmp, then prove dropping maps goes red."""
import os
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[3]
source = root / "internal/native/product_compile_test.go"
original = source.read_text()
seam = "\treturn mapped\n"
capture = "\t\t\t\tproduct[file.name] = data\n"
assert original.count(seam) == original.count(capture) == 1
logs = Path(tempfile.mkdtemp(prefix="adamic-product-prefix-mutant-"))
try:
    for name in ("mapped", "drop-maps", "absolute-source", "absolute-includes"):
        mutant = name != "mapped"
        evidence = logs / name
        evidence.mkdir()
        # Retain the exact bytes the regression compares, only in this temporary
        # test copy. No capture switch or mutant branch remains in the Go test.
        changed = original.replace(capture, capture + '''
                if err := os.WriteFile(filepath.Join(os.Getenv("ADAMIC_PREFIX_MAP_EVIDENCE"), fmt.Sprintf("%d-%s", len(products), file.name)), data, 0600); err != nil { t.Fatal(err) }
''')
        if name == "drop-maps":
            changed = changed.replace(seam, "\treturn arguments\n")
        elif name == "absolute-source":
            relative_seam = "\t\t\tmapped[index] = relative\n"
            assert changed.count(relative_seam) == 1
            changed = changed.replace(relative_seam, "\t\t\t_ = relative\n\t\t\tmapped[index] = absolute\n")
        elif name == "absolute-includes":
            for target in (
                '\t\t\tmapped[index+1] = alias(arguments[index+1], fmt.Sprintf("include-%d", include))\n',
                '\t\t\tmapped[index] = "-I" + alias(argument[2:], fmt.Sprintf("include-%d", include))\n',
            ):
                assert changed.count(target) == 1
                assignment = "mapped[index+1] = arguments[index+1]" if "index+1" in target else "mapped[index] = argument"
                changed = changed.replace(target, target + "\t\t\t" + assignment + "\n")
        source.write_text(changed)
        log = evidence / "test.log"
        with log.open("w") as output:
            result = subprocess.run(
                ["go", "test", "./internal/native", "-run", "^TestProductCompilerBytesAreIndependentOfDirectory$",
                 "-count=1", "-timeout", "90s", "-parallel", "4", "-v"],
                cwd=root, env={**os.environ, "GOMAXPROCS": "4", "ADAMIC_PREFIX_MAP_EVIDENCE": str(evidence)},
                stdout=output, stderr=subprocess.STDOUT, timeout=120, check=False,
            )
        observation = log.read_text()
        print(observation, end="")
        assert result.returncode == int(mutant), f"unexpected {name} test exit"
        for file in ("heap.o", "runtime.a", "harness", "main.c"):
            comparison = subprocess.run(
                ["cmp", "-s", str(evidence / f"0-{file}"), str(evidence / f"1-{file}")],
                timeout=10, check=False,
            )
            expected = int(mutant and file != "main.c")
            assert comparison.returncode == expected, f"unexpected cmp result for {name} {file}"
            print(f"{name}: cmp {file}: {'different' if expected else 'identical'}")
            if expected:
                assert f"{file} differs at byte" in observation, f"byte comparison did not catch {file}"
        assert "fixture and mutant must compile:" not in observation
        assert "clang:" not in observation
    print("caught drop-maps, absolute-source and absolute-includes mutants by object, archive and harness byte comparisons")
    print(f"evidence: {logs}")
finally:
    source.write_text(original)
