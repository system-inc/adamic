"""Regenerate pinned Go metadata and the helper corpus without changing the cohere submodule."""
import json,pathlib,subprocess,tempfile
here=pathlib.Path(__file__).resolve().parent
root=here.parents[4]
cohere=root/'cohere'
with tempfile.TemporaryDirectory(prefix='adamic-helper-metadata-') as directory:
 directory=pathlib.Path(directory)
 virtual=cohere/'adamic_helper_oracle.go'
 descriptors=cohere/'adamic_helper_descriptors.go'
 overlay={'Replace':{str(virtual):str(here/'oracle.go'),str(descriptors):str(here/'descriptors.go'),str(cohere/'policy/adamic_helpers.go'):str(here/'catalog.go')}}
 path=directory/'overlay.json';path.write_text(json.dumps(overlay))
 with (directory/'build.log').open('wb') as log:
  subprocess.run(['go','build','-overlay='+str(path),'-o',str(directory/'oracle'),str(virtual),str(descriptors)],cwd=cohere,stdout=log,stderr=subprocess.STDOUT,check=True)
 for command,name in [('catalog','catalog.json'),('descriptors','descriptors.json')]:
  with (here/name).open('wb') as output, (directory/(command+'.log')).open('wb') as log:
   subprocess.run([str(directory/'oracle'),command],stdout=output,stderr=log,check=True)
subprocess.run(['python3',str(here/'generate.py')],check=True)
