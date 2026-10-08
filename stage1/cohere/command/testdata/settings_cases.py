"""Settings controls transcribed from the pinned Go tests plus refusal reductions."""
import json
import pathlib
root = pathlib.Path(__file__).resolve().parent
cases = []
def case(name, files, queries=None, source='cohere/internal/format/formatoptions/resolve_test.go'):
    cases.append(dict(name=name, source=source, files=files, queries=queries or ['source/probe.css']))
case('zero config', {})
case('nearest ancestor', {'CohereSettings.json':'{"extends":"./NexusCohereSettings.json","rules":{}}', 'NexusCohereSettings.json':'{"format":{"tabWidth":4,"printWidth":120,"singleQuote":true,"bracketSameLine":true}}', 'libraries/structure/package.json':'{"name":"structure"}'}, ['libraries/structure/source/probe.css', 'source/probe.css'])
case('house chain', {'CohereSettings.json':'{"extends":"./structure/StructureCohereSettings.json","rules":{}}', 'structure/StructureCohereSettings.json':'{"extends":"../nexus/NexusCohereSettings.json","rules":{}}', 'nexus/NexusCohereSettings.json':'{"rules":{},"format":{"tabWidth":4,"useTabs":false,"semi":true,"singleQuote":true,"printWidth":120}}'})
case('settings no format', {'CohereSettings.json':'{"rules":{}}'})
case('own chain no format', {'CohereSettings.json':'{"extends":"./base/Base.json"}', 'base/Base.json':'{"rules":{}}'})
case('unified project block', {'CohereSettings.json':'{"extends":"cohere:system-inc/base","format":{}}'})
case('unified intermediate block', {'CohereSettings.json':'{"extends":"./shared/Shared.json"}', 'shared/Shared.json':'{"extends":"cohere:system-inc/structure","format":{"printWidth":100}}'})
case('outsider empty', {'CohereSettings.json':'{"extends":"cohere:typescript","format":{}}'})
case('outsider own', {'CohereSettings.json':'{"format":{"printWidth":100,"ignore":["generated/"]}}'})
case('outsider base', {'CohereSettings.json':'{"extends":"./base/Base.json"}', 'base/Base.json':'{"format":{"semi":false}}'})
case('own block replaces base', {'CohereSettings.json':'{"extends":"./base/Base.json","format":{"tabWidth":8}}', 'base/Base.json':'{"format":{"semi":false}}'})
case('ignore order', {'CohereSettings.json':'{"extends":"./base.json","ignorePatterns":["own/"],"format":{"ignore":["generated/"]}}', 'base.json':'{"ignorePatterns":["base/"]}'})
case('per file nearest', {'CohereSettings.json':'{"format":{"tabWidth":8}}', 'nested/CohereSettings.json':'{"format":{}}'}, ['source/probe.css','nested/source/probe.css','source/../source/again.css'])
case('leftover below settings', {'CohereSettings.json':'{"format":{}}','source/.prettierrc':'{}'})
case('leftover above winning settings ignored', {'.prettierrc':'{}','source/CohereSettings.json':'{"format":{}}'})
case('manifest first before legacy', {'package.json':'{"prettier":null}', '.prettierrc':'{}'})
legacy=['package.yaml','.prettierrc','.prettierrc.json','.prettierrc.yaml','.prettierrc.yml','.prettierrc.json5','.prettierrc.js','prettier.config.js','.prettierrc.ts','prettier.config.ts','.prettierrc.mjs','prettier.config.mjs','.prettierrc.mts','prettier.config.mts','.prettierrc.cjs','prettier.config.cjs','.prettierrc.cts','prettier.config.cts','.prettierrc.toml']
for name in legacy: case('legacy '+name, {name:'{}','CohereSettings.json':'{"format":{}}'})
case('legacy ordered', {name:'{}' for name in reversed(legacy)})
for raw in ['{}','null','[]','0','"x"','true','{"prettier":false}','{"prettier":{}}','\ufeff{}','{','{"a":}','{"a" 1}','{"a":1,}','[1,]','{}{}','{"a":"\\q"}','{"a":"\\uQQQQ"}','{"a":"\n"}','01','1.','1e+','truX']:
    case('manifest '+repr(raw), {'package.json':raw},source='cohere/internal/format/formatoptions/resolve.go:414')
