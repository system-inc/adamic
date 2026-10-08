import json, subprocess, pathlib, sys
prefix='/tmp/class-construction-census-source/src/compiler/'
rows=[
('binder.ts:2121:20','string | NodeArray<JSDocComment> | undefined',23),
('commandLineParser.ts:2599:13','"boolean" | "list" | "listOrElement" | "number" | "object" | "string" | Map<string, string | number>',4),
('checker.ts:16202:62','NodeArray<ParameterDeclaration> | readonly JSDocParameterTag[]',4),
('checker.ts:47945:13','string | number | undefined',3),
('tsbuildPublic.ts:656:12','AnyBuildOrder | undefined',2),
('moduleNameResolver.ts:512:46','boolean | (() => boolean) | undefined',2),
('moduleNameResolver.ts:2235:23','false | string[] | undefined',2),
('executeCommandLine.ts:375:21','"boolean" | "list" | "number" | "object" | "string" | Map<string, string | number>',1),
('commandLineParser.ts:1896:13','"boolean" | "number" | "object" | "string" | Map<string, string | number>',1),
('checker.ts:39371:13','0 | boolean | undefined',1),
('watchUtilities.ts:741:23','boolean | (() => boolean)',1),
('moduleNameResolver.ts:2418:12','string | false',1),
('moduleNameResolver.ts:2415:9','string | false | undefined',1),
('transformers/utilities.ts:441:33',None,1)]
phase=sys.argv[1]
for i,(where,typ,count) in enumerate(rows,1):
 reason='a field of type '+typ if typ else "a base that isn't a declared class"
 args=['/tmp/class-construction-trace-replay' if phase=='trace' else '/tmp/class-construction-after-replay' if phase=='after' else '/tmp/class-construction-replay','-project','/tmp/class-construction-census-source/src/tsc/tsc.ts','-where',prefix+where,'-kind','NotYet','-reason',reason]
 stem=f'/tmp/class-construction-{phase}-{i:02d}'
 with open(stem+'.json','w') as out,open(stem+'.log','w') as err:
  run=subprocess.run(args,stdout=out,stderr=err,env=__import__("os").environ|{"LATENT_TRACE_WHERE":prefix+where})
 print(i,count,run.returncode,where,flush=True)
pathlib.Path('/tmp/class-construction-kinds.json').write_text(json.dumps(rows,indent=2))
