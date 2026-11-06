function draw(lanes: number): string {
	const enum Box { lr = "─", ud = "│", udlr = "╫" }
	const enum Link { None = 0, Up = 1 << 0, Down = 1 << 1, Left = 1 << 2, Right = 1 << 3, UpDown = Up | Down, LeftRight = Left | Right, All = UpDown | LeftRight }
	function character(link: Link): string {
		switch (link) {
			case Link.UpDown: return Box.ud;
			case Link.LeftRight: return Box.lr;
			case Link.All: return Box.udlr;
		}
		return " ";
	}
	const links: Link[] = [];
	for (let lane = 0; lane < lanes; lane++) {
		let link: Link = Link.None;
		if (lane % 2 === 0) link |= Link.UpDown;
		else link |= Link.LeftRight;
		if (lane === 2) link |= Link.All;
		links.push(link);
	}
	return links.map((link) => character(link)).join("");
}
console.log(draw(4));
