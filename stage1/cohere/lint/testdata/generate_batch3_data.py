"""Copy policy vocabulary and acronym data, never computed lint verdicts."""
import json
import re
from pathlib import Path
root = Path(__file__).resolve().parents[4]
policy = json.loads((root / 'cohere/policy/Abbreviations.json').read_text())
rows = []
for entry in policy['abbreviations']:
    whole, prefix, suffix = entry.get('whole', {}), entry.get('prefix', {}), entry.get('suffix', {})
    rows.append([entry['abbreviation'], entry.get('expansion', ''), entry.get('advice', ''), whole.get('style', ''), prefix.get('phase', ''), prefix.get('style', ''), 'suffix' in entry, suffix.get('matcher', ''), suffix.get('replacement', ''), suffix.get('advice', ''), 'segment' in entry])
output = '// Generated policy data. Regenerate with testdata/generate_batch3_data.py.\n'
output += 'export const vocabulary: string[][] = [\n'
for row in rows:
    output += '    ' + json.dumps([str(value).lower() if isinstance(value, bool) else value for value in row], ensure_ascii=False) + ',\n'
output += '];\n'
for key in ['allowedNames', 'allowedSegments']:
    values = [entry['name' if key == 'allowedNames' else 'segment'] for entry in policy[key]]
    output += f'export const {key}: string[] = {json.dumps(values)};\n'
source = (root / 'cohere/internal/lint/rules/nexus/shouting.go').read_text()
for key in ['allowedUppercaseTokens', 'currencyCodes', 'shoutedTwoLetterWords']:
    body = re.search(r'var ' + key + r' = map\[string\]bool\{(.*?)\n\}', source, re.S).group(1)
    values = re.findall(r'"([^"\n]+)":\s*true', body)
    output += f'export const {key}: string[] = {json.dumps(values)};\n'
(root / 'stage1/cohere/lint/rules/policy_data.ts').write_text(output)
