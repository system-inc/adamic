from pathlib import Path
import json
import sys
rows = json.loads(Path(sys.argv[1] if len(sys.argv) > 1 else 'review/compiler/eep-presence/results.json').read_text())
for row in rows:
    for mode in ['native', 'sanitized', 'js']:
        result = row[mode + '_compile']
        assert result['exit'] == 1, (row['name'], mode, result['exit'])
        assert result['stdout'] == '', (row['name'], mode)
        assert row['name'] + '.a:' in result['stderr'], (row['name'], mode)
        assert "stage 0 can't lower" in result['stderr'] and '; ' in result['stderr'], (row['name'], mode)
        assert 'compiler bug:' not in result['stderr'], (row['name'], mode)
print('All seven witnesses refused with located workarounds in native, sanitized native and JavaScript')
