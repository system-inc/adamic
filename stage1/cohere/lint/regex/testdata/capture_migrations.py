"""Future esregexp option migration, through owned overlays, plus trace-only hooks."""
import json,pathlib,subprocess,tempfile,os,hashlib
root=pathlib.Path(__file__).resolve().parents[1];repo=root.parents[3];cohere=repo/'cohere';pkg=cohere/'internal/lint/rules/core'
def replace(text,old,new):
 assert old in text,old
 return text.replace(old,new)
with tempfile.TemporaryDirectory(prefix='regex-capture-') as temporary:
 temp=pathlib.Path(temporary);mapping={}
 for name in ['id_length','no_inline_comments','no_warning_comments']:
  source=(pkg/(name+'.go')).read_text()
  if name=='id_length':
   source=replace(source,'"regexp"','esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"')
   source=replace(source,'*regexp.Regexp','*esregexp.RegExp')
   source=replace(source,'regexp.Compile(pattern)','esregexp.Compile(pattern, "u")')
   source=replace(source,'pattern.MatchString(name)','waveRegexTest("id-length", pattern, name)')
  else:
   source=replace(source,'"regexp"','"regexp"\n esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"')
   source=replace(source,'*regexp.Regexp','*esregexp.RegExp')
   if name=='no_inline_comments':
    source=replace(source,'var ignorePattern *esregexp.RegExp','var ignorePattern *esregexp.RegExp\n waveRawPattern := \"\"')
    source=replace(source,'if compiled, err := regexp.Compile(resolved.IgnorePattern)','waveRawPattern = resolved.IgnorePattern; if compiled, err := regexp.Compile(resolved.IgnorePattern)')
    source=replace(source,'regexp.Compile(resolved.IgnorePattern)','esregexp.Compile(resolved.IgnorePattern, "u")')
    source=replace(source,'ignorePattern.MatchString(body)','waveRegexTest("no-inline-comments", ignorePattern, body)')
    source=replace(source,'fileComments := comments.ForFile(ctx)','fileComments := comments.ForFile(ctx)\n waveInlineGeometry(ctx,fileComments,waveRawPattern)')
   else:
    source=replace(source,'regexp.Compile("(?i)" + prefix + escaped + suffix)','waveCompileWarning(prefix + escaped + suffix, term, location, decoration)')
    source=replace(source,'matcher.MatchString(value)','waveWarningTest(matcher, value)')
  path=temp/(name+'.go');path.write_text(source);mapping[str(pkg/(name+'.go'))]=str(path)
 mapping[str(pkg/'adamic_inline_geometry.go')]=str(root/'testdata/inline_geometry.go')
 mapping[str(pkg/'adamic_regex_capture.go')]=str(root/'testdata/migration_capture.go')
 mapping[str(pkg/'adamic_regex_controls_test.go')]=str(root/'testdata/migration_controls.go.txt')
 testing_file=cohere/'internal/lint/testing/rule_testing.go'
 testing_source=replace(testing_file.read_text(),'expectEachFixParses(t, sourceFile, diagnostics)','waveFixtureCapture(subject, fileName, sourceText, options, diagnostics)\n expectEachFixParses(t, sourceFile, diagnostics)')
 testing_overlay=temp/'rule_testing.go';testing_overlay.write_text(testing_source)
 mapping[str(testing_file)]=str(testing_overlay)
 mapping[str(cohere/'internal/lint/testing/adamic_regex_fixtures.go')]=str(root/'testdata/fixture_capture.go')
 overlay=temp/'overlay.json';overlay.write_text(json.dumps({'Replace':mapping}))
 capture=temp/'cases.jsonl';fixtures=temp/'fixtures.jsonl';geometry=temp/'inline.jsonl';env=dict(os.environ,ADAMIC_REGEX_CAPTURE=str(capture),ADAMIC_REGEX_FIXTURES=str(fixtures),ADAMIC_INLINE_GEOMETRY=str(geometry))
 with (root/'evidence/migration-capture.log').open('wb') as log:
  result=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/rules/core','-run','^Test(IdLength|NoInlineComments|NoWarningComments|WaveRegexControls)','-count=1','-v','-timeout=10m'],cwd=cohere,env=env,stdout=log,stderr=subprocess.STDOUT)
 if result.returncode:raise SystemExit('upstream capture gate failed; see evidence/migration-capture.log')
 cases={}
 for line in capture.read_text().splitlines():
  obj=json.loads(line);key=json.dumps(obj,ensure_ascii=False,sort_keys=True);cases[key]=obj
 rows=[cases[key]for key in sorted(cases)]
 (root/'testdata/migration_cases.json').write_text(json.dumps(rows,ensure_ascii=False,indent=2)+'\n')
 # Generated .a input data are replayed independently by each migration.
 (root/'testdata/migration_cases.a').write_text('export interface MigrationCase { rule: string; input: string; pattern: string; term: string; location: string; decoration: string[]; matched: boolean; }\nexport function migrationCases(): MigrationCase[] {\n const cases: MigrationCase[] = [];\n'+''.join(' cases.push('+json.dumps(dict(r, decoration=r['decoration']or[]),ensure_ascii=False).replace('\u2028','\\u2028').replace('\u2029','\\u2029')+');\n'for r in rows)+' return cases;\n}\n')
 snapshots={}
 for line in fixtures.read_text().splitlines():
  row=json.loads(line);key=json.dumps(row,sort_keys=True,ensure_ascii=False);snapshots[key]=row
 warning_rows=[snapshots[key]for key in sorted(snapshots) if snapshots[key]['rule']=='no-warning-comments']
 id_rows=[snapshots[key]for key in sorted(snapshots) if snapshots[key]['rule']=='id-length']
 (root/'testdata/id_fixtures.json').write_text(json.dumps(id_rows,indent=2,ensure_ascii=False)+'\n')
 id_inputs=[]
 for r in id_rows:
  options=dict(r['options']or{})
  if isinstance(options.get('Exceptions'),dict):options['Exceptions']=[name for name,allowed in options['Exceptions'].items()if allowed]
  id_inputs.append(dict(source=r['source'],file=r['file'],options=json.dumps(options,ensure_ascii=False)))
 (root/'testdata/id_fixtures.a').write_text('export interface IdFixture { source: string; file: string; options: string; }\nexport function idFixtures(): IdFixture[] {\n const cases: IdFixture[] = [];\n'+''.join('cases.push('+json.dumps(r,ensure_ascii=False).replace('\u2028','\\u2028').replace('\u2029','\\u2029')+');\n'for r in id_inputs)+'return cases; }\n')
 print('complete id fixtures',len(id_rows))
 inline_rows=[snapshots[key]for key in sorted(snapshots) if snapshots[key]['rule']=='no-inline-comments']
 (root/'testdata/inline_fixtures.json').write_text(json.dumps(inline_rows,indent=2,ensure_ascii=False)+'\n')
 geometries={}
 for line in geometry.read_text().splitlines():
  g=json.loads(line);geometries[(g['source'],g['pattern'])]=g
 inline_inputs=[]
 for row in inline_rows:
  opts=row['options']or{};pattern=opts.get('IgnorePattern')or''
  view=geometries[(row['source'],pattern)]
  inline_inputs.append(view)
 (root/'testdata/inline_fixtures.a').write_text("import type { InlineFileView } from '../inline_rule.a';\nexport interface InlineFixture extends InlineFileView { pattern: string; }\nexport function inlineFixtures(): InlineFixture[] {\n const cases: InlineFixture[] = [];\n"+''.join('cases.push('+json.dumps(r,ensure_ascii=False).replace('\u2028','\\u2028').replace('\u2029','\\u2029')+');\n'for r in inline_inputs)+' return cases; }\n')
 print('projected inline fixtures',len(inline_rows))

 (root/'testdata/warning_fixtures.json').write_text(json.dumps(warning_rows,indent=2,ensure_ascii=False)+'\n')
 input_rows=[]
 for row in warning_rows:
  options=row['options']or{}
  terms=options.get('Terms');terms=['todo','fixme','xxx']if terms is None else terms
  location=options.get('Location')or'Start'
  normalized={'terms':terms,'location':'start'if location.lower()=='start'else'anywhere','decoration':options.get('Decoration')or[]}
  input_rows.append({'source':row['source'],'file':row['file'],'options':json.dumps(normalized,ensure_ascii=False)})
 (root/'testdata/warning_fixtures.a').write_text('export interface WarningFixture { source: string; file: string; options: string; }\nexport function warningFixtures(): WarningFixture[] { return '+json.dumps(input_rows,ensure_ascii=False).replace('\u2028','\\u2028').replace('\u2029','\\u2029')+'; }\n')
 print('complete warning fixtures',len(warning_rows))
 print('captured',len(rows),'unique matcher observations')
 from collections import Counter
 print(dict(Counter(r['rule']for r in rows)))

# Fixture snapshots contain only Go-owned geometry and findings for comparison,
# never answers consumed by the production matcher.
