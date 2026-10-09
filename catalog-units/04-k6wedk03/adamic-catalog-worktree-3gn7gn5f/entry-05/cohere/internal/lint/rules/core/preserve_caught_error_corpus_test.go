package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * ESLint 10.8.1's preserve-caught-error, run through the installed rule under @typescript-eslint/parser
 * over its whole test corpus and over the edges this port's own reading raised (#jjfa7qb), with every
 * finding's id, span and suggestion written here as ESLint answered.
 *
 * The corpus rows are ESLint's tests/lib/rules/preserve-caught-error.js as the registry's corpus file
 * holds them, trimmed. The edge rows were written for this port: parentheses at every position,
 * spreads before and after the options slot, an arrow and a static block inside the catch, every
 * built-in constructor the old oxc port left out, a shorthand cause under a catch parameter spelled
 * cause, the two-cause and quoted-key shapes the suggestion has to get right, and errorClassNames on
 * a member callee, an optional call and a position past the arguments given.
 *
 * One edge is left out because the rule departs from ESLint there on purpose: type arguments with no
 * arguments, where ESLint's suggestion writes inside the type arguments. Its own test says what this
 * writes instead.
 *
 * The suggestion column is the whole file after applying the finding's one suggestion, or empty where
 * ESLint offers none.
 */

// preserveCaughtErrorFinding is one ESLint finding: its id, the text it spans, and the file after its
// suggestion is applied.
type preserveCaughtErrorFinding struct {
	id         string
	text       string
	suggestion string
}

