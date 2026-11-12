#!/usr/bin/env python3
"""Partition independent file scans by intersecting unit ranges; omit body attempts."""
import json,pathlib,sys
src,out=map(pathlib.Path,sys.argv[1:]);out.mkdir(parents=True,exist_ok=True);m=json.loads((src/'overlay.json').read_text());r=m['Replace']
k=next(k for k in r if k.endswith('/internal/lower/latent_hook.go'));s=pathlib.Path(r[k]).read_text();a='fresh().latentScan(file)';assert s.count(a)==1;s=s.replace(a,'if lwpScanSelected(len(file.Statements.Nodes)) {fresh().latentScan(file)}');a='if !lwpSelected() { continue }';assert s.count(a)==1;s=s.replace(a,a+'; if os.Getenv("LWP_SCAN_ONLY")=="1" {continue}');p=out/'hook.go';p.write_text(s);r[k]=str(p.resolve())
k=next(k for k in r if k.endswith('/internal/lower/latent_whole_passes.go'));s=pathlib.Path(r[k]).read_text();s+='\nfunc lwpScanSelected(n int) bool {start,_:=strconv.Atoi(os.Getenv("LWP_START"));end,_:=strconv.Atoi(os.Getenv("LWP_END"));return lwpOrdinal+n>start&&(end==0||lwpOrdinal<end)}\n';p=out/'passes.go';p.write_text(s);r[k]=str(p.resolve());(out/'overlay.json').write_text(json.dumps(m,indent=2)+'\n')
