"""Classify exactly the pinned 31 open adapted-only casts, retaining original evidence."""
import collections,json,pathlib,subprocess,sys
root=pathlib.Path(__file__).resolve().parent
rows=json.loads((root/'evidence/ledger-additions.json').read_text())
rows=[r for r in rows if r['kind']=='as_cast' and r['disposition']=='open']
assert len(rows)==31
for r in rows:
 f,t=r['file'],r['text']
 if f.endswith('moduleNameResolver.ts'):
  owner='32-indexed-reads-program';why='Admit explicitly undefined directoryExists at the local helper call without editing the utilities owner.';choice='truthful internal declaration'
  typ='(directoryName: string, host: { directoryExists?: ((directoryName: string) => boolean) | undefined; }) => boolean'
  proposal='Change directoryProbablyExists owning host parameter to { directoryExists?: ((directoryName: string) => boolean) | undefined; }; remove this call view and its erased assertion parentheses. Its unchanged presence test handles explicit undefined. Cross-owner type-only change.'
  proof='Upstream assignability is true, but exactOptionalPropertyTypes changes callable parameter variance. The ledger records Adamic refusing; this is not a proven Adamic upcast.'
 elif f.endswith(('emitter.ts','parser.ts','semver.ts')):
  owner='45-regex-captures';why='Capture-free split separator produces a dense array of strings; avoid optional capture elements.';choice='truthful builtin specialization'
  typ='string[]';proposal='Remove this redundant cast under the upstream String.split contract. Under Adamic the capture-free separator needs a builtin proof returning string[] at this site; do not globally narrow regex split, which can insert undefined captures.'
  proof='Upstream resolved source and target are string[] and assignable. Adamic probe refused. Semantic density follows only for the reviewed capture-free separator; native builtin specialization remains required.'
 elif f.endswith('moduleSpecifiers.ts'):
  owner='32-indexed-reads-program';why='Expose the first preference-array element as defined to the nonempty consumers.';choice='truthful internal declaration'
  typ='[ModuleSpecifierEnding, ...ModuleSpecifierEnding[]]';proposal='Give ModuleSpecifierPreferences.getAllowedEndingsInPreferredOrder(syntaxImpliedNodeFormat?: ResolutionMode) the return type [ModuleSpecifierEnding, ...ModuleSpecifierEnding[]] and remove both consumer casts. All normal private factory returns contain one to four elements; default assertNever does not return. Preserve mutable tuple because the fresh result is mutable.'
  proof='Array-to-nonempty-tuple is a refinement, not an upcast. The local factory enumerates nonempty returns; no public override proof is claimed outside the reviewed private callers.'
 elif f.endswith('checker.ts'):
  owner='41-explicit-any-remaining';why='Rule 30 changes the allocator to an incomplete AllocatedSignature and asserts completion after field assignments.';choice='construction proof or checked view'
  typ='AllocatedSignature & Pick<Signature, "parameters" | "minArgumentCount">';proposal='The constructor must remain AllocatedSignature. After parameters and minArgumentCount are assigned, the completed value has AllocatedSignature & Pick<Signature, "parameters" | "minArgumentCount"> (a Signature). A local annotation claiming Signature at allocation would lie. Require definite field initialization proof or a checked completion view; do not change constructor return to Signature.'
  proof='The source optionalizes required fields; it is not assignable to Signature. The observed assignments establish completion semantically, but a source-only annotation does not carry a checker proof.'
 elif f.endswith('commandLineParser.ts'):
  owner='43-any-returns';why='Concrete recursive recovery return replaced any; the config converter asserts its root conversion is an object.';choice='root correlation proof or checked view'
  typ='AdamicJsonRecoveryObject';proposal='An object-root overload can return AdamicJsonRecoveryObject when rootExpression: ObjectLiteralExpression | undefined and returnValue: true. Keep the general return AdamicJsonRecoveryValue | undefined. firstObject is an object; rootExpression is only Expression | undefined after the kind test, so the second site also needs a checker-recognized root proof or checked result view. Do not globally give convertToJson an object return.'
  proof='Primitive and array roots legitimately return primitive/array recovery values. These two callers constrain roots to object or absent, and returnValue is true; the generic converter signature alone cannot prove that relationship.'
 elif f.endswith('factory/nodeFactory.ts') and 'localSymbol' in t:
  owner='60-temporary-factory-local-symbol';why='Permit writing undefined to an optional generic field without claiming presence in the returned T.';choice='generic write and construction proof'
  typ='Omit<Mutable<T>, "localSymbol"> & { localSymbol: undefined }';proposal='After this write the truthful field is localSymbol: undefined; the local post-write shape is Omit<Mutable<T>, "localSymbol"> & { localSymbol: undefined }. It cannot be returned as every T: a T may require a Symbol. An undefined-admitting base constraint alone does not exclude that subtype. Check writes against the original specialization and prove initialization at callers, or change the private factory result protocol; a checked cast alone cannot make the returned T true.'
  proof='The receiver view loses T-specific field restrictions. The wide view can write a value that a narrower T cannot hold; this is not an upcast proof.'
 elif f.endswith('factory/nodeFactory.ts'):
  owner='61-temporary-factory-type-expression';why='Permit the optional typeExpression input to be written through the generic factory receiver.';choice='generic write and construction proof'
  typ='Omit<Mutable<T>, "typeExpression"> & { typeExpression: JSDocTypeExpression | undefined }';proposal='Truthful post-write shape: Omit<Mutable<T>, "typeExpression"> & { typeExpression: JSDocTypeExpression | undefined }. Widening the base constraint to admit undefined does not prove all T accept it: a T can require a particular expression. Require checked writes for the actual specialization plus a construction/caller proof, or change the private factory result protocol. Do not claim Mutable<T> without that proof.'
  proof='Writing through the erased generic restriction needs more than checking property presence; the original specialization governs what values may be stored.'
 elif f.endswith('tracing.ts'):
  owner='62-temporary-tracing-write' if 'writeSync(' in t else '63-temporary-tracing-legend';why='Allow JSON.stringify result string | undefined at a Node fs boundary that throws for undefined.';choice='truthful throwing host binding'
  typ=r['target_type'];proposal='The target callable signature is truthful only for a binding that accepts string | undefined and preserves Node ERR_INVALID_ARG_TYPE/TypeError for undefined. Add that typed throwing boundary contract; do not cast a string-only callable to one accepting undefined or replace undefined with a string. A callable checked cast cannot prove external function behavior. Native Node binding/exception support remains required.'
  proof='Contravariant parameter widening is not a safe upcast; Node accepts the invocation and throws for undefined data. The source adaptation supplies no native implementation.'
 else:
  assert f.endswith('watchUtilities.ts')
  owner='32-indexed-reads-program';why='Preserve the key-specific watcher callback/flags/arguments protocol across union-typed .call receivers.';choice='correlated generic protocol proof'
  typ='WatchFactory<X, Y>[K]' if not t.startswith('cb as ') else 'Parameters<WatchFactory<X, Y>[K]>[1]'
  proposal='Carry K extends keyof WatchFactory<X, Y> through the selected receiver WatchFactory<X, Y>[K] and its argument tuple Parameters<WatchFactory<X, Y>[K]>. The callback is Parameters<WatchFactory<X, Y>[K]>[1]; forwarded eventArgs are Parameters<Parameters<WatchFactory<X, Y>[K]>[1]>. Instantiate separately for watchFile and watchDirectory. These are exact correlated types, not a single function that accepts arbitrary callback/flag combinations. A typed local alone does not discharge the generic .call body; requires correlation proof or code specialization.'
  proof='A union of key-specific callables cannot safely accept every combination of unioned parameters. Runtime callable/signature checking cannot recover the missing key correlation by itself.'
 r.update(checked_cast_route=('No cast needed after the verified truthful declaration proposal.' if choice=='truthful internal declaration' else 'Capture-free builtin proof can eliminate the cast; otherwise an array element checked view is required under the sound split library.' if choice=='truthful builtin specialization' else 'Checked completion/refinement view or an equivalent static construction/root proof; no unchecked assertion.' if choice in ['construction proof or checked view','root correlation proof or checked view'] else 'A checked cast alone is insufficient: the original specialization, key protocol or host implementation must also be proved.'), adaptation='stage3/adapt/'+owner+'/',why_added=why,decision=choice,truthful_type=typ,proposal=proposal,proof=proof,proven_adamic_upcast=False)
counts=dict(collections.Counter(r['decision'] for r in rows))
out=dict(ledger='855bcfaacb8919a0d61ec34eda785fce9fde8ba2',classification='ea1b2359',population='open adapted-only as_cast entries',count=31,counts=counts,rows=rows)
(root/'contracts.json').write_text(json.dumps(out,indent=2)+'\n')
print('PASS: 31 classified rows',counts)
