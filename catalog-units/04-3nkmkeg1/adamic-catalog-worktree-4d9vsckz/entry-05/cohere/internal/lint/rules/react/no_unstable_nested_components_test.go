package react

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestNoUnstableNestedComponentsStaysSilent holds every passing case upstream ships,
// imported verbatim from
// eslint-plugin-react/tests/lib/rules/no-unstable-nested-components.js at 7.37.5.
//
// Several carry an upstream comment marking them a known false negative rather than a
// deliberate exemption. Those are reproduced as silent because that is what upstream does,
// and the comment is preserved at the case so the next reader does not read the silence as a
// judgment we made.
func TestNoUnstableNestedComponentsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		source  string
		options *NoUnstableNestedComponentsOptions
	}{
		{
			name: "upstream valid 0",
			source: `
        function ParentComponent() {
          return (
            <div>
              <OutsideDefinedFunctionComponent />
            </div>
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 1",
			source: `
        function ParentComponent() {
          return React.createElement(
            "div",
            null,
            React.createElement(OutsideDefinedFunctionComponent, null)
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 2",
			source: `
        function ParentComponent() {
          return (
            <SomeComponent
              footer={<OutsideDefinedComponent />}
              header={<div />}
              />
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 3",
			source: `
        function ParentComponent() {
          return React.createElement(SomeComponent, {
            footer: React.createElement(OutsideDefinedComponent, null),
            header: React.createElement("div", null)
          });
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 4",
			source: `
        function ParentComponent() {
          const MemoizedNestedComponent = React.useCallback(() => <div />, []);

          return (
            <div>
              <MemoizedNestedComponent />
            </div>
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 5",
			source: `
        function ParentComponent() {
          const MemoizedNestedComponent = React.useCallback(
            () => React.createElement("div", null),
            []
          );

          return React.createElement(
            "div",
            null,
            React.createElement(MemoizedNestedComponent, null)
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 6",
			source: `
        function ParentComponent() {
          const MemoizedNestedFunctionComponent = React.useCallback(
            function () {
              return <div />;
            },
            []
          );

          return (
            <div>
              <MemoizedNestedFunctionComponent />
            </div>
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 7",
			source: `
        function ParentComponent() {
          const MemoizedNestedFunctionComponent = React.useCallback(
            function () {
              return React.createElement("div", null);
            },
            []
          );

          return React.createElement(
            "div",
            null,
            React.createElement(MemoizedNestedFunctionComponent, null)
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 8",
			source: `
        function ParentComponent(props) {
          // Should not interfere handler declarations
          function onClick(event) {
            props.onClick(event.target.value);
          }

          const onKeyPress = () => null;

          function getOnHover() {
            return function onHover(event) {
              props.onHover(event.target);
            }
          }

          return (
            <div>
              <button
                onClick={onClick}
                onKeyPress={onKeyPress}
                onHover={getOnHover()}

                // These should not be considered as components
                maybeComponentOrHandlerNull={() => null}
                maybeComponentOrHandlerUndefined={() => undefined}
                maybeComponentOrHandlerBlank={() => ''}
                maybeComponentOrHandlerString={() => 'hello-world'}
                maybeComponentOrHandlerNumber={() => 42}
                maybeComponentOrHandlerArray={() => []}
                maybeComponentOrHandlerObject={() => {}} />
            </div>
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 9",
			source: `
        function ParentComponent() {
          function getComponent() {
            return <div />;
          }

          return (
            <div>
              {getComponent()}
            </div>
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 10",
			source: `
        function ParentComponent() {
          function getComponent() {
            return React.createElement("div", null);
          }

          return React.createElement("div", null, getComponent());
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 11",
			source: `
        function ParentComponent() {
            return (
              <RenderPropComponent>
                {() => <div />}
              </RenderPropComponent>
            );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 12",
			source: `
        function ParentComponent() {
            return (
              <RenderPropComponent children={() => <div />} />
            );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 13",
			source: `
        function ParentComponent() {
          return (
            <ComplexRenderPropComponent
              listRenderer={data.map((items, index) => (
                <ul>
                  {items[index].map((item) =>
                    <li>
                      {item}
                    </li>
                  )}
                </ul>
              ))
              }
            />
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 14",
			source: `
        function ParentComponent() {
          return React.createElement(
              RenderPropComponent,
              null,
              () => React.createElement("div", null)
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 15",
			source: `
        function ParentComponent(props) {
          return (
            <ul>
              {props.items.map(item => (
                <li key={item.id}>
                  {item.name}
                </li>
              ))}
            </ul>
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 16",
			source: `
        function ParentComponent(props) {
          return (
            <List items={props.items.map(item => {
              return (
                <li key={item.id}>
                  {item.name}
                </li>
              );
            })}
            />
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 17",
			source: `
        function ParentComponent(props) {
          return React.createElement(
            "ul",
            null,
            props.items.map(() =>
              React.createElement(
                "li",
                { key: item.id },
                item.name
              )
            )
          )
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 18",
			source: `
        function ParentComponent(props) {
          return (
            <ul>
              {props.items.map(function Item(item) {
                return (
                  <li key={item.id}>
                    {item.name}
                  </li>
                );
              })}
            </ul>
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 19",
			source: `
        function ParentComponent(props) {
          return React.createElement(
            "ul",
            null,
            props.items.map(function Item() {
              return React.createElement(
                "li",
                { key: item.id },
                item.name
              );
            })
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 20",
			source: `
        function createTestComponent(props) {
          return (
            <div />
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 21",
			source: `
        function createTestComponent(props) {
          return React.createElement("div", null);
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 22",
			source: `
        function ParentComponent() {
          return (
            <ComponentWithProps footer={() => <div />} />
          );
        }
      `,
			options: &NoUnstableNestedComponentsOptions{AllowAsProps: true},
		},
		{
			name: "upstream valid 23",
			source: `
        function ParentComponent() {
          return React.createElement(ComponentWithProps, {
            footer: () => React.createElement("div", null)
          });
        }
      `,
			options: &NoUnstableNestedComponentsOptions{AllowAsProps: true},
		},
		{
			name: "upstream valid 24",
			source: `
        function ParentComponent() {
          return (
            <SomeComponent item={{ children: () => <div /> }} />
          )
        }
      `,
			options: &NoUnstableNestedComponentsOptions{AllowAsProps: true},
		},
		{
			name: "upstream valid 25",
			source: `
      function ParentComponent() {
        return (
          <SomeComponent>
            {
              thing.match({
                renderLoading: () => <div />,
                renderSuccess: () => <div />,
                renderFailure: () => <div />,
              })
            }
          </SomeComponent>
        )
      }
      `,
			options: nil,
		},
		{
			name: "upstream valid 26",
			source: `
      function ParentComponent() {
        const thingElement = thing.match({
          renderLoading: () => <div />,
          renderSuccess: () => <div />,
          renderFailure: () => <div />,
        });
        return (
          <SomeComponent>
            {thingElement}
          </SomeComponent>
        )
      }
      `,
			options: nil,
		},
		{
			name: "upstream valid 27",
			source: `
      function ParentComponent() {
        return (
          <SomeComponent>
            {
              thing.match({
                loading: () => <div />,
                success: () => <div />,
                failure: () => <div />,
              })
            }
          </SomeComponent>
        )
      }
      `,
			options: &NoUnstableNestedComponentsOptions{AllowAsProps: true},
		},
		{
			name: "upstream valid 28",
			source: `
      function ParentComponent() {
        const thingElement = thing.match({
          loading: () => <div />,
          success: () => <div />,
          failure: () => <div />,
        });
        return (
          <SomeComponent>
            {thingElement}
          </SomeComponent>
        )
      }
      `,
			options: &NoUnstableNestedComponentsOptions{AllowAsProps: true},
		},
		{
			name: "upstream valid 29",
			source: `
        function ParentComponent() {
          return (
            <ComponentForProps renderFooter={() => <div />} />
          );
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 30",
			source: `
        function ParentComponent() {
          return React.createElement(ComponentForProps, {
            renderFooter: () => React.createElement("div", null)
          });
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 31",
			source: `
        function ParentComponent() {
          useEffect(() => {
            return () => null;
          });

          return <div />;
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 32",
			source: `
        function ParentComponent() {
          return (
            <SomeComponent renderers={{ Header: () => <div /> }} />
          )
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 33",
			source: `
        function ParentComponent() {
          return (
            <SomeComponent renderMenu={() => (
              <RenderPropComponent>
                {items.map(item => (
                  <li key={item}>{item}</li>
                ))}
              </RenderPropComponent>
            )} />
          )
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 34",
			source: `
        const ParentComponent = () => (
          <SomeComponent
            components={[
              <ul>
                {list.map(item => (
                  <li key={item}>{item}</li>
                ))}
              </ul>,
            ]}
          />
        );
     `,
			options: nil,
		},
		{
			name: "upstream valid 35",
			source: `
        function ParentComponent() {
          const rows = [
            {
              name: 'A',
              render: (props) => <Row {...props} />
            },
          ];

          return <Table rows={rows} />;
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 36",
			source: `
        function ParentComponent() {
          return <SomeComponent renderers={{ notComponent: () => null }} />;
        }
      `,
			options: nil,
		},
		{
			name: "upstream valid 37",
			source: `
        const ParentComponent = createReactClass({
          displayName: "ParentComponent",
          statics: {
            getSnapshotBeforeUpdate: function () {
              return null;
            },
          },
          render() {
            return <div />;
          },
        });
      `,
			options: nil,
		},
		{
			name: "upstream valid 38",
			source: `
        function ParentComponent() {
          const rows = [
            {
              name: 'A',
              notPrefixedWithRender: (props) => <Row {...props} />
            },
          ];

          return <Table rows={rows} />;
        }
      `,
			options: &NoUnstableNestedComponentsOptions{AllowAsProps: true},
		},
		{
			name: "upstream valid 39",
			source: `
        function ParentComponent() {
          return <Table
            rowRenderer={(rowData) => <Row data={data} />}
          />
        }
      `,
			options: &NoUnstableNestedComponentsOptions{PropNamePattern: "*Renderer"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := DefaultNoUnstableNestedComponentsOptions()
			if testCase.options != nil {
				options = *testCase.options
				if options.PropNamePattern == "" {
					options.PropNamePattern = "render*"
				}
			}
			result := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents, "Component.tsx", testCase.source, options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoUnstableNestedComponentsFires holds every failing case upstream ships, imported
// verbatim, with the finding count upstream asserts for each.
func TestNoUnstableNestedComponentsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		source  string
		count   int
		options *NoUnstableNestedComponentsOptions
	}{
		{
			name: "upstream invalid 0",
			source: `
        function ParentComponent() {
          function UnstableNestedFunctionComponent() {
            return <div />;
          }

          return (
            <div>
              <UnstableNestedFunctionComponent />
            </div>
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 1",
			source: `
        function ParentComponent() {
          function UnstableNestedFunctionComponent() {
            return React.createElement("div", null);
          }

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedFunctionComponent, null)
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 2",
			source: `
        function ParentComponent() {
          const UnstableNestedVariableComponent = () => {
            return <div />;
          }

          return (
            <div>
              <UnstableNestedVariableComponent />
            </div>
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 3",
			source: `
        function ParentComponent() {
          const UnstableNestedVariableComponent = () => {
            return React.createElement("div", null);
          }

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedVariableComponent, null)
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 4",
			source: `
        const ParentComponent = () => {
          function UnstableNestedFunctionComponent() {
            return <div />;
          }

          return (
            <div>
              <UnstableNestedFunctionComponent />
            </div>
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 5",
			source: `
        const ParentComponent = () => {
          function UnstableNestedFunctionComponent() {
            return React.createElement("div", null);
          }

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedFunctionComponent, null)
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 6",
			source: `
        export default () => {
          function UnstableNestedFunctionComponent() {
            return <div />;
          }

          return (
            <div>
              <UnstableNestedFunctionComponent />
            </div>
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 7",
			source: `
        export default () => {
          function UnstableNestedFunctionComponent() {
            return React.createElement("div", null);
          }

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedFunctionComponent, null)
          );
        };
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 8",
			source: `
        const ParentComponent = () => {
          const UnstableNestedVariableComponent = () => {
            return <div />;
          }

          return (
            <div>
              <UnstableNestedVariableComponent />
            </div>
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 9",
			source: `
        const ParentComponent = () => {
          const UnstableNestedVariableComponent = () => {
            return React.createElement("div", null);
          }

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedVariableComponent, null)
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 10",
			source: `
        function ParentComponent() {
          class UnstableNestedClassComponent extends React.Component {
            render() {
              return <div />;
            }
          };

          return (
            <div>
              <UnstableNestedClassComponent />
            </div>
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 11",
			source: `
        function ParentComponent() {
          class UnstableNestedClassComponent extends React.Component {
            render() {
              return React.createElement("div", null);
            }
          }

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedClassComponent, null)
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 12",
			source: `
        class ParentComponent extends React.Component {
          render() {
            class UnstableNestedClassComponent extends React.Component {
              render() {
                return <div />;
              }
            };

            return (
              <div>
                <UnstableNestedClassComponent />
              </div>
            );
          }
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 13",
			source: `
        class ParentComponent extends React.Component {
          render() {
            class UnstableNestedClassComponent extends React.Component {
              render() {
                return React.createElement("div", null);
              }
            }

            return React.createElement(
              "div",
              null,
              React.createElement(UnstableNestedClassComponent, null)
            );
          }
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 14",
			source: `
        class ParentComponent extends React.Component {
          render() {
            function UnstableNestedFunctionComponent() {
              return <div />;
            }

            return (
              <div>
                <UnstableNestedFunctionComponent />
              </div>
            );
          }
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 15",
			source: `
        class ParentComponent extends React.Component {
          render() {
            function UnstableNestedClassComponent() {
              return React.createElement("div", null);
            }

            return React.createElement(
              "div",
              null,
              React.createElement(UnstableNestedClassComponent, null)
            );
          }
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 16",
			source: `
        class ParentComponent extends React.Component {
          render() {
            const UnstableNestedVariableComponent = () => {
              return <div />;
            }

            return (
              <div>
                <UnstableNestedVariableComponent />
              </div>
            );
          }
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 17",
			source: `
        class ParentComponent extends React.Component {
          render() {
            const UnstableNestedClassComponent = () => {
              return React.createElement("div", null);
            }

            return React.createElement(
              "div",
              null,
              React.createElement(UnstableNestedClassComponent, null)
            );
          }
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 18",
			source: `
        function ParentComponent() {
          function getComponent() {
            function NestedUnstableFunctionComponent() {
              return <div />;
            };

            return <NestedUnstableFunctionComponent />;
          }

          return (
            <div>
              {getComponent()}
            </div>
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 19",
			source: `
        function ParentComponent() {
          function getComponent() {
            function NestedUnstableFunctionComponent() {
              return React.createElement("div", null);
            }

            return React.createElement(NestedUnstableFunctionComponent, null);
          }

          return React.createElement("div", null, getComponent());
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 20",
			source: `
        function ComponentWithProps(props) {
          return <div />;
        }

        function ParentComponent() {
          return (
            <ComponentWithProps
              footer={
                function SomeFooter() {
                  return <div />;
                }
              } />
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 21",
			source: `
        function ComponentWithProps(props) {
          return React.createElement("div", null);
        }

        function ParentComponent() {
          return React.createElement(ComponentWithProps, {
            footer: function SomeFooter() {
              return React.createElement("div", null);
            }
          });
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 22",
			source: `
        function ComponentWithProps(props) {
          return <div />;
        }

        function ParentComponent() {
            return (
              <ComponentWithProps footer={() => <div />} />
            );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 23",
			source: `
        function ComponentWithProps(props) {
          return React.createElement("div", null);
        }

        function ParentComponent() {
          return React.createElement(ComponentWithProps, {
            footer: () => React.createElement("div", null)
          });
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 24",
			source: `
        function ParentComponent() {
            return (
              <RenderPropComponent>
                {() => {
                  function UnstableNestedComponent() {
                    return <div />;
                  }

                  return (
                    <div>
                      <UnstableNestedComponent />
                    </div>
                  );
                }}
              </RenderPropComponent>
            );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 25",
			source: `
        function RenderPropComponent(props) {
          return props.render({});
        }

        function ParentComponent() {
          return React.createElement(
            RenderPropComponent,
            null,
            () => {
              function UnstableNestedComponent() {
                return React.createElement("div", null);
              }

              return React.createElement(
                "div",
                null,
                React.createElement(UnstableNestedComponent, null)
              );
            }
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 26",
			source: `
        function ComponentForProps(props) {
          return <div />;
        }

        function ParentComponent() {
          return (
            <ComponentForProps notPrefixedWithRender={() => <div />} />
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 27",
			source: `
        function ComponentForProps(props) {
          return React.createElement("div", null);
        }

        function ParentComponent() {
          return React.createElement(ComponentForProps, {
            notPrefixedWithRender: () => React.createElement("div", null)
          });
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 28",
			source: `
        function ParentComponent() {
          return (
            <ComponentForProps someMap={{ Header: () => <div /> }} />
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 29",
			source: `
        class ParentComponent extends React.Component {
          render() {
            const List = (props) => {
              const items = props.items
                .map((item) => (
                  <li key={item.key}>
                    <span>{item.name}</span>
                  </li>
                ));

              return <ul>{items}</ul>;
            };

            return <List {...this.props} />;
          }
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 30",
			source: `
      function ParentComponent() {
        return (
          <SomeComponent>
            {
              thing.match({
                loading: () => <div />,
                success: () => <div />,
                failure: () => <div />,
              })
            }
          </SomeComponent>
        )
      }
      `,
			count:   3,
			options: nil,
		},
		{
			name: "upstream invalid 31",
			source: `
      function ParentComponent() {
        const thingElement = thing.match({
          loading: () => <div />,
          success: () => <div />,
          failure: () => <div />,
        });
        return (
          <SomeComponent>
            {thingElement}
          </SomeComponent>
        )
      }
      `,
			count:   3,
			options: nil,
		},
		{
			name: "upstream invalid 32",
			source: `
      function ParentComponent() {
        const rows = [
          {
            name: 'A',
            notPrefixedWithRender: (props) => <Row {...props} />
          },
        ];

        return <Table rows={rows} />;
      }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 33",
			source: `
        function ParentComponent() {
          const UnstableNestedComponent = React.memo(() => {
            return <div />;
          });

          return (
            <div>
              <UnstableNestedComponent />
            </div>
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 34",
			source: `
        function ParentComponent() {
          const UnstableNestedComponent = React.memo(
            () => React.createElement("div", null),
          );

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedComponent, null)
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 35",
			source: `
        function ParentComponent() {
          const UnstableNestedComponent = React.memo(
            function () {
              return <div />;
            }
          );

          return (
            <div>
              <UnstableNestedComponent />
            </div>
          );
        }
      `,
			count:   1,
			options: nil,
		},
		{
			name: "upstream invalid 36",
			source: `
        function ParentComponent() {
          const UnstableNestedComponent = React.memo(
            function () {
              return React.createElement("div", null);
            }
          );

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedComponent, null)
          );
        }
      `,
			count:   1,
			options: nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := DefaultNoUnstableNestedComponentsOptions()
			if testCase.options != nil {
				options = *testCase.options
				if options.PropNamePattern == "" {
					options.PropNamePattern = "render*"
				}
			}
			result := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents, "Component.tsx", testCase.source, options)
			expected := make([]string, testCase.count)
			for index := range expected {
				expected[index] = "unstableNestedComponent"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// upstreamMessage builds the exact text upstream produces, so the assertions below compare against
// a measured string rather than a paraphrase. Copied from the installed 7.37.5 build's output on
// 2026-08-27 rather than retyped from the rule source.
func upstreamMessage(parentName string, asProps bool) string {
	name := " "
	if parentName != "" {
		name = " “" + parentName + "” "
	}
	message := "Do not define components during render. React will see a new component type on " +
		"every render and destroy the entire subtree’s DOM nodes and state " +
		"(https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, " +
		"move this component definition out of the parent component" + name + "and pass data as props."
	if asProps {
		message += " If you want to allow component creation in props, set allowAsProps option to true."
	}
	return message
}

// TestNoUnstableNestedComponentsSpansAndMessages asserts WHERE each finding points and WHAT it
// says, neither of which ExpectFindings can see.
//
// Every expectation is the installed 7.37.5 build's own output, captured on 2026-08-27 by driving
// the real rule over these exact sources through the ESLint Linter API. Upstream reports columns;
// the reported text is sliced out of the source here, which is the same claim in the form this
// harness can check.
func TestNoUnstableNestedComponentsSpansAndMessages(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		// source is the whole file.
		source string
		// reported is the exact text the finding must span.
		reported string
		// message is the exact rendered description.
		message string
	}{
		{
			name: "names a function declaration parent",
			source: `function ParentComponent() {
  function Nested() { return <div />; }
  return <div><Nested /></div>;
}`,
			reported: `function Nested() { return <div />; }`,
			message:  upstreamMessage("ParentComponent", false),
		},
		{
			// Upstream's resolveComponentName falls back to the binding name only for an ARROW
			// parent, so this one names the parent.
			name: "names an arrow parent through its binding",
			source: `const ParentComponent = () => {
  function Nested() { return <div />; }
  return <div><Nested /></div>;
};`,
			reported: `function Nested() { return <div />; }`,
			message:  upstreamMessage("ParentComponent", false),
		},
		{
			// The same file written with a function expression parent produces the NO-NAME wording,
			// because upstream's fallback is guarded on ArrowFunctionExpression specifically.
			// Measured: this is upstream's behaviour, not ours, and reproducing the narrower guard
			// is what keeps the message text right.
			name: "omits the name for a function expression parent",
			source: `const ParentComponent = function () {
  function Nested() { return <div />; }
  return <div><Nested /></div>;
};`,
			reported: `function Nested() { return <div />; }`,
			message:  upstreamMessage("", false),
		},
		{
			name: "appends the as-props note for a component in a JSX prop",
			source: `function ParentComponent() {
  return <SomeComponent footer={() => <div />} />;
}`,
			reported: `() => <div />`,
			message:  upstreamMessage("ParentComponent", true),
		},
		{
			// Our parser keeps the parenthesis node estree folds away. Upstream reports this at
			// column 34, one past the unparenthesized case, which is the arrow itself rather than
			// the parenthesized expression. Without semanticParentOf skipping the parens, the
			// walk up to the JSX attribute fails and the rule goes silent.
			name: "reports through a parenthesized prop value",
			source: `function ParentComponent() {
  return <SomeComponent footer={(() => <div />)} />;
}`,
			reported: `() => <div />`,
			message:  upstreamMessage("ParentComponent", true),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
				"Component.tsx", testCase.source, DefaultNoUnstableNestedComponentsOptions())
			rule_testing.ExpectFindings(t, result, "unstableNestedComponent")

			// RunTyped writes strings.TrimSpace(source)+"\n" to disk, so slice what the harness
			// actually wrote rather than the literal above.
			onDisk := strings.TrimSpace(testCase.source) + "\n"
			diagnostic := result.Diagnostics[0]
			reported := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != testCase.reported {
				t.Errorf("finding spans %q, want %q", reported, testCase.reported)
			}
			if diagnostic.Message.Description != testCase.message {
				t.Errorf("message\n got %q\nwant %q", diagnostic.Message.Description, testCase.message)
			}
		})
	}
}

// TestNoUnstableNestedComponentsLowercaseParentIsSilent pins the parent-name capitalization test,
// which is NOT the same predicate as this package's hasComponentCapitalization.
//
// Upstream indexes character zero and compares it to its own lowercase, with no underscore
// stripping, so `_ParentComponent` is a lowercase parent and nothing inside it reports. Measured
// silent against the installed build on 2026-08-27. `no-multi-comp` counts `_Foo` AS a component
// through the other predicate, so the two look interchangeable and are not.
func TestNoUnstableNestedComponentsLowercaseParentIsSilent(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		`function _ParentComponent() {
  function Nested() { return <div />; }
  return <div><Nested /></div>;
}`,
		`function createTestComponent() {
  function Nested() { return <div />; }
  return <div><Nested /></div>;
}`,
	} {
		result := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
			"Component.tsx", source, DefaultNoUnstableNestedComponentsOptions())
		rule_testing.ExpectClean(t, result)
	}
}

// TestNoUnstableNestedComponentsRenderMethodGuard pins the claim isInsideRenderMethod's doc comment
// makes: our parser produces no node whose semantic parent is a method declaration, so upstream's
// `node.parent.type === 'MethodDefinition'` guard has no input here.
//
// Written as a structural assertion over the kinds this rule validates rather than as a findings
// check, because a findings check would pass whether the claim held or not.
func TestNoUnstableNestedComponentsRenderMethodGuard(t *testing.T) {
	t.Parallel()

	source := `class ParentComponent extends React.Component {
  render() {
    class Nested extends React.Component {
      render() { return <div />; }
    }
    return <div><Nested /></div>;
  }
}`

	probe := rule.Rule{
		Name:             "react/no-unstable-nested-components",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(sourceFile *ast.Node) {
					validated := 0
					offenders := 0
					var visit func(*ast.Node)
					visit = func(node *ast.Node) {
						switch node.Kind {
						case ast.KindFunctionDeclaration, ast.KindArrowFunction,
							ast.KindFunctionExpression, ast.KindClassDeclaration,
							ast.KindCallExpression:
							validated++
							if parent := semanticParentOf(node); parent != nil &&
								parent.Kind == ast.KindMethodDeclaration {
								offenders++
							}
						}
						node.ForEachChild(func(child *ast.Node) bool {
							visit(child)
							return false
						})
					}
					visit(sourceFile)

					// The control: if the walk found nothing, the zero below would be meaningless.
					if validated == 0 {
						t.Error("no validated nodes found, so the offender count proves nothing")
					}
					if offenders != 0 {
						t.Errorf("%d of %d validated nodes have a method declaration as their "+
							"semantic parent, so the render-method guard is reachable after all",
							offenders, validated)
					}
				},
			}
		},
	}
	rule_testing.RunTyped(t, probe, "Component.tsx", source)
}

// TestDecodeNoUnstableNestedComponentsOptions routes configuration through the rule's own decoder,
// which is where a default inversion would live. The pattern defaults to a NON-EMPTY string, so a
// zero-value struct is a different rule rather than the documented one.
func TestDecodeNoUnstableNestedComponentsOptions(t *testing.T) {
	t.Parallel()

	// A bare "error" arrives as empty input and must land on the documented defaults.
	decoded, err := DecodeNoUnstableNestedComponentsOptions(nil)
	if err != nil {
		t.Fatalf("decoding empty input: %v", err)
	}
	options, ok := decoded.(NoUnstableNestedComponentsOptions)
	if !ok {
		t.Fatalf("decoder returned %T", decoded)
	}
	if options.PropNamePattern != "render*" {
		t.Errorf("empty input gave pattern %q, want render*", options.PropNamePattern)
	}
	if options.AllowAsProps {
		t.Error("empty input enabled allowAsProps")
	}

	// An explicit pattern replaces the default.
	decoded, err = DecodeNoUnstableNestedComponentsOptions([]byte(`{"propNamePattern":"*Renderer"}`))
	if err != nil {
		t.Fatalf("decoding a pattern: %v", err)
	}
	if pattern := decoded.(NoUnstableNestedComponentsOptions).PropNamePattern; pattern != "*Renderer" {
		t.Errorf("pattern %q, want *Renderer", pattern)
	}

	// An explicit empty pattern folds back onto the default, matching upstream's `|| 'render*'`,
	// where the empty string is falsy.
	decoded, err = DecodeNoUnstableNestedComponentsOptions([]byte(`{"propNamePattern":""}`))
	if err != nil {
		t.Fatalf("decoding an empty pattern: %v", err)
	}
	if pattern := decoded.(NoUnstableNestedComponentsOptions).PropNamePattern; pattern != "render*" {
		t.Errorf("empty pattern gave %q, want render*", pattern)
	}

	// customValidators is in upstream's schema and unread by this rule. It must decode rather than
	// error, or a configuration carrying it would fail to load.
	if _, err := DecodeNoUnstableNestedComponentsOptions(
		[]byte(`{"customValidators":["./x"],"allowAsProps":true}`)); err != nil {
		t.Fatalf("decoding customValidators: %v", err)
	}
}

// TestNoUnstableNestedComponentsNilOptions covers the path a bare "error" takes into Run, which
// bypasses the decoder entirely: config.OptionsRegistry hands a non-required rule nil, and a type
// assertion on nil yields the zero value. Without the fallback in Run the empty pattern would match
// no prop name and every render prop would start reporting.
func TestNoUnstableNestedComponentsNilOptions(t *testing.T) {
	t.Parallel()

	source := `function ParentComponent() {
  return <SomeComponent renderFooter={() => <div />} />;
}`
	result := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
		"Component.tsx", source, nil)
	rule_testing.ExpectClean(t, result)
}

// TestNoUnstableNestedComponentsAllowAsProps pins the option against the same source with it off,
// so the case cannot pass by the rule being silent for an unrelated reason.
func TestNoUnstableNestedComponentsAllowAsProps(t *testing.T) {
	t.Parallel()

	source := `function ParentComponent() {
  return <SomeComponent footer={() => <div />} />;
}`

	reporting := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
		"Component.tsx", source, DefaultNoUnstableNestedComponentsOptions())
	rule_testing.ExpectFindings(t, reporting, "unstableNestedComponent")

	allowed := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
		"Component.tsx", source, NoUnstableNestedComponentsOptions{
			AllowAsProps:    true,
			PropNamePattern: "render*",
		})
	rule_testing.ExpectClean(t, allowed)
}

// TestNoUnstableNestedComponentsHasNoFileSuffixGate pins the absence of a `.tsx` gate.
//
// Upstream has no filename condition anywhere, and three rules in this package carried one
// inherited from an oxc port. A React component in a `.ts` file is ordinary and legal, so a gate
// here would cost every finding in every `.ts` file. The two runs differ only in the extension.
func TestNoUnstableNestedComponentsHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	source := `function ParentComponent() {
  const Nested = () => React.createElement("div", null);
  return React.createElement("div", null, React.createElement(Nested, null));
}`

	for _, fileName := range []string{"Component.tsx", "Component.ts"} {
		result := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
			fileName, source, DefaultNoUnstableNestedComponentsOptions())
		if len(result.Diagnostics) != 1 {
			t.Errorf("%s reported %d findings, want 1", fileName, len(result.Diagnostics))
		}
	}
}

// TestNoUnstableNestedComponentsDoesNotPanicOnAwkwardShapes covers the panic class the brief names:
// the walk recovers per FILE, so one nil dereference costs every rule its verdict on that file.
//
// The first shape is the one that actually fired during this port. `createElementCallCounts`
// forwards to an unchecked `node.AsCallExpression()`, and an ancestor walk hands it a
// VariableDeclaration on the way up. Upstream's own passing case 34 produces exactly that.
//
// The rest are shapes whose names cannot be read with Node.Text(): a destructured parameter, a JSX
// attribute whose value is a string rather than an expression container, a spread attribute with no
// name at all, a computed object key, and a namespaced JSX attribute.
func TestNoUnstableNestedComponentsDoesNotPanicOnAwkwardShapes(t *testing.T) {
	t.Parallel()

	sources := []string{
		`function ParentComponent() {
  const rows = [{ notPrefixedWithRender: (props) => <Row {...props} /> }];
  return <Table rows={rows} />;
}`,
		`function ParentComponent() {
  function Nested({ a, b }: Options) { return <div />; }
  return <div><Nested a={1} b={2} /></div>;
}`,
		`function ParentComponent() {
  return <SomeComponent title="plain" footer={() => <div />} />;
}`,
		`function ParentComponent() {
  return <SomeComponent {...rest} footer={() => <div />} />;
}`,
		`function ParentComponent() {
  const handlers = { [computed]: () => <div /> };
  return <SomeComponent handlers={handlers} />;
}`,
		`function ParentComponent() {
  return <svg xmlns:xlink="x"><use xlink:href="#a" /></svg>;
}`,
		`function ParentComponent() {
  return <div>{items.map(([first, ...rest]) => <li key={first} />)}</div>;
}`,
	}

	for index, source := range sources {
		// A panic here fails the test rather than being recovered, which is the assertion.
		result := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
			"Component.tsx", source, DefaultNoUnstableNestedComponentsOptions())
		t.Logf("shape %d produced %d findings", index, len(result.Diagnostics))
	}
}

// TestNoUnstableNestedComponentsHookNamePattern pins upstream's HOOK_REGEXP, `^use[A-Z0-9].*$`.
//
// The hook exemption is reached through the closest enclosing CALL's callee identifier, so a call
// whose name is not hook-shaped must NOT exempt what it returns. Upstream's corpus writes only
// hook-named calls in this position, so a mutation widening the pattern to match every name
// survived the whole imported set.
//
// Every row measured against the installed 7.37.5 build on 2026-08-27. Note `use` alone reports:
// the pattern requires a capital or a digit AFTER `use`, so the bare name does not match.
func TestNoUnstableNestedComponentsHookNamePattern(t *testing.T) {
	t.Parallel()

	cases := []struct {
		calleeName string
		findings   int
	}{
		{calleeName: "useThing", findings: 0},
		{calleeName: "use", findings: 1},
		{calleeName: "uselower", findings: 1},
		{calleeName: "notAHook", findings: 1},
		// A digit satisfies the character class as surely as a capital does.
		{calleeName: "use2", findings: 0},
	}

	for _, testCase := range cases {
		t.Run(testCase.calleeName, func(t *testing.T) {
			t.Parallel()
			source := `function ParentComponent() {
  const X = ` + testCase.calleeName + `(() => {
    return () => <div />;
  });
  return <div><X /></div>;
}`
			result := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
				"Component.tsx", source, DefaultNoUnstableNestedComponentsOptions())
			if len(result.Diagnostics) != testCase.findings {
				t.Errorf("%s gave %d findings, want %d",
					testCase.calleeName, len(result.Diagnostics), testCase.findings)
			}
		})
	}
}

// TestNoUnstableNestedComponentsPropNamePattern pins the option against the same source under two
// patterns, so neither row can pass by the rule being silent for an unrelated reason.
//
// The corpus configures propNamePattern exactly once, on a passing case, so nothing imported can
// see the pattern being consulted rather than ignored.
func TestNoUnstableNestedComponentsPropNamePattern(t *testing.T) {
	t.Parallel()

	source := `function ParentComponent() {
  return <SomeComponent someRenderer={() => <div />} />;
}`

	// Under the default `render*`, `someRenderer` is not a render prop, so it reports.
	def := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
		"Component.tsx", source, DefaultNoUnstableNestedComponentsOptions())
	rule_testing.ExpectFindings(t, def, "unstableNestedComponent")

	// Under `*Renderer` it is one, so it is silent. This is upstream's own valid case 39's option.
	configured := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
		"Component.tsx", source, NoUnstableNestedComponentsOptions{PropNamePattern: "*Renderer"})
	rule_testing.ExpectClean(t, configured)
}

// TestNoUnstableNestedComponentsAttributeValueKinds covers JSX attribute values that are not
// expression containers: a string value and a bare boolean attribute with no value at all.
//
// Both rows are here for the panic surface rather than for the predicate. A bare attribute has NO
// value node, so reading it through anything but the typed accessor is a nil dereference, and the
// walk recovers per FILE rather than per rule.
//
// # What these rows do NOT prove, stated because my first version of this comment claimed they did
//
// I wrote them expecting them to kill a mutation that drops the container check. They do not, and
// re-running found that out. The reason is that `Nested` is declared in the parent's BODY, so its
// ancestry is Block > FunctionDeclaration > SourceFile and it never walks near a JSX attribute at
// all. The predicate is not consulted for either row.
//
// Measured against the installed build on 2026-08-27, the guard's real job is to stop the walk
// CLIMBING PAST a string-valued attribute toward an outer one that is a container:
//
//	<Outer footer={() => <div />} />                        reports, with the as-props note
//	<Outer footer={<Inner title="x" render={...} />} />     silent
//	<Outer footer={<Inner title="x">{...}</Inner>} />       silent
//
// Both of the shapes that would exercise the climb are silent upstream for a different reason, so
// no fixture built from them can separate the two versions either. The mutation stays a live
// survivor and is recorded as such in the rule's own comment rather than papered over here.
func TestNoUnstableNestedComponentsAttributeValueKinds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name: "string valued attribute",
			source: `function ParentComponent() {
  function Nested() { return <div />; }
  return <SomeComponent title="x"><Nested /></SomeComponent>;
}`,
		},
		{
			name: "bare boolean attribute with no value",
			source: `function ParentComponent() {
  function Nested() { return <div />; }
  return <SomeComponent flag><Nested /></SomeComponent>;
}`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
				"Component.tsx", testCase.source, DefaultNoUnstableNestedComponentsOptions())
			rule_testing.ExpectFindings(t, result, "unstableNestedComponent")
			if strings.Contains(result.Diagnostics[0].Message.Description, "allowAsProps") {
				t.Errorf("a component declared in the parent body must not be read as a prop, but "+
					"the finding carried the as-props note: %q",
					result.Diagnostics[0].Message.Description)
			}
		})
	}
}

// TestNoUnstableNestedComponentsAttributeClimbIsBounded pins the three shapes above as a group.
//
// The first reports and the other two are silent, all three measured against the installed 7.37.5
// build. Held together in one test because the contrast is the content: the only difference between
// the first and the rest is whether an element with a string attribute sits between the component
// and the attribute that would exempt it.
func TestNoUnstableNestedComponentsAttributeClimbIsBounded(t *testing.T) {
	t.Parallel()

	reporting := `function ParentComponent() {
  return <Outer footer={() => <div />} />;
}`
	result := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
		"Component.tsx", reporting, DefaultNoUnstableNestedComponentsOptions())
	rule_testing.ExpectFindings(t, result, "unstableNestedComponent")

	for _, source := range []string{
		`function ParentComponent() {
  return <Outer footer={<Inner title="x" render={() => <div />} />} />;
}`,
		`function ParentComponent() {
  return <Outer footer={<Inner title="x">{() => <div />}</Inner>} />;
}`,
	} {
		silent := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
			"Component.tsx", source, DefaultNoUnstableNestedComponentsOptions())
		rule_testing.ExpectClean(t, silent)
	}
}

// TestNoUnstableNestedComponentsInsideClassComponents pins the class recovery path, which upstream
// needs because its detection pass misses a function component declared inside a class.
//
// Both rows measured against the installed 7.37.5 build on 2026-08-27, and the second is the one
// worth having: upstream does NOT confine the recovery to `render`, so a component built in an
// ordinary method reports too.
func TestNoUnstableNestedComponentsInsideClassComponents(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name: "arrow component in a render method",
			source: `class ParentComponent extends React.Component {
  render() {
    const Nested = () => <div />;
    return <div><Nested /></div>;
  }
}`,
		},
		{
			name: "arrow component in a method that is not render",
			source: `class ParentComponent extends React.Component {
  helper() {
    const Nested = () => <div />;
    return <Nested />;
  }
  render() { return <div>{this.helper()}</div>; }
}`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
				"Component.tsx", testCase.source, DefaultNoUnstableNestedComponentsOptions())
			rule_testing.ExpectFindings(t, result, "unstableNestedComponent")
		})
	}
}

// TestNoUnstableNestedComponentsCreateElementPropsPosition pins the props argument of a
// createElement call, which is the SECOND one.
//
// Measured against the installed build on 2026-08-27: a component in the props object reports with
// the as-props note, and one reached through the type slot does too, because a direct property
// value is claimed by an earlier branch before the createElement route is consulted.
//
// A mutation changing the argument index from 1 to 0 SURVIVES, and that is not a gap these rows
// close. Instrumented over all 77 corpus cases plus these: the createElement route is reached 92
// times, exactly one of those reaches a call carrying two or more arguments, and for that one both
// indices answer false. Every other reach short-circuits in `isComponentInProp`'s property branch
// first. A distinguishing input needs a DETECTED component inside a props object that is not itself
// a direct property value, and I could not construct one upstream also detects. Recorded as
// unresolved rather than as an equivalence verdict.
func TestNoUnstableNestedComponentsCreateElementPropsPosition(t *testing.T) {
	t.Parallel()

	source := `function ParentComponent() {
  return React.createElement(Some, { footer: () => React.createElement("div", null) });
}`
	result := rule_testing.RunTypedWithOptions(t, NoUnstableNestedComponents,
		"Component.tsx", source, DefaultNoUnstableNestedComponentsOptions())
	rule_testing.ExpectFindings(t, result, "unstableNestedComponent")
	if !strings.Contains(result.Diagnostics[0].Message.Description, "allowAsProps") {
		t.Errorf("a component in a createElement props object must carry the as-props note, got %q",
			result.Diagnostics[0].Message.Description)
	}
}

// TestPropNamePatternIsCompiledOncePerPattern holds why compilePropNamePattern keeps its matchers:
// `Run` asks for one per file. A pattern seen before must come back without compiling again, which
// is an allocation-free lookup, and the matcher it returns must still be the pattern's.
//
// Not parallel: testing.AllocsPerRun reads the process-wide allocation count, so a test running
// beside it would add its allocations to this one's.
func TestPropNamePatternIsCompiledOncePerPattern(t *testing.T) {
	compiled := testing.AllocsPerRun(10, func() {
		buildPropNamePattern("render*")
	})
	if compiled < 1 {
		t.Fatalf("compiling a pattern allocated %.0f times; the comparison below cannot fail", compiled)
	}
	compilePropNamePattern("cached*")
	repeated := testing.AllocsPerRun(10, func() {
		compilePropNamePattern("cached*")
	})
	if repeated != 0 {
		t.Errorf("asking again for a pattern already seen allocated %.0f times, want 0", repeated)
	}
	matches := compilePropNamePattern("cached*")
	if !matches("cachedValue") || matches("renderValue") {
		t.Error("the kept matcher does not match its own pattern")
	}
}
