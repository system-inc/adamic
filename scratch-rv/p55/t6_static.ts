interface Maker {
    make(): string;
}
class Factory {
    static make(): string {
        return "made";
    }
}
const maker: Maker = Factory;
console.log(maker.make ? maker.make() : "no make");
