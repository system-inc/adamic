#!/usr/bin/env python3
"""Compare owned .a candidates through a scratch compatibility overlay."""
import argparse, json, os, shutil, subprocess
from pathlib import Path
owned = Path(__file__).resolve().parent
repository = owned.parents[4]
parser = argparse.ArgumentParser()
parser.add_argument('--scratch', type=Path, required=True)
parser.add_argument('--run', default='^TestNextSupported$')
args = parser.parse_args()
scratch = args.scratch.resolve()
scratch.mkdir(parents=True, exist_ok=True)
paths = ['stage1/cohere/lint/registry/registry.go', 'stage1/cohere/lint/lint_test.go']
for relative in paths:
    target = scratch / relative
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(repository / relative, target)
with (scratch/'patch.log').open('w') as log:
    subprocess.run(['patch', '--batch', '-p1', '-d', str(scratch), '-i', str(owned.parent/'structure-tailwind-no-physical-direction'/'compatibility.patch')], stdout=log, stderr=subprocess.STDOUT, check=True)
test = scratch / paths[1]
s = test.read_text()
s = s.replace('Rule, Source, Outcome, FixedSource string', 'Rule, Source, Outcome, FixedSource, File string')
s = s.replace('fmt.Sprintf("case-%03d.ts", i)', 'fmt.Sprintf("case-%03d%s", i, filepath.Ext(row.File))')
s = s.replace('side, err := filepath.Abs("testdata/oracle.go")', 'side, err := filepath.Abs(os.Getenv("WAVE07_ORACLE"))')
# The original testcase filename gates the Tailwind rule. Keep its extension.
# Cohere tests intentionally include malformed source; the rule API still visits it.
oracle = (repository/'stage1/cohere/lint/testdata/oracle.go').read_text()
oracle = oracle.replace('file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: path, Path: tspath.Path(path)}, source, core.ScriptKindTS)', 'script := core.ScriptKindTS; if strings.HasSuffix(path, ".tsx") { script = core.ScriptKindTSX }; file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: path, Path: tspath.Path(path)}, source, script)')
start = oracle.index('\tif len(file.Diagnostics()) != 0 {')
end = oracle.index('\n\tselected :=', start)
oracle = oracle[:start] + oracle[end:]
oracle = oracle.replace('result, err := edit.FixText(path, source,', 'if fields[1] == "@eslint-community/eslint-comments/require-description" || fields[1] == "structure/tailwind-no-physical-direction" { fmt.Fprintf(out, "fixed\\t%s\\n", written(source)); return len(diagnostics) }; result, err := edit.FixText(path, source,')
(scratch/'oracle.go').write_text(oracle)
test.write_text(s)
replace = {str(repository/p):str(scratch/p) for p in paths}
replace[str(repository/'stage1/cohere/lint/wave07_overlay_test.go')] = str(owned/'validation_test.go.txt')
overlay = scratch/'overlay.json'
overlay.write_text(json.dumps({'Replace':replace}))
environment = os.environ.copy()
environment['WAVE07_ORACLE'] = str(scratch/'oracle.go')
environment['ADAMIC_GATE_UNCACHED'] = '1'
command = ['go','test','-overlay='+str(overlay),'./stage1/cohere/lint','-count=1','-v','-timeout=20m','-run',args.run]
print(' '.join(command), flush=True)
with (scratch/'validation.log').open('w') as log:
    result = subprocess.run(command, cwd=repository, env=environment, stdout=log, stderr=subprocess.STDOUT)
print('exit='+str(result.returncode)+' log='+str(scratch/'validation.log'),flush=True)
raise SystemExit(result.returncode)
