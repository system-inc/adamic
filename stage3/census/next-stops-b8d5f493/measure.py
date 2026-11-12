"""Recipe entry point. Invoke each piece separately to keep processes bounded.
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source
python3 run.py apply bash stage3/apply.sh /workspace/scratch/next-stops-adapted
python3 build-binaries.py
python3 run.py groups 1
python3 run.py groups 2
python3 run.py groups 3
python3 run.py groups 4
python3 run.py analysis python3 stage3/census/next-stops-b8d5f493/analyze.py
python3 entry-run.py
python3 single-files.py
python3 finalize.py
python3 package-evidence.py
See status JSON files for exact executed argv and child exit codes.
"""
