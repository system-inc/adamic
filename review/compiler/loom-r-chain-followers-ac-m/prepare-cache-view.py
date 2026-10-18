"""Reuse immutable local cache objects; all new atomic writes stay on /tmp."""
import json,os
from pathlib import Path
old=Path('/home/agent/.cache/adamic-gocache');new=Path('/tmp/adamic-gate/loom-chain-go-cache');linked=0
for area in ['blobs','refs']:
 for directory,dirs,files in os.walk(old/area):
  target=new/Path(directory).relative_to(old);target.mkdir(parents=True,exist_ok=True)
  for name in files:
   if name.startswith('.tmp-'):continue
   destination=target/name
   try:destination.symlink_to(Path(directory)/name);linked+=1
   except FileExistsError:pass
Path('review/compiler/loom-r-chain-followers-ac-m/cache-view.json').write_text(json.dumps({'old_cache':str(old),'new_cache':str(new),'linked_existing_files':linked,'old_files_modified_or_removed':False,'new_writes':'Go cache atomicFile creates and renames only in new_cache; existing cache bodies are read via symlinks'},indent=2))
print('Added read-through links:',linked)
