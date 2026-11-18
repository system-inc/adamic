#!/usr/bin/env python3
from pathlib import Path
import json,gzip,hashlib,collections,subprocess
W=Path('/workspace');E=W/'depthgate2-base/stage3/typescript/depthgate-census-2/evidence';R=W/'depthgate2-runs'
checks=[]
def good(test,msg):
 assert test,msg
 checks.append('PASS '+msg)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def git(root,*args):return subprocess.check_output(['git','-C',str(root),*args],text=True).strip()
expected={'base':'c56b022995e1e960bad71626fd2074bc46aee485','head':'26d208af47f81143ff1db2280a5b43292deb96ba'}
manifest={name:val for val,name in (l.split('  ',1) for l in (E/'adapted-sources.sha256').read_text().splitlines())}
good(len(manifest)==81 and all(sha(W/'depthgate2-adapted'/p)==v for p,v in manifest.items()),'81-source adapted manifest unchanged')
all_events=[];runs=list(R.iterdir());good(len(runs)==12,'12 run directories')
for r in sorted(runs):
 s=json.loads((r/'status.json').read_text());records=[json.loads(l) for l in (r/'latent.jsonl').read_text().splitlines()];h=records[0];label=s['label'];env=s['environment']
 good(s['exit_code']==0 and not s['timed_out'] and s['wall_seconds']<1800 and records[-1]['status']=='relation_report',r.name+': normal complete run, under 30 minutes')
 good(h['latent_mode']=='full' and h['module_format']=='CommonJSBundle' and not h['checker_rejected'] and not h['diagnostics'] and len(h['sources'])==81,r.name+': full/CommonJSBundle/checker-clean 81-source header')
 good(h['roots']==['/workspace/depthgate2-adapted/src/tsc/tsc.ts'] and {p.removeprefix('/workspace/depthgate2-adapted/') for p in h['sources']}==set(manifest),r.name+': shared entry root and source set')
 good(all(env.get(k)=='1' for k in ['LATENT_FULL','LATENT_ASSERT_NO_OUTPUT','LATENT_TAIL']) and env['CENSUS_MODULE_FORMAT']=='CommonJSBundle' and not any(k.startswith('ADAMIC_PROBE_') for k in env) and Path(env['LATENT_ONLY_FILE']).is_absolute(),r.name+': requested exact-file environment; no probe flag')
 good(s['sha']==expected[label] and s['binary_sha256']==sha(W/('depthgate2-tools-'+label)/'entry-census') and s['selected_source_sha256']==sha(Path(env['LATENT_ONLY_FILE'])),r.name+': pinned own-side tool and unchanged selected source')
 good(not any(word in (r/'stderr.log').read_bytes() for word in [b'DEPTHGATE',b'DEPTHWALK']),r.name+': no probe output')
 for name in ['latent.jsonl','stderr.log','stdout.log']:good(gzip.decompress((E/'runs'/r.name/(name+'.gz')).read_bytes())==(r/name).read_bytes(),r.name+': entire '+name+' archive is lossless')
 events=[f for rec in records for f in rec.get('findings',[])]
 if label=='head':good(not any(f['reason']=='a generic comparison deeper than 4 positions that can reach a Map' for f in events),r.name+': no identity-bound finding anywhere')
 all_events.extend(events)
ordered=sorted((json.loads((r/'status.json').read_text()) for r in runs),key=lambda x:x['started_utc'])
good(len({s['hostname'] for s in ordered})==1,'same hostname for all runs')
good(all(a['finished_utc']<=b['started_utc'] for a,b in zip(ordered,ordered[1:])),'all 12 runs sequential')
expected_order=[(n,l,i) for n in ['utilitiesPublic','transformers-utilities','utilities'] for l,i in [('base',1),('head',1),('head',2),('base',2)]]
good([(s['name'],s['label'],s['repeat']) for s in ordered]==expected_order,'ABBA order per file')
good(not any(f['reason']=='a fit walk past 4096 distinct pairs' for f in all_events),'no cap finding in target or non-target records')
ss=json.loads((E/'site-outcomes.json').read_text());t=json.loads((E/'targets.json').read_text());good(len(ss)==40 and len({s['where'] for s in ss})==40 and all(x['where'] and len(x['matches'])==1 for x in t),'40 unique exactly mapped sites')
good(all(sum(x['repeats'][i]['base_gate_confirmed'] for x in ss)==38 for i in [0,1]),'both baseline repeats confirm 38 gate targets; other two are real Map')
good(collections.Counter(x['classification'] for x in ss)=={'lowers at mapped site':34,'later stop in owning attempt':4,'real Map check':2},'head classification: 34 lower, 4 later, 2 real Map')
good(collections.Counter(x['classification'] for x in ss if x['probe_54210cd4_outcome']=='identity bound')=={'lowers at mapped site':30,'later stop in owning attempt':4},'prior 34 identity-bound: 30 lower, 4 later')
repeat=json.loads((E/'repeat-stability.json').read_text());good(len(repeat)==6 and all(v['both_complete'] and v['normalized_events_equal'] and v['site_unit_events_equal'] for v in repeat.values()),'six same-side repeat comparisons agree by normalized and exact site/unit event multisets')
for label in ['base','head']:
 wt=W/('depthgate2-'+label)
 good(not git(wt,'diff','--name-only'),label+': all tracked checkout sources unchanged')
 good(not git(wt,'diff','--name-only',expected[label],'--','internal','stage3/adaptations','stage3/ledger','cohere'),label+': protected sources unchanged')
 good(not git(wt,'diff','--name-only','--cached'),label+': no staged source changes before report staging')
 checks.append('INFO '+label+' working-tree status: '+git(wt,'status','--short').replace('\n','; '))
 good(git(wt/'cohere','rev-parse','HEAD')=='5360753e19e4417675a90e92a1823bea66b13918' and git(wt/'cohere/TypeScript','rev-parse','HEAD')=='d92d9bfee114c80be2c375d72edae966176e3a4f',label+': submodule gitlinks verified')
 good(git(wt,'rev-parse','HEAD')==expected[label],label+': checkout still at tool input SHA before report commit')
(E/'validation.txt').write_text('\n'.join(c.rstrip() for c in checks)+'\n')
print('Validated',len(checks),'checks/observations')
