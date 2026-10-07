#!/usr/bin/env python3
"""Read fetched Git objects and emit the wave-07 ranking/claim audit."""
import json,pathlib,re,subprocess
root=pathlib.Path(__file__).resolve().parents[4]
def git(*args):return subprocess.check_output(['git',*args],cwd=root,text=True)
counts={}
for population in ('compiler','repository'):
 for line in (root/f'stage1/cohere/typeaware/validation-volume/{population}-all.counts').read_text().splitlines():
  if '\t' not in line:continue
  name,count=line.split('\t');counts.setdefault(name,{'compiler':0,'repository':0})[population]=int(count)
refs=git('for-each-ref','--format=%(refname:short)','refs/remotes/origin').splitlines()
claims={name:[] for name in counts};blobs={}
for ref in refs:
 for file in git('ls-tree','-r','--name-only',ref,'stage1/cohere/typeaware/claims/').splitlines():
  if not file.endswith('.md'):continue
  oid=git('rev-parse',ref+':'+file).strip()
  if oid in blobs:continue
  blobs[oid]=ref+':'+file
  text=git('show',ref+':'+file)
  for name in counts:
   if re.search(r'(?<![\w/@-])'+re.escape(name)+r'(?![\w/-])',text):claims[name].append(ref+':'+file)
base='origin/codex/tsgo-c-library'
volume=git('show',base+':stage1/cohere/typeaware/testdata/oracle_volume.go')
identifiers=set(re.findall(r'rules\.(\w+)',volume));ported=set()
for file in (root/'cohere/internal/lint/rules').rglob('*.go'):
 text=file.read_text()
 for identifier in identifiers:
  match=re.search(r'var '+identifier+r'\s*=\s*rule.Rule\s*\{[\s\S]*?Name:\s*"([^"]+)"',text)
  if match:
   name=match.group(1)
   if name not in counts and '@typescript-eslint/'+name in counts:name='@typescript-eslint/'+name
   ported.add(name)
coverage=git('show',base+':stage1/cohere/typeaware/testdata/oracle_coverage.go')
coverage_map=re.search(r'var coverageNames = map\[string\]bool\{([^}]+)\}',coverage).group(1)
ported.update(re.findall(r'"([^"]+)": true',coverage_map))
if len(ported)!=26:raise RuntimeError('original port registry must contain 26 rules: '+repr(sorted(ported)))
for ref in ['origin/main',base]:
 for file in git('ls-tree','-r','--name-only',ref,'stage1').splitlines():
  if not file.endswith(('.a','.ts')) or '/testdata/' in file:continue
  stem=pathlib.PurePosixPath(file).stem
  for name in counts:
   if stem==name.split('/')[-1].replace('-','_'):ported.add(name)
ranking=sorted(counts,key=lambda name:(-sum(counts[name].values()),name))
# Exclude this worker's third claim to reconstruct the pre-claim decision.
selected=['react-hooks/set-state-in-effect','react-hooks/set-state-in-render','react-hooks/static-components']
remaining=[name for name in ranking if name not in ported and (not claims[name] or name in selected and all('wave-07.md' in location for location in claims[name]))]
if remaining[:3]!=selected:raise RuntimeError('claimed trio no longer matches first eligible names: '+repr(remaining[:3]))
print(json.dumps({'refs':{ref:git('rev-parse',ref).strip() for ref in refs},'claim_blobs':blobs,'ported':sorted(ported),'selection':selected,'ranking':[dict(name=name,**counts[name],claims=claims[name],ported=name in ported) for name in ranking]},indent=2))
