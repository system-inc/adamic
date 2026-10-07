#!/usr/bin/env python3
"""Prepare a scratch Go overlay without modifying shared infrastructure."""
from pathlib import Path
import argparse,json,subprocess
owned=Path(__file__).resolve().parent
repository=owned.parents[4]
parser=argparse.ArgumentParser()
parser.add_argument('--scratch',type=Path,required=True)
args=parser.parse_args();scratch=args.scratch.resolve();scratch.mkdir(parents=True,exist_ok=True)
paths=['stage1/cohere/lint/registry/registry.go','stage1/cohere/lint/lint_test.go','stage1/cohere/lint/profile_test.go']
for relative in paths:
 target=scratch/relative;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes((repository/relative).read_bytes())
with (scratch/'patch.log').open('w') as log:
 subprocess.run(['patch','--batch','-p1','-d',str(scratch),'-i',str(owned/'compatibility.patch')],stdout=log,stderr=subprocess.STDOUT,check=True)
# Preserve upstream test filename extensions; a .js exemption is not a .ts case.
test=scratch/paths[1];text=test.read_text().replace('Rule, Source, Outcome, FixedSource string','Rule, File, Source, Outcome, FixedSource string').replace('fmt.Sprintf("%s\\t%+v\\t%s", row.Rule, row.Options, row.Source)','fmt.Sprintf("%s\\t%s\\t%+v\\t%s", row.Rule, row.File, row.Options, row.Source)').replace('fmt.Sprintf("case-%03d.ts", i)','fmt.Sprintf("case-%03d%s", i, filepath.Ext(row.File))');test.write_text(text)
# The old Go driver forced TS even for TSX input. Keep this fix isolated too.
relative='stage1/cohere/lint/testdata/oracle.go';target=scratch/relative;target.parent.mkdir(parents=True,exist_ok=True)
text=(repository/relative).read_text().replace('file := parser.ParseSourceFile(', 'kind := core.ScriptKindTS\n if strings.HasSuffix(path, ".tsx") { kind = core.ScriptKindTSX }\n file := parser.ParseSourceFile(').replace('source, core.ScriptKindTS)', 'source, kind)');target.write_text(text);paths.append(relative)
# Nested Go builds need the corrected oracle as their actual overlay target.
text=test.read_text().replace('filepath.Abs("testdata/oracle.go")', 'filepath.Abs('+json.dumps(str(target))+')');test.write_text(text)
replacements={str(repository/relative):str(scratch/relative) for relative in paths}
replacements[str(repository/'stage1/cohere/lint/wave05_overlay_test.go')]=str(owned/'validation_test.go.txt')
(scratch/'overlay.json').write_text(json.dumps({'Replace':replacements}))
print(scratch/'overlay.json')
