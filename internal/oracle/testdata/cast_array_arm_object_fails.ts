// A lying arm: the value is a CircularBuildOrder, so Array.isArray says false where the cast claims the array
// side, and the cast fails with its own message. Node runs on without a check and prints object.
interface CircularBuildOrder { readonly buildOrder: readonly string[]; readonly circularDiagnostics: readonly string[] }
function cast(order: readonly string[] | CircularBuildOrder): readonly string[] { return order as readonly string[]; }
console.log('casting');
console.log(typeof cast({ buildOrder: ['a'], circularDiagnostics: [] }));
