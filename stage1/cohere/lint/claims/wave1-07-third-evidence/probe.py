"""Run independent Go suggestion geometry and the shared oracle rejection."""
import json, subprocess, argparse
from pathlib import Path
owned=Path(__file__).resolve().parent
root=owned.parents[4]
parser=argparse.ArgumentParser();parser.add_argument('--scratch',type=Path,required=True);args=parser.parse_args()
scratch=args.scratch.resolve();scratch.mkdir(parents=True,exist_ok=True)
for slug in ['typescript-no-non-null-asserted-optional-chain','typescript-no-non-null-assertion']:
 virtual=root/'cohere/adamic_wave07_suggestion.go'
 source=owned/slug/'testdata/suggestion-probe.go.txt'
 overlay=scratch/(slug+'.json');overlay.write_text(json.dumps({'Replace':{str(virtual):str(source)}}))
 with (scratch/(slug+'.txt')).open('w') as log:
  result=subprocess.run(['go','run','-overlay='+str(overlay),str(virtual)],cwd=root/'cohere',stdout=log,stderr=subprocess.STDOUT)
 if result.returncode:raise SystemExit(result.returncode)
# Preserve the actual shared output protocol: do not weaken the shape assertion.
source=(root/'stage1/cohere/lint/testdata/oracle.go').read_text()
source=source.replace('"github.com/system-inc/cohere/internal/lint/rule"','"github.com/system-inc/cohere/internal/lint/rule"\n rules "github.com/system-inc/cohere/internal/lint/rules/typescript"')
source+='\ntype registeredRule struct {subject rule.Rule; options func([]string) any}\nfunc registeredRules() []registeredRule {return []registeredRule{{rules.NoNonNullAssertedOptionalChain, func([]string)any{return nil}},{rules.NoNonNullAssertion,func([]string)any{return nil}}}}\n'
path=scratch/'oracle.go';path.write_text(source)
virtual=root/'cohere/adamic_wave07_oracle.go';overlay=scratch/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(path)}}))
for slug,name in [('typescript-no-non-null-asserted-optional-chain','@typescript-eslint/no-non-null-asserted-optional-chain'),('typescript-no-non-null-assertion','@typescript-eslint/no-non-null-assertion')]:
 fixture=scratch/(slug+'.ts');fixture.write_bytes((owned/slug/'testdata/witness.ts.txt').read_bytes())
 manifest=scratch/(slug+'-manifest.txt');manifest.write_text(str(fixture)+'\t'+name+'\n')
 logpath=scratch/(slug+'-shared-oracle.txt')
 with logpath.open('w') as log:
  result=subprocess.run(['go','run','-overlay='+str(overlay),str(virtual),'--manifest',str(manifest)],cwd=root/'cohere',stdout=log,stderr=subprocess.STDOUT)
 if result.returncode==0 or 'panic: unexpected suggestion shape' not in logpath.read_text():raise SystemExit('oracle gap changed')
 print(name+': Go finding and shared-oracle suggestion rejection confirmed')
