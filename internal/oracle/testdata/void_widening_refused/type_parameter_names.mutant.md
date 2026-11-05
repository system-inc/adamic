# Mutant: a type-argument step printed by position (#v97ejyb)

Edit: internal/lower/void_widening.go, in voidWidened's reference arm, put the position back:
`found.path = "<" + strconv.Itoa(index+1) + ">" + found.path`.

Red: TestVoidWideningIsRefused (internal/oracle) in contained_promise.a, container_interface.a,
container_promise.a, iterable_return.a and iterator_return.a, each with "diagnostic: got ... <1> ..."
where its .refused file names the type parameter (<T>, <Result>, <TReturn>).
