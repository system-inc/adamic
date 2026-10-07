from pathlib import Path
import subprocess,json,hashlib
root=Path('/workspace/lint-cost-controls');root.mkdir(exist_ok=True)
cases=[
('octal','no-octal-escape',r"const value = '\251';"),
('multiline_call','no-unexpected-multiline','const x = f\n(a);'),
('multiline_element','no-unexpected-multiline','const x = f\n[a];'),
('multiline_tagged','no-unexpected-multiline','const x = f\n`a`;'),
('multiline_division','no-unexpected-multiline','const x = a\n/b/g;'),
('private_class','no-unused-private-class-members','class C { #x=0; m(){this.#x=1;} }'),
('private_expression','no-unused-private-class-members','const C = class { #x=0; };'),
('constructor','no-useless-constructor','class C { constructor(){} }'),
('quoted_constructor','no-useless-constructor',"class C { 'constructor'(){} }"),
('template_string','prefer-template',"const x = 'a' + value;"),
('template_no_substitution','prefer-template','const x = `a` + value;'),
('template_expression','prefer-template','const x = `a${v}` + value;'),
('forward_ref','react/forward-ref-uses-ref','const C = forwardRef(props => props);'),
('find_dom_node','react/no-find-dom-node',"ReactDOM['findDOMNode'](node);"),
('is_mounted','react/no-is-mounted','class C { m(){this.isMounted();} }'),
('pure_class','react/no-redundant-should-component-update','class C extends PureComponent {shouldComponentUpdate(){}}'),
('pure_expression','react/no-redundant-should-component-update','const C = class extends PureComponent {shouldComponentUpdate(){}};'),
('jsx_text','react/jsx-no-comment-textnodes','const view = <div>// comment</div>;'),
]
oracle='/workspace/lint-cost-baseline/oracle';rows=[];counts={}
for name,rule,source in cases:
 p=root/(name+'.a');p.write_text(source+'\n');row=str(p)+'\t'+rule
 if name=='jsx_text':
  alias=root/(name+'.a.tsx')
  if not alias.exists():alias.symlink_to(p.name)
  with (root/'jsx-spans.stdout').open('wb') as out,(root/'jsx-spans.stderr').open('wb') as err:r=subprocess.run([oracle,'--jsx-spans',str(alias)],stdout=out,stderr=err)
  assert r.returncode==0 and not (root/'jsx-spans.stderr').read_bytes();spans=','.join((root/'jsx-spans.stdout').read_text().splitlines());row=str(alias)+'\t'+rule+'\t'+spans
 manifest=root/(name+'.txt');manifest.write_text(row+'\n')
 with (root/(name+'-Go.stdout')).open('wb') as out,(root/(name+'-Go.stderr')).open('wb') as err:r=subprocess.run([oracle,'--manifest',str(manifest),'--count'],stdout=out,stderr=err)
 assert r.returncode==0 and not (root/(name+'-Go.stderr')).read_bytes();count=int((root/(name+'-Go.stdout')).read_text());assert count>0,(name,'not a positive Go control');counts[name]=count;rows.append(row)
manifest=root/'controls.txt';manifest.write_text('\n'.join(rows)+'\n');records={}
for label,cmd in [('Go',[oracle]),('today',['/workspace/lint-cost-baseline/scanner']),('prototype',['/workspace/lint-cost-prototype/scanner']),('Node',['node','--disable-warning=ExperimentalWarning','/workspace/adamic/oracle/node.mjs','/workspace/lint-cost-prototype/main.ts'])]:
 with (root/(label+'.stdout')).open('wb') as out,(root/(label+'.stderr')).open('wb') as err:r=subprocess.run(cmd+['--manifest',str(manifest)],stdout=out,stderr=err)
 data=(root/(label+'.stdout')).read_bytes();err=(root/(label+'.stderr')).read_bytes();assert r.returncode==0 and not err,(label,r.returncode,err[:1000])
 if label=='Go':want=data
 else:assert data==want,(label,'fixture byte mismatch')
 records[label]={'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest()};print(label,records[label],flush=True)
Path('/workspace/lint-cost-controls.json').write_text(json.dumps({'positive_counts':counts,'outputs':records,'jsx_alias':'source authored as .a; .a.tsx symlink selects TSX parsing in the Go oracle'},indent=2)+'\n')
