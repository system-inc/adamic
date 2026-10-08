#!/usr/bin/env python3
"""Attach checked source sites to every saved reduction and count repeated witnesses.
Emission sites always identify the actual rule implementations. Extra causal sites
are recorded only for the inspected witness shapes; they do not imply a shared fix.
"""
import argparse, collections, gzip, json, pathlib, re
p=argparse.ArgumentParser();p.add_argument('reductions');p.add_argument('out');a=p.parse_args()
root=pathlib.Path(__file__).resolve().parents[3];out=pathlib.Path(a.out);out.mkdir(parents=True,exist_ok=True)
def site(path,anchor):
 text=(root/path).read_text();at=text.find(anchor)
 if at<0:raise SystemExit('missing attribution anchor: '+path+' '+anchor)
 return path+':'+str(text[:at].count('\n')+1)
P='stage1/typescript/parser/parser.ts';S='stage1/typescript/parser/statements.ts';G='cohere/TypeScript/tsc/internal/parser/parser.go';E='stage1/cohere/tsprinter/expressions.ts';C='cohere/internal/lint/ecmascript/comments/comments.go'
def causal(r):
 text=r['deletion_minimal_input'];name=r['rule'];sig=r['signature'];port=[];go=[];cause='rule emission/fix sites; underlying tree difference requires owner investigation'
 def add(label,pp,pa,gp,ga):
  nonlocal cause
  cause=label;port.append(site(pp,pa));go.append(site(gp,ga))
 if name=='format/typescript':
  if sig.startswith('comment-loss'):
   add('whole-program printer has no comment attachment',E,'export function formatProgram','cohere/internal/format/javascript/comments_hooks.go','func printComment(')
  elif sig=='BOM preservation':
   add('whole-program printer omits file BOM wrapper',E,'export function formatProgram','cohere/internal/format/native/native.go','hasByteOrderMark :=')
  elif '\n\n' in text:
   add('statement sequence emits one hardline and loses the source empty line',E,'if(parts.length > 0) parts.push(this.docs.hardline());','cohere/internal/format/javascript/print_block.go','if isNextLineEmptyAfter(statement, options)')
  elif text.startswith('#!'):
   add('whole-program entry omits shebang preservation',E,'export function formatProgram','cohere/internal/format/estree/parse.go','func ReplaceHashbang(')
  elif 'using' in text:
   add('VariableStatement prints child declaration list without statement modifiers',E,"case 'VariableStatement':",'cohere/internal/format/javascript/print_expressions.go','func printVariableDeclaration(')
  elif 'yield' in text:
   add('yield expression/parenthesization printer branch',E,"case 'YieldExpression':",'cohere/internal/format/javascript/print.go','case "YieldExpression":')
  elif '<>' in text:
   add('call type-argument printing branch',E,"case 'CallExpression':",'cohere/internal/format/javascript/print_call_expression.go','func printCallExpression(')
  elif text.startswith('`'):
   add('template raw text/line-ending branch',E,"case 'NoSubstitutionTemplateLiteral':",'cohere/internal/format/javascript/print_template_literal.go','func printTemplateLiteral(')
  elif re.search(r'\{\d+n[:(]',text):
   add('property name parser omits bigint literal branch',P,'propertyName(): number {',G,'func (p *Parser) parsePropertyNameWorker(')
  else:
   add('program statement/expression printer dispatch',E,'statementBodyDoc(index:','cohere/internal/format/javascript/print.go','func')
 elif name=='format/json':
  add('trailing comment text retained verbatim instead of trimming line end','stage1/cohere/json/formatter.ts',"before.push(documents.text(' '));",'cohere/internal/format/javascript/comments_hooks.go','estree.TrimEndJavaScript(options.OriginalText')
 elif pathlib.Path(r['file']).suffix in {'.js','.jsx'} and '>(' in text and name in {'no-sequences','@typescript-eslint/no-wrapper-object-types','nexus/consistency-no-abbreviated-identifier','nexus/consistency-no-ambiguous-identifier','lint/syntax-fixes'}:
  add('JavaScript suffix path accepts TypeScript type arguments',P,"if(this.kind() === 'LessThanToken' && typeArgumentsAhead",G,'func (p *Parser) tryParseTypeArgumentsInExpression(')
 elif 'for(var of ' in text or 'for(await using of ' in text:
  add('for-header empty variable list recovery consumes of as binding',S,'variableList(): number {',G,'func (p *Parser) parseVariableDeclarationList(')
 elif 'set f(){}' in text or 'get f(){}' in text:
  add('type literal accessor parser leaves recovery body outside member',P,'typeLiteral(): number {',G,'func (p *Parser) parseAccessorDeclaration(')
 elif (name=='no-useless-rename' or name=='lint/syntax-fixes') and '...rest:rest' in text:
  add('binding rename listener skips rest-token elements','stage1/cohere/lint/rules/no-useless-rename/rule.a',"c.kind(parts[0] ?? -1) === 'DotDotDotToken'",'cohere/internal/lint/rules/core/no_useless_rename.go','func checkUselessRenameInBindingPattern(')
 elif 'interface I{[a,]' in text:
  add('index signature lookahead omits comma recovery',P,'indexSignatureAhead(): boolean {',G,'func (p *Parser) nextIsUnambiguouslyIndexSignature(')
 elif 't=({' in text and 'n:()=>{' in text:
  add('speculative parenthesized arrow accepts recovered parameter diagnostics',P,'arrowCandidate(allowReturn:',G,'func (p *Parser) tryParseParenthesizedArrowFunctionExpression(')
 elif '<{-readonly' in text:
  add('type-argument lookahead rejects mapped-type minus modifier','stage1/typescript/parser/lookahead.ts','export function typeArgumentsAhead(',G,'func (p *Parser) tryParseTypeArgumentsInExpression(')
 elif 'new?' in text:
  add('type member new keyword treated as construct signature without lookahead',P,"if(this.kind() === 'NewKeyword' || this.kind() === 'OpenParenToken'",G,'func (p *Parser) parseTypeMember(')
 elif '...' in text and '<' in text and (name=='react/jsx-curly-brace-presence' or name=='lint/syntax-fixes'):
  add('JSX expression helper selects spread token instead of expression child','stage1/cohere/lint/rules/react-jsx-curly-brace-presence/rule.a','expression(index: number): number {','cohere/internal/lint/rules/react/jsx_curly_brace_presence.go','container.AsJsxExpression().Expression')
 elif name=='lint/syntax-fixes' and '#!' in text:
  add('comment guard stops on newline after skipped shebang','stage1/cohere/lint/rules/nexus-consistency-no-long-line-comment/comments.a',"if(anchor === 0 && context.source.startsWith('#!'))",C,'func canBeginAt(')
 elif (name=='lint/syntax-fixes' or name=='@typescript-eslint/no-inferrable-types') and '!' in text and ':number' in text:
  add('annotation fix endpoint/definite-assignment token edits','stage1/cohere/lint/rules/typescript-eslint-no-inferrable-types/rule.a','if(token >= 0) edits.push','cohere/internal/lint/rules/typescript/no_inferrable_types.go','func reportNoInferrableType(')
 elif 'function*yield' in text or 'function await' in text:
  add('function name rejected before generator/async body context is entered',P,'!(async && this.kind()',G,'func (p *Parser) parseFunctionExpression(')
 elif 'using{}' in text:
  add('using declaration lookahead omits eager object binding recovery',S,"(this.parser.kind() === 'UsingKeyword' && this.parser.nextIdentifierSameLine())",G,'func (p *Parser) isUsingDeclaration(')
 elif 'catch(' in text and '=' in text:
  add('catch variable parser omits initializer and ends declaration early',S,"if(this.parser.kind() === 'CatchKeyword')",G,'func (p *Parser) parseCatchClause(')
 elif 'extends import.' in text:
  add('heritage element start gate excludes import meta property','stage1/typescript/parser/grammar.ts','export function heritageTokenStart(',G,'func (p *Parser) parseTypeHeritageClauseElement(')
 elif name=='no-empty' and 'n()' in text:
  add('property name parser omits bigint literal branch',P,'propertyName(): number {',G,'func (p *Parser) parsePropertyNameWorker(')
 elif '=>' in text and any(token in text for token in ['{e:y=>','{y:s=>','{m:e=>']):
  add('speculative parenthesized arrow accepts recovered parameter diagnostics',P,'arrowCandidate(allowReturn:',G,'func (p *Parser) tryParseParenthesizedArrowFunctionExpression(')
 elif name=='nexus/consistency-no-ambiguous-identifier' and 'handle:' in text:
  add('event context treats final type child as initializer','stage1/cohere/lint/rules/nexus-consistency-no-ambiguous-identifier/helpers.a',"if(node.kind === 'VariableDeclaration' && last === cursor)",'cohere/internal/lint/rules/nexus/consistency_no_ambiguous_identifier.go','declaration.Initializer == current')
 elif '{[' in text and ' in ' in text and (';' in text or '\nl()' in text or '\nl:' in text):
  add('mapped type parser omits recovery members after the value type',P,'mappedType(): number {',G,'func (p *Parser) parseMappedType(')
 elif name=='no-unused-private-class-members' and 'typeof' in text:
  add('private-use scan treats type-only qualified names as value reads','stage1/cohere/lint/rules/no-unused-private-class-members/rule.ts','function used(','cohere/internal/lint/rules/core/no_unused_private_class_members.go','func markPrivateMemberUses(')
 elif name=='nexus/consistency-no-long-line-comment' and '\ufeff' in text:
  add('comment fold emptiness uses different whitespace sets','stage1/cohere/lint/rules/nexus-consistency-no-long-line-comment/rule.a',"trim(text) === ''",'cohere/internal/lint/rules/nexus/consistency_no_long_line_comment.go','if text.TrimWhitespace(body) == ""')
 elif re.search(r'import(?:\s|/\*.*?\*/)+source',text,re.S):
  add('import phase parser recognizes type/defer but not source',S,'importDeclaration(pos:',G,'func (p *Parser) parseImportDeclarationOrImportEqualsDeclaration(')
 elif 'constructor' in text:
  add('constructor keyword overrides accessor/generator member kind',S,"const constructor = this.parser.kind() === 'ConstructorKeyword';",G,'func (p *Parser) tryParseConstructorDeclaration(')
 elif 'static{' in text:
  add('static block recognition depends on first modifier',S,"this.parser.node(member[0] ?? -1).kind === 'StaticKeyword'",G,'func (p *Parser) parseClassStaticBlockDeclaration(')
 elif '#!' in text and ('comment' in name or 'shouting' in name):
  slug='nexus-consistency-no-shouting' if 'shouting' in name else 'nexus-consistency-no-long-line-comment'
  add('comment guard stops on newline after skipped shebang','stage1/cohere/lint/rules/'+slug+'/comments.a',("if(anchor === 0 && source.startsWith('#!'))" if 'shouting' in name else "if(anchor === 0 && context.source.startsWith('#!'))"),C,'func canBeginAt(')
 elif '/**@' in text:
  add('whole-file parser omits attached lazy JSDoc tag trees',P,'file(): number {','cohere/TypeScript/tsc/internal/parser/jsdoc.go','func (p *Parser) parseJSDocComment(')
 elif name=='@typescript-eslint/no-inferrable-types' and '!' in text:
  add('annotation fix endpoint/definite-assignment token edits','stage1/cohere/lint/rules/typescript-eslint-no-inferrable-types/rule.a','if(token >= 0) edits.push','cohere/internal/lint/rules/typescript/no_inferrable_types.go','func reportNoInferrableType(')
 elif ('warning' in name or 'shouting' in name or 'jsdoc' in name) and '<' in text:
  if name=='no-warning-comments':pp='stage1/cohere/lint/rules/no-warning-comments/rule.ts';anchor='commentAnchors(index: number)'
  elif 'shouting' in name:pp='stage1/cohere/lint/rules/nexus-consistency-no-shouting/comments.a';anchor='function'
  else:pp='stage1/cohere/lint/rules/nexus-consistency-no-single-line-jsdoc/rule.a';anchor='export class'
  add('JSX interior comment anchoring/scan boundary',pp,anchor,C,'func collectListInteriors(')
 elif '({' in text and re.search(r':(?:[A-Za-z_$][\w$]*|\(\))=>',text):
  add('speculative parenthesized arrow accepts recovered parameter diagnostics',P,'arrowCandidate(allowReturn:',G,'func (p *Parser) tryParseParenthesizedArrowFunctionExpression(')
 elif name=='nexus/consistency-no-ambiguous-identifier' and 'const{...a:b}' in text:
  add('foreign binding key detection assumes first child is the key','stage1/cohere/lint/rules/nexus-consistency-no-ambiguous-identifier/helpers.a',"case 'BindingElement':",'cohere/internal/lint/ecmascript/binding/ownership.go','func IsForeignName(')
 elif 'export const;await' in text:
  add('whole-file parser does not reparse top-level await after module detection',P,'file(): number {',G,'func (p *Parser) reparseTopLevelAwait(')
 elif 'yield\n/' in text:
  add('yield newline/regexp token boundary',P,"if(operator === 'YieldKeyword'",G,'func (p *Parser) parseYieldExpression(')
 elif '\nis' in text:
  add('type predicate lookahead ignores the preceding line break',P,'returnType(): number {',G,'func (p *Parser) parseTypeOrTypePredicate(')
 elif 'asserts' in text:
  add('return-type predicate recognition on recovery input',P,'returnType(): number {',G,'func (p *Parser) parseReturnType(')
 elif '?' in text and '=>' in text:
  add('conditional branch arrow return-type disambiguation',P,"if(operator === 'QuestionToken') {",G,'func (p *Parser) parseConditionalExpressionRest(')
 elif '<' in text and ('no-bitwise' in name or 'no-new' in name):
  add('generic call/new expression versus binary expression parsing',P,'if(this.kind() === \'LessThanToken\' && typeArgumentsAhead',G,'func (p *Parser) parseLeftHandSideExpressionOrHigher(')
 elif name=='lint/syntax-fixes' and 'rejected prefer-template' in r['go'].get('stdout',''):
  add('fix reparse continues without a diagnostic guard','stage1/cohere/lint/lint.ts','next.run();','cohere/internal/edit/engine.go','parses, reason, rewrittenTree := parsesWithTree')
 return cause,port,go
