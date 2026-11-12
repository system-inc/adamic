"""Recipe. Run each piece separately from /workspace/adamic, using the full directory path for run.py.
Single-root scheduling bounds scheduler processes as well as children; the ordinary
census command and duplicate agreement checks are unchanged from 09998fc3.

source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source
python3 stage3/census/next-stops-3007955b/run.py setup bash cloud/setup.sh
python3 stage3/census/next-stops-3007955b/run.py typescript git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/typescript-source
python3 stage3/census/next-stops-3007955b/run.py npm npm ci --prefix stage3/api
python3 stage3/census/next-stops-3007955b/run.py apply bash stage3/apply.sh /workspace/scratch/next-stops-adapted
python3 stage3/census/next-stops-3007955b/run.py build go build -o /workspace/scratch/next-stops-census ./stage3/census/tool
python3 stage3/census/next-stops-3007955b/run-one.py 1
python3 stage3/census/next-stops-3007955b/run-one.py 2
python3 stage3/census/next-stops-3007955b/run-one.py 3
python3 stage3/census/next-stops-3007955b/run-one.py 4
python3 stage3/census/next-stops-3007955b/run-one.py 5
python3 stage3/census/next-stops-3007955b/run-one.py 6
python3 stage3/census/next-stops-3007955b/run-one.py 7
python3 stage3/census/next-stops-3007955b/run-one.py 8
python3 stage3/census/next-stops-3007955b/run-one.py 9
python3 stage3/census/next-stops-3007955b/run-one.py 10
python3 stage3/census/next-stops-3007955b/run-one.py 11
python3 stage3/census/next-stops-3007955b/run-one.py 12
python3 stage3/census/next-stops-3007955b/run-one.py 13
python3 stage3/census/next-stops-3007955b/run-one.py 14
python3 stage3/census/next-stops-3007955b/run-one.py 15
python3 stage3/census/next-stops-3007955b/run-one.py 16
python3 stage3/census/next-stops-3007955b/run-one.py 17
python3 stage3/census/next-stops-3007955b/run-one.py 18
python3 stage3/census/next-stops-3007955b/run-one.py 19
python3 stage3/census/next-stops-3007955b/run-one.py 20
python3 stage3/census/next-stops-3007955b/run-one.py 21
python3 stage3/census/next-stops-3007955b/run-one.py 22
python3 stage3/census/next-stops-3007955b/run-one.py 23
python3 stage3/census/next-stops-3007955b/run-one.py 24
python3 stage3/census/next-stops-3007955b/run-one.py 25
python3 stage3/census/next-stops-3007955b/run-one.py 26
python3 stage3/census/next-stops-3007955b/run-one.py 27
python3 stage3/census/next-stops-3007955b/run-one.py 28
python3 stage3/census/next-stops-3007955b/run-one.py 29
python3 stage3/census/next-stops-3007955b/run-one.py 30
python3 stage3/census/next-stops-3007955b/run-one.py 31
python3 stage3/census/next-stops-3007955b/run-one.py 32
python3 stage3/census/next-stops-3007955b/run-one.py 33
python3 stage3/census/next-stops-3007955b/run-one.py 34
python3 stage3/census/next-stops-3007955b/run-one.py 35
python3 stage3/census/next-stops-3007955b/run-one.py 36
python3 stage3/census/next-stops-3007955b/run-one.py 37
python3 stage3/census/next-stops-3007955b/run-one.py 38
python3 stage3/census/next-stops-3007955b/run-one.py 39
python3 stage3/census/next-stops-3007955b/run-one.py 40
python3 stage3/census/next-stops-3007955b/run-one.py 41
python3 stage3/census/next-stops-3007955b/run-one.py 42
python3 stage3/census/next-stops-3007955b/run-one.py 43
python3 stage3/census/next-stops-3007955b/run-one.py 44
python3 stage3/census/next-stops-3007955b/run-one.py 45
python3 stage3/census/next-stops-3007955b/run-one.py 46
python3 stage3/census/next-stops-3007955b/run-one.py 47
python3 stage3/census/next-stops-3007955b/run-one.py 48
python3 stage3/census/next-stops-3007955b/run-one.py 49
python3 stage3/census/next-stops-3007955b/run-one.py 50
python3 stage3/census/next-stops-3007955b/run-one.py 51
python3 stage3/census/next-stops-3007955b/run-one.py 52
python3 stage3/census/next-stops-3007955b/run-one.py 53
python3 stage3/census/next-stops-3007955b/run-one.py 54
python3 stage3/census/next-stops-3007955b/run-one.py 55
python3 stage3/census/next-stops-3007955b/run-one.py 56
python3 stage3/census/next-stops-3007955b/run-one.py 57
python3 stage3/census/next-stops-3007955b/run-one.py 58
python3 stage3/census/next-stops-3007955b/run-one.py 59
python3 stage3/census/next-stops-3007955b/run-one.py 60
python3 stage3/census/next-stops-3007955b/run-one.py 61
python3 stage3/census/next-stops-3007955b/run-one.py 62
python3 stage3/census/next-stops-3007955b/run-one.py 63
python3 stage3/census/next-stops-3007955b/run-one.py 64
python3 stage3/census/next-stops-3007955b/run-one.py 65
python3 stage3/census/next-stops-3007955b/run-one.py 66
python3 stage3/census/next-stops-3007955b/run-one.py 67
python3 stage3/census/next-stops-3007955b/run-one.py 68
python3 stage3/census/next-stops-3007955b/run-one.py 69
python3 stage3/census/next-stops-3007955b/run-one.py 70
python3 stage3/census/next-stops-3007955b/run-one.py 71
python3 stage3/census/next-stops-3007955b/run-one.py 72
python3 stage3/census/next-stops-3007955b/run-one.py 73
python3 stage3/census/next-stops-3007955b/run-one.py 74
python3 stage3/census/next-stops-3007955b/run-one.py 75
python3 stage3/census/next-stops-3007955b/run-one.py 76
python3 stage3/census/next-stops-3007955b/run-one.py 77
python3 stage3/census/next-stops-3007955b/run-one.py 78
python3 stage3/census/next-stops-3007955b/run-one.py 79
python3 stage3/census/next-stops-3007955b/run.py assemble python3 stage3/census/next-stops-3007955b/assemble.py
python3 stage3/census/next-stops-3007955b/run.py analysis python3 stage3/census/next-stops-3007955b/analyze.py
python3 stage3/census/next-stops-3007955b/run.py source-excerpts python3 stage3/census/next-stops-3007955b/source-excerpts.py
python3 stage3/census/next-stops-3007955b/run.py write-report python3 stage3/census/next-stops-3007955b/write-report.py
python3 stage3/census/next-stops-3007955b/run.py validation python3 stage3/census/next-stops-3007955b/validate.py
"""
