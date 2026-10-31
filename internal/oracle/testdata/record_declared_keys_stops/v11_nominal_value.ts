interface Shape { x: number; describe(): string }
class Point { x = 1; describe(): string { return "point"; } }
class Other { x = 2; label = "other"; describe(): string { return this.label; } }
type Places = { names: Point; [k: string]: Point | Shape };
const places: Places = { names: new Point() };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
places[key] = new Other();
console.log(places.names.describe().toUpperCase());
