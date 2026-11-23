interface Greeter {
    greet(): string;
}
class Person {
    greet(): string {
        return "hi";
    }
}
class Field {
    greet = (): string => "field";
}
const people: Greeter[] = [new Person(), new Field(), { greet: (): string => "literal" }];
for (const person of people) {
    if (person.greet) {
        console.log(person.greet());
    }
    console.log(String(!person.greet));
}
