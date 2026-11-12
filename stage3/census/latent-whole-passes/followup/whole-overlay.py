import json,pathlib,sys
repo,out=map(pathlib.Path,sys.argv[1:]);out.mkdir(parents=True,exist_ok=True);r={}
def put(target,text):
 p=out/target.replace('/','_');p.write_text(text);r[str(repo/target)]=str(p.resolve())
s=(repo/'internal/lower/view_widened.go').read_text();s=s.replace('import (','import (\n "os"',1);needle='\t\tif demanded {\n\t\t\tthrough := "an unnamed type"';assert s.count(needle)==1;s=s.replace(needle,'\t\tif demanded {\n if os.Getenv("LWP_TRACE")!="" {l.lwp2Trace(read,reaches,inside,lies,unentered)}\n\t\t\tthrough := "an unnamed type"');s=s.replace('func (l *lowering) checkViewWidenedReads() error {','func (l *lowering) checkViewWidenedReads() error {\n if os.Getenv("LWP_TRACE")!="" {lwp2Write(map[string]any{"event":"production_widened_reached"})}')
put('internal/lower/view_widened.go',s)
s=(repo/'internal/lower/lower.go').read_text();needle='\tlowering.checkOverloadResults(modules)';assert s.count(needle)==1;s=s.replace(needle,'\tif os.Getenv("LWP_TRACE")!="" {lwp2Write(map[string]any{"event":"production_tail_enter"})}\n'+needle);put('internal/lower/lower.go',s)
put('internal/lower/lwp2_trace.go',pathlib.Path(__file__).with_name('trace.go.txt').read_text()+'\nvar latentFindingOwner = "whole-program"\n')
put('stage3/census/tool/main.go',pathlib.Path(__file__).with_name('whole-main.go.txt').read_text())
(out/'overlay.json').write_text(json.dumps({'Replace':r},indent=2)+'\n')
