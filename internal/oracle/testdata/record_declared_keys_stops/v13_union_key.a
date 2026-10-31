interface Shape { x: number; describe(): string }
class Point { x = 1; describe(): string { return "point"; } }
class Other { x = 2; label = "other"; describe(): string { return this.label; } }
type Places = { names: Point; [k: string]: Point | Shape };
function put(places: Places, key: "names" | "other"): void { places[key] = new Other(); }
const places: Places = { names: new Point() };
put(places, "names");
console.log(places.names.describe().toUpperCase());
