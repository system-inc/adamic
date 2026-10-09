from pathlib import Path
import subprocess
old=subprocess.check_output(['git','show','a3f28d97:internal/oracle/counts.md'],text=True).splitlines(True)
new=Path('internal/oracle/counts.md').read_text().splitlines(True)
added=[line for line in new if line.startswith('| Program region: ')]
assert len(added)==11
assert [line for line in new if not line.startswith('| Program region: ')]==old
print('PASS: all existing count text byte identical; eleven explicit Program region rows added')
