exec(open('/tmp/u148-run.py').read().split('for row in rows:')[0])
records=json.loads((out/'runs.json').read_text())
for id,f in [('P1','stage1/cohere/values/values.ts'),('W1','stage1/cohere/values/values_test.go')]:
 s=(root/f).read_text()
 (out/(id+'.log')).rename(out/(id+'-initial.log'))
 try:
  subprocess.run(['git','apply',str(out/(id+'.diff'))],check=True)
  run(id,'timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/values/ -run .',{'ADAMIC_BUILD_CACHE_DIR':'/tmp/u148/cache/replay-'+id})
 finally:(root/f).write_text(s)
