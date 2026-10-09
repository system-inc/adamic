package next

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The fail cases are oxc's own, plus the paired `<img></img>` form, which our tree reaches at a
// different depth than the self-closing one and which oxc's corpus does not contain.
func TestNoImgElementReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name: "inside a div",
			source: `export class MyComponent {
  render() {
    return (
      <div>
        <img src="/test.png" alt="Test picture" width={500} height={500} />
      </div>
    );
  }
}`,
		},
		{
			name: "returned directly",
			source: `export class MyComponent {
  render() {
    return <img src="/test.png" alt="Test picture" width={500} height={500} />;
  }
}`,
		},
		{
			name: "src is not a string literal",
			source: `import somePicture from './foo.png';
export const MyComponent = () => <img src={somePicture.src} alt='foo' />;`,
		},
		{
			// The paired form reaches its ancestors one level deeper than the self-closing form. A
			// fixed-hop exemption is right for one and wrong for the other, so both are pinned.
			name:   "paired form",
			source: `export const C = () => <div><img src="/test.png"></img></div>;`,
		},
		{
			// A <picture> above a wrapping element is not the art-direction shape the exemption is
			// for. This pins the walk stopping at the nearest enclosing element.
			name:   "inside a div inside a picture",
			source: `export const C = () => <picture><div><img src="/test.png" /></div></picture>;`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoImgElement, "Component.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, messageNoImgElement.Id)
			if result.Diagnostics[0].Message.Id != "noImgElement" {
				t.Fatalf("unexpected message id %q", result.Diagnostics[0].Message.Id)
			}
		})
	}
}

// The pass cases are oxc's own. The <picture> ones are the exemption, and the `<Image />` one is
// the replacement the rule exists to push toward, which must never itself report.
func TestNoImgElementIsSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name: "next/image component",
			source: `import { Image } from 'next/image';
export class MyComponent {
  render() {
    return (
      <div>
        <Image src="/test.png" alt="Test picture" width={500} height={500} />
      </div>
    );
  }
}`,
		},
		{
			name: "directly inside a picture",
			source: `export class MyComponent {
  render() {
    return (
      <picture>
        <img src="/test.png" alt="Test picture" width={500} height={500} />
      </picture>
    );
  }
}`,
		},
		{
			name: "inside a picture with a source sibling",
			source: `export class MyComponent {
  render() {
    return (
      <div>
        <picture>
          <source media="(min-width:650px)" srcSet="/test.jpg" />
          <img src="/test.png" alt="Test picture" />
        </picture>
      </div>
    );
  }
}`,
		},
		{
			// A component named through a member expression never emits an <img>, so matching on
			// the text alone would flag code that renders nothing of the kind.
			name:   "member expression name",
			source: `export const C = () => <Foo.img src="/test.png" />;`,
		},
		{
			name:   "paired form inside a picture",
			source: `export const C = () => <picture><img src="/test.png"></img></picture>;`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoImgElement, "Component.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}