func TestPreserveCaughtErrorAgreesWithESLint(t *testing.T) {
	t.Parallel()

	cases := []struct {
		origin   string
		options  string
		code     string
		findings []preserveCaughtErrorFinding
	}{
		{"corpus", "", "try {\n        throw new Error(\"Original error\");\n    } catch (error) {\n        throw new Error(\"Failed to perform error prone operations\", { cause: error });\n    }\n", nil},
		{"corpus", "", "try {\n\t\tdoSomething();\n\t} catch (error) {\n\t\tthrow new Error(\"Something failed\", { 'cause': error });\n\t}\n", nil},
		{"corpus", "", "try {\n\t\tdoSomething();\n\t} catch (error) {\n\t\tthrow new Error(\"Something failed\", { \"cause\": error });\n\t}\n", nil},
		{"corpus", "", "try {\n\t\tdoSomething();\n\t} catch (error) {\n\t\tthrow new Error(\"Something failed\", { ['cause']: error });\n\t}\n", nil},
		{"corpus", "", "try {\n\t\tdoSomething();\n\t} catch (error) {\n\t\tthrow new Error(\"Something failed\", { [\"cause\"]: error });\n\t}\n", nil},
		{"corpus", "", "try {\n\t\tdoSomething();\n\t} catch (error) {\n\t\tthrow new Error(\"Something failed\", { [`cause`]: error });\n\t}\n", nil},
		{"corpus", "", "try {\n        doSomething();\n    } catch (e) {\n        console.error(e);\n    }\n", nil},
		{"corpus", "", "try {\n        doSomething();\n    } catch (err) {\n        throw new Error(\"Failed\", { cause: err, extra: 42 });\n    }\n", nil},
		{"corpus", "", "try {\n        doSomething();\n    } catch (error) {\n        switch (error.code) {\n            case \"A\":\n                throw new Error(\"Type A\", { cause: error });\n            case \"B\":\n                throw new Error(\"Type B\", { cause: error });\n            default:\n                throw new Error(\"Other\", { cause: error });\n        }\n    }\n", nil},
		{"corpus", "", "try {\n\t\t// ...\n\t} catch (err) {\n\t\tconst opts = { cause: err }\n\t\tthrow new Error(\"msg\", { ...opts });\n\t}\n", nil},
		{"corpus", "", "try {\n\t} catch (error) {\n\t\tfoo = {\n\t\t\tbar() {\n\t\t\t\tthrow new Error();\n\t\t\t}\n\t\t};\n\t}\n", nil},
		{"corpus", "", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tconst args = [];\n\t\t\t\tthrow new Error(...args);\n\t\t}\n", nil},
		{"corpus", "", "import { Error } from \"./my-custom-error.js\";\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow Error(\"Failed to perform error prone operations\");\n\t\t\t}\n", nil},
		{"corpus", "{\"requireCatchParameter\": false}", "try {\n\t\tdoSomething();\n\t} catch {\n\t\tthrow new Error(\"Something went wrong\");\n\t}\n", nil},
		{"corpus", "", "try {\n\t\t\tdoSomething();\n\t\t} catch (error) {\n\t\t\tthrow new Error(\"Something failed\", { cause: anotherError, cause: error });\n\t\t}\n", nil},
		{"corpus", "", "try {\n\t\t\tdoSomething();\n\t\t} catch (error) {\n\t\t\tthrow new Error(\"Something failed\", { \"cause\": anotherError, \"cause\": error });\n\t\t}\n", nil},
		{"corpus", "", "try {\n\t\t\tdoSomething();\n\t\t} catch (error) {\n\t\t\tthrow new Error(\"Something failed\", { cause: anotherError, \"cause\": error });\n\t\t}\n", nil},
		{"corpus", "{\"errorClassNames\": [\"AppError\"]}", "const errors = { AppError: class AppError extends Error {} };\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new errors.AppError(\"Something failed\", { cause: error });\n\t\t\t}\n", nil},
		{"corpus", "{\"errorClassNames\": [\"APIError\"]}", "const lib = { APIError: class APIError extends Error {} };\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new lib.APIError(\"Something failed\", { cause: error });\n\t\t\t}\n", nil},
		{"corpus", "{\"errorClassNames\": [\"CustomApplicationError\"]}", "class CustomApplicationError extends Error {}\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new CustomApplicationError(\"Cause not provided\", { cause: err });\n\t\t\t}\n", nil},
		{"corpus", "{\"errorClassNames\": [\"APIError\"]}", "class APIError extends Error {}\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new APIError(\"API failed\", { cause: error });\n\t\t\t}\n", nil},
		{"corpus", "{\"errorClassNames\": []}", "class CustomError extends Error {}\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new CustomError(\"No cause\");\n\t\t\t}\n", nil},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"MyError\", \"argumentPosition\": 3}]}", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new MyError(\"failed\", someData, { cause: err });\n\t\t\t}\n", nil},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"SimpleError\", \"argumentPosition\": 1}]}", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new SimpleError({ cause: err });\n\t\t\t}\n", nil},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"AppError\", \"argumentPosition\": 2}, {\"name\": \"AppError\", \"argumentPosition\": 3}]}", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new AppError(\"Message\", {}, { cause: err });\n\t\t\t}\n", nil},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"Error\", \"argumentPosition\": 3}]}", "try {\n\t\t\t    doSomething();\n\t\t\t} catch (err) {\n\t\t\t    throw new context.Error(\"failed\", someData, { cause: err });\n\t\t\t}\n", nil},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"Error\", \"argumentPosition\": 3}]}", "try {\n\t\t\t    doSomething();\n\t\t\t} catch (err) {\n\t\t\t    throw new Error(\"failed\", { cause: err });\n\t\t\t}\n", nil},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"AggregateError\", \"argumentPosition\": 1}]}", "import { AggregateError } from \"some-module\";\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new AggregateError({ cause: err });\n\t\t\t}\n", nil},
		{"corpus", "", "try {\n            doSomething();\n        } catch (err) {\n            throw new Error(\"Something failed\");\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"Something failed\");", "try {\n            doSomething();\n        } catch (err) {\n            throw new Error(\"Something failed\", { cause: err });\n        }\n"}}},
		{"corpus", "", "try {\n            doSomething();\n        } catch (err) {\n            const unrelated = new Error(\"other\");\n            throw new Error(\"Something failed\", { cause: unrelated });\n        }\n", []preserveCaughtErrorFinding{{"incorrectCause", "unrelated", "try {\n            doSomething();\n        } catch (err) {\n            const unrelated = new Error(\"other\");\n            throw new Error(\"Something failed\", { cause: err });\n        }\n"}}},
		{"corpus", "", "try {\n            doSomething();\n        } catch (err) {\n            const unrelated = new Error(\"other\");\n            throw new Error(\"Something failed\", { \"cause\": unrelated });\n        }\n", []preserveCaughtErrorFinding{{"incorrectCause", "unrelated", "try {\n            doSomething();\n        } catch (err) {\n            const unrelated = new Error(\"other\");\n            throw new Error(\"Something failed\", { \"cause\": err });\n        }\n"}}},
		{"corpus", "", "try {\n            doSomething();\n        } catch (err) {\n            const e = err;\n            throw new Error(\"Failed\", { cause: e });\n        }\n", []preserveCaughtErrorFinding{{"incorrectCause", "e", "try {\n            doSomething();\n        } catch (err) {\n            const e = err;\n            throw new Error(\"Failed\", { cause: err });\n        }\n"}}},
		{"corpus", "", "try {\n            doSomething();\n        } catch (error) {\n            throw new Error(\"Failed\", { cause: error.message });\n        }\n", []preserveCaughtErrorFinding{{"incorrectCause", "error.message", "try {\n            doSomething();\n        } catch (error) {\n            throw new Error(\"Failed\", { cause: error });\n        }\n"}}},
		{"corpus", "", "try {\n            doSomething();\n        } catch (error) {\n            if (shouldThrow) {\n                while (true) {\n                    if (Math.random() > 0.5) {\n                        throw new Error(\"Failed without cause\");\n                    }\n                }\n            }\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"Failed without cause\");", "try {\n            doSomething();\n        } catch (error) {\n            if (shouldThrow) {\n                while (true) {\n                    if (Math.random() > 0.5) {\n                        throw new Error(\"Failed without cause\", { cause: error });\n                    }\n                }\n            }\n        }\n"}}},
		{"corpus", "", "try {\n            doSomething();\n        } catch (error) {\n            switch (error.code) {\n                case \"A\":\n                    throw new Error(\"Type A\");\n                case \"B\":\n                    throw new Error(\"Type B\", { cause: error });\n                default:\n                    throw new Error(\"Other\", { cause: error });\n            }\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"Type A\");", "try {\n            doSomething();\n        } catch (error) {\n            switch (error.code) {\n                case \"A\":\n                    throw new Error(\"Type A\", { cause: error });\n                case \"B\":\n                    throw new Error(\"Type B\", { cause: error });\n                default:\n                    throw new Error(\"Other\", { cause: error });\n            }\n        }\n"}}},
		{"corpus", "", "try {\n            doSomething();\n        } catch (error) {\n            throw new Error(`The certificate key \"${chalk.yellow(keyFile)}\" is invalid.\n${err.message}`);\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(`The certificate key \"${chalk.yellow(keyFile)}\" is invalid.\n${err.message}`);", "try {\n            doSomething();\n        } catch (error) {\n            throw new Error(`The certificate key \"${chalk.yellow(keyFile)}\" is invalid.\n${err.message}`, { cause: error });\n        }\n"}}},
		{"corpus", "", "try {\n            doSomething();\n        } catch (error) {\n            const errorMessage = \"Operation failed\";\n            throw new Error(errorMessage);\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(errorMessage);", "try {\n            doSomething();\n        } catch (error) {\n            const errorMessage = \"Operation failed\";\n            throw new Error(errorMessage, { cause: error });\n        }\n"}}},
		{"corpus", "", "try {\n            doSomething();\n        } catch (error) {\n            const errorMessage = \"Operation failed\";\n            throw new Error(errorMessage, { existingOption: true, complexOption: { moreOptions: {} } });\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(errorMessage, { existingOption: true, complexOption: { moreOptions: {} } });", "try {\n            doSomething();\n        } catch (error) {\n            const errorMessage = \"Operation failed\";\n            throw new Error(errorMessage, { existingOption: true, complexOption: { moreOptions: {} }, cause: error });\n        }\n"}}},
		{"corpus", "", "try {\n            doSomething();\n        } catch (err) {\n            if (err.code === \"A\") {\n                throw new Error(\"Type A\");\n            }\n            throw new TypeError(\"Fallback error\");\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"Type A\");", "try {\n            doSomething();\n        } catch (err) {\n            if (err.code === \"A\") {\n                throw new Error(\"Type A\", { cause: err });\n            }\n            throw new TypeError(\"Fallback error\");\n        }\n"}, {"missingCause", "throw new TypeError(\"Fallback error\");", "try {\n            doSomething();\n        } catch (err) {\n            if (err.code === \"A\") {\n                throw new Error(\"Type A\");\n            }\n            throw new TypeError(\"Fallback error\", { cause: err });\n        }\n"}}},
		{"corpus", "", "try {\n            doSomething();\n        } catch (err) {\n            throw Error(\"Something failed\");\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw Error(\"Something failed\");", "try {\n            doSomething();\n        } catch (err) {\n            throw Error(\"Something failed\", { cause: err });\n        }\n"}}},
		{"corpus", "", "try {\n        } catch (err) {\n            my_label:\n            throw new Error(\"Failed without cause\");\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"Failed without cause\");", "try {\n        } catch (err) {\n            my_label:\n            throw new Error(\"Failed without cause\", { cause: err });\n        }\n"}}},
		{"corpus", "", "try {\n        } catch (err) {\n            {\n                throw new Error(\"Something went wrong\");\n            }\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"Something went wrong\");", "try {\n        } catch (err) {\n            {\n                throw new Error(\"Something went wrong\", { cause: err });\n            }\n        }\n"}}},
		{"corpus", "", "try {\n        } catch (err) {\n            {\n                throw new Error();\n            }\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error();", "try {\n        } catch (err) {\n            {\n                throw new Error(\"\", { cause: err });\n            }\n        }\n"}}},
		{"corpus", "", "try {\n        } catch (err) {\n            {\n                throw new AggregateError([], \"Lorem ipsum\");\n            }\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new AggregateError([], \"Lorem ipsum\");", "try {\n        } catch (err) {\n            {\n                throw new AggregateError([], \"Lorem ipsum\", { cause: err });\n            }\n        }\n"}}},
		{"corpus", "", "try {\n        } catch (err) {\n            {\n                throw new AggregateError();\n            }\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new AggregateError();", "try {\n        } catch (err) {\n            {\n                throw new AggregateError([], \"\", { cause: err });\n            }\n        }\n"}}},
		{"corpus", "", "try {\n        } catch (err) {\n            {\n                throw new AggregateError([]);\n            }\n        }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new AggregateError([]);", "try {\n        } catch (err) {\n            {\n                throw new AggregateError([], \"\", { cause: err });\n            }\n        }\n"}}},
		{"corpus", "{\"requireCatchParameter\": true}", "try {\n\t\t\tdoSomething();\n\t\t} catch {\n\t\t\tthrow new Error(\"Something went wrong\");\n\t\t}\n", []preserveCaughtErrorFinding{{"missingCatchErrorParam", "throw new Error(\"Something went wrong\");", ""}}},
		{"corpus", "", "try {\n            doSomething();\n        } catch (err) {\n            throw new Error(\"Something failed\", { cause });\n        }\n", []preserveCaughtErrorFinding{{"incorrectCause", "cause", "try {\n            doSomething();\n        } catch (err) {\n            throw new Error(\"Something failed\", { cause: err });\n        }\n"}}},
		{"corpus", "", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch ({ message }) {\n\t\t\t\tthrow new Error(message);\n\t\t\t}\n", []preserveCaughtErrorFinding{{"partiallyLostError", "catch ({ message }) {\n\t\t\t\tthrow new Error(message);\n\t\t\t}", ""}}},
		{"corpus", "", "try {\n\t\t\t\tdoSomethingElse();\n\t\t\t} catch ({ ...error }) {\n\t\t\t\tthrow new Error(error.message);\n\t\t\t}\n", []preserveCaughtErrorFinding{{"partiallyLostError", "catch ({ ...error }) {\n\t\t\t\tthrow new Error(error.message);\n\t\t\t}", ""}}},
		{"corpus", "", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tif (whatever) {\n\t\t\t\t\tconst error = anotherError;\n\t\t\t\t\tthrow new Error(\"Something went wrong\", { cause: error });\n\t\t\t\t}\n\t\t\t}\n", []preserveCaughtErrorFinding{{"caughtErrorShadowed", "throw new Error(\"Something went wrong\", { cause: error });", ""}}},
		{"corpus", "", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new Error(\n\t\t\t\t\t\"Something went wrong\" // some comments\n\t\t\t\t);\n\t\t\t}\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\n\t\t\t\t\t\"Something went wrong\" // some comments\n\t\t\t\t);", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new Error(\n\t\t\t\t\t\"Something went wrong\", { cause: error } // some comments\n\t\t\t\t);\n\t\t\t}\n"}}},
		{"corpus", "", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new Error(\"Something failed\", {});\n\t\t\t}\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"Something failed\", {});", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new Error(\"Something failed\", {cause: err});\n\t\t\t}\n"}}},
		{"corpus", "", "try {\n\t\t\tdoSomething();\n\t\t} catch (error) {\n\t\t\tconst cause = \"desc\";\n\t\t\tthrow new Error(\"Something failed\", { [cause]: \"Some error\" });\n\t\t}\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"Something failed\", { [cause]: \"Some error\" });", "try {\n\t\t\tdoSomething();\n\t\t} catch (error) {\n\t\t\tconst cause = \"desc\";\n\t\t\tthrow new Error(\"Something failed\", { [cause]: \"Some error\", cause: error });\n\t\t}\n"}}},
		{"corpus", "", "try {\n\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\tthrow new Error(\"Something failed\", { cause() { /* do something */ }  });\n\t\t\t}\n", []preserveCaughtErrorFinding{{"incorrectCause", "() { /* do something */ }", "try {\n\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\tthrow new Error(\"Something failed\", { cause: error  });\n\t\t\t}\n"}}},
		{"corpus", "", "try {} catch (error) {\n\t\t\t\tthrow new Error(\"Something failed\", { cause: error, cause: anotherError });\n\t\t\t}\n", []preserveCaughtErrorFinding{{"incorrectCause", "anotherError", ""}}},
		{"corpus", "", "try {} catch (error) {\n\t\t\t\tthrow new Error(\"Something failed\", { cause: error, \"cause\": anotherError });\n\t\t\t}\n", []preserveCaughtErrorFinding{{"incorrectCause", "anotherError", ""}}},
		{"corpus", "", "try {} catch (error) {\n\t\t\t\tthrow new Error(\"Something failed\", { get cause() { } });\n\t\t\t}\n", []preserveCaughtErrorFinding{{"incorrectCause", "() { }", "try {} catch (error) {\n\t\t\t\tthrow new Error(\"Something failed\", { cause: error });\n\t\t\t}\n"}}},
		{"corpus", "", "try {} catch (error) {\n\t\t\t\tthrow new Error(\"Something failed\", { set cause(value) { } });\n\t\t\t}\n", []preserveCaughtErrorFinding{{"incorrectCause", "(value) { }", "try {} catch (error) {\n\t\t\t\tthrow new Error(\"Something failed\", { cause: error });\n\t\t\t}\n"}}},
		{"corpus", "", "try {} catch (error) {\n\t\t\t\tthrow new Error(\"Something failed\", {\n\t\t\t\t\tget cause() { return error; },\n\t\t\t\t\tset cause(value) { error = value; },\n\t\t\t\t});\n\t\t\t}\n", []preserveCaughtErrorFinding{{"incorrectCause", "(value) { error = value; }", ""}}},
		{"corpus", "", "try { doSomething(); } catch (err) { throw new Error((\"Something failed\")); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error((\"Something failed\"));", "try { doSomething(); } catch (err) { throw new Error((\"Something failed\"), { cause: err }); }\n"}}},
		{"corpus", "", "try { doSomething(); } catch (err) { throw new Error((\"Something failed\"),); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error((\"Something failed\"),);", "try { doSomething(); } catch (err) { throw new Error((\"Something failed\"), { cause: err },); }\n"}}},
		{"corpus", "", "try { doSomething(); } catch (err) { throw new AggregateError((errors)); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new AggregateError((errors));", "try { doSomething(); } catch (err) { throw new AggregateError((errors), \"\", { cause: err }); }\n"}}},
		{"corpus", "", "try { doSomething(); } catch (err) { throw new AggregateError(errors, (\"message\")); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new AggregateError(errors, (\"message\"));", "try { doSomething(); } catch (err) { throw new AggregateError(errors, (\"message\"), { cause: err }); }\n"}}},
		{"corpus", "{\"errorClassNames\": [\"CustomError\"]}", "try { doSomething(); } catch (err) { throw new CustomError((foo)); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new CustomError((foo));", "try { doSomething(); } catch (err) { throw new CustomError((foo), { cause: err }); }\n"}}},
		{"corpus", "{\"errorClassNames\": [\"CustomApplicationError\"]}", "class CustomApplicationError extends Error {}\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new CustomApplicationError(\"Cause not provided\");\n\t\t\t}\n", []preserveCaughtErrorFinding{{"missingCause", "throw new CustomApplicationError(\"Cause not provided\");", "class CustomApplicationError extends Error {}\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new CustomApplicationError(\"Cause not provided\", { cause: err });\n\t\t\t}\n"}}},
		{"corpus", "{\"errorClassNames\": [\"APIError\"]}", "class APIError extends Error {}\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new APIError(\"API failed\", { cause: wrong });\n\t\t\t}\n", []preserveCaughtErrorFinding{{"incorrectCause", "wrong", "class APIError extends Error {}\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new APIError(\"API failed\", { cause: error });\n\t\t\t}\n"}}},
		{"corpus", "{\"errorClassNames\": [\"AppError\"]}", "const errors = { AppError: class AppError extends Error {} };\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new errors.AppError(\"Something failed\");\n\t\t\t}\n", []preserveCaughtErrorFinding{{"missingCause", "throw new errors.AppError(\"Something failed\");", "const errors = { AppError: class AppError extends Error {} };\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new errors.AppError(\"Something failed\", { cause: error });\n\t\t\t}\n"}}},
		{"corpus", "{\"errorClassNames\": [\"APIError\"]}", "const lib = { APIError: class APIError extends Error {} };\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new lib.APIError(\"Something failed\", { cause: wrong });\n\t\t\t}\n", []preserveCaughtErrorFinding{{"incorrectCause", "wrong", "const lib = { APIError: class APIError extends Error {} };\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new lib.APIError(\"Something failed\", { cause: error });\n\t\t\t}\n"}}},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"MyError\", \"argumentPosition\": 3}]}", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new MyError(\"failed\", someData);\n\t\t\t}\n", []preserveCaughtErrorFinding{{"missingCause", "throw new MyError(\"failed\", someData);", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new MyError(\"failed\", someData, { cause: err });\n\t\t\t}\n"}}},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"MyError\", \"argumentPosition\": 3}]}", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new MyError(\"failed\", someData, { cause: wrong });\n\t\t\t}\n", []preserveCaughtErrorFinding{{"incorrectCause", "wrong", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new MyError(\"failed\", someData, { cause: err });\n\t\t\t}\n"}}},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"SimpleError\", \"argumentPosition\": 1}]}", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new SimpleError();\n\t\t\t}\n", []preserveCaughtErrorFinding{{"missingCause", "throw new SimpleError();", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new SimpleError({ cause: err });\n\t\t\t}\n"}}},
		{"corpus", "{\"errorClassNames\": [\"MyError\"]}", "try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new MyError();\n\t\t\t}\n", []preserveCaughtErrorFinding{{"missingCause", "throw new MyError();", ""}}},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"Error\", \"argumentPosition\": 3}]}", "try {\n\t\t\t    doSomething();\n\t\t\t} catch (err) {\n\t\t\t    throw new context.Error(\"failed\", someData);\n\t\t\t}\n", []preserveCaughtErrorFinding{{"missingCause", "throw new context.Error(\"failed\", someData);", "try {\n\t\t\t    doSomething();\n\t\t\t} catch (err) {\n\t\t\t    throw new context.Error(\"failed\", someData, { cause: err });\n\t\t\t}\n"}}},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"Error\", \"argumentPosition\": 3}]}", "try {\n\t\t\t    doSomething();\n\t\t\t} catch (err) {\n\t\t\t    throw new Error(\"failed\");\n\t\t\t}\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"failed\");", "try {\n\t\t\t    doSomething();\n\t\t\t} catch (err) {\n\t\t\t    throw new Error(\"failed\", { cause: err });\n\t\t\t}\n"}}},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"AggregateError\", \"argumentPosition\": 1}]}", "import { AggregateError } from \"some-module\";\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new AggregateError();\n\t\t\t}\n", []preserveCaughtErrorFinding{{"missingCause", "throw new AggregateError();", "import { AggregateError } from \"some-module\";\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new AggregateError({ cause: err });\n\t\t\t}\n"}}},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"AggregateError\", \"argumentPosition\": 1}]}", "import { AggregateError } from \"some-module\";\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new AggregateError({ cause: wrong });\n\t\t\t}\n", []preserveCaughtErrorFinding{{"incorrectCause", "wrong", "import { AggregateError } from \"some-module\";\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new AggregateError({ cause: err });\n\t\t\t}\n"}}},
		{"corpus", "", "try {} catch (err) { throw new Error; }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error;", "try {} catch (err) { throw new Error(\"\", { cause: err }); }\n"}}},
		{"corpus", "", "try {} catch (err) { throw new AggregateError; }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new AggregateError;", "try {} catch (err) { throw new AggregateError([], \"\", { cause: err }); }\n"}}},
		{"corpus", "{\"errorClassNames\": [{\"name\": \"CustomError\", \"argumentPosition\": 1}]}", "try {} catch (err) { throw new CustomError; }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new CustomError;", "try {} catch (err) { throw new CustomError({ cause: err }); }\n"}}},
		{"corpus", "", "try {} catch (err) { throw new (Error); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new (Error);", "try {} catch (err) { throw new (Error)(\"\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause<T>() {} }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "<T>() {}", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { async cause() {} }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "() {}", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { *cause() {} }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "() {}", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new (Error)(\"m\"); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new (Error)(\"m\");", "try { a(); } catch (err) { throw new (Error)(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw (new Error(\"m\")); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw (new Error(\"m\"));", "try { a(); } catch (err) { throw (new Error(\"m\", { cause: err })); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error; }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error;", "try { a(); } catch (err) { throw new Error(\"\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new (Error); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new (Error);", "try { a(); } catch (err) { throw new (Error)(\"\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: (other) }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "other", "try { a(); } catch (err) { throw new Error(\"m\", { cause: (err) }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: (err) }); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err! }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "err!", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", o); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", {}, 3); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"m\", {}, 3);", "try { a(); } catch (err) { throw new Error(\"m\", {cause: err}, 3); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: other }, ...rest); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "other", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err }, ...rest); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(...a, {}); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { const f = () => { throw new Error(\"m\"); }; f(); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { class K { static { throw new Error(\"m\"); } } }\n", nil},
		{"edge", "", "try { a(); } catch (err) { throw new RangeError(\"m\"); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new RangeError(\"m\");", "try { a(); } catch (err) { throw new RangeError(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (cause) { throw new Error(\"m\", { cause }); }\n", nil},
		{"edge", "", "try { a(); } catch ({ message }) { throw new Error(message); throw new TypeError(message); }\n", []preserveCaughtErrorFinding{{"partiallyLostError", "catch ({ message }) { throw new Error(message); throw new TypeError(message); }", ""}, {"partiallyLostError", "catch ({ message }) { throw new Error(message); throw new TypeError(message); }", ""}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { \"cause\": other }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "other", "try { a(); } catch (err) { throw new Error(\"m\", { \"cause\": err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: other, cause: other2 }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "other2", ""}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\",); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"m\",);", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err },); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error((\"m\")); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error((\"m\"));", "try { a(); } catch (err) { throw new Error((\"m\"), { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", {}); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"m\", {});", "try { a(); } catch (err) { throw new Error(\"m\", {cause: err}); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { }); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"m\", { });", "try { a(); } catch (err) { throw new Error(\"m\", {cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { a: 1, }); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"m\", { a: 1, });", "try { a(); } catch (err) { throw new Error(\"m\", { a: 1, cause: err, }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw Error(); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw Error();", "try { a(); } catch (err) { throw Error(\"\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err } as any); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\") as any; }\n", nil},
		{"edge", "", "try { a(); } catch (err) { { let err = 1; throw new Error(\"m\", { cause: err }); } }\n", []preserveCaughtErrorFinding{{"caughtErrorShadowed", "throw new Error(\"m\", { cause: err });", ""}}},
		{"edge", "", "try { a(); } catch (err) { { throw new Error(\"m\", { cause: err }); } }\n", nil},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err, ...rest }); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { ...rest, cause: other }); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { try {} finally { throw new Error(\"m\"); } }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"m\");", "try { a(); } catch (err) { try {} finally { throw new Error(\"m\", { cause: err }); } }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { get cause() { return err; } }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "() { return err; }", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new AggregateError([], \"m\", o); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { throw new AggregateError([err], \"m\", \"x\"); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err.cause }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "err.cause", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error/* ( */(); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error/* ( */();", "try { a(); } catch (err) { throw new Error/* ( */(\"\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { [`cause`]: other }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "other", "try { a(); } catch (err) { throw new Error(\"m\", { [`cause`]: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { [\"cause\"]: other }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "other", "try { a(); } catch (err) { throw new Error(\"m\", { [\"cause\"]: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: other, [\"cause\"]: err }); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err, cause: other }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "other", ""}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: other, cause: err }); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "cause", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new AggregateError([err]); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new AggregateError([err]);", "try { a(); } catch (err) { throw new AggregateError([err], \"\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new AggregateError; }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new AggregateError;", "try { a(); } catch (err) { throw new AggregateError([], \"\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: other } satisfies object); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: <any>other }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "<any>other", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { set cause(v) {} }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "(v) {}", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", { cause: function () {} }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "function () {}", "try { a(); } catch (err) { throw new Error(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw globalThis.Error(\"m\"); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { throw Error?.(\"m\"); }\n", nil},
		{"edge", "", "try { a(); } catch (err) { function g() {} throw new Error(\"m\", { cause: err as Error }); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "err as Error", "try { a(); } catch (err) { function g() {} throw new Error(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch { throw new Error(\"m\"); const f = () => { throw new Error(\"n\"); }; }\n", nil},
		{"edge", "", "try { a(); } catch (err: unknown) { throw new Error(\"m\"); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"m\");", "try { a(); } catch (err: unknown) { throw new Error(\"m\", { cause: err }); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { var err; throw new Error(\"m\", { cause: err }); }\n", nil},
		{"edge", "{\"errorClassNames\": [\"MyError\"]}", "try { a(); } catch (err) { throw a?.MyError(); }\n", nil},
		{"edge", "{\"errorClassNames\": [\"MyError\"]}", "try { a(); } catch (err) { throw new errors.MyError(\"m\"); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new errors.MyError(\"m\");", "try { a(); } catch (err) { throw new errors.MyError(\"m\", { cause: err }); }\n"}}},
		{"edge", "{\"errorClassNames\": [\"MyError\"]}", "try { a(); } catch (err) { throw new errors[\"MyError\"](\"m\"); }\n", nil},
		{"edge", "{\"errorClassNames\": [{\"name\": \"MyError\", \"argumentPosition\": 3}]}", "try { a(); } catch (err) { throw new MyError(\"m\"); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new MyError(\"m\");", ""}}},
		{"edge", "{\"errorClassNames\": [{\"name\": \"MyError\", \"argumentPosition\": 3}]}", "try { a(); } catch (err) { throw new MyError(\"a\", \"b\"); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new MyError(\"a\", \"b\");", "try { a(); } catch (err) { throw new MyError(\"a\", \"b\", { cause: err }); }\n"}}},
		{"edge", "{\"errorClassNames\": [{\"name\": \"MyError\", \"argumentPosition\": 1}]}", "try { a(); } catch (err) { throw new MyError(); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new MyError();", "try { a(); } catch (err) { throw new MyError({ cause: err }); }\n"}}},
		{"edge", "{\"errorClassNames\": [{\"name\": \"MyError\", \"argumentPosition\": 1}]}", "try { a(); } catch (err) { throw new MyError; }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new MyError;", "try { a(); } catch (err) { throw new MyError({ cause: err }); }\n"}}},
		{"edge", "{\"errorClassNames\": [\"Error\"]}", "import { Error } from 'x';\ntry { a(); } catch (err) { throw new Error(\"m\"); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"m\");", "import { Error } from 'x';\ntry { a(); } catch (err) { throw new Error(\"m\", { cause: err }); }\n"}}},
		{"edge", "{\"errorClassNames\": [{\"name\": \"Error\", \"argumentPosition\": 3}]}", "try { a(); } catch (err) { throw new Error(\"m\", \"x\", {}); }\n", nil},
		{"edge", "{\"errorClassNames\": [\"MyError\"]}", "try { a(); } catch (err) { throw new this.MyError(\"m\"); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new this.MyError(\"m\");", "try { a(); } catch (err) { throw new this.MyError(\"m\", { cause: err }); }\n"}}},
		{"edge", "{\"errorClassNames\": [\"MyError\"]}", "try { a(); } catch (err) { throw new (errors.MyError)(\"m\"); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new (errors.MyError)(\"m\");", "try { a(); } catch (err) { throw new (errors.MyError)(\"m\", { cause: err }); }\n"}}},
		{"edge", "{\"requireCatchParameter\": true}", "try { a(); } catch { throw new Error(\"m\"); throw new TypeError(\"n\"); }\n", []preserveCaughtErrorFinding{{"missingCatchErrorParam", "throw new Error(\"m\");", ""}, {"missingCatchErrorParam", "throw new TypeError(\"n\");", ""}}},
		{"edge", "{\"requireCatchParameter\": true}", "try { a(); } catch { const f = () => { throw new Error(\"n\"); }; }\n", nil},
		{"edge", "{\"requireCatchParameter\": true}", "try { a(); } catch { throw err; }\n", nil},
		{"edge", "{\"requireCatchParameter\": true}", "try { a(); } catch ({ message }) { throw new Error(message); }\n", []preserveCaughtErrorFinding{{"partiallyLostError", "catch ({ message }) { throw new Error(message); }", ""}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", ({ cause: other })); }\n", []preserveCaughtErrorFinding{{"incorrectCause", "other", "try { a(); } catch (err) { throw new Error(\"m\", ({ cause: err })); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", ({})); }\n", []preserveCaughtErrorFinding{{"missingCause", "throw new Error(\"m\", ({}));", "try { a(); } catch (err) { throw new Error(\"m\", ({cause: err})); }\n"}}},
		{"edge", "", "try { a(); } catch (err) { throw new Error(\"m\", ({ cause: err })); }\n", nil},
	}

	decode := rule.DecodeOptionsInto[PreserveCaughtErrorOptions]()
	for index, testCase := range cases {
		var options any
		if testCase.options != "" {
			decoded, err := decode([]byte(testCase.options))
			if err != nil {
				t.Fatalf("case %d (%s): ESLint accepts %s and the decoder refused it: %v", index, testCase.origin, testCase.options, err)
			}
			options = decoded
		}
		result := rule_testing.RunTypedWithOptions(t, PreserveCaughtError, "input.ts", testCase.code, options)
		source := result.SourceFile.Text()
		if source != testCase.code {
			t.Fatalf("case %d: the harness changed the text, so the spans below would be read off different bytes", index)
		}
		if len(result.Diagnostics) != len(testCase.findings) {
			t.Errorf("case %d (%s): %d findings, ESLint reports %d\n%s", index, testCase.origin, len(result.Diagnostics), len(testCase.findings), testCase.code)
			continue
		}
		for position, diagnostic := range result.Diagnostics {
			want := testCase.findings[position]
			text := source[diagnostic.Range.Pos():diagnostic.Range.End()]
			if diagnostic.Message.Id != want.id || text != want.text {
				t.Errorf("case %d (%s): finding %d is %s at %q, ESLint reports %s at %q\n%s",
					index, testCase.origin, position, diagnostic.Message.Id, text, want.id, want.text, testCase.code)
				continue
			}
			if len(diagnostic.Fixes) != 0 {
				t.Errorf("case %d: finding %d carries an autofix, and every repair here is a suggestion", index, position)
			}
			if want.suggestion == "" {
				if len(diagnostic.Suggestions) != 0 {
					t.Errorf("case %d (%s): finding %d offers a suggestion where ESLint offers none\n%s", index, testCase.origin, position, testCase.code)
				}
				continue
			}
			if len(diagnostic.Suggestions) != 1 {
				t.Errorf("case %d (%s): finding %d offers %d suggestions, ESLint offers one\n%s", index, testCase.origin, position, len(diagnostic.Suggestions), testCase.code)
				continue
			}
			if diagnostic.Suggestions[0].Message.Id != "includeCause" {
				t.Errorf("case %d: the suggestion is %s, ESLint's is includeCause", index, diagnostic.Suggestions[0].Message.Id)
			}
			if applied := applySuggestion(t, source, diagnostic.Suggestions[0]); applied != want.suggestion {
				t.Errorf("case %d (%s): the suggestion writes\n  %q\nESLint writes\n  %q", index, testCase.origin, applied, want.suggestion)
			}
		}
	}
}
