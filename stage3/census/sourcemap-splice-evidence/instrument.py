from pathlib import Path
import sys
p=Path(sys.argv[1])/'src/compiler/sourcemap.ts';s=p.read_text()
s=s.replace('mappingCharCodes.length','((globalThis as any).__smRead("length", mappingCharCodes)).length')
assert s.count('mappingCharCodes.push(charCode)')==1
s=s.replace('mappingCharCodes.push(charCode)','(globalThis as any).__smPush(mappingCharCodes, charCode)')
s=s.replace('String.fromCharCode.apply(undefined, mappingCharCodes)','String.fromCharCode.apply(undefined, (globalThis as any).__smRead("apply", mappingCharCodes))')
s=s.replace('function flushMappingBuffer(): void {','function flushMappingBuffer(): void {\n        (globalThis as any).__smFlush(mappingCharCodes);')
assert s.count('mappingCharCodes.splice(0);')==1
s=s.replace('mappingCharCodes.splice(0);','(globalThis as any).__smClearBefore(mappingCharCodes);\n            mappingCharCodes.splice(0);\n            (globalThis as any).__smClearAfter(mappingCharCodes);')
p.write_text(s)
