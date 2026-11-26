class Scale {
	readonly label = "t";
	weigh(): number {
		return 20;
	}
}
function read(scale: Scale): number {
	return scale.weigh();
}
const scales: Scale[] = [new Scale(), new Scale()];
let total = 0;
for (const scale of scales) {
	total += read(scale);
}
console.log(String(total));
