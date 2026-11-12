import re,sys
from pathlib import Path
repo=Path('/workspace/adamic');source=repo/'stage3/census/latent/make_overlay.py'
text=source.read_text().replace("scratch / name.replace('/', '_')", "scratch / (name.replace('/', '_') + '.txt')")
text=re.sub(r"((?:scratch|output) / '[^']+\.go)'",r"\1.txt'",text)
sys.argv=[str(source),str(repo),'/workspace/scratch/path-components-proof/latent-overlay']
exec(compile(text,str(source),'exec'),{'__file__':str(source),'__name__':'__main__'})
assert not list(Path(sys.argv[2]).rglob('*.go')),'all physical Go evidence must be .go.txt'
