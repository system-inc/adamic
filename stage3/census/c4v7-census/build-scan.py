"""Build the unchanged scan-only wrapper once per measured compiler commit."""
import hashlib,json,subprocess,sys
from pathlib import Path
BASE=Path('/workspace/scratch/c4v7-census');OUT=Path(__file__).resolve().parent
label=sys.argv[1];assert label in ('main','head');repo=BASE/label;overlay=BASE/f'overlay-{label}'
subprocess.run([sys.executable,str(repo/'stage3/census/latent/make_overlay.py'),str(repo),str(overlay)],cwd=repo,check=True)
mapfile=overlay/'overlay.json';m=json.loads(mapfile.read_text());replace=m['Replace']
wrapper=overlay/'internal_lower_latent_scan_only.go';wrapper.write_bytes((OUT/'scan-only.go.txt').read_bytes())
replace[str(repo/'internal/lower/latent_scan_only.go')]=str(wrapper)
main=Path(replace[str(repo/'stage3/census/latent/tool/main.go')]);s=main.read_text();old='lower.Latent(context.Background(), program,';assert s.count(old)==1
main.write_text(s.replace(old,'lower.LatentScanOnly(context.Background(), program,'))
mapfile.write_text(json.dumps(m,indent=2)+'\n')
subprocess.run(['gofmt','-w',*map(str,overlay.glob('*.go'))],check=True)
subprocess.run(['go','build','-buildvcs=false',f'-overlay={mapfile}','-o',str(BASE/f'scan-{label}'),'./stage3/census/latent/tool'],cwd=repo,check=True)
(OUT/f'build-{label}.json').write_text(json.dumps({'label':label,'binary_sha256':hashlib.sha256((BASE/f'scan-{label}').read_bytes()).hexdigest(),'wrapper_sha256':hashlib.sha256(wrapper.read_bytes()).hexdigest(),'build_once':True},indent=2)+'\n')
