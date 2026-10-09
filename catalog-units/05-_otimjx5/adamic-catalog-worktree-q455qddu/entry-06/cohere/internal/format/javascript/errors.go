package javascript

// common/errors.js, the one error the JavaScript printer throws and catches itself.

// argExpansionBailout is upstream's ArgExpansionBailout. Upstream throws it from deep inside printing
// an argument with expandFirstArg or expandLastArg (print/function-parameters.js and
// print/arrow-function.js) and catches it in printCallArguments (print/call-arguments.js), which then
// prints every argument broken out instead. Go has no exceptions, so the throw is
// `panic(argExpansionBailout{})` and the catch is a deferred recover in printCallArguments that
// re-panics anything else, as upstream rethrows.
type argExpansionBailout struct{}

func (argExpansionBailout) Error() string { return "ArgExpansionBailout" }
