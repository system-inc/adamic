interface Handler {
	handle(x: string): string;
}
const h: Handler = {
	handle(x: never): string {
		return "literal " + typeof x;
	},
};
console.log(h.handle("hi"));
