"""Accept only the complete API text mechanically reconstructed by generate-api."""
import pathlib,sys,json,subprocess
root=pathlib.Path(sys.argv[1]);reconstruction=json.load(open(sys.argv[2]));expected=reconstruction['expected'].encode()
reference=root/'tests/baselines/reference/api/typescript.d.ts'
assert (root/'built/local/typescript.d.ts').read_bytes()==expected
assert reference.read_bytes() in [reconstruction['original'].encode(),expected]
allowed=['tests/baselines/reference/api/typescript.d.ts']
changed=lambda:subprocess.check_output(['git','-C',str(root),'diff','--name-only','--','tests/baselines/reference'],text=True).splitlines()
assert changed() in [[],allowed]
if reference.read_bytes()!=expected:
 reference.write_bytes(expected)
assert changed()==allowed
print(json.dumps({'generatedEqualsSanctionedExpected':True,'changedReferenceFiles':changed(),'acceptedOnlyProvenApi':True},indent=2))
