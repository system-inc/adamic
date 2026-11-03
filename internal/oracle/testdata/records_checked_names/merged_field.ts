// A field an interface merged with a class promises and the class doesn't initialize stays checked:
// its receiver is the merged class, never plain, and the read stops on the missing field where Node
// reads undefined.
interface Box { readonly label: string }
class Box {
	size = 1;
}
function show(box: Box): string {
	return box.label;
}
console.log(String(new Box().size));
console.log(show(new Box()));
