import json
import re
from pathlib import Path

root = Path(__file__).resolve().parents[3]
review = Path(__file__).resolve().parent
text = (root / 'internal/lower/prototype_test.go').read_text()
members = json.loads((review / 'node-members.json').read_text())
declared = {'constructor', 'toString', 'toLocaleString', 'valueOf', 'hasOwnProperty', 'propertyIsEnumerable', 'isPrototypeOf'}
rows = []

def add(test, row, why):
    rows.append((test, row, 'refuse', 'Preserved unchanged', why))

parts = dict(re.findall(r'func (Test\w+)\(t \*testing.T\) \{(.*?)(?=\nfunc |\Z)', text, re.S))
test = 'TestInheritedLibraryReadsNeverLoadOwnFields'
receivers = re.findall(r'\{"(\w+)", ("(?:[^"\\]|\\.)*")\}', parts[test])
for receiver, source in receivers:
    for member in members:
        for bracket in (False, True):
            access = f"value['{member}']" if bracket else f'value.{member}'
            reason = 'Requires Refused and no expression from the direct lowering entry point.' if member in declared else 'Legacy member absent from es2024: requires checker rejection, otherwise lowering Refused.'
            add(test, f'{receiver}/{member}/bracket={str(bracket).lower()}: ' + json.loads(source) + ' ' + access + ';', reason)

test = 'TestNullishPrototypeReadsAreRejectedByChecker'
for value in ('null', 'undefined'):
    for member in members:
        add(test, f'const value = {value}; value.{member};', 'Requires load.CheckError; no acceptance assertion.')

for test in ('TestPrototypeMethodsAreRefusedWithReasons', 'TestPrototypeHazardsBehindObjectViewsAreNotYet', 'TestUnrepresentedPrototypeCallsAreNotYet', 'TestIsPrototypeOfReadsExplainThePrototypeRefusal'):
    body = parts[test]
    array = re.search(r'\[\]string\{(.*?)\n\t\} \{', body, re.S).group(1)
    for index, literal in enumerate(re.findall(r'"(?:[^"\\]|\\.)*"', array), 1):
        source = json.loads(literal)
        reason = 'Requires NotYet (including a declare-field-specific diagnostic).' if test == 'TestPrototypeHazardsBehindObjectViewsAreNotYet' else 'Requires NotYet.' if test == 'TestUnrepresentedPrototypeCallsAreNotYet' else 'Requires Refused with the specified explanation.'
        add(test, f'{index}: {source}', reason)
    if test == 'TestPrototypeMethodsAreRefusedWithReasons':
        for member in ('toString', 'toLocaleString', 'valueOf', 'hasOwnProperty', 'propertyIsEnumerable', 'isPrototypeOf', 'constructor'):
            add(test, 'destructure/' + member + ': const value = {}; const { ' + member + ': detached } = value;', 'Requires Refused for a destructured inherited member.')

out = ['# Prototype source-row classification', '', 'Task #41bkfdw, wave 2. Base: 62286991 (current origin/main at branch creation).', '', f'{len(rows)} source rows: 0 accept, {len(rows)} refuse, 0 IR-only. Each generated source is a separate row, including both sweep access forms. Node membership comes from node-members.json.', '', 'Classification follows the final assertion, not the setup calls that successfully lower receiver declarations. The direct-entry sweep requires Refused and a nil expression; that is a refusal contract, not an ownership or representation assertion.', '', 'No test or source row was changed or deleted. No IR golden snapshots or extra acceptance IR assertions exist in this file. No new fixture requires counts.md regeneration.', '', 'The requested converted-row mutation proofs are inapplicable: there are no acceptance rows to convert. M20 belongs to escapeRegexSource in regexp.go; TestRegExpSourceNode is in regexp_test.go, outside this unit. We preserve the refusal scope rather than introduce an unrelated acceptance row. See REPORT.md for the mutation runs and their limits.', '', '| test | row | class | what you did | why |', '|---|---|---|---|---|']
for row in rows:
    out.append('| ' + ' | '.join(value.replace('|', '&#124;').replace('\n', '<br>') for value in row) + ' |')
out.extend(['', '## Authorized regex acceptance additions', '', 'Three new acceptance rows live beside the existing regex tests in internal/lower/regexp_test.go. Each prints source, flags and a slash-match result, runs lowersAndAgreesWithNode, and compares sanitized native output with source Node. The prototype census above stays unchanged.', '', '| test | row | class | what you did | why |', '|---|---|---|---|---|'])
regex_text = (root / 'internal/lower/regexp_test.go').read_text()
for test in ('TestRegExpUnicodeClassSourceAgreesWithNode', 'TestRegExpSlashClassSourceAgreesWithNode', 'TestRegExpEscapedSlashSourceAgreesWithNode'):
    literal = re.search(r'func ' + test + r'\(t \*testing.T\) \{.*?source := ("(?:[^"\\]|\\.)*")', regex_text, re.S).group(1)
    source = json.loads(literal).replace('|', '&#124;')
    out.append('| ' + test + ' | ' + source + ' | accept | Added Node agreement and sanitized native behavior comparison | JavaScript recomputes source metadata; native output exposes M20 and overescaping. |')
(review / 'rows.md').write_text('\n'.join(out) + '\n')
print(f'{len(rows)} refuse rows; receivers={len(receivers)}, Node members={len(members)}')
