import pathlib,json,argparse,subprocess,os
parser=argparse.ArgumentParser();parser.add_argument('directory');args=parser.parse_args()
source=pathlib.Path(__file__).resolve().parent;repository=source.parents[3];root=pathlib.Path(args.directory).resolve();root.mkdir(parents=True,exist_ok=True)
replacements={}
for name in ['correctness_no_process_exit_after_output_test.go','correctness_require_blocking_standard_streams_test.go']:
 original=repository/'cohere/internal/lint/rules/nexus'/name;text=original.read_text().replace('rule_testing.RunTypedFiles(t,','wave08Capture(t,').replace('rule_testing.RunTypedFilesWithSetup(t,','wave08CaptureWithSetup(t,');replacement=root/name;replacement.write_text(text);replacements[str(original)]=str(replacement)
virtual=repository/'cohere/internal/lint/rules/nexus/adamic_wave08_capture_test.go';replacements[str(virtual)]=str(source/'testdata/capture_process_test.go');overlay=root/'overlay.json';overlay.write_text(json.dumps({'Replace':replacements}))
with open(root/'capture.stdout','wb') as out,open(root/'capture.stderr','wb') as err:
 p=subprocess.run(['go','test','-overlay',str(overlay),'./internal/lint/rules/nexus','-run','^TestCorrectness(NoProcessExitAfterOutput|RequireBlockingStandardStreams)','-count=1','-v'],cwd=repository/'cohere',stdout=out,stderr=err,env=dict(os.environ,ADAMIC_WAVE08_CAPTURE=str(root/'inputs')))
assert p.returncode==0,(root/'capture.stderr').read_text()
for path in sorted((root/'inputs').glob('*.json')):
 data=json.loads(path.read_text());directory=root/'cases'/path.stem
 for filename,text in data['Files'].items():
  target=directory/filename.lstrip('/');target.parent.mkdir(parents=True,exist_ok=True);target.write_text(text)
 roots=[str(directory/name.lstrip('/')) for name in sorted(data['Files']) if name.endswith('.ts') and not name.endswith('.d.ts')]
 (directory/'roots.manifest').write_text(''.join(p+'\n' for p in roots));(directory/'tsconfig.json').write_text(data['Config'])
 (directory/'metadata.json').write_text(json.dumps({k:v for k,v in data.items() if k!='Files'},indent=2))
print('captured',len(list((root/'inputs').glob('*.json'))),'upstream fixture programs',flush=True)
