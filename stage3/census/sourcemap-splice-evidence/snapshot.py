from pathlib import Path
import hashlib,json,sys
root=Path(sys.argv[1]);out=Path(sys.argv[2]);rows={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((root/'src').rglob('*')) if p.is_file()}
out.write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps({'files':len(rows),'manifest':str(out)}))
