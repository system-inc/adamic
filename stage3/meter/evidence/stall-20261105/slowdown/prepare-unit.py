from pathlib import Path
import json,sys
r=Path('/tmp/meter-stall-20261105');dest=Path(sys.argv[1]);base=Path(sys.argv[2]);dest.mkdir(parents=True,exist_ok=True)
m=json.loads((base/'overlay.json').read_text())
def change(suffix,fn):
 k=next(k for k in m['Replace'] if k.endswith('/'+suffix));p=dest/suffix.replace('/','_');p.write_text(fn(Path(m['Replace'][k]).read_text()));m['Replace'][k]=str(p)
profile='''
func init() {
 if path := os.Getenv("METER_CPU_PROFILE"); path != "" {
  go func() {
   time.Sleep(600*time.Second)
   f,err:=os.Create(path);if err!=nil {panic(err)}
   if err:=pprof.StartCPUProfile(f);err!=nil {panic(err)}
   time.Sleep(60*time.Second)
   pprof.StopCPUProfile();if err:=f.Close();err!=nil {panic(err)}
  }()
 }
}
'''
change('stage3/census/latent/tool/main.go',lambda s:s.replace('import (','import (\n "runtime/pprof"\n "time"',1)+profile)
def full(s):
 needle='func (l *lowering) latentSnapshot() *lowering {'
 assert needle in s
 s=s.replace(needle,'func (l *lowering) latentSnapshotRaw() *lowering {',1)
 s=s.replace('import (','import (\n "time"',1)
 return s+'''
var meterSnapshotWall, meterSnapshotCPU float64
func (l *lowering) latentSnapshot() *lowering {
 started,cpu:=time.Now(),processCPUSeconds()
 defer func(){meterSnapshotWall+=time.Since(started).Seconds();meterSnapshotCPU+=processCPUSeconds()-cpu}()
 return l.latentSnapshotRaw()
}
'''
change('internal/lower/latent_full.go',full)
def timing(s):
 s=s.replace('import (','import (\n "os"\n "runtime/pprof"',1)
 s=s.replace('wall, cpu := time.Now(), processCPUSeconds()', 'var meterProfile *os.File\n if path:=os.Getenv("METER_UNIT_CPU_PROFILE");path!="" {\n  var err error\n  meterProfile,err=os.Create(path);if err!=nil{panic(err)}\n  if err=pprof.StartCPUProfile(meterProfile);err!=nil{panic(err)}\n }\n wall, cpu := time.Now(), processCPUSeconds()',1)
 s=s.replace('if err := timer.encoder.Encode(record); err != nil {', 'if meterProfile!=nil { defer func(){pprof.StopCPUProfile();if err:=meterProfile.Close();err!=nil{panic(err)}}() }\n  if err := timer.encoder.Encode(record); err != nil {',1)
 s=s.replace('type UnitTime struct {','type UnitTime struct {\n SnapshotWallSeconds float64 `json:"snapshot_wall_seconds"`\n SnapshotCPUSeconds float64 `json:"snapshot_cpu_seconds"`',1)
 s=s.replace('wall, cpu := time.Now(), processCPUSeconds()','wall, cpu := time.Now(), processCPUSeconds()\n snapshotWall,snapshotCPU:=meterSnapshotWall,meterSnapshotCPU',1)
 s=s.replace('record := UnitTime{file, where, time.Since(wall).Seconds(), processCPUSeconds() - cpu}','record := UnitTime{meterSnapshotWall-snapshotWall,meterSnapshotCPU-snapshotCPU,file, where, time.Since(wall).Seconds(), processCPUSeconds() - cpu}',1)
 return s
change('internal/lower/latent_timing.go',timing)
# Add only a timing wrapper around the existing replay selection and lowering path.
k=next(k for k in m['Replace'] if k.endswith('/internal/lower/latent_replay.go'))
s=Path(m['Replace'][k]).read_text().replace('import (','import (\n "io"',1)
s+='''
func MeterTimedReplay(ctx context.Context,p *load.Program,where string,out io.Writer)(LatentFile,error){
 latentUnitTimer=NewUnitTimer(out)
 defer func(){latentUnitTimer=nil}()
 return LatentReplay(ctx,p,where,"NotYet","slowdown-measurement-no-signature")
}
'''
p=dest/'replay.go';p.write_text(s);m['Replace'][k]=str(p)
(dest/'overlay.json').write_text(json.dumps(m,indent=2)+'\n')
