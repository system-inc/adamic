"""Audit the continuing predicate seam and diagnosed-body boundary independently."""
import json,pathlib,subprocess,os,tempfile,sys
binary=pathlib.Path(sys.argv[1]).resolve()
with tempfile.TemporaryDirectory(prefix='latent-predicate-audit-') as directory:
 root=pathlib.Path(directory);source=root/'source';source.mkdir()
 (source/'proof.a').write_text('''function first(value: unknown): value is number { return true; }
function second(value: unknown): value is number { return typeof value === "number" && value > 0; }
function good(value: unknown): value is number { return typeof value === "number"; }
function accept(value: number, callback: (value: number) => value is 1): boolean { return callback(value); }
function fake(value: number): value is 1 { return true; }
accept(2, fake);
function broken(value: unknown): value is number { const wrong: number = "wrong"; return true; }
''')
 records=root/'records.jsonl'
 result=subprocess.run([str(binary),str(source),str(records)],env=dict(os.environ,LATENT_ASSERT_NO_OUTPUT='1'),capture_output=True,text=True)
 assert result.returncode==0,(result.returncode,result.stdout,result.stderr)
 raw=[json.loads(line) for line in records.read_text().splitlines()]
 assert raw[0]['checker_rejected']
 row=raw[1];refusals=[f for f in row['findings'] if f['phase']=='refusal_scan']
 by_line=lambda line:[f for f in refusals if f['where'].split(':')[-2]==str(line)]
 assert by_line(1) and by_line(2) and by_line(5),refusals
 assert not by_line(3),by_line(3)
 assert any(f['kind']=='Refused' and 'unproven predicate argument for parameter callback' in f['reason'] for f in by_line(6)),refusals
 broken=next(u for u in row['units'] if u['where'].split(':')[-2]=='7')
 assert broken['status']=='skipped_checker_body',broken
 assert not any(f['unit']==broken['where'] for f in row['findings']),row
 print(json.dumps(dict(continuing_predicate_refusals='pass',positive_body_proof='pass',argument_contract='pass',diagnosed_body_boundary='pass',output_guards='pass',refusals=refusals),indent=2))
