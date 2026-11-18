from pathlib import Path
import json,hashlib,gzip,subprocess,re,ast
W=Path('/workspace');root=W/'depthgate2-base/stage3/typescript/depthgate-census-2';out=root/'profiles';checks=[]
def good(t,m):
 assert t,m
 checks.append('PASS '+m)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def git(p,*args):return subprocess.check_output(['git','-C',str(p),*args],text=True).strip()
for line in (root/'evidence/adapted-sources.sha256').read_text().splitlines():
 h,p=line.split('  ',1);good(sha(W/'depthgate2-adapted'/p)==h,'adapted source unchanged: '+p)
for side,expected in [('base','c56b022995e1e960bad71626fd2074bc46aee485'),('head','26d208af47f81143ff1db2280a5b43292deb96ba')]:
 source=W/f'depthgate2-{side}'
 paths=git(source,'diff','--name-only',expected).splitlines();good(all(p.startswith('stage3/typescript/depthgate-census-2/') for p in paths),side+': compiler/source tree matches requested SHA')
 good(not git(source,'diff','--name-only'),side+': tracked checkout files unchanged before report staging')
 original=json.loads((W/f'depthgate2-tools-{side}/entry-overlay/overlay.json').read_text())['Replace'];modified=json.loads((W/f'depthgate2-profiles/{side}/overlay.json').read_text())['Replace']
 good(set(original)==set(modified) and [k for k in original if original[k]!=modified[k]]==[str(source/'stage3/census/latent/tool/main.go')],side+': ONLY scratch overlay main replaced')
 p=W/'depthgate2-profile-runs'/f'{side}-utilities-r1';s=json.loads((p/'status.json').read_text());rs=[json.loads(l) for l in (p/'latent.jsonl').read_bytes().splitlines()];h=rs[0]
 good(s['sha']==expected and s['binary_sha256']==sha(W/f'depthgate2-profiles/{side}/entry-census'),side+': expected SHA and exact rebuilt binary hash')
 good(s['exit_code']==0 and not s['timed_out'] and s['wall_seconds']<1800 and rs[-1]['status']=='relation_report',side+': complete normal run under 30 minutes')
 good(h['latent_mode']=='full' and h['module_format']=='CommonJSBundle' and not h['checker_rejected'] and not h['diagnostics'] and len(h['sources'])==81,side+': full/CommonJSBundle/checker-clean shared tree')
 expected_env=json.loads((W/'depthgate2-runs'/f'{side}-utilities-r1/status.json').read_text())['environment'];env=s['environment'].copy();profile_path=env.pop('ADAMIC_CPU_PROFILE');good(env==expected_env,side+': environment identical to census-2 except CPU profile path')
 good(profile_path==str(p/'cpu.pprof') and s['selected_source_sha256']==sha(W/'depthgate2-adapted/src/compiler/utilities.ts'),side+': exact selected source and intended profile path')
 good((out/(side+'.pprof')).read_bytes()==(p/'cpu.pprof').read_bytes(),side+': committed pprof bytes equal original capture')
 for name in ['latent.jsonl','stderr.log','stdout.log']:good(gzip.decompress((out/'runs'/side/(name+'.gz')).read_bytes())==(p/name).read_bytes(),side+': whole '+name+' archive lossless')
 for mode in ['top','top-cum']:good((out/'analysis'/side/(mode+'-first40.txt')).read_bytes()==b''.join((out/'analysis'/side/(mode+'.txt')).read_bytes().splitlines(keepends=True)[:40]),side+': literal first 40 lines of '+mode)
 for name in (W/'depthgate2-profile-analysis'/side).glob('*.raw.txt'):good(gzip.decompress((out/'analysis'/side/(name.name+'.gz')).read_bytes())==name.read_bytes(),side+': raw pprof report lossless: '+name.name)
 good(not any(w in (p/'stderr.log').read_bytes() for w in [b'DEPTHGATE',b'DEPTHWALK']),side+': no probe logging')
good((W/'depthgate2-profiles/base/entry_main.go').read_bytes()==(W/'depthgate2-profiles/head/entry_main.go').read_bytes(),'identical profiling hook and entire overlay main on both sides')
b=json.loads((out/'behavior-comparison.json').read_text());good(set(b)=={'base','head'} and all(len(v['comparisons'])==2 and all(c['byte_identical'] and c['sha256']==v['sha256'] for c in v['comparisons']) for v in b.values()),'each profiled findings-array byte stream equals BOTH census-2 repeats, ordered and unnormalized')
s=[json.loads((out/'runs'/side/'status.json').read_text()) for side in ['base','head']];good(s[0]['hostname']==s[1]['hostname'] and s[0]['finished_utc']<=s[1]['started_utc'],'same-machine sequential base then head capture')
a=json.loads((out/'attribution.json').read_text())
for side in ['base','head']:
 v=a[side]['buckets'];good(v['mentionsHeld']['cpu_ns']+v['walk_excluding_mentionsHeld']['cpu_ns']+v['outside_walk_and_mentionsHeld']['cpu_ns']==v['total']['cpu_ns'],side+': disjoint sampled CPU partition exact')
 good(v['typeNesting_in_walk_excluding_mentionsHeld']['cpu_ns']+v['walk_excluding_mentionsHeld_and_typeNesting']['cpu_ns']==v['walk_excluding_mentionsHeld']['cpu_ns'],side+': nesting subset partition exact')
 good(v['pair_anchors_excluding_mentionsHeld']['cpu_ns']==v['walk_excluding_mentionsHeld']['cpu_ns']+v['pair_anchors_without_root_excluding_mentionsHeld']['cpu_ns'],side+': root-omitting pair-frame diagnostic exact')
 good(a[side]['profile_sha256']==sha(out/(side+'.pprof')),side+': analyzed profile is captured profile')
for p in out.rglob('*.py'):ast.parse(p.read_text())
checks.append('PASS all packaged Python scripts parse')
for line in (root/'evidence/SHA256SUMS').read_text().splitlines():
 h,p=line.split('  ',1);assert sha(root/'evidence'/p)==h,p
checks.append('PASS all earlier census-2 evidence checksums unchanged')
(out/'validation.txt').write_text('\n'.join(c.rstrip() for c in checks)+'\n')
print('Validated',len(checks),'checks')
