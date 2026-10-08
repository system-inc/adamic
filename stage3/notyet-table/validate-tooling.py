"""Run inherited census audits with witnesses valid on the measured base.

Truthy conditions are supported here. Use debugger and ==, whose actual lowering
returns NotYet, beside var's Refused. Keep continuation, rollback, checker-body
eligibility, ordered equality and exact-signature assertions. Production compiler
and the merged measurement implementation remain untouched.
"""
from pathlib import Path
import sys

repository = Path(__file__).resolve().parents[2]
mode = sys.argv.pop(1)
if mode == 'full':
    original = repository / 'stage3/census/latent/full_audit.py'
elif mode == 'replay':
    original = repository / 'stage3/census/latent/replay/replay_test.py'
else:
    raise SystemExit('usage: validate-tooling.py full|replay [original arguments]')
text = original.read_text()
assert "'a string as a condition', 'a number as a condition', 'var'" in text
assert 'if (text) { }' in text and 'if (count) { }' in text
text = text.replace('if (text) { }', 'debugger;').replace('if (count) { }', 'const loose = count == 1;')
if mode == 'full':
    text = text.replace("'a string as a condition', 'a number as a condition', 'var'", "'a DebuggerStatement', 'a BinaryExpression with a number and a number', 'var'")
    text = text.replace("finding['kind'] == 'Refused'", "finding['kind'] in ('NotYet', 'Refused')")
    # This base refuses a dependency on a diagnosed grandchild at signature time.
    # Keep the diagnosed declaration independently catalogued, as a sibling.
    text = text.replace('        function grandBad(): number { return "wrong"; }\n', '')
    text = text.replace('export function outerOnly(): void {\n', 'export function outerOnly(): void {\n    function grandBad(): number { return "wrong"; }\n')
    text = text.replace('three distinct lowering refusals', 'three distinct lowering failures')
else:
    text = text.replace("'a string as a condition', 'a number as a condition', 'var'", "'var'")
    text = text.replace("assert not any(finding['kind'] == 'NotYet' for finding in cls.expected), cls.expected", "assert {finding['reason'] for finding in cls.expected if finding['kind'] == 'NotYet'} == {'a DebuggerStatement', 'a BinaryExpression with a number and a number'}, cls.expected")
sys.argv[0] = str(original)
__file__ = str(original)
exec(compile(text, str(original), 'exec'), globals())
