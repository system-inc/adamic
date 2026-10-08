"""Run source mutants independently, restoring the witness on every exit."""
from pathlib import Path
import subprocess

source = Path(__file__).with_name("repeated-destructure.a")
original = source.read_text()
directive = "// @ts-expect-error The first destructuring diagnoses the undefined tuple.\n"
mutants = {
    "remove-undefined": original.replace(" | undefined", "").replace(directive, ""),
    "remove-first": "\n".join(line for index, line in enumerate(original.split("\n")) if index not in (2, 3)),
    "extra-assignment": original + '\nconst wrong: number = "wrong";\n',
}
try:
    for name, text in mutants.items():
        source.write_text(text)
        log = Path("/tmp") / ("ts2488-mutant-" + name + ".log")
        with log.open("w") as output:
            result = subprocess.run(["node", str(source.with_name("reproduce.cjs"))], stdout=output, stderr=subprocess.STDOUT)
        assert result.returncode != 0 and "unexpected" in log.read_text(), name
        print(name + ": caught by checker observation assertion")
finally:
    source.write_text(original)
