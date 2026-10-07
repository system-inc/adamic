#!/usr/bin/env python3
"""Require no introduced strict diagnostic code/message pairs after owner edits."""
import collections
import json
from pathlib import Path
import sys
before, after = [json.loads(Path(p).read_text()) for p in sys.argv[1:]]
options = ['strict', 'exactOptionalPropertyTypes', 'noUncheckedIndexedAccess',
           'verbatimModuleSyntax', 'erasableSyntaxOnly']
assert all(before['options'][key] is True and after['options'][key] is True for key in options)
def bag(report):
    return collections.Counter((d['code'], d['text']) for d in report['diagnostics'])
b, a = bag(before), bag(after)
assert not a - b, 'owner edits introduced strict diagnostics: ' + repr(a - b)
print(json.dumps({'before':sum(b.values()), 'after':sum(a.values()),
                  'introduced':0, 'removed':sum((b-a).values()),
                  'comparison':'diagnostic code/message multiset, independent of shifted locations',
                  'options':{key:True for key in options}}, indent=2))
