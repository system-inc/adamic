from pathlib import Path
import re,sys
repo=Path('/workspace/path-components-adapt');source=repo/'stage3/census/latent/make_overlay.py';scratch=Path('/workspace/scratch/path-adapt-overlay')
s=source.read_text().replace("scratch / name.replace('/', '_')", "scratch / (name.replace('/', '_') + '.txt')")
s=re.sub(r"((?:scratch|output) / '[^']+\.go)'",r"\1.txt'",s)
sys.argv=[str(source),str(repo),str(scratch)]
exec(compile(s,str(source),'exec'),{'__file__':str(source),'__name__':'__main__'})
