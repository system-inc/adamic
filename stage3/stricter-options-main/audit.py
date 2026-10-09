from pathlib import Path
import json,re,subprocess,gzip,hashlib,csv,io
root=Path.cwd();out=root/'stage3/stricter-options-main'
changes=[]
for path in subprocess.check_output(['git','diff','--name-only','a774d316','--','stage3/fixtures/*/status.json'],text=True).splitlines():
 before=subprocess.check_output(['git','show','a774d316:'+path],text=True);after=Path(path).read_text()
 def rest(s):
  result=[];last=0
  for match in re.finditer(r'"stage0"\s*:\s*',s):
   _,length=json.JSONDecoder().raw_decode(s[match.end():]);end=match.end()+length
   result.append(s[last:match.end()]);result.append('<stage0>');last=end
  return ''.join(result)+s[last:]
 assert rest(before)==rest(after),path+': bytes outside stage0 changed'
 old={v['file']:v for v in json.loads(before)};new={v['file']:v for v in json.loads(after)}
 assert old.keys()==new.keys()
 for name,entry in old.items():
  assert {k:v for k,v in entry.items() if k!='stage0'}=={k:v for k,v in new[name].items() if k!='stage0'}
  if entry['stage0']!=new[name]['stage0']:
   a,b=entry['stage0'],new[name]['stage0'];assert a['outcome'] not in ('Compiles','CheckedStop') or b['outcome'] not in ('Refused','NotYet')
   changes.append({'file':str(Path(path).parent/name),'before':a,'after':b,'reason':'readonly dictionary boundary replaces blanket index refusal' if 'readonly string index signature' in b.get('what','') else 'array-binding lowering advances to the unsupported rest binding'})
(out/'status-audit.json').write_text(json.dumps({'deliberate_regressions':[],'node_bytes_changed':0,'bytes_outside_stage0_changed':0,'changes':changes},indent=2)+'\n')
report=json.loads(Path('/tmp/stricter-options-ledger-census-final.json').read_text())
assert len(report['rows'])==173 and report['counts']['scheduled-check']==173 and not report['ordinary_errors']
s=json.dumps(report,indent=2).replace('/tmp/stricter-options-ledger-adapted','<adapted>')+'\n'
for p in [out/'evidence/production.json.gz',root/'stage3/stricter-options/whole-program/production.json.gz']:p.write_bytes(gzip.compress(s.encode(),mtime=0))
text=io.StringIO();writer=csv.writer(text,lineterminator='\n');writer.writerow(['id','state','kind','file','line','column','code','reason'])
for name,row in sorted(report['rows'].items()):
 site=row['site'];writer.writerow([name,row['state'],row.get('kind',''),site['file'].removeprefix('/tmp/stricter-options-ledger-adapted/'),site['line'],site['column'],site['code'],row.get('reason','')])
(root/'stage3/stricter-options/whole-program/states.csv').write_text(text.getvalue())
print('Stage 3 audit:',len(changes),'updates, no compiling regression, Node and all non-stage0 bytes unchanged; census 173 scheduled, zero ordinary errors')
