// A presence test of an optional field holding a function, in a program where a class owns it as a
// getter: finishAccessors makes the read a dispatcher's call, which calls the getter on the owner and on
// a subclass that inherits it, and reads a plain literal's slot as absent or present (#p2n17c4 shape 1,
// on #4gpbn7w's dispatcher). Node finds the getter up the prototype chain; native's adamic_accessor_find
// finds it up the base chain.
interface H {
    readonly n: number;
    readonly m?: () => number;
}
class G {
    readonly n: number = 1;
    get m(): () => number {
        return (): number => 7;
    }
}
class S extends G {
    readonly k: number = 4;
}
function probe(h: H): string {
    return h.m ? 'has' : 'none';
}
console.log(probe(new S()));
console.log(probe(new G()));
console.log(probe({ n: 2 }));
console.log(probe({ n: 3, m: (): number => 1 }));
