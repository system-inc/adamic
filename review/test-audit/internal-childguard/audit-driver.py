import pathlib, subprocess, difflib,json,time
root=pathlib.Path('/workspace/adamic'); p=root/'internal/childguard/childguard.go'; base=p.read_text(); out=root/'review/test-audit/internal-childguard'
# Menu fixed before any mutant execution. Every target is production code.
menu=[
('M01',39,'if e.FirstOutput {','if !e.FirstOutput {','flip condition'),
('M02',44,'return fmt.Sprintf("ceiling: %s", e.Elapsed)','return fmt.Sprintf("limit: %s", e.Elapsed)','change constant'),
('M03',61,'w.activity.last = time.Now()','w.activity.last = time.Time{}','change constant'),
('M04',62,'w.activity.seen = true','w.activity.seen = false','change constant'),
('M05',69,'return w.dst.Write(p)','return len(p), nil','return early'),
('M06',77,'options.FirstOutput = DefaultFirstOutput','options.FirstOutput = time.Nanosecond','change option'),
('M07',80,'options.Stall = DefaultStall','options.Stall = time.Nanosecond','change option'),
('M08',83,'options.Ceiling = DefaultCeiling','options.Ceiling = time.Nanosecond','change option'),
('M09',91,'cmd.SysProcAttr.Setpgid = true','cmd.SysProcAttr.Setpgid = false','change option'),
('M10',94,'defer func() { cmd.Stdout, cmd.Stderr = originalOut, originalErr }()','','drop statement'),
('M11',103,'cmd.Stdout = wrap(originalOut)','cmd.Stdout = wrap(originalErr)','swap argument'),
('M12',108,'cmd.Stderr = wrap(originalErr)','cmd.Stderr = wrap(originalOut)','swap argument'),
('M13',120,'return err','return fmt.Errorf("%v", err)','change return option'),
('M14',132,'firstOutput := !a.seen','firstOutput := a.seen','flip condition'),
('M15',135,'window = options.FirstOutput','window = options.Stall','change option'),
('M16',140,'if elapsed >= options.Ceiling {','if elapsed < options.Ceiling {','flip condition'),
('M17',146,'syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)','syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)','change PID multiplier constant from -1 to +1'),
('M18',153,'load = fields[0]','load = "unavailable"','change constant'),
('P01',75,'func Run(cmd *exec.Cmd, options Options) error {','func Run(cmd *exec.Cmd, options Options) error {\n if true { return nil }','empty-answer probe')]
records=[]; switched=base
for mid,line,old,new,kind in menu:
 lines=base.splitlines(True); assert old in lines[line-1],(mid,lines[line-1]); lines[line-1]=lines[line-1].replace(old,new); mutated=''.join(lines)
 diff=''.join(difflib.unified_diff(base.splitlines(True),mutated.splitlines(True),fromfile='a/internal/childguard/childguard.go',tofile='b/internal/childguard/childguard.go'))
 (out/(mid+'.diff')).write_text(diff)
 records.append(dict(id=mid,line=line,old=old,new=new,kind=kind))
 if mid=='P01': replacement=old+'\n if os.Getenv("ADAMIC_MUTANT") == "P01" { return nil }'
 elif old.startswith('if '): replacement='if (os.Getenv("ADAMIC_MUTANT") != "'+mid+'" && ('+old[3:-2]+')) || (os.Getenv("ADAMIC_MUTANT") == "'+mid+'" && ('+new[3:-2]+')) {'
 elif mid=='M17': replacement='syscall.Kill(func() int { if os.Getenv("ADAMIC_MUTANT") == "M17" { return cmd.Process.Pid }; return -cmd.Process.Pid }(), syscall.SIGKILL)'
 elif mid=='M14': replacement='firstOutput := !a.seen\n if os.Getenv("ADAMIC_MUTANT") == "M14" { firstOutput = a.seen }'
 else: replacement='if os.Getenv("ADAMIC_MUTANT") == "'+mid+'" { '+new+' } else { '+old+' }'
 # replace on exact origin line in reverse afterward to avoid shifting
 records[-1]['switch']=replacement
(out/'menu.json').write_text(json.dumps(records,indent=2))
# validate every standalone against original before installing switch
for rec in records:
 lines=base.splitlines(True); lines[rec['line']-1]=lines[rec['line']-1].replace(rec['old'],rec['new']); p.write_text(''.join(lines))
 with (out/(rec['id']+'.vet.log')).open('w') as log: r=subprocess.run(['go','vet','./internal/childguard/'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
 if r.returncode: p.write_text(base); raise SystemExit('vet failed '+rec['id'])
lines=base.splitlines(True)
for rec in records: lines[rec['line']-1]=lines[rec['line']-1].replace(rec['old'],rec['switch'])
p.write_text(''.join(lines)); subprocess.run(['gofmt','-w',str(p)],check=True)
with (out/'switch-build.log').open('w') as log: subprocess.run(['go','test','-c','-o','/tmp/u016-childguard.test','./internal/childguard/'],cwd=root,stdout=log,stderr=subprocess.STDOUT,check=True)
(out/'switch.diff').write_text(''.join(difflib.unified_diff(base.splitlines(True),p.read_text().splitlines(True),fromfile='a/internal/childguard/childguard.go',tofile='b/internal/childguard/childguard.go')))
rows=['TestChild','TestProgress','TestStalled','TestCeiling','TestExitIsNotGuardError','TestKillsProcessGroup','TestNoFirstOutput']
timing=[]
for row in rows:
 for i in range(3):
  with (out/(row+f'.time{i+1}.log')).open('w') as log:
   start=time.monotonic(); r=subprocess.run(['go','test','-count=1','-timeout','90s','./internal/childguard/','-run','^'+row+'$'],cwd=root,stdout=log,stderr=subprocess.STDOUT); timing.append(dict(test=row,run=i+1,wall=time.monotonic()-start,exit=r.returncode))
(out/'timing.json').write_text(json.dumps(timing,indent=2))
import os
runs=[]
for rec in records:
 env=os.environ.copy(); env['ADAMIC_MUTANT']=rec['id']
 start=time.monotonic()
 with (out/(rec['id']+'.log')).open('w') as log: r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/childguard/','-run','.'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 runs.append(dict(id=rec['id'],wall=time.monotonic()-start,exit=r.returncode)); (out/'runs.json').write_text(json.dumps(runs,indent=2))
 print(rec['id'],r.returncode,flush=True)
p.write_text(base)
