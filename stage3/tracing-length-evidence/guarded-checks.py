from run import run,raw
run('after-build-wrapper',['python3',raw/'build-side.py','after'])
run('after-checks-wrapper',['python3',raw/'check-side.py','after'])
run('mutants-wrapper',['python3',raw/'mutants.py'])
run('census-summary',['python3',raw/'census-summary.py'])
