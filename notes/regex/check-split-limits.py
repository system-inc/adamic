"""Observe V8's omitted/explicit limit paths without involving Adamic."""
import subprocess
from pathlib import Path

source = r"""const input="a🌍b";
for(const regex of [/\B/u,/\B/v,new RegExp("\\B(?<part>a?)","u"),new RegExp("(?!\\W)","v")]) {
 console.log(JSON.stringify(input.split(regex)));
 console.log(JSON.stringify(input.split(regex,4294967295)));
}
"""
result = subprocess.run(['node', '-e', source], capture_output=True, check=True)
Path('/tmp/regex-coverage/split-limit-control.stdout').write_bytes(result.stdout)
print(result.stdout.decode(), end='')