for key in ['printWidth','tabWidth','useTabs','semi','singleQuote','trailingComma','bracketSpacing','bracketSameLine','arrowParens','endOfLine','ignore','quoteProps','plugins']:
    values=['null','{}','[]','true','"x"','2','2.5']
    if key in ['printWidth','tabWidth']: values += ['2e1','-2','9223372036854775808','-9223372036854775809']
    if key == 'ignore': values += ['["generated/",null]','[2]']
    for value in values: case('format '+key+' '+value, {'CohereSettings.json':'{"format":{'+json.dumps(key)+':'+value+'}}'})
for raw in ['null','[]','0','"x"','true','{}','{"format":null}','{"format":[]}','{"format":true}','{"root":true,"parser":{}}','{"overrides":[{"files":["*.ts"],"excludedFiles":["*.d.ts"]}]}','{"extends":2}','{"extends":[""]}','{"extends":[null]}','{"extends":[2]}','{"extends":"cohere:unknown"}','{"extends":"./missing.json"}','{"extends":"./CohereSettings.json"}','{"extends":["cohere:react","cohere:next"]}','{"format":{"tabWidth":8,"tabWidth":null}}','{"format":{"tabWidth":2.5,"tabWidth":2}}','{"format":{"endOfLine":"crlf","quoteProps":true}}']:
    case('settings '+raw, {'CohereSettings.json':raw}, source='cohere/internal/lint/configuration/configuration.go:666')
for key in ['departures','reasons','plugins','rules','ignorePatterns','overrides','cohere','settings']:
    for raw in ['2','"x"','[]','{}','[2]','null']:
        case('raw shape '+key+' '+raw, {'CohereSettings.json':'{'+json.dumps(key)+':'+raw+'}'},source='cohere/internal/lint/configuration/configuration.go:964')
for raw in ['{"overrides":[2]}','{"overrides":[{"files":2}]}','{"overrides":[{"files":[2]}]}','{"overrides":[{"rules":2}]}','{"overrides":[{"reason":2}]}','{"reasons":{"x":2}}','{"departures":{"x":2}}']:
    case('nested '+raw, {'CohereSettings.json':raw})
case('two-node cycle', {'CohereSettings.json':'{"extends":"./base.json"}','base.json':'{"extends":"./CohereSettings.json"}'})
case('diamond order', {'CohereSettings.json':'{"extends":["./a.json","./b.json"]}','a.json':'{"extends":"./shared.json","ignorePatterns":["a"]}','b.json':'{"extends":"./shared.json","ignorePatterns":["b"]}','shared.json':'{"ignorePatterns":["shared"]}'})
case('ignore unknown order', {'CohereSettings.json':'{"format":{"z":false,"a":false}}'})
for raw in ['{"format":{"arrowParens":"\\ud800"}}', '{"format":{"trailingComma":"\\udc00"}}', '{"format":{"arrowParens":"\\ud83d\\ude00"}}', '{"\\ue000":true,"\\ud800\\udc00":true}', '{"\\u2028":true}', '{"format":{"\\u0007":true}}']:
    case('Unicode settings '+raw, {'CohereSettings.json':raw})
for raw in ['{"extends":"./missing.json","extends":null}', '{"extends":"./missing.json","extends":""}', '{"extends":"./missing.json","extends":[]}', '{"extends":"cohere:react","extends":["cohere:next"]}']:
    case('duplicate extends '+raw, {'CohereSettings.json':raw})
(root/'settings-cases.json').write_text(json.dumps(cases, indent=2,ensure_ascii=True)+'\n')
print('settings cases',len(cases),'queries',sum(len(c['queries']) for c in cases))
