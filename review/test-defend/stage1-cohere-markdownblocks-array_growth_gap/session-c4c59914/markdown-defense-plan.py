import pathlib,json,difflib,subprocess,re
R=pathlib.Path('/workspace/adamic');P=R/'review/test-defend/stage1-cohere-markdownblocks-array_growth_gap/session-c4c59914'
plan=[dict(id='D1',test='TestMarkdownASTPreprocessing',file='stage1/cohere/markdownblocks/astPreprocess.ts',old="                    node.originalAlt = alt ?? '';\n",new='',menu='drop statement',difference='AstPreprocessor.run original image alt projection; parser representation probes compile independent gap programs and never import this port.'),dict(id='D2',test='TestMicromarkInputChunks',file='stage1/cohere/markdownblocks/inputChunks.ts',old="    return output.join(';');",new="    return output.join(',');",menu='change constant',difference='serializeChunks is an exported port function reached by chunks_probe only. Tokenizer events call inputChunks but never serializeChunks.'),dict(id='D3',test='TestMarkdownSourceDecoding',file='stage1/cohere/markdownblocks/decodeString.ts',old='(code > 64975 && code < 65008)',new='(code > 64974 && code < 65008)',menu='off by one lower bound',difference='numericReference newly replaces valid U+FDCF (64975). Exhaustive numeric decoder inputs include it; independent parser gap programs do not call the decoder. Mdast programs also reach the decoder and must be tested.')]
for m in plan:
 s=(R/m['file']).read_text();assert s.count(m['old'])==1;m['line']=s[:s.index(m['old'])].count('\n')+1
 (P/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(m['old'],m['new'],1).splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
 (P/(m['id']+'-apply-check.log')).write_text(subprocess.check_output(['git','apply','--check',str(P/(m['id']+'.diff'))],cwd=R).decode())
(P/'plan.json').write_text(json.dumps(plan,indent=2))
cmd=['rg','-n','ast_probe|chunks_probe|decode_probe|mdast_probe|AstPreprocessor|serializeChunks|decodeString','stage1/cohere/markdownblocks','--glob','*.ts','--glob','*test.go','--glob','!events_units_test.go']
p=subprocess.run(cmd,cwd=R,capture_output=True,text=True);(P/'callers.txt').write_text(p.stdout)
print(json.dumps(plan,indent=2))
