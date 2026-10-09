exec(open('/tmp/u159/audit.py').read().split('for name,pattern in rows.items():')[0])
meta=json.loads((E/'runs.json').read_text())
base=subprocess.check_output(['git','rev-parse','origin/main'],cwd=R,text=True).strip()
menu=json.loads((E/'menu.json').read_text())
for id,path,old,new,kind in menu:
 if id not in ('M1','M2','M3','P1'):continue
 f=R/path;original=f.read_text()
 if id=='P1':
  start=original.index('function run(');end=original.index('\nconst args =',start)
  changed=original[:start]+'function run(path: string, mode: string, countOnly: boolean): number {\n    return 0;\n}\n'+original[end:]
 else:changed=original.replace(old,new)
 (E/'diffs'/(id+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),'a/'+path,'b/'+path)))
 (E/(id+'.log')).rename(E/(id+'-invalid.log'))
 f.write_text(changed)
 try:
  subprocess.run(['git','add',path],cwd=R,check=True)
  subprocess.run(['git','commit','--quiet','-m','Temporary audit '+id],cwd=R,check=True)
  head=subprocess.check_output(['git','rev-parse','HEAD'],cwd=R,text=True).strip();(E/(id+'-commit.txt')).write_text(head+'\n')
  run(id,'.',id+'-replay')
 finally:
  subprocess.run(['git','reset','--hard',base],cwd=R,check=True,stdout=subprocess.DEVNULL)
print('RESTORED',flush=True)
