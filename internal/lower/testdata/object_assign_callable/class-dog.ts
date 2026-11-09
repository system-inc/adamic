class Dog {
 name = "rex";
 bark() { return "woof"; }
}
const copy: Dog = Object.assign({}, new Dog());
console.log(copy.name);
console.log(copy.bark());
