"""One-file full census, bounded at 25 minutes, with an actual child exit."""
import os,sys
from pathlib import Path
from run import run,scratch
name=sys.argv[1];assert name in ['head','parent']
os.environ['CENSUS_MODULE_FORMAT']='CommonJSBundle'
os.environ['LATENT_FULL']='1'
os.environ['LATENT_ONLY_FILE']=str(scratch/'next-stops-516c3ded-adapted/src/compiler/sys.ts')
os.environ['LATENT_ASSERT_NO_OUTPUT']='1'
run('sys-'+name+'-full',[str(scratch/('next-stops-516c3ded-latent-'+name+'-census')),os.environ['LATENT_ONLY_FILE'],str(Path(__file__).resolve().parent/('sys-'+name+'-full.jsonl'))],limit=1500)
