// #6c4swvt's record-to-fixed direction through a parameter: a function over Options seen as one over a record.
// A view drops or adds no contract through a parameter (parameterStorage), so it stops at the view, as the
// drop's parameter roads do.
type Value = string | number | boolean | undefined;
interface Options {
    strict?: boolean;
    [option: string]: Value;
}
type Bag = Record<string, Value>;
const read: (bag: Bag) => string = (options: Options) => String(options.strict);
const bag: Bag = {};
bag["strict"] = "yes";
console.log(read(bag));
