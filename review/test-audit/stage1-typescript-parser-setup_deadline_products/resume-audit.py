exec(open('/tmp/u157.py').read().split('# Timings grouped')[0])
records=json.loads((p/'runs.json').read_text())
exec(open('/tmp/u157.py').read().split('# Witness for the count-only behavior, before mutations.')[1])
