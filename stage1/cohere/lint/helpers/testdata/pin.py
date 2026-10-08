"""Capture provenance comes from the repository gitlink, never a literal revision."""
import subprocess

def capture_pin(root):
    fields = subprocess.check_output(['git', '-C', str(root), 'ls-tree', 'HEAD', 'cohere'], text=True).split()
    assert len(fields) == 4 and fields[:2] == ['160000', 'commit'] and fields[3] == 'cohere', 'invalid cohere gitlink'
    pin = fields[2]
    actual = subprocess.check_output(['git', '-C', str(root / 'cohere'), 'rev-parse', 'HEAD'], text=True).strip()
    assert actual == pin, 'Go cohere checkout differs from gitlink: ' + actual + ' != ' + pin
    return pin
