"""Join exact stock-tsc evidence to each initial refusal's named first blocker."""
import collections,json,re
from pathlib import Path
root=Path(__file__).resolve().parent
source=json.loads((root/'exact-tsc.json').read_text())
expected=json.loads((root/'classification-reasons.json').read_text())
rows=[]
for row in source['results']:
 reason=row['reason']
 direct=re.search(r'(?:refuses |not yet: )(Object\.\w+)',reason)
 builtin=re.search(r'not yet: (?:reading )?(Object|Array|String|Number|Boolean|Date|RegExp|Error|EvalError|RangeError|ReferenceError|SyntaxError|TypeError|URIError|Math|JSON|Function)(?: as a value|$|\.prototype)',reason)
 feature=direct[1] if direct else builtin[1]+' builtin value' if builtin else 'boxed/builtin constructors' if reason.startswith('not yet: new an Identifier') else 'Object.prototype.isPrototypeOf' if reason=='refuses isPrototypeOf' else None
 assert not row['tscCodes'],row
 rows.append({**row,'bucket':'library' if feature else 'language','feature':feature or reason})
counts={bucket:dict(collections.Counter(r['feature'] for r in rows if r['bucket']==bucket)) for bucket in ['library','language']}
assert counts==expected,(counts,expected)
(root/'classification.json').write_text(json.dumps(rows,indent=2)+'\n')
print('stock tsc accepts all',len(rows),'first library blockers',sum(counts['library'].values()),'first language blockers',sum(counts['language'].values()))
