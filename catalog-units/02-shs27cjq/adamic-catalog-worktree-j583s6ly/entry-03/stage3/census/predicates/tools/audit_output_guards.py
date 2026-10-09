"""Build two scratch guard mutants; rejected-program measurement must never expose output."""
import json, os, pathlib, shutil, subprocess, sys
repository=pathlib.Path(sys.argv[1]).resolve(); baseline=pathlib.Path(sys.argv[2]).resolve(); scratch=pathlib.Path(sys.argv[3]).resolve();scratch.mkdir(parents=True,exist_ok=True)
for name,filename,old,new,expected in [
 ('IR','internal_lower_lower.go','return nil, fmt.Errorf("latent census: measurement only; no IR output")','return &ir.Program{}, fmt.Errorf("latent census: measurement only; no IR output")','measurement returned usable IR'),
 ('loader','internal_load_load.go','return nil, fmt.Errorf("latent census: measurement loader only; no output path")','return LatentLoad(paths)','measurement loader exposed an output program')]:
 folder=scratch/name;shutil.copytree(baseline,folder,dirs_exist_ok=True)
 path=folder/filename;s=path.read_text();assert old in s;s=s.replace(old,new,1);path.write_text(s)
 overlay=json.loads((baseline/'overlay.json').read_text());overlay['Replace']={k:str(folder/pathlib.Path(v).name) for k,v in overlay['Replace'].items()}
 (folder/'overlay.json').write_text(json.dumps(overlay,indent=2)+'\n')
 binary=folder/'census'
 with (folder/'build.log').open('w') as log:
  subprocess.run(['go','build','-buildvcs=false','-overlay='+str(folder/'overlay.json'),'-o',str(binary),'./stage3/census/latent/tool'],cwd=repository,stdout=log,stderr=subprocess.STDOUT,check=True)
 with (folder/'audit.log').open('w') as log:
  result=subprocess.run(['python3',str(pathlib.Path(__file__).parent/'audit.py'),str(binary)],stdout=log,stderr=subprocess.STDOUT)
 # audit.py's inner run log is temporary; run a small diagnosed input to preserve exact guard panic evidence.
 source=folder/'source';source.mkdir(exist_ok=True);(source/'bad.a').write_text('function wrong(): number { return "wrong"; }\n')
 with (folder/'guard.log').open('w') as log:
  guard=subprocess.run([str(binary),str(source),str(folder/'mutant.jsonl')],env=dict(os.environ,LATENT_ASSERT_NO_OUTPUT='1'),stdout=log,stderr=subprocess.STDOUT)
 assert result.returncode!=0 and guard.returncode!=0 and expected in (folder/'guard.log').read_text(), name
 print(name+' output guard mutant caught: '+expected,flush=True)
