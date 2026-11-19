// A relation no lowering meets: admit is generic and nothing calls it, so its body is never instantiated and
// never lowered, and the view rule 3300 recorded at source is never taken. checkRelationViewsTaken refuses
// it in that rule's words once lowering is done. It's the post-pass's own witness (mutant (a) in
// relation_views.mutants.md), now that shorthand refuses its own road.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function admit<Tag>(tag: Tag, source: Source): Target {
    const target: Target = source;
    return target;
}
const plain: Source = { kind: 1 };
console.log(plain.kind);
