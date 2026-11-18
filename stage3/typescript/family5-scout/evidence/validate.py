from pathlib import Path
import json,gzip,hashlib,subprocess,collections,re,ast
W=Path('/workspace');root=W/'family5-report/stage3/typescript/family5-scout';E=root/'evidence';A=W/'family5-analysis';R=W/'family5-run/base-tsc-entry-r1';checks=[]
def good(t,m):
 assert t,m
 checks.append('PASS '+m)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def git(p,*args):return subprocess.check_output(['git','-C',str(p),*args],text=True).strip()
record=json.loads(gzip.decompress((E/'record-report.json.gz').read_bytes()));good(record['tree_commit']==record['adamic_commit']=='c56b022995e1e960bad71626fd2074bc46aee485','record compiler/tree commits are c56b0229')
rs=[json.loads(l) for l in (R/'latent.jsonl').read_bytes().splitlines()];status=json.loads((R/'status.json').read_text());header=rs[0]
good(status['exit_code']==0 and not status['timed_out'] and rs[-1]['status']=='relation_report','regenerated entry pass completed normally')
good(len(rs)==83 and len(header['sources'])==81 and {r['file'] for r in rs[1:-1]}==set(header['sources']),'all 81 source records present exactly once plus header/trailer')
good(header['latent_mode']=='first-error' and header['module_format']=='CommonJSBundle' and not header['checker_rejected'] and not header['diagnostics'],'first-error/CommonJSBundle/checker-clean header')
env=status['environment'];good(env['LATENT_FULL']=='0' and env['LATENT_ASSERT_NO_OUTPUT']==env['LATENT_TAIL']=='1' and env['CENSUS_MODULE_FORMAT']=='CommonJSBundle' and 'LATENT_ONLY_FILE' not in env and not any(k.startswith('ADAMIC_PROBE_') for k in env),'exact unrestricted entry mode/tail/no-output flags; no probe flag')
good(header['roots']==['/workspace/depthgate2-adapted/src/tsc/tsc.ts'] and status['sha']=='c56b022995e1e960bad71626fd2074bc46aee485','requested SHA and tsc entry root')
good(sha(W/'depthgate2-tools-base/entry-census')==status['binary_sha256']=='909c659b0238b08320d3bd51717a04ce7119ad4b022685999e2d1be538ce5776','original uninstrumented c56 entry binary unchanged')
for name in ['latent.jsonl','stderr.log','stdout.log']:good(gzip.decompress((E/'run'/(name+'.gz')).read_bytes())==(R/name).read_bytes(),'entire '+name+' archive lossless')
for line in (E/'adapted-sources.sha256').read_text().splitlines():
 h,p=line.split('  ',1);good(sha(W/'depthgate2-adapted'/p)==h,'frozen adapted source unchanged: '+p);good(sha(W/'family5-candidate'/p)==h,'scratch candidate reset: '+p)
pattern=re.compile(r'^optional property (\S+) in (.+) absent from structural source (.+), which can hide fields$');allfs=[f for r in rs for f in r.get('findings',[]) if pattern.fullmatch(f['reason'])];body=[f for f in allfs if f.get('phase')!='tail'];u={tuple(f[k] for k in ['kind','where','reason','text']) for f in body};rc=collections.Counter(x[2] for x in u)
expected={r['reason']:r['count'] for r in record['tsc_entry']['lowering_census']['reason_rows'] if pattern.fullmatch(r['reason'])}
good(len(allfs)==len(body)==len(u)==131 and len({x[1] for x in u})==131 and len(rc)==45 and dict(rc)==expected,'131 positions / meter sites / events, 45 exact reason counts match record; no family tail extras')
sites=json.loads((E/'sites.json').read_text());contexts=json.loads(gzip.decompress((E/'site-contexts.json.gz').read_bytes()));groups=json.loads((E/'group-decisions.json').read_text());ids={s['id'] for s in sites};good(len(sites)==131 and len(ids)==131 and {tuple(s[k] for k in ['kind','where','reason','text']) for s in sites}==u,'complete exact site list')
good(len(contexts)==131 and {s['id'] for s in contexts}==ids and all(s['exact_node_candidates'] for s in contexts),'all sites mapped to exact AST nodes, no nearest-line fallback')
good(collections.Counter(s['relation_position'] for s in contexts)=={'argument':100,'assignment':12,'return':7,'property in object literal':4,'spread':8},'all five outer relation-position counts verified')
seen=[];counts=collections.Counter();groupcounts=collections.Counter()
for g in groups:
 good(g['regenerated_sites']==len(g['site_ids'])==g['meter_sites'],g['id']+': group count matches record and sites')
 flattened=[x for xs in g['by_relation_position'].values() for x in xs];good(set(flattened)==set(g['site_ids']) and len(flattened)==len(g['site_ids']),g['id']+': relation-position subgroups partition sites')
 seen.extend(g['site_ids']);counts[g['road']]+=g['regenerated_sites'];groupcounts[g['road']]+=1
 good(bool(g['evidence']) and bool(g['premise']),g['id']+': road has construction evidence and premise')
