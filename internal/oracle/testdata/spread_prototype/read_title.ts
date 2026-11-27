interface Titled {
	readonly label: string;
	readonly title: string;
}
class Page {
	readonly label = "p";
	get title(): string {
		return "home";
	}
}
const built: Titled = new Page();
const copy: Titled = { ...built };
console.log(`${built.title} ${copy.title}`);
console.log(typeof copy.title + " " + String(copy.title === undefined));
