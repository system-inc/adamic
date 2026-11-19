// rulings.md:72's second slice (#0b9xqhh), a generic call's argument: generic_deferred.ts's relation in a body
// a call instantiates. admit<number> is lowered, but rule 3300 judges the body as the checker types it, where
// show's resolved slot at the call is Box<Tag>, admit's own type parameter still in it. So it refuses in its
// own words rather than view the argument as Box<number> under the type mapper (Lower's hook 2). Its values
// are honest: without that condition (mutant (k)) it lowers and prints what Node prints.
interface Plain<Tag> { readonly tag: Tag; readonly kind: number; }
interface Box<Tag> { readonly tag: Tag; readonly kind: number; readonly label?: string; }

function show<Tag>(box: Box<Tag>): string {
    return box.label ?? "absent";
}
function admit<Tag>(plain: Plain<Tag>): string {
    return show(plain);
}
console.log(admit({ tag: 1, kind: 1 }));
