package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestSuppressionDirectivesAreRefused(test *testing.T) {
	test.Parallel()
	for _, directive := range []string{"@ts-ignore", "@ts-expect-error"} {
		for _, probe := range []struct{ name, prefix, body, location string }{
			{"number.a", "", "const n: number = 'wrong';\nconsole.log(`${n + 1}`);\n", "main.a:1:1:"},
			{"boolean.a", "", "const b: boolean = 2;\nconsole.log(`${b}`);\n", "main.a:1:1:"},
			{
				"call.a",
				"function value(): number { return 2; }\n",
				"const b: boolean = value();\nconsole.log(`${b}`);\n",
				"main.a:2:1:",
			},
		} {
			for _, comment := range []string{
				"// " + directive + "\n",
				"/* " + directive + " */\n",
				"/** " + directive + " */\n",
				"/** prose\n * " + directive + " */\n",
			} {
				test.Run(directive+"/"+probe.name+"/"+comment, func(test *testing.T) {
					test.Parallel()
					_, err := lowerSource(test, probe.prefix+comment+probe.body)
					var refused *Refused
					if !errors.As(err, &refused) {
						test.Fatalf("got %v, want directive refusal", err)
					}
					location := probe.location
					if strings.Contains(comment, "prose") {
						if probe.prefix == "" {
							location = "main.a:2:1:"
						} else {
							location = "main.a:3:1:"
						}
					}
					for _, text := range []string{location, directive, "remove it and fix the type error"} {
						if !strings.Contains(refused.Error(), text) {
							test.Errorf("got %v, want %q", err, text)
						}
					}
				})
			}
		}
	}
}

func TestSuppressionDirectiveSoundNeighbors(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"// This mentions ts-ignore in prose.\nconsole.log('@ts-ignore');\n",
		"// Explain @ts-ignore here.\nconsole.log('@ts-expect-error');\n",
		"/* Explain @ts-ignore here. */\nconsole.log('ok');\n",
		"/** Explain @ts-expect-error here. */\nconsole.log('ok');\n",
		"class Box { n: number = 2; }\nconsole.log(`${new Box().n + 1}`);\n",
		"let n: number;\nn = 2;\nconsole.log(`${n + 1}`);\n",
	} {
		_, err := lowerSource(test, source)
		if err != nil {
			test.Errorf("sound neighbor: %v", err)
		}
	}
}
