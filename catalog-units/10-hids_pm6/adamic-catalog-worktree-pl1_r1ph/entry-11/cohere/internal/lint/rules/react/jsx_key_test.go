package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// jsxKeyFile is where the fixtures pretend to live.
//
// A `.tsx` name because every case is JSX and needs a parser that reads it. It is NOT load-bearing:
// this rule has no file gate, and `TestJsxKeyHasNoFileGate` pins that by running one reporting
// source under three JSX-capable extensions.
const jsxKeyFile = "/repository/source/JsxKey.tsx"

// The corpus is `eslint-plugin-react`'s own, extracted rather than retyped.
//
// `/tmp/lint-sources/eslint-plugin-react/tests/lib/rules/jsx-key.js` holds 37 valid and 30 invalid
// cases. Every string below was pulled out of that file by loading it with a stubbed `RuleTester`
// and serializing the captured object, then verified byte against byte by extracting the same file
// a second time through an independent path that writes raw bytes rather than JSON. All 67 codes
// were byte-identical across the two extractions.
//
// # This rule's two authorities DISAGREE, and the corpus wins
//
// All 67 were run against the installed build, version 7.37.5, and it disagreed with the corpus on
// SIX of them. Every one is a block-bodied callback returning a conditional or a logical: the clone
// descends into those and the installed build does not. The clone carries unreleased commits on top
// of the 7.37.5 tag while still reporting that version string, so the installed build is a snapshot
// that is behind rather than a different opinion.
//
// The six are marked in the table below. Anyone comparing this rule's output against the ESLint
// this repository currently runs should expect exactly those six shapes to differ, in the direction
// of us reporting and ESLint being silent. The full reasoning is on the rule.
//
// The ARROW-BODY twins of those same shapes agree on both builds and are in the tables unmarked,
// which is what makes the divergence easy to misread as a defect here.
//
// # Five cases carry a pragma through settings, and every one of them moves
//
// Upstream configures `pragma: Act, fragment: Frag` on five cases. cohere reads no React
// settings, so neither can reach a rule here. Each was re-run against the installed build
// with its settings REMOVED and the table records that answer. Four of the five report the same
// thing under either configuration; `valid-33` is the one that genuinely moves, because its
// `Act.Children.toArray` suppression depends on the pragma being `Act`.
//
// # The options column is RAW JSON
//
// Every default here is false, so no inversion is needed, but the tables still route through
// `DecodeJsxKeyOptions` because that is the function the config calls and because it has a nil path
// the generic decoder does not.
//
// Two cases needed a TypeScript parser that upstream's JSX harness does not select, so they came
// back as parse errors from the oracle rather than as verdicts. Our parser reads TypeScript
// unconditionally, and both are clean cases upstream, so both are recorded as clean and their
// silence here is a real measurement of our rule rather than an inherited one.

func TestJsxKeyFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantIds    []string
	}{
		{"upstream valid-33", "\n        import Act from 'react';\n        import { Children as ReactChildren } from 'react';\n\n        const { Children } = Act;\n        const { toArray } = Children;\n\n        Act.Children.toArray([1, 2 ,3].map(x => <App />));\n        Act.Children.toArray(Array.from([1, 2 ,3], x => <App />));\n        Children.toArray([1, 2 ,3].map(x => <App />));\n        Children.toArray(Array.from([1, 2 ,3], x => <App />));\n        // ReactChildren.toArray([1, 2 ,3].map(x => <App />));\n        // ReactChildren.toArray(Array.from([1, 2 ,3], x => <App />));\n        // toArray([1, 2 ,3].map(x => <App />));\n        // toArray(Array.from([1, 2 ,3], x => <App />));\n      ", "", []string{"missingIterKey", "missingIterKey"}}, // upstream sets a pragma; cohere reads no React settings, so this is the unconfigured answer
		{"upstream invalid-0", "\n        [1, 2, 3].map((item) => {\n          return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;\n        })", "", []string{"missingIterKey", "missingIterKey"}},                                                                                                                                                                                                                                                           // CLONE behavior; the installed 7.37.5 is silent here (see the rule doc)
		{"upstream invalid-1", "\n        [1, 2, 3].map(function(item) {\n          return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;\n        })", "", []string{"missingIterKey", "missingIterKey"}},                                                                                                                                                                                                                                                      // CLONE behavior; the installed 7.37.5 is silent here (see the rule doc)
		{"upstream invalid-2", "\n        Array.from([1, 2, 3], (item) => {\n          return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;\n        })", "", []string{"missingIterKey", "missingIterKey"}},                                                                                                                                                                                                                                                   // CLONE behavior; the installed 7.37.5 is silent here (see the rule doc)
		{"upstream invalid-3", "\n        import { Fragment } from 'react';\n\n        const ITEMS = ['bar', 'foo'];\n\n        export default function BugIssue() {\n          return (\n            <Fragment>\n              {ITEMS.map((item) => {\n                return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;\n              })}\n            </Fragment>\n          );\n        }\n      ", "", []string{"missingIterKey", "missingIterKey"}}, // CLONE behavior; the installed 7.37.5 is silent here (see the rule doc)
		{"upstream invalid-4", "[<App />];", "", []string{"missingArrayKey"}},
		{"upstream invalid-5", "[<App {...key} />];", "", []string{"missingArrayKey"}},
		{"upstream invalid-6", "[<App key={0}/>, <App />];", "", []string{"missingArrayKey"}},
		{"upstream invalid-7", "[1, 2 ,3].map(function(x) { return <App /> });", "", []string{"missingIterKey"}},
		{"upstream invalid-8", "[1, 2 ,3].map(x => <App />);", "", []string{"missingIterKey"}},
		{"upstream invalid-9", "[1, 2 ,3].map(x => x && <App x={x} />);", "", []string{"missingIterKey"}},
		{"upstream invalid-10", "[1, 2 ,3].map(x => x ? <App x={x} key=\"1\" /> : <OtherApp x={x} />);", "", []string{"missingIterKey"}},
		{"upstream invalid-11", "[1, 2 ,3].map(x => x ? <App x={x} /> : <OtherApp x={x} key=\"2\" />);", "", []string{"missingIterKey"}},
		{"upstream invalid-12", "[1, 2 ,3].map(x => { return <App /> });", "", []string{"missingIterKey"}},
		{"upstream invalid-13", "Array.from([1, 2 ,3], function(x) { return <App /> });", "", []string{"missingIterKey"}},
		{"upstream invalid-14", "Array.from([1, 2 ,3], (x => { return <App /> }));", "", []string{"missingIterKey"}},
		{"upstream invalid-15", "Array.from([1, 2 ,3], (x => <App />));", "", []string{"missingIterKey"}},
		{"upstream invalid-16", "[1, 2, 3]?.map(x => <BabelEslintApp />)", "", []string{"missingIterKey"}},
		{"upstream invalid-17", "[1, 2, 3]?.map(x => <TypescriptEslintApp />)", "", []string{"missingIterKey"}},
		{"upstream invalid-18", "[1, 2, 3].map(x => <>{x}</>);", "{\"checkFragmentShorthand\": true}", []string{"missingIterKeyUsePrag"}},          // upstream sets a pragma; cohere reads no React settings, so this is the unconfigured answer
		{"upstream invalid-19", "[<></>];", "{\"checkFragmentShorthand\": true}", []string{"missingArrayKeyUsePrag"}},                              // upstream sets a pragma; cohere reads no React settings, so this is the unconfigured answer
		{"upstream invalid-20", "[<App {...obj} key=\"keyAfterSpread\" />];", "{\"checkKeyMustBeforeSpread\": true}", []string{"keyBeforeSpread"}}, // upstream sets a pragma; cohere reads no React settings, so this is the unconfigured answer
		{"upstream invalid-21", "[<div {...obj} key=\"keyAfterSpread\" />];", "{\"checkKeyMustBeforeSpread\": true}", []string{"keyBeforeSpread"}}, // upstream sets a pragma; cohere reads no React settings, so this is the unconfigured answer
		{"upstream invalid-22", "\n        const spans = [\n          <span key=\"notunique\"/>,\n          <span key=\"notunique\"/>,\n        ];\n      ", "{\"warnOnDuplicates\": true}", []string{"nonUniqueKeys", "nonUniqueKeys"}},
		{"upstream invalid-23", "\n        const div = (\n          <div>\n            <span key=\"notunique\"/>\n            <span key=\"notunique\"/>\n          </div>\n        );\n      ", "{\"warnOnDuplicates\": true}", []string{"nonUniqueKeys", "nonUniqueKeys"}},
		{"upstream invalid-24", "\n        const Test = () => {\n          const list = [1, 2, 3, 4, 5];\n\n          return (\n            <div>\n              {list.map(item => {\n                if (item < 2) {\n                  return <div>{item}</div>;\n                }\n\n                return <div />;\n              })}\n            </div>\n          );\n        };\n      ", "", []string{"missingIterKey", "missingIterKey"}},
		{"upstream invalid-25", "\n        const TestO = () => {\n          const list = [1, 2, 3, 4, 5];\n\n          return (\n            <div>\n              {list.map(item => {\n                if (item < 2) {\n                  return <div>{item}</div>;\n                } else if (item < 5) {\n                  return <div></div>\n                }  else {\n                  return <div></div>\n                }\n\n                return <div />;\n              })}\n            </div>\n          );\n        };\n      ", "", []string{"missingIterKey", "missingIterKey", "missingIterKey", "missingIterKey"}},
		{"upstream invalid-26", "\n        const TestCase = () => {\n          const list = [1, 2, 3, 4, 5];\n\n          return (\n            <div>\n              {list.map(item => {\n                if (item < 2) return <div>{item}</div>;\n                else if (item < 5) return <div />;\n                else return <div />;\n              })}\n            </div>\n          );\n        };\n      ", "", []string{"missingIterKey", "missingIterKey", "missingIterKey"}},
		{"upstream invalid-27", "\n        const TestCase = () => {\n          const list = [1, 2, 3, 4, 5];\n\n          return (\n            <div>\n              {list.map(x => <div {...spread} key={x} />)}\n            </div>\n          );\n        };\n      ", "{\"checkKeyMustBeforeSpread\": true}", []string{"keyBeforeSpread"}},
		{"upstream invalid-28", "[1, 2, 3].map(x => { return x && <App />; });", "", []string{"missingIterKey"}},      // CLONE behavior; the installed 7.37.5 is silent here (see the rule doc)
		{"upstream invalid-29", "[1, 2, 3].map(x => { return x || y || <App />; });", "", []string{"missingIterKey"}}, // CLONE behavior; the installed 7.37.5 is silent here (see the rule doc)

		// --- cases upstream does not cover, each measured against the installed build 7.37.5 ---
		//
		// Three of them sit on the branch where the clone and the installed build disagree, and
		// they are marked. Their verdict here is the clone's, derived from the same rule the six
		// marked corpus cases above establish, rather than from a run: the installed build reports
		// nothing for any block-bodied return of a conditional or a logical.
		// An element writing `key` TWICE reports twice, because upstream's loop runs once per
		// matching attribute rather than once per element. Written for a mutant that returned only
		// the first match, which survived the whole imported corpus because nothing in it doubles
		// an attribute.
		{"two key attributes, key after spread", "[<App {...o} key=\"a\" key=\"b\" />];", "{\"checkKeyMustBeforeSpread\":true}", []string{"keyBeforeSpread", "keyBeforeSpread"}},
		{"two identical key attributes on one element", "[<App key=\"a\" key=\"a\" />];", "{\"warnOnDuplicates\":true}", []string{"nonUniqueKeys", "nonUniqueKeys"}},

		{"map arrow", "[1,2,3].map(x => <App />);", "", []string{"missingIterKey"}},
		{"map block", "[1,2,3].map(x => { return <App /> });", "", []string{"missingIterKey"}},
		{"map fn expr", "[1,2,3].map(function(x) { return <App /> });", "", []string{"missingIterKey"}},
		{"Array.from", "Array.from([1,2,3], x => <App />);", "", []string{"missingIterKey"}},
		{"optional map", "[1,2,3]?.map(x => <App />)", "", []string{"missingIterKey"}},
		{"uppercase KEY attr in array", "[<App KEY={1} />];", "", []string{"missingArrayKey"}},
		{"spread satisfies map", "[1,2,3].map(x => <App {...p} />);", "", []string{"missingIterKey"}},
		{"spread in array", "[<App {...p} />];", "", []string{"missingArrayKey"}},
		{"dup array warn", "const s = [<span key=\"a\"/>, <span key=\"a\"/>];", "{\"warnOnDuplicates\": true}", []string{"nonUniqueKeys", "nonUniqueKeys"}},
		{"dup children warn", "const d = (<div><span key=\"a\"/><span key=\"a\"/></div>);", "{\"warnOnDuplicates\": true}", []string{"nonUniqueKeys", "nonUniqueKeys"}},
		{"triple dup warn", "const s = [<span key=\"a\"/>, <span key=\"a\"/>, <span key=\"a\"/>];", "{\"warnOnDuplicates\": true}", []string{"nonUniqueKeys", "nonUniqueKeys", "nonUniqueKeys"}},
		{"keyBeforeSpread array", "[<App {...obj} key=\"k\" />];", "{\"checkKeyMustBeforeSpread\": true}", []string{"keyBeforeSpread"}},
		{"keyBeforeSpread map", "[1,2,3].map(x => <div {...s} key={x} />);", "{\"checkKeyMustBeforeSpread\": true}", []string{"keyBeforeSpread"}},
		{"fragment shorthand map opt", "[1,2,3].map(x => <>{x}</>);", "{\"checkFragmentShorthand\": true}", []string{"missingIterKeyUsePrag"}},
		{"nested element only reports dup", "<div><App key=\"a\"/><App key=\"a\"/></div>;", "{\"warnOnDuplicates\": true}", []string{"nonUniqueKeys", "nonUniqueKeys"}},
		{"kbs children span", "<div><App {...obj} key=\"k\" /></div>;", "{\"checkKeyMustBeforeSpread\": true}", []string{"keyBeforeSpread"}},
		{"dup span", "[<span key=\"a\"/>, <span key=\"a\"/>];", "{\"warnOnDuplicates\": true}", []string{"nonUniqueKeys", "nonUniqueKeys"}},
		{"block if return", "[1,2,3].map(x => { if (x) { return <App />; } return <B />; });", "", []string{"missingIterKey", "missingIterKey"}},
		{"block if else return", "[1,2,3].map(x => { if (x) return <A />; else return <B />; });", "", []string{"missingIterKey", "missingIterKey"}},
		{"block ternary return", "[1,2,3].map(x => { return x ? <A /> : <B />; });", "", []string{"missingIterKey", "missingIterKey"}}, // CLONE behavior; the installed 7.37.5 is silent here
		{"block logical return", "[1,2,3].map(x => { return x && <A />; });", "", []string{"missingIterKey"}},                          // CLONE behavior; the installed 7.37.5 is silent here
		{"arrow ternary body", "[1,2,3].map(x => x ? <A /> : <B />);", "", []string{"missingIterKey", "missingIterKey"}},
		{"arrow logical body", "[1,2,3].map(x => x && <A />);", "", []string{"missingIterKey"}},
		{"block nested fn return", "[1,2,3].map(x => { function f() { return <A />; } return <B />; });", "", []string{"missingIterKey"}},
		{"deep if nesting", "[1,2,3].map(x => { if (a) { if (b) { return <A />; } } });", "", []string{"missingIterKey"}},
		{"children toArray wrong ns", "Foo.Children.toArray([1,2,3].map(x => <App />));", "", []string{"missingIterKey"}},
		{"toArray bare fn", "toArray([1,2,3].map(x => <App />));", "", []string{"missingIterKey"}},
		{"Children toArray then outside", "React.Children.toArray([1,2,3].map(x => <App />)); [<B />];", "", []string{"missingArrayKey"}},
		{"array in array", "[[<App />]];", "", []string{"missingArrayKey"}},
		{"array holes", "[, <App />];", "", []string{"missingArrayKey"}},
		{"member map on obj", "obj.list.map(x => <App />);", "", []string{"missingIterKey"}},
		{"optional member map", "obj?.list?.map(x => <App />);", "", []string{"missingIterKey"}},
		{"fragment long form in array", "[<React.Fragment></React.Fragment>];", "", []string{"missingArrayKey"}},
		{"paren arrow from", "Array.from([1,2,3], (x => <App />));", "", []string{"missingIterKey"}},
		{"paren arrow map", "[1,2,3].map((x => <App />));", "", []string{"missingIterKey"}},
		{"paren fn expr map", "[1,2,3].map((function(x) { return <App /> }));", "", []string{"missingIterKey"}},
		{"double paren arrow", "[1,2,3].map(((x => <App />)));", "", []string{"missingIterKey"}},
		{"paren array literal", "([<App />]);", "", []string{"missingArrayKey"}},
		{"paren callee map", "([1,2,3].map)(x => <App />);", "", []string{"missingIterKey"}},
		{"paren element in array", "[(<App />)];", "", []string{"missingArrayKey"}},
		{"paren return arg", "[1,2,3].map(x => { return (<App />); });", "", []string{"missingIterKey"}},
		{"paren arrow body", "[1,2,3].map(x => (<App />));", "", []string{"missingIterKey"}},
		{"paren conditional return", "[1,2,3].map(x => { return (c ? <A/> : <B/>); });", "", []string{"missingIterKey", "missingIterKey"}}, // CLONE behavior; the installed 7.37.5 is silent here
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := decodeJsxKeyOptionsForTest(t, testCase.rawOptions)
			result := rule_testing.RunWithOptions(t, JsxKey, jsxKeyFile, testCase.sourceText, options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

func TestJsxKeyStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"upstream valid-0", "\n        [1, 2, 3].map((item) => {\n         return item === 'bar' ? <div key={item}>{item}</div> : <span key={item}>{item}</span>;\n        })", ""},
		{"upstream valid-1", "fn()", ""},
		{"upstream valid-2", "[1, 2, 3].map(function () {})", ""},
		{"upstream valid-3", "<App />;", ""},
		{"upstream valid-4", "[<App key={0} />, <App key={1} />];", ""},
		{"upstream valid-5", "[1, 2, 3].map(function(x) { return <App key={x} /> });", ""},
		{"upstream valid-6", "[1, 2, 3].map(x => <App key={x} />);", ""},
		{"upstream valid-7", "[1, 2 ,3].map(x => x && <App x={x} key={x} />);", ""},
		{"upstream valid-8", "[1, 2 ,3].map(x => x ? <App x={x} key=\"1\" /> : <OtherApp x={x} key=\"2\" />);", ""},
		{"upstream valid-9", "[1, 2, 3].map(x => { return <App key={x} /> });", ""},
		{"upstream valid-10", "Array.from([1, 2, 3], function(x) { return <App key={x} /> });", ""},
		{"upstream valid-11", "Array.from([1, 2, 3], (x => <App key={x} />));", ""},
		{"upstream valid-12", "Array.from([1, 2, 3], (x => {return <App key={x} />}));", ""},
		{"upstream valid-13", "Array.from([1, 2, 3], someFn);", ""},
		{"upstream valid-14", "Array.from([1, 2, 3]);", ""},
		{"upstream valid-15", "[1, 2, 3].foo(x => <App />);", ""},
		{"upstream valid-16", "var App = () => <div />;", ""},
		{"upstream valid-17", "[1, 2, 3].map(function(x) { return; });", ""},
		{"upstream valid-18", "foo(() => <div />);", ""},
		{"upstream valid-19", "foo(() => <></>);", ""},
		{"upstream valid-20", "<></>;", ""},
		{"upstream valid-21", "<App {...{}} />;", ""},
		{"upstream valid-22", "<App key=\"keyBeforeSpread\" {...{}} />;", "{\"checkKeyMustBeforeSpread\": true}"},
		{"upstream valid-23", "<div key=\"keyBeforeSpread\" {...{}} />;", "{\"checkKeyMustBeforeSpread\": true}"},
		{"upstream valid-24", "\n        const spans = [\n          <span key=\"notunique\"/>,\n          <span key=\"notunique\"/>,\n        ];\n      ", ""},
		{"upstream valid-25", "\n        function Component(props) {\n          return hasPayment ? (\n            <div className=\"stuff\">\n              <BookingDetailSomething {...props} />\n              {props.modal && props.calculatedPrice && (\n                <SomeOtherThing items={props.something} discount={props.discount} />\n              )}\n            </div>\n          ) : null;\n        }\n      ", ""},
		{"upstream valid-26", "\n        import React, { FC, useRef, useState } from 'react';\n\n        import './ResourceVideo.sass';\n        import VimeoVideoPlayInModal from '../vimeoVideoPlayInModal/VimeoVideoPlayInModal';\n\n        type Props = {\n          videoUrl: string;\n          videoTitle: string;\n        };\n        const ResourceVideo: FC<Props> = ({\n          videoUrl,\n          videoTitle,\n        }: Props): JSX.Element => {\n          return (\n            <div className=\"resource-video\">\n              <VimeoVideoPlayInModal videoUrl={videoUrl} />\n              <h3>{videoTitle}</h3>\n            </div>\n          );\n        };\n\n        export default ResourceVideo;\n      ", ""},
		{"upstream valid-27", "\n        // testrule.jsx\n        const trackLink = () => {};\n        const getAnalyticsUiElement = () => {};\n\n        const onTextButtonClick = (e, item) => trackLink([, getAnalyticsUiElement(item), item.name], e);\n      ", ""},
		{"upstream valid-28", "\n        function Component({ allRatings }) {\n          return (\n            <RatingDetailsStyles>\n              {Object.entries(allRatings)?.map(([key, value], index) => {\n                const rate = value?.split(/(?=[%, /])/);\n\n                if (!rate) return null;\n\n                return (\n                  <li key={`${entertainment.tmdbId}${index}`}>\n                    <img src={`/assets/rating/${key}.png`} />\n                    <span className=\"rating-details--rate\">{rate?.[0]}</span>\n                    <span className=\"rating-details--rate-suffix\">{rate?.[1]}</span>\n                  </li>\n                );\n              })}\n            </RatingDetailsStyles>\n          );\n        }\n      ", ""},
		{"upstream valid-29", "\n        const baz = foo?.bar?.()?.[1] ?? 'qux';\n\n        qux()?.map()\n\n        const directiveRanges = comments?.map(tryParseTSDirective)\n      ", ""},
		{"upstream valid-30", "\n        import { observable } from \"mobx\";\n\n        export interface ClusterFrameInfo {\n          frameId: number;\n          processId: number;\n        }\n\n        export const clusterFrameMap = observable.map<string, ClusterFrameInfo>();\n      ", ""},
		{"upstream valid-31", "React.Children.toArray([1, 2 ,3].map(x => <App />));", ""},
		{"upstream valid-32", "\n        import { Children } from \"react\";\n        Children.toArray([1, 2 ,3].map(x => <App />));\n      ", ""},
		{"upstream valid-34", "[1, 2, 3].map(x => { return x && <App key={x} />; });", ""},
		{"upstream valid-35", "[1, 2, 3].map(x => { return x && y && <App key={x} />; });", ""},
		{"upstream valid-36", "[1, 2, 3].map(x => { return x && foo(); });", ""},

		// --- cases upstream does not cover, each measured against the installed build 7.37.5 ---
		//
		// Three of them sit on the branch where the clone and the installed build disagree, and
		// they are marked. Their verdict here is the clone's, derived from the same rule the six
		// marked corpus cases above establish, rather than from a run: the installed build reports
		// nothing for any block-bodied return of a conditional or a logical.
		// A BARE call is never an iterator here. Upstream's selectors require a member-expression
		// callee, so `map(...)` and `from(...)` written as free functions match nothing. Written
		// for a mutant that accepted a bare identifier callee, which survived the imported corpus
		// because every case in it calls through a member.
		{"bare map call", "map(x => <App />);", ""},
		{"bare from call", "from(a, x => <App />);", ""},

		// Only the FIRST argument of `map` is inspected, and only the SECOND of `Array.from`.
		// `map` takes a `thisArg` second argument, so a callback written there is not the iteratee
		// and upstream does not look at it. Written for a mutant that read the last argument
		// instead, which survived the whole imported corpus because nothing in it calls `map` with
		// two arguments.
		{"map with a thisArg second argument", "[1,2,3].map(someFn, x => <App />);", ""},

		// A logical whose RIGHT side is not JSX falls through to nothing rather than to the general
		// arm, because upstream's chain is `else if (isLogical && isJSX(right))` followed by an
		// `else`. So a JSX element on the LEFT is invisible. Written for a mutant that removed the
		// `continue`, letting the general arm see the whole logical expression.
		{"logical with jsx on the left only", "[1,2,3].map(x => { return <App /> && y; });", ""},

		// A nullish coalescing IS a logical expression to espree, and in a block-bodied return it
		// therefore takes the logical arm. Its right side is JSX here, so under the clone it
		// reports; recorded as reporting in the fires table rather than here.
		{"array literal has key", "[<App key={1} />];", ""},
		{"map arrow key", "[1,2,3].map(x => <App key={x} />);", ""},
		{"map fn decl ref", "[1,2,3].map(someFn);", ""},
		{"Array.from one arg", "Array.from([1,2,3]);", ""},
		{"not map", "[1,2,3].foo(x => <App />);", ""},
		{"nested jsx child", "<div><App /></div>;", ""},
		{"uppercase KEY attr in map", "[1,2,3].map(x => <App KEY={x} />);", ""},
		{"children toArray", "React.Children.toArray([1,2,3].map(x => <App />));", ""},
		{"Children toArray bare", "Children.toArray([1,2,3].map(x => <App />));", ""},
		{"fragment shorthand in map", "[1,2,3].map(x => <>{x}</>);", ""},
		{"fragment shorthand in array", "[<></>];", ""},
		{"dup array no warn", "const s = [<span key=\"a\"/>, <span key=\"a\"/>];", ""},
		{"dup different values", "const s = [<span key=\"a\"/>, <span key=\"b\"/>];", "{\"warnOnDuplicates\": true}"},
		{"keyBeforeSpread ok", "[<App key=\"k\" {...obj} />];", "{\"checkKeyMustBeforeSpread\": true}"},
		{"keyBeforeSpread off", "[<App {...obj} key=\"k\" />];", ""},
		{"fragment in children opt", "<div><></></div>;", "{\"checkFragmentShorthand\": true}"},
		{"nested jsx element child missing", "<div><App /><App /></div>;", ""},
		{"block return nothing", "[1,2,3].map(x => { return; });", ""},
		{"block no return", "[1,2,3].map(x => { const y = <A />; });", ""},
		{"block loop return", "[1,2,3].map(x => { for (;;) { return <A />; } });", ""},
		{"block try return", "[1,2,3].map(x => { try { return <A />; } catch(e) {} });", ""},
		{"children toArray pragma", "/** @jsx Act */ Act.Children.toArray([1,2,3].map(x => <App />));", ""},
		{"Children toArray deep nest", "React.Children.toArray([[<App />]]);", ""},
		{"array non jsx", "[1, 2, 3];", ""},
		{"array empty", "[];", ""},
		{"map with no args", "[1,2,3].map();", ""},
		{"map with non fn", "[1,2,3].map(x);", ""},
		{"Array.from second non fn", "Array.from([1,2,3], someFn);", ""},
		{"fragment long form key", "[<React.Fragment key={1}></React.Fragment>];", ""},
		// A keyed element in parentheses, at each position an iterator callback reaches it. ESLint's
		// parser keeps no paren node, so these read as bare elements there; here the check skipped the
		// paren to see an element and then read the paren's attributes, and reported a missing key on
		// all 103 such sites in ahra and www (#ge8899e). Each measured silent on eslint-plugin-react 7.37.5.
		{"keyed element in a parenthesized arrow body", "[1,2,3].map((x) => (<div key={x}>{x}</div>));", ""},
		{"keyed elements in parenthesized arms of a returned conditional", "[1,2,3].map(function(x) { return c ? (<div key={x} />) : (<span key={x} />); });", ""},
		{"keyed elements in parenthesized arms of an arrow conditional", "[1,2,3].map((x) => c ? (<div key={x} />) : (<span key={x} />));", ""},
		{"keyed element doubly parenthesized", "[1,2,3].map((x) => ((<div key={x} />)));", ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := decodeJsxKeyOptionsForTest(t, testCase.rawOptions)
			result := rule_testing.RunWithOptions(t, JsxKey, jsxKeyFile, testCase.sourceText, options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// decodeJsxKeyOptionsForTest routes a fixture's raw JSON through the rule's own decoder.
//
// The three options all default to false, so this changes no value today. It is here because the
// decoder owns the nil path, which is what a bare `"error"` configuration produces, and because a
// fixture building the struct directly would stop exercising the decoder the moment one of these
// options grows a non-zero default.
func decodeJsxKeyOptionsForTest(t *testing.T, raw string) any {
	t.Helper()
	options, err := DecodeJsxKeyOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding options %q: %v", raw, err)
	}
	return options
}

// TestJsxKeySpans pins WHERE each of the six findings points, and three of them are surprising.
//
// `ExpectFindings` asserts ids and counts and nothing else, and this rule has six ids across four
// listeners, so there is a lot of room for a finding to be correct and anchored wrong. Three spans
// here are not the element the reader would guess:
//
//   - `keyBeforeSpread` from the COLLECTION arm points at the CONTAINER, so the whole array literal
//     or the whole parent element, while the same id from the ITERATOR arm points at the element.
//     One id, two anchors, and only a span assertion can tell them apart.
//   - `nonUniqueKeys` points at the KEY ATTRIBUTE rather than at the element carrying it.
//
// Every want below is the slice upstream underlines, taken from the installed build's reported
// columns rather than from reading the source.
func TestJsxKeySpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		want       []string
	}{
		{"missingArrayKey points at the element", "[<App />];", "", []string{"<App />"}},
		{"missingArrayKey skips the keyed sibling", "[<App key={0}/>, <App />];", "",
			[]string{"<App />"}},
		{"missingIterKey points at the element", "[1,2,3].map(x => <App />);", "",
			[]string{"<App />"}},
		// The element inside the parentheses, not the paren, which is where ESLint's column is.
		{"missingIterKey points inside a paren", "[1,2,3].map(x => (<App />));", "",
			[]string{"<App />"}},

		// The container, not the element. This is the collection arm.
		{"keyBeforeSpread from an array points at the array",
			"[<App {...obj} key=\"k\" />];", "{\"checkKeyMustBeforeSpread\":true}",
			[]string{"[<App {...obj} key=\"k\" />]"}},
		{"keyBeforeSpread from children points at the parent",
			"<div><App {...obj} key=\"k\" /></div>;", "{\"checkKeyMustBeforeSpread\":true}",
			[]string{"<div><App {...obj} key=\"k\" /></div>"}},

		// The element, not the container. Same id, iterator arm.
		{"keyBeforeSpread from an iterator points at the element",
			"[1,2,3].map(x => <div {...s} key={x} />);", "{\"checkKeyMustBeforeSpread\":true}",
			[]string{"<div {...s} key={x} />"}},

		{"fragment in an iterator", "[1,2,3].map(x => <>{x}</>);",
			"{\"checkFragmentShorthand\":true}", []string{"<>{x}</>"}},
		{"fragment in an array", "[<></>];", "{\"checkFragmentShorthand\":true}",
			[]string{"<></>"}},

		// The key attribute, not the element.
		{"nonUniqueKeys points at each key attribute",
			"[<span key=\"a\"/>, <span key=\"a\"/>];", "{\"warnOnDuplicates\":true}",
			[]string{"key=\"a\"", "key=\"a\""}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := decodeJsxKeyOptionsForTest(t, testCase.rawOptions)
			result := rule_testing.RunWithOptions(t, JsxKey, jsxKeyFile, testCase.sourceText, options)
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("want %d findings, got %d: %v", len(testCase.want),
					len(result.Diagnostics), result.MessageIds())
			}
			// `rule_testing.RunWithOptions` does not trim, so the source on disk is this literal.
			for index, diagnostic := range result.Diagnostics {
				reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != testCase.want[index] {
					t.Errorf("finding %d span: got %q, want %q", index, reported, testCase.want[index])
				}
			}
		})
	}
}

