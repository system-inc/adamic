# Assertion and group syntax helpers

Each claimed helper is in one `.a` file. Inputs representing Go strings are readonly arrays of bytes, each an integer from zero through 255. `errNothingToRepeat` requires the separately owned exact Go-equivalent quantifier-width callback, returning a byte count within the input. `checkGroupQuantifier` accepts Go group kinds plain=0, lookahead=1, lookbehind=2 and the Unicode flag. `checkGroupConstruct` accepts arbitrary raw bytes; its error precision follows Go UTF-8 rune decoding, including malformed bytes.

Errors are exposed as readonly raw byte arrays: empty means nil/success for the two checks, and every nonempty result represents ErrUnsupportedSyntax with the rendered wrapped-error bytes. `errNothingToRepeat` always returns that error representation. Rule and regexp callers must preserve that error category when wrapping it. These are isolated helper ports, not a regexp compiler or matcher. See REPORT.md for tested inputs, consumers, mutants, dependencies and limits.

Run with the setup environment sourced:

```
python3 stage1/cohere/lint/helpers/slot04_wave13/testdata/capture.py > /tmp/slot04-wave13-capture.log 2>&1
go test -count=1 -v ./stage1/cohere/lint/helpers/slot04_wave13 > /tmp/slot04-wave13-tests.log 2>&1
```
