"""Use the standard census overlay with physical evidence filenames ending .go.txt."""
import pathlib,json,subprocess,sys
repo=pathlib.Path('/workspace/adamic');scratch=pathlib.Path('/workspace/scratch/sys-timer-slot/overlay')
# make_overlay.py already generated the unmodified virtual source replacements.
# Only physical filenames change, exactly as the cited evidence branch does.
p=scratch/'overlay.json';data=json.loads(p.read_text())
for virtual,physical in list(data['Replace'].items()):
    old=pathlib.Path(physical)
    if old.suffix=='.go':
        new=old.with_name(old.name+'.txt');old.rename(new);data['Replace'][virtual]=str(new)
p.write_text(json.dumps(data,indent=2)+'\n')
print(str(p),flush=True)