// TestJsxKeyDuplicateOrderIsStable pins that repeated findings come out in source order.
//
// The duplicate arm groups elements by their key's source text, and a Go map has no order. Reading
// the groups straight out of the map would leave the COUNT fixed and the ORDER varying run to run,
// which no fixture asserting ids can see because every id in the group is the same. The rule keeps
// an insertion-ordered list of keys beside the map for exactly this, and this test is what would
// catch its removal.
//
// Three distinct duplicate groups, so a map with three entries has six possible orders and only one
// is correct.
func TestJsxKeyDuplicateOrderIsStable(t *testing.T) {
	t.Parallel()

	const source = "[<a key=\"one\"/>, <b key=\"two\"/>, <c key=\"one\"/>, <d key=\"three\"/>," +
		" <e key=\"two\"/>, <f key=\"three\"/>];"
	want := []string{
		"key=\"one\"", "key=\"one\"",
		"key=\"two\"", "key=\"two\"",
		"key=\"three\"", "key=\"three\"",
	}

	// Run it several times, because a map that happens to iterate in insertion order once proves
	// nothing. Go randomises map order per run, and with three entries a single run agrees with
	// insertion order about one time in six.
	for attempt := 0; attempt < 12; attempt++ {
		options := decodeJsxKeyOptionsForTest(t, "{\"warnOnDuplicates\":true}")
		result := rule_testing.RunWithOptions(t, JsxKey, jsxKeyFile, source, options)
		if len(result.Diagnostics) != len(want) {
			t.Fatalf("want %d findings, got %d", len(want), len(result.Diagnostics))
		}
		for index, diagnostic := range result.Diagnostics {
			reported := source[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != want[index] {
				t.Fatalf("attempt %d, finding %d: got %q, want %q", attempt, index, reported, want[index])
			}
		}
	}
}

// TestJsxKeyHandlesNilOptions pins the path that bypasses the decoder.
//
// A rule can reach `Run` with nil, and the comma-ok assertion then yields the zero struct, which is
// upstream's defaults here because all three options are false. The test exists because that
// coincidence is not obvious from the code and a later option with a true default would break it
// silently.
func TestJsxKeyHandlesNilOptions(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, JsxKey, jsxKeyFile, "[<App />];", nil)
	rule_testing.ExpectFindings(t, result, "missingArrayKey")

	// And the three option-gated judgments stay OFF, which is the half a nil-options bug would
	// turn on rather than off.
	for _, source := range []string{
		"[1,2,3].map(x => <>{x}</>);",
		"[<App {...obj} key=\"k\" />];",
		"[<span key=\"a\"/>, <span key=\"a\"/>];",
	} {
		rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, JsxKey, jsxKeyFile, source, nil))
	}
}

// TestJsxKeyHasNoFileGate pins that the extension decides nothing.
//
// Upstream registers no filename predicate. Only the three JSX-capable extensions are exercised,
// because a `.ts` file cannot parse `[<App />]` as JSX at all and its silence would say nothing
// about the rule.
func TestJsxKeyHasNoFileGate(t *testing.T) {
	t.Parallel()

	const source = "[<App />];"
	for _, fileName := range []string{
		"/repository/source/Probe.tsx",
		"/repository/source/Probe.jsx",
		"/repository/source/Probe.js",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, JsxKey, fileName, source, nil)
			rule_testing.ExpectFindings(t, result, "missingArrayKey")
		})
	}
}
