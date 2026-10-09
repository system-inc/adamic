#!/usr/bin/env python3
"""Run isolated testgrain mutants; invoke from any directory with Go on PATH."""
from pathlib import Path
import subprocess,os,tempfile,shutil
source=Path(__file__).resolve().parent; original=(source/'grain.go').read_text()
clock=original.replace('var realClock','var mutantStarts sync.Map\nvar realClock').replace('func beginSetup(t testing.TB, key string, c clock) func() {\n\tstarted := c.now()', 'func beginSetup(t testing.TB, key string, c clock) func() {\n\tstarted := c.now()\n mutantStarts.Store(t, started)').replace('func begin(t testing.TB, label string, c clock) func() {\n\tstarted := c.now()', 'func begin(t testing.TB, label string, c clock) func() {\n\tstarted := c.now()\n if prior, ok := mutantStarts.Load(t); ok { started = prior.(time.Time) }')
setupkill=original.replace('func beginSetup(t testing.TB, key string, c clock) func() {\n\tstarted := c.now()', 'func beginSetup(t testing.TB, key string, c clock) func() {\n\tstarted := c.now()\n timer := c.after(Kill, func() { panic("mutant setup kill") })').replace('t.Logf("grain setup %s: %.3f s", key, c.now().Sub(started).Seconds())', 'timer.Stop()\n t.Logf("grain setup %s: %.3f s", key, c.now().Sub(started).Seconds())')
mutants=[('clock before setup',clock,'TestUnitClock'),('kill leader only',original.replace('syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)','syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)'),'TestKill'),('skip test.list',original.replace('unit(t, realClock)\n\tcmd := Command','return\n unit(t, realClock)\n\tcmd := Command'),'TestUnion'),('setup without once',original.replace('p, exists := preparations.values[key]','p, exists := preparations.values[key]\n exists = false'),'TestSetupOnce'),('setup given kill',setupkill,'TestSetupUnbounded'),('context kills leader only',original.replace('err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)','err := syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)'),'TestCommandContext')]
with tempfile.TemporaryDirectory(prefix='testgrain-mutants-') as directory:
 root=Path(directory);(root/'go.mod').write_text('module github.com/system-inc/adamic/internal/testgrain\n\ngo 1.27\n')
 shutil.copy(source/'grain_test.go', root/'grain_test.go')
 for name, mutated, test in mutants:
  assert mutated != original,name
  (root/'grain.go').write_text(mutated)
  proc=subprocess.run(['timeout','-s','KILL','90','go','test','-timeout=90s','-count=1','-run','^'+test+'$'],cwd=root,env={**os.environ,'GOWORK':'off'},stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
  (root/(test+'.log')).write_bytes(proc.stdout)
  print(name,'KILLED' if proc.returncode else 'SURVIVED','exit',proc.returncode,flush=True)
  if proc.returncode != 1 or ('--- FAIL: '+test).encode() not in proc.stdout:
   print(proc.stdout.decode(errors='replace'))
   raise SystemExit('mutant did not produce the expected test failure')
