package load

import "testing"

func TestRegExpCaptureTypes(test *testing.T) {
	test.Parallel()
	for _, expression := range []string{
		"'ab'.split(/(a)|(z)/).map(value => value.toUpperCase())",
		"'ab'.match(/(a)|(z)/)?.map(value => value.toUpperCase())",
		"/(a)|(z)/.exec('ab')?.map(value => value.toUpperCase())",
		"/(?<x>a)/.exec('a')?.groups?.x.toUpperCase()",
	} {
		test.Run(expression, func(test *testing.T) {
			test.Parallel()
			test.Log(expression)
			diagnostics := checkErrors(test, writeProgram(test, [2]string{"main.a", expression + ";"}))
			if len(diagnostics) == 0 {
				test.Fatal("undefined capture accepted as a string")
			}
		})
	}
}
