package load

import "testing"

func TestRegExpOmittedPattern(t *testing.T) {
	_, err := Load(writeProgram(t, [2]string{"main.a", "const first = new RegExp(); const second = RegExp(); first.test(''); second.test('a'); new RegExp(undefined, 'gy').test('a'); RegExp(undefined).test('a');"}))
	if err != nil {
		t.Fatal(err)
	}
}

func TestRegExpCaptureTypes(t *testing.T) {
	t.Parallel()
	for _, expression := range []string{"'ab'.split(/(a)|(z)/).map(value => value.toUpperCase())", "'ab'.match(/(a)|(z)/)?.map(value => value.toUpperCase())", "/(a)|(z)/.exec('ab')?.map(value => value.toUpperCase())", "/(?<x>a)/.exec('a')?.groups?.x.toUpperCase()"} {
		t.Run(expression, func(t *testing.T) {
			t.Parallel()
			t.Log(expression)
			diagnostics := checkErrors(t, writeProgram(t, [2]string{"main.a", expression + ";"}))
			if len(diagnostics) == 0 {
				t.Fatal("undefined capture accepted as a string")
			}
		})
	}
}
