const ts = require('typescript');
function proven(t, checker, seen = new Set()) {
 if (seen.has(t)) return false;
 seen = new Set(seen).add(t);
 if (t.flags & (ts.TypeFlags.Any|ts.TypeFlags.Unknown|ts.TypeFlags.Null|ts.TypeFlags.Undefined|ts.TypeFlags.Void)) return false;
 if (t.isUnion()) return t.types.every(p => proven(p,checker,seen));
 if (t.flags & ts.TypeFlags.TypeParameter) {
  const constraint = checker.getBaseConstraintOfType(t);
  return !!constraint && proven(constraint,checker,seen);
 }
 if (t.isIntersection()) return t.types.some(p => proven(p,checker,seen));
 return true;
}
module.exports = proven;
