package next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoImgElement = rule.Message{
	Id: "noImgElement",
	Description: "This is a raw <img> element. Use `<Image />` from `next/image` instead, which " +
		"serves a correctly sized image, defers what is offscreen, and reserves the space before " +
		"the bytes arrive. A raw <img> does none of that, so it costs a slower Largest " +
		"Contentful Paint and more bandwidth without failing anything at build time.",
}

// NoImgElement flags a raw <img> element outside a <picture>.
//
//	valid:   <Image src={picture} alt="" />
//	valid:   an <img> inside <picture>
//	invalid: <img src="/test.png" alt="" />
//
// Ported from `@next/next/no-img-element`, read against oxc's `no_img_element.rs`.
//
// # The <picture> exemption, and why the ancestor hop is measured rather than copied
//
// Upstream exempts an <img> whose enclosing element is a <picture>, because art direction through
// <source> siblings is the one case the framework's own component cannot express. That much is
// Vercel's judgment and is preserved.
//
// How the ancestor is reached is not portable. oxc reaches it with `ancestor_kinds(node).nth(1)`,
// a fixed two-hop walk that is correct in its tree and wrong in ours. Measured on our AST:
//
//	<picture><img /></picture>       img.Parent is the JsxElement wrapping the img itself,
//	                                 and its Parent is the <picture> JsxElement
//	<picture><img></img></picture>   the opening element sits one level deeper again
//
// A self-closing <img> and a paired <img></img> therefore reach <picture> at different fixed
// depths, so any hard-coded hop count is right for one form and wrong for the other. The rule walks
// to the nearest enclosing JSX element instead and asks what it is, which is what the fixed hop was
// approximating in a tree where the two forms happen to agree.
var NoImgElement = rule.Rule{
	// No family prefix. The config writes `nextjs/no-img-element` and matching strips the namespace
	// on a `/` boundary, so a rule named `next-no-img-element` matches nothing, runs on no files,
	// and passes every one of its own tests. That is not hypothetical: the first rule in this
	// package shipped that way and was inert.
	Name: "@next/next/no-img-element",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		report := func(node *ast.Node, name *ast.Node) {
			// An intrinsic element is a lowercase, unqualified identifier. A member-expression name
			// (`<Foo.img />`) or a namespaced one is a component reference that never emits an
			// <img>, so comparing the text alone would flag code that renders no image element.
			if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "img" {
				return
			}
			if enclosedByPictureElement(node) {
				return
			}
			ctx.ReportNode(node, messageNoImgElement)
		}

		// Both opening forms are listened to. `<img />` parses as a JsxSelfClosingElement and never
		// produces a JsxOpeningElement, and it is the shape an image is almost always written in,
		// so a rule listening only to the latter is silent on the common case.
		return rule.Listeners{
			ast.KindJsxOpeningElement: func(node *ast.Node) {
				tagName, _ := jsx.ElementParts(node)
				report(node, tagName)
			},
			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				tagName, _ := jsx.ElementParts(node)
				report(node, tagName)
			},
		}
	},
}

// enclosedByPictureElement reports whether the nearest JSX element containing this one is <picture>.
//
// The walk stops at the first enclosing element rather than searching upward for a <picture>
// anywhere above. An <img> nested deeper inside a <picture>, under a wrapping <div>, is not the
// art-direction shape the exemption exists for, and an unbounded search would exempt every image
// on a page whose layout happens to sit inside one.
func enclosedByPictureElement(node *ast.Node) bool {
	// The first hop leaves the JsxElement that wraps this opening element itself, which is why the
	// enclosing element is found at the second. That difference is the whole reason this is a walk
	// rather than a fixed index.
	for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if ancestor.Kind != ast.KindJsxElement {
			continue
		}
		opening := ancestor.AsJsxElement().OpeningElement
		if opening == nil {
			return false
		}
		// The opening form specifically, reached from a JsxElement that already resolved it, so
		// there is no self-closing case here for `jsx.ElementParts` to disambiguate.
		name := opening.AsJsxOpeningElement().TagName
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		// The element wrapping the img itself is skipped: its own tag name is `img`, and treating
		// it as the enclosing element would compare the element to itself.
		if name.Text() == "img" {
			continue
		}
		return name.Text() == "picture"
	}
	return false
}
