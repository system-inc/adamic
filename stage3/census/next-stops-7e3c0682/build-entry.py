from pathlib import Path
import re
import sys
repo = Path('/workspace/adamic')
scratch = Path('/workspace/scratch/next-stops')
# Preserve the meter overlay's bytes and virtual Go paths. Only physical
# scratch filenames gain .txt, as required by this census task.
for relative, arguments in [
    ('stage3/census/latent/make_overlay.py', [str(repo), str(scratch / 'latent-overlay')]),
    ('stage3/meter/entry_overlay.py', [str(scratch / 'latent-overlay'), str(scratch / 'entry-overlay')]),
]:
    source = repo / relative
    text = source.read_text()
    text = text.replace("scratch / name.replace('/', '_')", "scratch / (name.replace('/', '_') + '.txt')")
    text = re.sub(r"((?:scratch|output) / '[^']+\.go)'", r"\1.txt'", text)
    sys.argv = [str(source), *arguments]
    exec(compile(text, str(source), 'exec'), {'__file__': str(source), '__name__': '__main__'})
