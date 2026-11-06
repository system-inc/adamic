// The narrow view in .ts, where a member cast is checked at run time: the namespace cast still names a
// table the view's value type doesn't hold, so it stays as on main.
import * as ts from './barrel.a';
const view = (ts as Readonly<Record<"SyntaxKind", Readonly<Record<string, number | "Unknown">>>>).SyntaxKind;
console.log(Object.keys(view).join(","));
