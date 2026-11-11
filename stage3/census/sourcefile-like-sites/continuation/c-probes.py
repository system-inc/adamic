import pathlib,sys
sys.dont_write_bytecode=True
from variants import invoke
p=pathlib.Path(__file__).resolve().parent
invoke(['/workspace/scratch/baobab-first-stop',str(p/'c/probes/required-optional-writer.ts'),str(p/'c/probes/required-optional-writer.first-stop.jsonl')],p/'c/probes/required-optional-writer.log',120)
