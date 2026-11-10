from pathlib import Path
import shutil,sys
source=Path(sys.argv[1]);output=Path(sys.argv[2]);output.mkdir()
for p in source.iterdir():
 if p.name not in ['built','node_modules']: (output/p.name).symlink_to(p,target_is_directory=p.is_dir())
(output/'built/local').mkdir(parents=True)
for p in (source/'built').iterdir():
 if p.name!='local':(output/'built'/p.name).symlink_to(p,target_is_directory=p.is_dir())
for p in (source/'built/local').iterdir():
 if p.is_file() and p.suffix=='.js':shutil.copyfile(p,output/'built/local'/p.name)
 else:(output/'built/local'/p.name).symlink_to(p,target_is_directory=p.is_dir())
