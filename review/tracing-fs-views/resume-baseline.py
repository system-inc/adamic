import importlib.util,pathlib,os
root=pathlib.Path('/workspace/tracing-evidence')
spec=importlib.util.spec_from_file_location('baseline_apply',root/'stage3/apply.py')
module=importlib.util.module_from_spec(spec)
import sys
sys.path.insert(0,str(root/'stage3'));spec.loader.exec_module(module)
dirs=sorted(p for p in (root/'stage3/adapt').iterdir() if p.is_dir() and p.name >= '99-correctness-axolotl-e')
env=dict(os.environ,NODE_PATH='/workspace/adamic/stage3/api/node_modules')
module.adapt(pathlib.Path('/workspace/tracing-before'),dirs,env)