good(len(seen)==len(set(seen))==131 and set(seen)==ids,'45 group indexes cover every site once')
good(counts=={'i':109,'ii':8,'iii':14} and groupcounts=={'i':33,'ii':6,'iii':6},'road totals: 109/8/14 sites in 33/6/6 groups')
chosen=json.loads((E/'selected-candidates.json').read_text());good(len(chosen)==14 and {g['candidate'] for g in groups if g['road']=='i'}=={c['id'] for c in chosen},'14 selected candidates cover all and only declaration groups')
for label,filename in [('project','stock-baseline.json'),('strict','stock-strict-baseline.json')]:
 b=json.loads((E/'stock'/filename).read_text());good(b['root_count']==79 and b['diagnostic_count']==0 and not b['diagnostics'],label+': 79-root baseline zero diagnostics')
 for c in chosen:
  result=json.loads((E/'stock'/label/(c['id']+'.json')).read_text());good(result['root_count']==79 and result['diagnostic_count']==0 and not result['diagnostics'],label+' '+c['id']+': zero stock diagnostics')
for c in chosen:
 good(c['emitted_js_byte_identical'] and c['strict_checker_result']['emitted_js_byte_identical'],c['id']+': changed-file emitted JS identical')
 for f in c.get('file_hashes',[{'file':c['file'],'original_sha256':c.get('original_sha256')}]):good(sha(W/'depthgate2-adapted/src/compiler'/f['file'])==f['original_sha256'],c['id']+': original hash matches unchanged source '+f['file'])
 good(bool(c['declaration_locations']),c['id']+': original declaration locations retained')
strict=json.loads((E/'stock/stock-strict-baseline.json').read_text())['options'];good(strict['strict'] and strict['noUncheckedIndexedAccess'] and strict['exactOptionalPropertyTypes'],'stricter options preserved, never relaxed')
projectsummary={c['id']:c for c in json.loads((E/'stock/project/summary.json').read_text())};strictsummary={c['id']:c for c in json.loads((E/'stock/strict/summary.json').read_text())}
good([d['code'] for d in strictsummary['C01']['diagnostics']]==[2379,2379] and projectsummary['C09']['diagnostics'][0]['code']==2320 and projectsummary['C14']['diagnostics'][0]['code']==2411,'all measured failed candidate variants retained')
good(all(projectsummary[k]['diagnostic_count']==strictsummary[k]['diagnostic_count']==0 and k not in {c['id'] for c in chosen} for k in ['C04','C05']),'zero-diagnostic but untruthful marker candidates retained and rejected')
h=json.loads((E/'hazards.json').read_text());d=json.loads((E/'diagnostic-hazard.json').read_text());good(h['packageInputs'][0]['readJsonResult']['name']==42 and h['packageInputs'][1]['cachedPackageJsonContent']['version']==7,'actual package input construction retains wrong-typed hidden fields')
good(h['buildInfoInput']['readBuilderProgramAccepted'] and h['buildInfoInput']['retainedDiagnostic']['messageText']['repopulateInfo']==42,'accepted build-info source chain retains hidden numeric repopulateInfo')
good(d['readBuilderProgramAccepted'] and d['constructedDiagnosticRelatedInformation']['reportsUnnecessary'] is None and d['constructedReusableDiagnosticRelatedInformation']['reportsUnnecessary'] is None,'accepted related-info cache and both exact spread constructors preserve hidden null marker')
good(all(k in d['retainedReusableRelatedInformation'] for k in ['category','code','file','start','length','messageText']),'marker source witness includes every required declared field')
readings=json.loads(gzip.decompress((E/'whole-readings.json.gz').read_bytes()));good(len(readings)==18 and readings[0]['group']=='G01' and readings[1]['group']=='G02' and all(r['function']['text'] for r in readings),'18 whole-function readings, largest groups first')
for label,p in [('report',W/'family5-report'),('tool source',W/'depthgate2-base')]:
 changes=git(p,'diff','--name-only','c56b022995e1e960bad71626fd2074bc46aee485').splitlines();good(all(s.startswith('stage3/typescript/') for s in changes),label+': no adaptation/ledger/internal or other code change relative to c56')
 good(not git(p,'diff','--name-only'),label+': no tracked source edits')
good(git(W/'depthgate2-base/cohere','rev-parse','HEAD')=='5360753e19e4417675a90e92a1823bea66b13918' and git(W/'depthgate2-base/cohere/TypeScript','rev-parse','HEAD')=='d92d9bfee114c80be2c375d72edae966176e3a4f','actual tool source submodules initialized at pinned commits')
for p in E.rglob('*.py'):ast.parse(p.read_text())
checks.append('PASS packaged Python scripts parse')
(E/'validation.txt').write_text('\n'.join(x.rstrip() for x in checks)+'\n');print('Validated',len(checks),'checks')
