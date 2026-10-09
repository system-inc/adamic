package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus below is oxc's own, copied from `rules_of_hooks.rs` through the extractor rather
// than transcribed, and verified byte against byte after writing. Every case keeps its upstream
// index so a disagreement can be looked up in one step.
//
// Twelve upstream diagnostics are not represented here: they belong to the `useEffectEvent`
// reference-escape family, which this port scopes out and the rule doc explains. The cases
// carrying them are omitted rather than recorded as clean, because recording them as clean would
// assert the opposite of upstream.
func TestRulesOfHooksStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{name: "upstream000", source: `
            function ComponentWithHook() {
              useHook();
            }
        `},
		{name: "upstream001", source: `
            function ComponentWithConditionalHook() {
               if (useHook() ? good() : bad()) {
                    check();
               }
            }
        `},
		{name: "upstream002", source: `
            function Component() {
              if (!useHasPermission()) {
                return null;
              }
              return <Content />;
            }
        `},
		{name: "upstream003", source: `
            function Component() {
              return useHasPermission() ? <Content /> : null;
            }
        `},
		{name: "upstream004", source: `
            function Component() {
              return useHasPermission() && <Content />;
            }
        `},
		{name: "upstream005", source: `
            function Component() {
              if (!useHasPermission()) {
                return null;
              }
              return <Content />;
            }
        `},
		{name: "upstream006", source: `
            function Component() {
              if (useHasPermission() && isAdmin()) {
                return <AdminContent />;
              }
              return <Content />;
            }
        `},
		{name: "upstream007", source: `
            function Component() {
              return (useHasPermission() && isAdmin()) ? <AdminContent /> : <Content />;
            }
        `},
		{name: "upstream008", source: `
            function createComponentWithHook() {
              return function ComponentWithHook() {
                useHook();
              };
            }
        `},
		{name: "upstream009", source: `
            function useHookWithHook() {
              useHook();
            }
        `},
		{name: "upstream010", source: `
            function use2FAMutation() {
              return useState(null);
            }
        `},
		{name: "upstream011", source: `
            function Component() {
              use2FAMutation();
            }
        `},
		{name: "upstream012", source: `
            function use3DEngine() {
              useEffect(() => {}, []);
            }
        `},
		{name: "upstream013", source: `
            function createHook() {
              return function useHookWithHook() {
                useHook();
              }
            }
        `},
		{name: "upstream014", source: `
            function ComponentWithNormalFunction() {
              doSomething();
            }
        `},
		{name: "upstream015", source: `
            function normalFunctionWithNormalFunction() {
              doSomething();
            }
        `},
		{name: "upstream016", source: `
            function normalFunctionWithConditionalFunction() {
              if (cond) {
                doSomething();
              }
            }
        `},
		{name: "upstream017", source: `
            function functionThatStartsWithUseButIsntAHook() {
              if (cond) {
                userFetch();
              }
            }
        `},
		{name: "upstream018", source: `
            function useUnreachable() {
              return;
              useHook();
            }
        `},
		{name: "upstream019", source: `
            function useHook() { useState(); }
            const whatever = function useHook() { useState(); };
            const useHook1 = () => { useState(); };
            let useHook2 = () => useState();
            useHook2 = () => { useState(); };
            ({useHook: () => { useState(); }});
            ({useHook() { useState(); }});
            const {useHook3 = () => { useState(); }} = {};
            ({useHook = () => { useState(); }} = {});
            Namespace.useHook = () => { useState(); };
        `},
		{name: "upstream020", source: `
            function useHook() {
              useHook1();
              useHook2();
            }
        `},
		{name: "upstream021", source: `
            function createHook() {
              return function useHook() {
                useHook1();
                useHook2();
              };
            }
        `},
		{name: "upstream022", source: `
            function useHook() {
              useState() && a;
            }
        `},
		{name: "upstream023", source: `
            function useHook() {
              return useHook1() + useHook2();
            }
        `},
		{name: "upstream024", source: `
            function useHook() {
              return useHook1(useHook2());
            }
        `},
		{name: "upstream025", source: `
            const FancyButton = React.forwardRef((props, ref) => {
              useHook();
              return <button {...props} ref={ref} />
            });
        `},
		{name: "upstream026", source: `
            const FancyButton = React.forwardRef(function (props, ref) {
              useHook();
              return <button {...props} ref={ref} />
            });
        `},
		{name: "upstream027", source: `
            const FancyButton = forwardRef(function (props, ref) {
              useHook();
              return <button {...props} ref={ref} />
            });
        `},
		{name: "upstream028", source: `
            const MemoizedFunction = React.memo(props => {
              useHook();
              return <button {...props} />
            });
        `},
		{name: "upstream029", source: `
            const MemoizedFunction = memo(function (props) {
              useHook();
              return <button {...props} />
            });
        `},
		{name: "upstream030", source: `
            class C {
              m() {
                this.useHook();
                super.useHook();
              }
            }
        `},
		{name: "upstream031", source: `
            jest.useFakeTimers();
            beforeEach(() => {
              jest.useRealTimers();
            })
        `},
		{name: "upstream032", source: `
            fooState();
            _use();
            _useState();
            use_hook();
            // also valid because it's not matching the PascalCase namespace
            jest.useFakeTimer()
        `},
		{name: "upstream033", source: `
            function makeListener(instance) {
              each(pixelsWithInferredEvents, pixel => {
                if (useExtendedSelector(pixel.id) && extendedButton) {
                  foo();
                }
              });
            }
        `},
		{name: "upstream034", source: `
            React.unknownFunction((foo, bar) => {
              if (foo) {
                useNotAHook(bar)
              }
            });
        `},
		{name: "upstream035", source: `
            unknownFunction(function(foo, bar) {
              if (foo) {
                useNotAHook(bar)
              }
            });
        `},
		{name: "upstream036", source: `
            function RegressionTest() {
              const foo = cond ? a : b;
              useState();
            }
        `},
		{name: "upstream037", source: `
            function RegressionTest() {
              if (page == null) {
                throw new Error('oh no!');
              }
              useState();
            }
        `},
		{name: "upstream038", source: `
            function RegressionTest(test) {
              while (test) {
                test = update(test);
              }
              React.useLayoutEffect(() => {});
            }
        `},
		{name: "upstream039", source: `
            function RegressionTest() {
              const res = [];
              const additionalCond = true;
              for (let i = 0; i !== 10 && additionalCond; ++i ) {
                res.push(i);
              }
              React.useLayoutEffect(() => {});
            }
        `},
		{name: "upstream040", source: `
            function MyComponent() {
              // 40 conditions
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}
              if (c) {} else {}

              // 10 hooks
              useHook();
              useHook();
              useHook();
              useHook();
              useHook();
              useHook();
              useHook();
              useHook();
              useHook();
              useHook();
            }
        `},
		{name: "upstream041", source: `
            const useSomeHook = () => {};

            const SomeName = () => {
              const filler = FILLER ?? FILLER ?? FILLER;
              const filler2 = FILLER ?? FILLER ?? FILLER;
              const filler3 = FILLER ?? FILLER ?? FILLER;
              const filler4 = FILLER ?? FILLER ?? FILLER;
              const filler5 = FILLER ?? FILLER ?? FILLER;
              const filler6 = FILLER ?? FILLER ?? FILLER;
              const filler7 = FILLER ?? FILLER ?? FILLER;
              const filler8 = FILLER ?? FILLER ?? FILLER;

              useSomeHook();

              if (anyConditionCanEvenBeFalse) {
                return null;
              }

              return (
                <React.Fragment>
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                  {FILLER ? FILLER : FILLER}
                </React.Fragment>
              );
            };
            `},
		{name: "upstream042", source: `
            function App(props) {
              const someObject = {propA: true};
              for (const propName in someObject) {
                if (propName === true) {
                } else {
                }
              }
              const [myState, setMyState] = useState(null);
            }
        `},
		{name: "upstream043", source: `
            function App() {
              const text = use(Promise.resolve('A'));
              return <Text text={text} />
            }
        `},
		{name: "upstream044", source: `
            import * as React from 'react';
            function App() {
              if (shouldShowText) {
                const text = use(query);
                const data = React.use(thing);
                const data2 = react.use(thing2);
                return <Text text={text} />
              }
              return <Text text={shouldFetchBackupText ? use(backupQuery) : "Nothing to see here"} />
            }
        `},
		{name: "upstream045", source: `
            function App() {
              let data = [];
              for (const query of queries) {
                const text = use(item);
                data.push(text);
              }
              return <Child data={data} />
            }
        `},
		{name: "upstream046", source: `
            function App() {
              const data = someCallback((x) => use(x));
              return <Child data={data} />
            }
        `},
		{name: "upstream047", source: `
            function useLabeledBlock() {
                label: {
                    useHook();
                    if (a) break label;
                }
            }
        `},
		{name: "upstream048", source: `
            export const FalsePositive = ({ editor, anchorElem, isLink, linkNodeUrl, close }: Props) => {
              // This custom hook invocation seems to trigger false positives below
              const [state, setState] = useCustomHook<State>({
                inputLinkUrl: linkNodeUrl ?? '',
                editable: !isLink,
                lastLinkUrl: '',
                lastSelection: null
              });

              const [someThing, setSomeThing] = useState(true);

              const onEdit = useCallback(() => setSomeThing(false), [inputLinkUrl, setSomeThing]);

              const updateLinkEditor = useCallback(() => {
                const rootElement = editor.getRootElement();

                if (nativeSelection.anchorNode === rootElement) {
                  let inner = rootElement;
                  while (inner.firstElementChild !== null) {
                    inner = inner.firstElementChild as HTMLElement;
                  }
                }
              }, [anchorElem, editor, setSomeThing]);

              return <div>test</div>;
            };
        `},
		{name: "upstream049", source: `
            function useLabeledBlock() {
                let x = () => {
                    if (some) {
                        noop();
                    }
                };
                useHook();
            }
        `},
		{name: "upstream050", source: `

            export const Component = () => {
                return {
                    Target: () => {
                        useEffect(() => {
                            return () => {
                                something.value = true;
                            };
                        }, []);
                        return <div></div>;
                    },
                    useTargetModule: (m) => {
                        useModule(m);
                    },
                };
            };
        `},
		{name: "upstream051", source: `
            test.beforeEach(async () => {
                timer = Sinon.useFakeTimers({
                    toFake: ['setInterval'],
                });
            });
    `},
		{name: "upstream052", source: `export default function App() {
       const [state, setState] = useState(0);

       useEffect(() => {
         console.log('Effect called');
       }, []);

       return <div>{state}</div>;
    }
    // https://github.com/toeverything/AFFiNE/blob/0ec1995addbb09fb5d4af765d84cc914b2905150/packages/frontend/core/src/hooks/use-query.ts#L46
    `},
		{name: "upstream053", source: `const createUseQuery =
    (immutable: boolean): useQueryFn =>
    (options, config) => {
        const configWithSuspense: SWRConfiguration = useMemo(
            () => ({
                suspense: true,
                ...config,
            }),
            [config],
        );

        const useSWRFn = immutable ? useSWRImutable : useSWR;
        return useSWRFn(options ? () => ['cloud', options.query.id, options.variables] : null, options ? () => fetcher(options) : null, configWithSuspense);
    };`},
		{name: "upstream054", source: `const MyComponent = makeComponent(() => { useHook(); });`},
		{name: "upstream055", source: `const MyComponent2 = makeComponent(function () { useHook(); });`},
		{name: "upstream056", source: `const MyComponent4 = makeComponent(function InnerComponent() { useHook(); });`},
		{name: "upstream057", source: `const Foo = hoc((props) => { if (props.cond) { const [_a, _b] = useState(false); } });`},
		{name: "upstream058", source: `
        async (_, use) => {
          await use();
        };
    `},
		{name: "upstream059", source: `
        function Foo() {
          try {
            f();
          } catch {}
          useState();
        }
    `},
		{name: "upstream060", source: `
            function MyComponent({ theme }) {
              const onClick = useEffectEvent(() => {
                showNotification(theme);
              });
              useEffect(() => {
                onClick();
              });
              React.useEffect(() => {
                onClick();
              });
            }
        `},
		{name: "upstream061", source: `
            function MyComponent({ theme }) {
              const onClick = useEffectEvent(() => {
                showNotification(theme);
              });
              const onClick2 = useEffectEvent(() => {
                debounce(onClick);
                debounce(() => onClick());
                debounce(() => { onClick() });
                deboucne(() => debounce(onClick));
              });
              useEffect(() => {
                let id = setInterval(() => onClick(), 100);
                return () => clearInterval(onClick);
              }, []);
              React.useEffect(() => {
                let id = setInterval(() => onClick(), 100);
                return () => clearInterval(onClick);
              }, []);
              return null;
            }
        `},
		{name: "upstream062", source: `
            function MyComponent({ theme }) {
              useEffect(() => {
                onClick();
              });
              const onClick = useEffectEvent(() => {
                showNotification(theme);
              });
            }
        `},
		{name: "upstream063", source: `
            function MyComponent({ theme }) {
              const onEvent = useEffectEvent((text) => {
                console.log(text);
              });
              useEffect(() => {
                onEvent('Hello world');
              });
              React.useEffect(() => {
                onEvent('Hello world');
              });
            }
        `},
		{name: "upstream064", source: `
            function MyComponent({ theme }) {
              const onClick = useEffectEvent(() => {
                showNotification(theme);
              });
              useLayoutEffect(() => {
                onClick();
              });
              React.useLayoutEffect(() => {
                onClick();
              });
            }
        `},
		{name: "upstream065", source: `
            function MyComponent({ theme }) {
              const onClick = useEffectEvent(() => {
                showNotification(theme);
              });
              useInsertionEffect(() => {
                onClick();
              });
              React.useInsertionEffect(() => {
                onClick();
              });
            }
        `},
		{name: "upstream066", source: `
            function MyComponent({ theme }) {
              const onClick = useEffectEvent(() => {
                showNotification(theme);
              });
              const onClick2 = useEffectEvent(() => {
                debounce(onClick);
                debounce(() => onClick());
                debounce(() => { onClick() });
                deboucne(() => debounce(onClick));
              });
              useLayoutEffect(() => {
                let id = setInterval(() => onClick(), 100);
                return () => clearInterval(onClick);
              }, []);
              React.useLayoutEffect(() => {
                let id = setInterval(() => onClick(), 100);
                return () => clearInterval(onClick);
              }, []);
              useInsertionEffect(() => {
                let id = setInterval(() => onClick(), 100);
                return () => clearInterval(onClick);
              }, []);
              React.useInsertionEffect(() => {
                let id = setInterval(() => onClick(), 100);
                return () => clearInterval(onClick);
              }, []);
              return null;
            }
        `},
		{name: "upstream067", source: `
            function notAComponent() {
                return new Promise.then(() => {
                    useState();
                });
            }
        `},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, RulesOfHooks, "component.jsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

func TestRulesOfHooksFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		ids    []string
	}{
		{name: "upstream068", source: `
        function ComponentWithConditionalHook() {
               if (cond) {
                 useConditionalHook();
               }
             }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream069", source: `
            function useHook() {
                a && useHook1();
                b && useHook2();
            }
        `, ids: []string{"rulesOfHooksConditional", "rulesOfHooksConditional"}},
		{name: "upstream070", source: `
            function Component() {
                if (condition) {
                    // This is invalid because the hook is called conditionally
                    useHook();
                }
            }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream071", source: `
            function Component() {
                condition ? useHook() : null;
            }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream072", source: `
        function Component() {
          if (Math.random()) {
            return null;
          } else if (!useHasPermission()) {
            return <Foo />
          }
          return <Content />;
        }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream073", source: `
        function Component() {
          switch (foo) {
            case 1:
              useCaseHook();
              break;
            default:
              break;
          }
        }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream074", source: `
        function Component() {
          switch (foo) {
            case 1:
              break;
            default:
              useDefaultHook();
          }
        }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream075", source: `
            Hook.useState();
            Hook._useState();
            Hook.use42();
            Hook.useHook();
            Hook.use_hook();
        `, ids: []string{"rulesOfHooksTopLevel", "rulesOfHooksTopLevel", "rulesOfHooksTopLevel"}},
		{name: "upstream076", source: `
            class C {
                 m() {
                     This.useHook();
                     Super.useHook();
                 }
            }
        `, ids: []string{"rulesOfHooksClassComponent", "rulesOfHooksClassComponent"}},
		{name: "upstream077", source: `
            class Foo extends Component {
                render() {
                    if (cond) {
                        FooStore.useFeatureFlag();
                    }
                }
            }
        `, ids: []string{"rulesOfHooksClassComponent"}},
		{name: "upstream078", source: `
            function ComponentWithConditionalHook() {
                if (cond) {
                    Namespace.useConditionalHook();
                }
            }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream079", source: `
                function createComponent() {
                    return function ComponentWithConditionalHook() {
                        if (cond) {
                            useConditionalHook();
                        }
                    }
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream080", source: `
                function useHookWithConditionalHook() {
                    if (cond) {
                        useConditionalHook();
                    }
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream081", source: `
                function createHook() {
                    return function useHookWithConditionalHook() {
                        if (cond) {
                            useConditionalHook();
                        }
                    }
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream082", source: `
                function ComponentWithTernaryHook() {
                    cond ? useTernaryHook() : null;
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream083", source: `
                function ComponentWithHookInsideCallback() {
                    useEffect(() => {
                        useHookInsideCallback();
                    });
                }
        `, ids: []string{"rulesOfHooksCallback"}},
		{name: "upstream084", source: `
                function createComponent() {
                    return function ComponentWithHookInsideCallback() {
                        useEffect(() => {
                            useHookInsideCallback();
                        });
                    }
                }
        `, ids: []string{"rulesOfHooksCallback"}},
		{name: "upstream085", source: `
                const ComponentWithHookInsideCallback = React.forwardRef((props, ref) => {
                    useEffect(() => {
                        useHookInsideCallback();
                    });
                    return <button {...props} ref={ref} />
                });
        `, ids: []string{"rulesOfHooksCallback"}},
		{name: "upstream086", source: `
                const ComponentWithHookInsideCallback = React.memo(props => {
                    useEffect(() => {
                        useHookInsideCallback();
                    });
                    return <button {...props} />
                });
        `, ids: []string{"rulesOfHooksCallback"}},
		{name: "upstream087", source: `
                function ComponentWithHookInsideCallback() {
                    function handleClick() {
                        useState();
                    }
                }
        `, ids: []string{"rulesOfHooksNotComponent"}},
		{name: "upstream088", source: `
                function createComponent() {
                    return function ComponentWithHookInsideCallback() {
                        function handleClick() {
                            useState();
                        }
                    }
                }
        `, ids: []string{"rulesOfHooksNotComponent"}},
		{name: "upstream089", source: `
                function ComponentWithHookInsideLoop() {
                    while (cond) {
                        useHookInsideLoop();
                    }
                }
        `, ids: []string{"rulesOfHooksLoop"}},
		{name: "upstream090", source: `
            function ComponentWithHookInsideLoop() {
              do {
                useHookInsideLoop();
              } while (cond);
            }
        `, ids: []string{"rulesOfHooksLoop"}},
		{name: "upstream091", source: `
            function ComponentWithHookInsideLoop() {
              do {
                foo();
              } while (useHookInsideLoop());
            }
        `, ids: []string{"rulesOfHooksLoop"}},
		{name: "upstream092", source: `
                function renderItem() {
                    useState();
                }

                function List(props) {
                    return props.items.map(renderItem);
                }
        `, ids: []string{"rulesOfHooksNotComponent"}},
		{name: "upstream093", source: `
                function normalFunctionWithHook() {
                    useHookInsideNormalFunction();
                }
        `, ids: []string{"rulesOfHooksNotComponent"}},
		{name: "upstream094", source: `
                function _normalFunctionWithHook() {
                    useHookInsideNormalFunction();
                }
                function _useNotAHook() {
                    useHookInsideNormalFunction();
                }
        `, ids: []string{"rulesOfHooksNotComponent", "rulesOfHooksNotComponent"}},
		{name: "upstream095", source: `
                function normalFunctionWithConditionalHook() {
                    if (cond) {
                        useHookInsideNormalFunction();
                    }
                }
        `, ids: []string{"rulesOfHooksNotComponent"}},
		{name: "upstream096", source: `
                function useHookInLoops() {
                    while (a) {
                        useHook1();
                        if (b) return;
                        useHook2();
                    }
                    while (c) {
                        useHook3();
                        if (d) return;
                        useHook4();
                    }
                }
        `, ids: []string{"rulesOfHooksLoop", "rulesOfHooksLoop", "rulesOfHooksLoop", "rulesOfHooksLoop"}},
		{name: "upstream097", source: `
            function useHookInLoops() {
                while (a) {
                    useHook1();
                    if (b) continue;
                    useHook2();
                }
            }
        `, ids: []string{"rulesOfHooksLoop", "rulesOfHooksLoop"}},
		{name: "upstream098", source: `
       function useHookInLoops() {
         do {
           useHook1();
           if (a) return;
           useHook2();
         } while (b);

         do {
           useHook3();
           if (c) return;
           useHook4();
         } while (d)
       }
       `, ids: []string{"rulesOfHooksLoop", "rulesOfHooksLoop", "rulesOfHooksLoop", "rulesOfHooksLoop"}},
		{name: "upstream099", source: `
        function useHookInLoops() {
          do {
            useHook1();
            if (a) continue;
            useHook2();
          } while (b);
        }
        `, ids: []string{"rulesOfHooksLoop", "rulesOfHooksLoop"}},
		{name: "upstream100", source: `
                function useLabeledBlock() {
                    label: {
                        if (a) break label;
                        useHook();
                    }
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream101", source: `
            function a() { useState(); }
            const whatever = function b() { useState(); };
            const c = () => { useState(); };
            let d = () => useState();
            e = () => { useState(); };
            ({f: () => { useState(); }});
            ({g() { useState(); }});
            const {j = () => { useState(); }} = {};
            ({k = () => { useState(); }} = {});
        `, ids: []string{"rulesOfHooksNotComponent", "rulesOfHooksNotComponent", "rulesOfHooksNotComponent", "rulesOfHooksNotComponent", "rulesOfHooksNotComponent", "rulesOfHooksNotComponent", "rulesOfHooksNotComponent", "rulesOfHooksNotComponent"}},
		{name: "upstream102", source: `
                function useHook() {
                    if (a) return;
                    useState();
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream103", source: `
                function useHook() {
                    if (a) return;
                    if (b) {
                        console.log('true');
                    } else {
                        console.log('false');
                    }
                    useState();
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream104", source: `
                function useHook() {
                    if (b) {
                        console.log('true');
                    } else {
                        console.log('false');
                    }
                    if (a) return;
                    useState();
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream105", source: `
                function useHook() {
                    a && useHook1();
                    b && useHook2();
                }
        `, ids: []string{"rulesOfHooksConditional", "rulesOfHooksConditional"}},
		{name: "upstream106", source: `
                function useHook() {
                    try {
                        f();
                        useState();
                    } catch {}
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream107", source: `
                function TestChild() {
                    let captured = null;
                    try {
                        captured = useTooltipContext();
                        return null;
                    } catch (error) {
                        return null;
                    }
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream108", source: `
                function ComponentWithHookInsideLoop() {
                    try {
                        while (cond) {
                            useState();
                        }
                    } catch {}
                }
        `, ids: []string{"rulesOfHooksLoop"}},
		{name: "upstream109", source: `
                function Foo() {
                    try {
                        const value = 1;
                        useState(value);
                    } catch {}
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream110", source: `
                function useHook() {
                    try {
                        const value = f();
                        useState(value);
                    } catch {}
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream111", source: `
                function useHook() {
                    try {
                        f(), useState();
                    } catch {}
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream112", source: `
                function useHook() {
                    try {
                        throw err;
                        useState();
                    } catch {}
                }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream113", source: `
                function App({p1, p2}) {
                    try {
                        use(p1);
                    } catch (error) {
                        console.error(error);
                    }
                    use(p2);
                    return <div>App</div>;
                }
        `, ids: []string{"rulesOfHooksTryCatchUse"}},
		{name: "upstream114", source: `
                function App({p1, p2}) {
                    try {
                        doSomething();
                    } catch {
                        use(p1);
                    }
                    use(p2);
                    return <div>App</div>;
                }
        `, ids: []string{"rulesOfHooksTryCatchUse"}},
		{name: "upstream115", source: `
                function useHook({ bar }) {
                    let foo1 = bar && useState();
                    let foo2 = bar || useState();
                    let foo3 = bar ?? useState();
                }
        `, ids: []string{"rulesOfHooksConditional", "rulesOfHooksConditional", "rulesOfHooksConditional"}},
		{name: "upstream116", source: `
                const FancyButton = React.forwardRef((props, ref) => {
                    if (props.fancy) {
                        useCustomHook();
                    }
                    return <button ref={ref}>{props.children}</button>;
                });
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream117", source: `
                const FancyButton = forwardRef(function(props, ref) {
                    if (props.fancy) {
                        useCustomHook();
                    }
                    return <button ref={ref}>{props.children}</button>;
                });
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream118", source: `
                const MemoizedButton = memo(function(props) {
                    if (props.fancy) {
                        useCustomHook();
                    }
                    return <button>{props.children}</button>;
                });
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream119", source: `
                React.unknownFunction(function notAComponent(foo, bar) {
                    useProbablyAHook(bar)
                });
        `, ids: []string{"rulesOfHooksNotComponent"}},
		{name: "upstream120", source: `
            useState();
            if (foo) {
                const foo = React.useCallback(() => {});
            }
            useCustomHook();
        `, ids: []string{"rulesOfHooksTopLevel", "rulesOfHooksTopLevel", "rulesOfHooksTopLevel"}},
		{name: "upstream121", source: `
            const {createHistory, useBasename} = require('history-2.1.2');
            const browserHistory = useBasename(createHistory)({
                basename: '/',
            });
        `, ids: []string{"rulesOfHooksTopLevel"}},
		{name: "upstream122", source: `
                class ClassComponentWithFeatureFlag extends React.Component {
                    render() {
                        if (foo) {
                            useFeatureFlag();
                        }
                    }
                }
        `, ids: []string{"rulesOfHooksClassComponent"}},
		{name: "upstream123", source: `
                class ClassComponentWithHook extends React.Component {
                    render() {
                        React.useState();
                    }
                }
        `, ids: []string{"rulesOfHooksClassComponent"}},
		{name: "upstream124", source: `(class {useHook = () => { useState(); }});`, ids: []string{"rulesOfHooksClassComponent"}},
		{name: "upstream125", source: `(class {useHook() { useState(); }});`, ids: []string{"rulesOfHooksClassComponent"}},
		{name: "upstream126", source: `(class {h = () => { useState(); }});`, ids: []string{"rulesOfHooksClassComponent"}},
		{name: "upstream127", source: `(class {i() { useState(); }});`, ids: []string{"rulesOfHooksClassComponent"}},
		{name: "upstream128", source: `
                async function AsyncComponent() {
                    useState();
                }
        `, ids: []string{"rulesOfHooksAsync"}},
		{name: "upstream129", source: `
                const AsyncComponent = async () => {
                    useState();
                }
        `, ids: []string{"rulesOfHooksAsync"}},
		{name: "upstream130", source: `
                async function Page() {
                  useId();
                  React.useId();
                }
        `, ids: []string{"rulesOfHooksAsync", "rulesOfHooksAsync"}},
		{name: "upstream131", source: `
                async function useAsyncHook() {
                    useState();
                }
        `, ids: []string{"rulesOfHooksAsync"}},
		{name: "upstream132", source: `
                async function notAHook() {
                  useId();
                }
        `, ids: []string{"rulesOfHooksNotComponent"}},
		{name: "upstream133", source: `
            Hook.use();
            Hook._use();
            Hook.useState();
            Hook._useState();
            Hook.use42();
            Hook.useHook();
            Hook.use_hook();
        `, ids: []string{"rulesOfHooksTopLevel", "rulesOfHooksTopLevel", "rulesOfHooksTopLevel", "rulesOfHooksTopLevel"}},
		{name: "upstream134", source: `
                function notAComponent() {
                    use(promise);
                }
        `, ids: []string{"rulesOfHooksNotComponent"}},
		{name: "upstream135", source: `
            const text = use(promise);
            function App() {
                return <Text text={text} />
            }
        `, ids: []string{"rulesOfHooksTopLevel"}},
		{name: "upstream136", source: `
            class C {
                m() {
                    use(promise);
                }
            }
        `, ids: []string{"rulesOfHooksClassComponent"}},
		{name: "upstream137", source: `
            async function AsyncComponent() {
                    use();
            }
        `, ids: []string{"rulesOfHooksAsync"}},
		{name: "upstream138", source: `
            const notAComponent = () => {
                useState();
            }
        `, ids: []string{"rulesOfHooksNotComponent"}},
		{name: "upstream139", source: `
            export default () => {
                if (isVal) {
                    useState(0);
                }
            }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream140", source: `
            export default function() {
                if (isVal) {
                    useState(0);
                }
            }
        `, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream141", source: `
            function Component() {
                return new Promise.then(() => {
                    useState();
                });
            }
        `, ids: []string{"rulesOfHooksCallback"}},
		{name: "upstream142", source: `const MyComponent3 = makeComponent(function foo () { useHook(); });`, ids: []string{"rulesOfHooksNotComponent"}},
		{name: "upstream143", source: `
            function Component() {
                return <Foo>{() => { useState(); }}</Foo>;
            }
        `, ids: []string{"rulesOfHooksCallback"}},
		{name: "upstream144", source: `
            function Component() {
                return <Foo>{props => { useMemo(() => {}, []); }}</Foo>;
            }
        `, ids: []string{"rulesOfHooksCallback"}},
		{name: "upstream145", source: `
            function Component() {
                return <Foo render={() => { useState(); }} />;
            }
        `, ids: []string{"rulesOfHooksCallback"}},
		{name: "upstream146", source: `
            function Component() {
                return <Foo render={props => { useCallback(() => {}, []); }} />;
            }
        `, ids: []string{"rulesOfHooksCallback"}},
		{name: "upstream147", source: `const Foo3 = hoc(function NamedComp(props) { if (props.cond) { const [_a, _b] = useState(false); } });`, ids: []string{"rulesOfHooksConditional"}},
		{name: "upstream150", source: `
            function MyComponent({ theme }) {
                return <Child onClick={useEffectEvent(() => {
                    showNotification(theme);
                })} />;
            }
        `, ids: []string{"rulesOfHooksEffectEventEscape"}},
		{name: "upstream157", source: `function notAComponent() {
  const onEvent = useEffectEvent(() => {});
  return onEvent;
}`, ids: []string{"rulesOfHooksNotComponent"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, RulesOfHooks, "component.jsx", testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

// TestRulesOfHooksSpans pins where each arm points.
//
// Upstream points two different places and `ExpectFindings` cannot see the difference: the
// `notComponent` arm covers the CALLEE alone while every other arm covers the whole call expression.
// Measured on the release oxlint binary with `--format=json` rather than read from the Rust, and
// then confirmed against `call.callee.span()` in `function_error` versus `call.span` elsewhere. A
// port that pointed at the call in every arm passes the whole 150-case fixture set above.
func TestRulesOfHooksSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		reported string
		id       string
	}{
		{
			name:     "conditionalCoversTheWholeCall",
			source:   "function Component() {\n  if (cond) { useHook(1); }\n}\n",
			reported: "useHook(1)",
			id:       "rulesOfHooksConditional",
		},
		{
			name:     "notComponentCoversTheCalleeAlone",
			source:   "function notAComponent() {\n  useHook(1);\n}\n",
			reported: "useHook",
			id:       "rulesOfHooksNotComponent",
		},
		{
			name:     "topLevelCoversTheWholeCall",
			source:   "useHook(1);\n",
			reported: "useHook(1)",
			id:       "rulesOfHooksTopLevel",
		},
		{
			name:     "loopCoversTheWholeCall",
			source:   "function Component() {\n  while (cond) { useHook(1); }\n}\n",
			reported: "useHook(1)",
			id:       "rulesOfHooksLoop",
		},
		{
			name:     "classComponentCoversTheWholeCall",
			source:   "class C {\n  m() { useHook(1); }\n}\n",
			reported: "useHook(1)",
			id:       "rulesOfHooksClassComponent",
		},
		{
			name:     "asyncCoversTheWholeCall",
			source:   "async function Component() {\n  useHook(1);\n}\n",
			reported: "useHook(1)",
			id:       "rulesOfHooksAsync",
		},
		{
			name:     "callbackCoversTheWholeCall",
			source:   "function Component() {\n  useEffect(() => { useHook(1); });\n}\n",
			reported: "useHook(1)",
			id:       "rulesOfHooksCallback",
		},
		{
			name:     "tryCatchUseCoversTheWholeCall",
			source:   "function Component() {\n  try { use(promise); } catch (e) {}\n}\n",
			reported: "use(promise)",
			id:       "rulesOfHooksTryCatchUse",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, RulesOfHooks, "component.jsx", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected exactly one finding, got %d: %v", len(result.Diagnostics), result.MessageIds())
			}
			diagnostic := result.Diagnostics[0]
			// Asserted against a literal typed here rather than against the rule's own message
			// constant, so a mutation moving the id cannot move both sides together.
			if diagnostic.Message.Id != testCase.id {
				t.Errorf("expected id %q, got %q", testCase.id, diagnostic.Message.Id)
			}
			reported := testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != testCase.reported {
				t.Errorf("expected the finding to cover %q, got %q", testCase.reported, reported)
			}
		})
	}
}

// TestRulesOfHooksMessages pins the message identifiers and their descriptions.
//
// `rule.Message` carries no format verbs, so there is nothing rendered and nothing to interpolate;
// the guard is an equality check on the id and a prefix check on the sentence a reader sees. Both
// literals are typed here rather than read off the rule.
func TestRulesOfHooksMessages(t *testing.T) {
	t.Parallel()

	pairs := []struct {
		id         string
		startsWith string
		actualId   string
		actualText string
	}{
		{"rulesOfHooksConditional", "This Hook is not called on every render path", messageRulesOfHooksConditional.Id, messageRulesOfHooksConditional.Description},
		{"rulesOfHooksLoop", "This Hook can run more than once in a single render", messageRulesOfHooksLoop.Id, messageRulesOfHooksLoop.Description},
		{"rulesOfHooksTopLevel", "A Hook called at module scope", messageRulesOfHooksTopLevel.Id, messageRulesOfHooksTopLevel.Description},
		{"rulesOfHooksClassComponent", "A class component does not render through the Hook dispatcher", messageRulesOfHooksClassComponent.Id, messageRulesOfHooksClassComponent.Description},
		{"rulesOfHooksCallback", "This Hook is called inside a callback rather than during render", messageRulesOfHooksCallback.Id, messageRulesOfHooksCallback.Description},
		{"rulesOfHooksNotComponent", "This Hook is called from a plain function", messageRulesOfHooksNotComponent.Id, messageRulesOfHooksNotComponent.Description},
		{"rulesOfHooksAsync", "An async function body resumes after its awaits", messageRulesOfHooksAsync.Id, messageRulesOfHooksAsync.Description},
		{"rulesOfHooksTryCatchUse", "`use` suspends by throwing a promise", messageRulesOfHooksTryCatchUse.Id, messageRulesOfHooksTryCatchUse.Description},
		{"rulesOfHooksEffectEventEscape", "`useEffectEvent` returns a function", messageRulesOfHooksEffectEventEscape.Id, messageRulesOfHooksEffectEventEscape.Description},
	}
	for _, pair := range pairs {
		if pair.actualId != pair.id {
			t.Errorf("expected id %q, got %q", pair.id, pair.actualId)
		}
		if !strings.HasPrefix(pair.actualText, pair.startsWith) {
			t.Errorf("expected the description for %q to start with %q, got %q", pair.id, pair.startsWith, pair.actualText)
		}
	}
}

// TestRulesOfHooksBeyondTheCorpus covers what upstream's corpus does not write, from reading the two
// implementations rather than from the fixtures.
//
// Each case says which reading produced it, because a case invented from the same belief as the code
// tests nothing.
func TestRulesOfHooksBeyondTheCorpus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		ids    []string
		reason string
	}{
		{
			name:   "hookInAFinallyThatCertainlyRuns",
			source: "function Component() {\n  try { f(); } finally { useHook(); }\n}\n",
			ids:    []string{"rulesOfHooksConditional"},
			reason: "Both upstream implementations report this, and the reason THIS rule reports " +
				"it is not the reason it first looks like. The graph lays a `finally` body out " +
				"twice with the same positions and the two copies disagree about whether the Hook " +
				"is on every final path, so which copy wins looks decisive. It is not: a `finally` " +
				"is syntactically inside a `TryStatement`, so the try arm answers above the graph " +
				"and neither copy is ever consulted. Recorded because a mutation flipping the copy " +
				"preference survived this fixture, which is what revealed the branch was subsumed.",
		},
		{
			name:   "bareUseIsAHook",
			source: "function notAComponent() {\n  use(promise);\n}\n",
			ids:    []string{"rulesOfHooksNotComponent"},
			reason: "`internal/utilities/react.IsHookName` answers false for a bare `use`, and both " +
				"upstream implementations answer true. No upstream case writes bare `use` inside a " +
				"badly-named function, so nothing in the imported corpus separates the shelf helper " +
				"from this rule's own predicate on this input.",
		},
		{
			name:   "digitAfterThePrefixIsAHook",
			source: "function notAComponent() {\n  use42();\n}\n",
			ids:    []string{"rulesOfHooksNotComponent"},
			reason: "Same predicate, the other half. The shelf helper tests `unicode.IsUpper` alone " +
				"and answers false for a digit. Upstream case 133 covers `Hook.use42()` through a " +
				"namespace; the bare spelling is not in the corpus.",
		},
		{
			name:   "objectLiteralMethodIsNotAClass",
			source: "({ g() { useState(); } });\n",
			ids:    []string{"rulesOfHooksNotComponent"},
			reason: "An object-literal method parses to the same KindMethodDeclaration a class " +
				"method does. This rule reported `classComponent` here on its first run. The " +
				"corpus does cover it, at case 101, and this case exists so a future edit to " +
				"`isClassMemberFunction` fails against the shape rather than against a nine-finding " +
				"multi-statement fixture where one wrong arm is hard to see.",
		},
		{
			name:   "unreachableHookIsNotThisRulesProblem",
			source: "function Component() {\n  return;\n  useHook();\n}\n",
			ids:    nil,
			reason: "Upstream's `useUnreachable` pass case says so, and this is the same shape " +
				"under a component name rather than a hook name so the silence cannot be coming " +
				"from the name test.",
		},
		{
			name:   "unreachableHookInsideATryStillReports",
			source: "function Component() {\n  try { throw err; useHook(); } catch {}\n}\n",
			ids:    []string{"rulesOfHooksConditional"},
			reason: "The pair to the case above, and the one that fixes their ordering. `try` is " +
				"answered before reachability, so these two shapes differ only in whether a `try` " +
				"wraps them and they reach opposite verdicts.",
		},
		{
			name:   "memoCallbackIsAComponent",
			source: "const C = memo(function () {\n  if (cond) { useHook(); }\n});\n",
			ids:    []string{"rulesOfHooksConditional"},
			reason: "A bare `memo` rather than `React.memo`, which the corpus writes. It must reach " +
				"the conditional arm rather than the callback arm, which is the only thing that " +
				"separates `isNonReactFunctionArgument` from a plain call-argument test.",
		},
		{
			name:   "destructuringAssignmentDefaultSuppliesNoName",
			source: "({k = () => { useState(); }} = {});\n",
			ids:    nil,
			reason: "The pair to the binding-declaration spelling below it. Upstream case 101 " +
				"writes both one line apart and reports only the declaration form, which is the " +
				"only statement in the whole corpus establishing it and is invisible inside a " +
				"nine-finding fixture. Our parser calls this a shorthand property assignment, " +
				"which is exactly the arm `declarationIdentifierOf` refuses to read.",
		},
		{
			name:   "destructuringDeclarationDefaultSuppliesAName",
			source: "const {j = () => { useState(); }} = {};\n",
			ids:    []string{"rulesOfHooksNotComponent"},
			reason: "The reporting half of the pair above, so a change collapsing the two shapes " +
				"into one answer fails whichever way it collapses them.",
		},
		{
			name:   "lowercaseNamespaceIsNotAHookCall",
			source: "function notAComponent() {\n  lowercase.useState();\n}\n",
			ids:    nil,
			reason: "The member spelling requires a component-named object. Upstream case 133 writes " +
				"only `Hook.` as the object, so nothing in the corpus shows what a lowercase " +
				"namespace does.",
		},
		{
			name:   "doWhileBodyIsALoopThoughTheGraphSaysOtherwise",
			source: "function Component() {\n  do { useHook(); } while (cond);\n}\n",
			ids:    []string{"rulesOfHooksLoop"},
			reason: "The graph answers cyclic=false and onEveryFinal=true for this body, both " +
				"correctly, because the back edge enters the test rather than the body. Only the " +
				"syntax says the body repeats. The corpus covers it at case 90; this duplicate " +
				"exists so the syntactic check has a fixture naming what it is for.",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, RulesOfHooks, "component.jsx", testCase.source)
			if len(testCase.ids) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

// TestRulesOfHooksAdditionalHooksChangesNothing pins upstream's inert option: the decoder accepts it,
// and the findings with it are the findings without it, which is what the installed build does.
func TestRulesOfHooksAdditionalHooksChangesNothing(t *testing.T) {
	t.Parallel()

	decoded, err := rule.DecodeOptionsInto[RulesOfHooksOptions]()([]byte(`{"additionalHooks": "useCustomEffect"}`))
	if err != nil {
		t.Fatalf("decoding additionalHooks: %v", err)
	}
	const source = "function MyComponent(props) { if (props.on) { useCustomEffect(() => {}); useState(0); } }"
	with := rule_testing.RunWithOptions(t, RulesOfHooks, rulesOfHooksOptionFile, source, decoded)
	without := rule_testing.Run(t, RulesOfHooks, rulesOfHooksOptionFile, source)
	if strings.Join(with.MessageIds(), ",") != strings.Join(without.MessageIds(), ",") || len(with.Diagnostics) == 0 {
		t.Errorf("with the option %v, without %v; want the same, non-empty", with.MessageIds(), without.MessageIds())
	}

	if _, err := rule.DecodeOptionsInto[RulesOfHooksOptions]()([]byte(`{"additionalEffectHooks": "x"}`)); err == nil {
		t.Error("the settings key decoded as an option; upstream's schema has only additionalHooks")
	}
}

const rulesOfHooksOptionFile = "/repository/source/RulesOfHooksOption.tsx"
