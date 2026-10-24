"""Run and restore the preflight-name and unused-import mutants. Usage: audit_overlay_drift.py LOG_DIRECTORY."""
from pathlib import Path
import subprocess
import sys
root = Path(__file__).resolve().parents[3]
logs = Path(sys.argv[1]).resolve()
logs.mkdir(parents=True, exist_ok=True)
def run(name, relative, change, command, expected):
    path = root / relative
    original = path.read_text()
    mutant = change(original)
    assert mutant != original
    try:
        path.write_text(mutant)
        with open(logs / ('mutant-' + name + '.log'), 'w') as log:
            result = subprocess.run(command, cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=180)
        text = Path(logs / ('mutant-' + name + '.log')).read_text()
        assert result.returncode != 0 and expected in text, (name, result.returncode, text)
        print(name + ': caught', flush=True)
    finally:
        path.write_text(original)
def names(source):
    start = source.index('\t\tcondition, ok := conditional.Cond.(*ast.BinaryExpr)')
    end = source.index('\n\t\tret, ok :=', start)
    return source[:start] + '''\t\tinit := printed(conditional.Init)
		if printed(conditional.Cond) != "err != nil" || (init != "err := l.referenceRefusal(module)" && init != "err := l.parallelPreflight(module)") { continue }''' + source[end:]
run('name-list', 'stage3/census/latent/refusalrewrite/rewrite.go', names,
    ['go','test','./stage3/census/latent/refusalrewrite','-run','^TestCompilerSlice3Shape$','-count=1'], 'unsupported outer refusal returns')
run('skip-imports', 'stage3/census/latent/importclean/main.go', lambda s: s.replace('return output.Bytes(), nil','return source, nil'),
    ['go','test','./stage3/census/latent/importclean','-run','^TestImportCleanDropsOnlyTheImportsARewriteLeftUnused$','-count=1'], '"errors" imported and not used')
