from pathlib import Path
import re,sys
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/core-cast');source=repo/'stage3/census/latent/make_overlay.py';text=source.read_text();text=text.replace("scratch / name.replace('/', '_')", "scratch / (name.replace('/', '_') + '.txt')");text=re.sub(r"((?:scratch|output) / '[^']+\.go)'",r"\1.txt'",text);sys.argv=[str(source),str(repo),str(s/'overlay')];exec(compile(text,str(source),'exec'),{'__file__':str(source),'__name__':'__main__'})
