"""Prepare a profiling-only overlay after make_overlay.py. No production writes."""
import json,pathlib,re,shutil,subprocess,sys
repo=pathlib.Path(sys.argv[1]).resolve()
scratch=pathlib.Path(sys.argv[2]).resolve()
here=pathlib.Path(__file__).resolve().parent
manifest=scratch/'overlay.json'
data=json.loads(manifest.read_text()); replace=data['Replace']; outputs=[]
if len(sys.argv)>3 and sys.argv[3]=='counts':
    for source in sorted((repo/'internal/lower').glob('*.go')):
        if source.name.endswith('_test.go'): continue
        key=str(source)
        if key not in replace:
            target=scratch/('profile_'+source.name); shutil.copyfile(source,target); replace[key]=str(target)
    for key,target in replace.items():
        if key.startswith(str(repo/'internal/lower')+'/') and not key.endswith('_test.go'): outputs.append(target)
    shutil.copyfile(here/'instrument.go.txt',scratch/'instrument.go')
    # Build separately so rewritten .go paths are program arguments, not go run inputs.
    subprocess.run(['go','build','-o',str(scratch/'instrument'),str(scratch/'instrument.go')],check=True)
    subprocess.run([str(scratch/'instrument'),*outputs],check=True)
# This probe does not run the production test suite. Keep unrelated tests out of
# its binary: make_overlay removes output-producing lowerChecked helpers that
# ordinary lower tests reference. This is not a defect in those main tests.
for source in sorted((repo/'internal/lower').glob('*_test.go')):
    target=scratch/('excluded_'+source.name)
    package=re.search(r'^package (\w+)',source.read_text(),re.M).group(1)
    target.write_text('package '+package+'\n')
    replace[str(source)]=str(target)
target=scratch/'profile_test.go';shutil.copyfile(here/'probe_test.go.txt',target)
replace[str(repo/'internal/lower/scan_profile_test.go')]=str(target)
manifest.write_text(json.dumps(data,indent=2)+'\n')