seen=set();counts=collections.Counter();examples={};rows=0
with gzip.open(out/'divergences-minimized.jsonl.gz','wt') as target:
 for f in sorted(pathlib.Path(a.reductions).glob('reductions-*.gz')):
  for line in gzip.open(f,'rt'):
   r=json.loads(line);key=(r['index'],r['rule'])
   if key in seen:raise SystemExit('duplicate reduction '+str(key))
   seen.add(key)
   if not r['verified_single_rune_deletions']:raise SystemExit('unverified reduction')
   # Retain exact rule emission sites in addition to the descriptor entry.
   if not r['rule'].startswith(('format/','lint/')):
    for loc in list(r['port_sites']):
     path=loc.rsplit(':',1)[0]
     for number,line in enumerate((root/path).read_text().splitlines(),1):
      if any(token in line for token in ['.reportNode(','.reportRange(','.reportNodeWith','.report(']):
       r['port_sites'].append(path+':'+str(number))
   for field in ['port_sites','go_sites']:
    for loc in r[field]:
     path,num=loc.rsplit(':',1)
     if not num.isdigit() or not 1<=int(num)<=len((root/path).read_text().splitlines()):raise SystemExit('invalid implementation site '+loc)
   cause,ports,gos=causal(r);r['cause']=cause;r['causal_port_sites']=ports;r['causal_go_sites']=gos
   key=(r['rule'],r['signature'],r['deletion_minimal_input'],cause);counts[key]+=1;examples[key]=r
   target.write(json.dumps(r,ensure_ascii=True,separators=(',',':'))+'\n');rows+=1
witnesses=[]
for key,n in counts.most_common():
 r=examples[key];witnesses.append(dict(rule=key[0],signature=key[1],input=key[2],cause=key[3],cells=n,example_index=r['index'],example_file=r['file'],port_sites=r['causal_port_sites'] or r['port_sites'],go_sites=r['causal_go_sites'] or r['go_sites']))
(out/'witnesses.json').write_text(json.dumps(dict(cells=rows,distinct_witnesses=len(witnesses),minimality='source-subsequence, same path, Go parse-valid for script files, single-rune deletion fixed point; not globally shortest',witnesses=witnesses),indent=2)+'\n')
print('attributed',rows,'divergences;',len(witnesses),'distinct witnesses')
