interface Raw { name: string; score?: number }
interface Scored { name: string; score: number }
function isScored(raw: Raw): raw is Scored {
	return raw.name !== "unscored";
}
const raws: Raw[] = [{ name: "a", score: 1 }, { name: "unscored" }, { name: "b", score: 2 }];
const scored = raws.filter(isScored);
let total = 0;
for (const item of scored) {
	total += item.score;
}
console.log(String(scored.length) + " " + String(total));
