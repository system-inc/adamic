import json,pathlib,subprocess
p=pathlib.Path('review/test-audit/internal-native-radix'); names=[n for _,ms in json.loads((p/'scope.json').read_text()) for n in ms];cmd=['timeout','120','go','test','-count=1','-timeout','90s','-coverprofile',str(p/'go.cover'), './internal/native/','-run','^('+'|'.join(names)+')$']
with (p/'coverage.log').open('w') as log:subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
with (p/'go-function-coverage.txt').open('w') as log:subprocess.run(['go','tool','cover','-func',str(p/'go.cover')],stdout=log,stderr=subprocess.STDOUT)
