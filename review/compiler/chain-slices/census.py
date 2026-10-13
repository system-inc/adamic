import subprocess,json,pathlib
root=pathlib.Path('review/compiler/chain-slices');root.mkdir(parents=True,exist_ok=True)
def git(*args): return subprocess.check_output(['git',*args],text=True)
rows=[('optional-chain-after-call-main','5028fa69'),('spread-shorthand','e82c9646'),('optional-calls-main','b6ef2cb4'),('c-portability-main','ebad12cf'),('generic-body-relations','2cc109e5'),('generics-scout-main','48391dcf'),('iteration-main','518eca83'),('namespace-value','26ccff9c'),('checked-any','57ea02f8'),('project-references-main','e2aa3750'),('refusal-rulings-main','8799f518'),('assignment-proofs-main','a6f05066'),('tsgo.go-cache','63f5bb07'),('miscompile-fxspptb-2b','6de2edb5'),('placeholder-nonnull-main','bafb9ef4'),('step24-parser-main','e503868c'),('miscompile-fxspptb-2a','373f0055'),('callback-widening','36472529'),('generators-main','6f0ebd15'),('exceptions-21-main','205586a0'),('inherit-guards','16a0b626'),('self-compare-main','387c2826'),('feature-set-link-main','e1efb527'),('per-backend-stops','7a151f7f'),('search-shrink','c0c4e102'),('optional-presence-next','897d0e79'),('eep-presence','21ceb23c'),('lint-features','35353870')]
result=[]
for name,sha in rows:
 base=git('merge-base','origin/main',sha).strip()
 if name in ['tsgo.go-cache','lint-features']:base=git('rev-parse',sha+'^').strip()
 patch=subprocess.check_output(['git','diff','--binary',base,sha,'--','.',':!review',':!docs'])
 p=subprocess.run(['git','apply','--check','--reverse'] if name=='c-portability-main' else ['git','apply','--check'],input=patch,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
 result.append(dict(member=name,source=git('rev-parse',sha).strip(),base=base,patch_check_exit=p.returncode,diagnostic=p.stderr.decode(),files=git('diff','--name-only',base,sha).splitlines()))
(root/'member-patch-checks.json').write_text(json.dumps(result,indent=2)+'\n')
(root/'merge-history.txt').write_text(git('log','--merges','--format=%H %P%n%B','88adf555..f5236b48'))
(root/'commit-history.txt').write_text(git('log','--format=%H %P%n%B','--reverse','88adf555..f5236b48'))
print('\n'.join(f"{r['member']} {r['source'][:8]} check={r['patch_check_exit']} files={len(r['files'])}" for r in result))
