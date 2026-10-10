"""Run each premature-check mutant separately and restore exact source bytes."""
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
cases = [
    ("native-order", "internal/native/view_callable_calls.go",
     "\tpacked, count := e.closureArguments(call)",
     '\te.line("if (%s.kind != adamic_view_union_function) adamic_panic(\\"early callable check\\", sizeof \\"early callable check\\" - 1);", candidate)\n\tpacked, count := e.closureArguments(call)'),
    ("javascript-order", "internal/javascript/view_callable_calls.go",
     '\tinvocation := b.String() +',
     '\traw = "((value) => { if (adamicTypeOf(value) !== \'function\') panic(\'early callable check\'); return value; })(" + raw + ")"\n\tinvocation := b.String() +'),
]
cases.append(("unsupported-domain", "internal/lower/view_lazy.go", "call.CallContract == 0 ||", "call.CallContract == -1 ||"))
cases.append(("unsupported-producer", "internal/lower/view_lazy.go", "graph.viewCallableUncheckableProducer(property, reaches)", "false"))
for name, relative, original, replacement in cases:
    path = root / relative
    data = path.read_bytes()
    text = data.decode()
    assert text.count(original) == 1, name
    log = root / "review/compiler/views-v4" / (name + ".log")
    try:
        path.write_text(text.replace(original, replacement))
        with log.open("w") as output:
            result = subprocess.run(["timeout", "90", "go", "test", "./internal/oracle", "-run", "^TestV4DirectProducerRefusal$" if name == "unsupported-producer" else "^TestV4DirectRecursiveRefusal$" if name == "unsupported-domain" else "^TestV4DirectCalleeOrder$", "-count=1", "-v", "-timeout", "90s"], cwd=root, stdout=output, stderr=subprocess.STDOUT, timeout=95)
        observation = log.read_text()
        if name.startswith("unsupported-"):
            assert result.returncode == 1 and "refusal with path and fix: <nil>" in observation, observation
            print(name + ": caught by compiler refusal assertion; unsupported recursive relation admitted")
        else:
            assert result.returncode == 1 and "early callable check" in observation and 'stdout=""' in observation and "Sanitizer" not in observation, observation
            print(name + ": caught by missing argument output; exit 70; no sanitizer kill")
    finally:
        path.write_bytes(data)
