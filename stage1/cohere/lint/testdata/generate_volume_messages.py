"""Generate pinned policy text and static Go rule messages, never rule verdicts."""
from pathlib import Path
import json
import re
import sys

root = Path(__file__).resolve().parents[4]
selected = json.loads(Path(__file__).with_name('volume_rules.json').read_text())
rules = {row[0] for row in selected}
entries = []
for path in sorted((root / 'cohere/policy/messages').glob('*.json')):
    data = json.loads(path.read_text())
    rule = data.get('rules', {}).get('TypeScript')
    if rule not in rules:
        continue
    for message_id, message in data['messages'].items():
        text = message['text']
        for key, terms in message.get('terms', {}).items():
            text = text.replace('[[' + key + ']]', terms['TypeScript'])
        for key, phrase in data.get('phrases', {}).items():
            if isinstance(phrase, str):
                text = text.replace('<<' + key + '>>', phrase)
        entries.append((rule + '/' + message_id, text))
static = []
for _, source, _, _, _ in selected:
    data = (root / 'cohere/internal/lint/rules' / (source + '.go')).read_text()
    # Every static description in this slice is a sequence of quoted Go literals.
    # Dynamic fmt/policy rendering stays in the rule implementation.
    for match in re.finditer(r'(?:var|func)\s+(\w+)\s*(?:\(\) rule.Message \{\s*return )?=?\s*rule.Message\{(.*?)\n\}', data, re.S):
        body = match[2]
        description = re.search(r'Description:\s*((?:"(?:[^"\\]|\\.)*"\s*(?:\+\s*)?)+),', body)
        if description:
            strings = re.findall(r'"(?:[^"\\]|\\.)*"', description[1])
            static.append((match[1], ''.join(json.loads(s) for s in strings)))
text = "// Generated from pinned Cohere policy and Go rule descriptions.\nimport { panic } from 'adamic';\n\nexport function policyMessage(rule: string, id: string, fields: readonly string[]): string {\n    let text: string;\n    switch(`${rule}/${id}`) {\n"
for key, value in entries:
    text += '        case ' + json.dumps(key) + ': text = ' + json.dumps(value) + '; break;\n'
text += "        default: return panic(`unknown policy message ${rule}/${id}`);\n    }\n    for(let index = 0; index < fields.length; index += 2) {\n        const name = fields[index] ?? panic('field name');\n        const value = fields[index + 1] ?? panic('field value');\n        text = text.split(`{{${name}}}`).join(value).split(`<<${name}>>`).join(value);\n    }\n    return text;\n}\n\nexport function cohereMessage(name: string): string {\n    switch(name) {\n"
for key, value in static:
    text += '        case ' + json.dumps(key) + ': return ' + json.dumps(value) + ';\n'
text += "        default: return panic(`unknown Go message ${name}`);\n    }\n}\n"
Path(sys.argv[1] if len(sys.argv) > 1 else root / 'stage1/cohere/lint/volume_messages.ts').write_text(text)
