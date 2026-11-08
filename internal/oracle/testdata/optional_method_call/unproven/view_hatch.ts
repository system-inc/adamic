/*
 * Check 49's road (#tvq1eqm): a checked cast lets in an object whose tag says Probe and whose enabled
 * is a number, where the type says a function or undefined. ?. passes it, since it's neither undefined
 * nor null, and Node calls it: "viewed.enabled is not a function". Adamic reads the method through the
 * view's checked read, which has no contract for an optional callable yet, so the call stops as not
 * yet, and the check has nothing to fire on.
 */
interface Probe {
    readonly kind: 'Probe';
    enabled?(): boolean;
}

function probe(raw: { readonly kind: string }): string {
    const viewed = raw as Probe;
    return String(viewed.enabled?.());
}

const real: Probe = { kind: 'Probe', enabled: (): boolean => true };
const hatch = { kind: ['Pro', 'be'].join(''), enabled: 7 };
console.log(probe(real));
console.log(probe(hatch));
