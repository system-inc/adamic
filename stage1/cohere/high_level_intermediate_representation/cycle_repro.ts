class HIRFunction { readonly functions: HIRFunction[] = []; }
function append(parent: HIRFunction, child: HIRFunction): void { parent.functions.push(child); }
append(new HIRFunction(), new HIRFunction());
