import json,difflib,subprocess,sys
from pathlib import Path
P=Path('review/test-audit/bridge-tsgo-registered_units');S=Path('/tmp/u004');plan=json.load(open(P/'plan.json'));base={m['file']:subprocess.check_output(['git','show','origin/main:'+m['file']],text=True) for m in plan}
def diff(file,text):return ''.join(difflib.unified_diff(base[file].splitlines(True),text.splitlines(True),fromfile='a/'+file,tofile='b/'+file))
for m in plan:
 source=base[m['file']]; assert m['before'] in source
 standalone=source.replace(m['before'],m['after'],1)
 (P/(m['id']+'.diff')).write_text(diff(m['file'],standalone));q=S/'standalone'/m['id'];q.mkdir(parents=True,exist_ok=True);(q/Path(m['file']).name).write_text(standalone)
 (q/'overlay.json').write_text(json.dumps({'Replace':{str(Path(m['file']).resolve()):str(q/Path(m['file']).name)}}))
if sys.argv[-1]=='install':
 switched=dict(base)
 for m in plan:
  file=m['file'];before=m['before'];after=m['after'];go=file.endswith('.go'); select='os.Getenv("ADAMIC_MUTANT") == "'+m['id']+'"' if go else '(getenv("ADAMIC_MUTANT") != NULL && strcmp(getenv("ADAMIC_MUTANT"), "'+m['id']+'") == 0)'
  if m['probe']:
   signature,statement=after.split('\n',1);statement=statement.strip()
   if go:statement='return nil, true, nil'
   replacement=signature+'\n '+('if '+select+' { '+statement+' }' if go else 'if '+select+' { '+statement+' }')
  elif go:replacement='if !l.program.TSGoEnabled() && !('+select+') {'
  elif before.startswith('if (timing())'):replacement='if (!'+select+' && timing()) { uint64_t elapsed = now_ns() - started;'
  elif m['id']=='M01':replacement='('+select+' ? (result.type.length == 0 ? 0 : result.type.length - 1) : result.type.length));'
  elif before.endswith(';'):replacement='if '+select+' { '+after+' } else { '+before+' }'
  else:replacement='('+select+' ? '+after+' : '+before+')'
  switched[file]=switched[file].replace(before,replacement,1)
 for file,source in switched.items():
  if file.endswith('.go'):source=source.replace('"fmt"','"fmt"\n "os"',1)
  Path(file).write_text(source)
 subprocess.run(['gofmt','-w','internal/lower/tsgo.go'],check=True)
 (P/'switch.diff').write_text(subprocess.check_output(['git','diff','--']+list(base),text=True))
print('saved',len(plan),'standalone diffs')
