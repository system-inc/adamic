import pathlib,subprocess,json
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-native-radix');a=p/'audit';a.mkdir(exist_ok=True);ref='origin/test-audit/internal-native-radix';base='review/test-audit/internal-native-radix/'
for f in ['REPORT.md','friction-and-limits.md','mutant-table.md','rows.json','summary.json','matrix.json','mutant-plan.json','scope.json']:(a/f).write_bytes(subprocess.check_output(['git','show',ref+':'+base+f]))
rows=json.loads((a/'rows.json').read_text());print([(r['test'],r['oracle']) for r in rows if r['test'] in ['TestRecordBenchmark','TestRuntimeReleasePaths','TestRuntimeStringEquality','TestRegExpIteratorResultShape']])
