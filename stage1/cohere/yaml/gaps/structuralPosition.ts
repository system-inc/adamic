class Span {
    start = 0;
    end = 1;
}
class Properties {
    start = 0;
    end = 1;
    errors: Span[] = [];
}
export function append(properties: Properties, other: Properties): void {
    for(const span of other.errors) properties.errors.push(span);
}
const properties = new Properties();
const other = new Properties();
other.errors.push(new Span());
append(properties, other);
console.log(String(properties.errors.length));
