// rulings.md:72's second slice (#0b9xqhh), a generic call's argument: the relation sits in a generic body,
// where show's resolved signature at the call is show<Tag>, so the argument's slot is Box<Tag>, the body's own
// type parameter still in it. Rule 3300 refuses it there in its own words rather than view it as one
// instantiation's type. Nothing calls admit, so the post-pass would refuse it in the same words too.
interface Plain<Tag> { readonly tag: Tag; readonly kind: number; }
interface Box<Tag> { readonly tag: Tag; readonly kind: number; readonly label?: string; }

function show<Tag>(box: Box<Tag>): string {
    return box.label ?? "absent";
}
function admit<Tag>(plain: Plain<Tag>): string {
    return show(plain);
}
console.log(show({ tag: 1, kind: 1 }));
