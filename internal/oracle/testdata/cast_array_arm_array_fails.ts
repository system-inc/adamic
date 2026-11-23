// A lying arm: the value is an array, so Array.isArray says true where the cast claims the other side, and the
// cast fails with its own message. Node runs on without a check and prints object.
interface CircularBuildOrder { readonly buildOrder: readonly string[]; readonly circularDiagnostics: readonly string[] }
function cast(order: readonly string[] | CircularBuildOrder): CircularBuildOrder { return order as CircularBuildOrder; }
console.log('casting');
console.log(typeof cast(['a', 'b']));
