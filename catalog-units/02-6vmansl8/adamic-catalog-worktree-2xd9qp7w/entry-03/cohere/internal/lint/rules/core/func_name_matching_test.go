package core

import (
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// funcNameMatchingFile is where the fixtures pretend to live.
//
// A `.ts` name: this rule has no file gate, and nothing in its corpus is JSX.
const funcNameMatchingFile = "/repository/source/FuncNameMatching.ts"

// funcNameMatchingCase is one row of upstream's corpus.
type funcNameMatchingCase struct {
	// name is the corpus list and index the row came from, so a failure names a case that
	// can be found in upstream's own file rather than a number local to this table.
	name string

	// source is upstream's `code`, byte for byte.
	source string

	// options is the RAW JSON of upstream's options object, routed through the rule's own
	// exported decoder rather than built as a struct, so the decoder's defaults and its
	// empty-input path are under test.
	options string

	// ids are the message ids upstream produced for this input, in order.
	ids []string
}

// funcNameMatchingFiresCases are the rows upstream reports on.
var funcNameMatchingFiresCases = []funcNameMatchingCase{
	{
		name:    "invalid-0",
		source:  "let foo = function bar() {};",
		options: "[\"always\"]",
		ids:     []string{"matchVariable"},
	},
	{
		name:    "invalid-1",
		source:  "let foo = function bar() {};",
		options: "",
		ids:     []string{"matchVariable"},
	},
	{
		name:    "invalid-2",
		source:  "foo = function bar() {};",
		options: "",
		ids:     []string{"matchVariable"},
	},
	{
		name:    "invalid-3",
		source:  "foo &&= function bar() {};",
		options: "",
		ids:     []string{"matchVariable"},
	},
	{
		name:    "invalid-4",
		source:  "obj.foo ||= function bar() {};",
		options: "",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-5",
		source:  "obj['foo'] ??= function bar() {};",
		options: "",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-6",
		source:  "obj.foo = function bar() {};",
		options: "",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-7",
		source:  "obj.bar.foo = function bar() {};",
		options: "",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-8",
		source:  "obj['foo'] = function bar() {};",
		options: "",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-9",
		source:  "let obj = {foo: function bar() {}};",
		options: "",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-10",
		source:  "let obj = {'foo': function bar() {}};",
		options: "",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-11",
		source:  "({['foo']: function bar() {}})",
		options: "",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-12",
		source:  "module.exports = function foo(name) {};",
		options: "[{\"includeCommonJSModuleExports\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-13",
		source:  "module.exports = function foo(name) {};",
		options: "[\"always\", {\"includeCommonJSModuleExports\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-14",
		source:  "module.exports = function exports(name) {};",
		options: "[\"never\", {\"includeCommonJSModuleExports\": true}]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-15",
		source:  "module['exports'] = function foo(name) {};",
		options: "[{\"includeCommonJSModuleExports\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-16",
		source:  "module['exports'] = function foo(name) {};",
		options: "[\"always\", {\"includeCommonJSModuleExports\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-17",
		source:  "module['exports'] = function exports(name) {};",
		options: "[\"never\", {\"includeCommonJSModuleExports\": true}]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-18",
		source:  "var foo = function foo(name) {};",
		options: "[\"never\"]",
		ids:     []string{"notMatchVariable"},
	},
	{
		name:    "invalid-19",
		source:  "obj.foo = function foo(name) {};",
		options: "[\"never\"]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-20",
		source:  "Object.defineProperty(foo, 'bar', { value: function baz() {} })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-21",
		source:  "Object.defineProperties(foo, { bar: { value: function baz() {} } })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-22",
		source:  "Object.create(proto, { bar: { value: function baz() {} } })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-23",
		source:  "var obj = { value: function foo(name) {} }",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-24",
		source:  "Object.defineProperty(foo, 'bar', { value: function bar() {} })",
		options: "[\"never\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-25",
		source:  "Object.defineProperties(foo, { bar: { value: function bar() {} } })",
		options: "[\"never\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-26",
		source:  "Object.create(proto, { bar: { value: function bar() {} } })",
		options: "[\"never\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-27",
		source:  "Reflect.defineProperty(foo, 'bar', { value: function baz() {} })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-28",
		source:  "Reflect.defineProperty(foo, 'bar', { value: function bar() {} })",
		options: "[\"never\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-29",
		source:  "foo({ value: function bar() {} })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-30",
		source:  "(obj?.aaa).foo = function bar() {};",
		options: "",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-31",
		source:  "Object?.defineProperty(foo, 'bar', { value: function baz() {} })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-32",
		source:  "(Object?.defineProperty)(foo, 'bar', { value: function baz() {} })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-33",
		source:  "Object?.defineProperty(foo, 'bar', { value: function bar() {} })",
		options: "[\"never\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-34",
		source:  "(Object?.defineProperty)(foo, 'bar', { value: function bar() {} })",
		options: "[\"never\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-35",
		source:  "Object?.defineProperties(foo, { bar: { value: function baz() {} } })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-36",
		source:  "(Object?.defineProperties)(foo, { bar: { value: function baz() {} } })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-37",
		source:  "Object?.defineProperties(foo, { bar: { value: function bar() {} } })",
		options: "[\"never\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-38",
		source:  "(Object?.defineProperties)(foo, { bar: { value: function bar() {} } })",
		options: "[\"never\", {\"considerPropertyDescriptor\": true}]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-39",
		source:  "class C { x = function y() {}; }",
		options: "[\"always\"]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-40",
		source:  "class C { x = function x() {}; }",
		options: "[\"never\"]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-41",
		source:  "class C { 'x' = function y() {}; }",
		options: "[\"always\"]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-42",
		source:  "class C { 'x' = function x() {}; }",
		options: "[\"never\"]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-43",
		source:  "class C { ['x'] = function y() {}; }",
		options: "[\"always\"]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-44",
		source:  "class C { ['x'] = function x() {}; }",
		options: "[\"never\"]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-45",
		source:  "class C { static x = function y() {}; }",
		options: "[\"always\"]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-46",
		source:  "class C { static x = function x() {}; }",
		options: "[\"never\"]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-47",
		source:  "(class { x = function y() {}; })",
		options: "[\"always\"]",
		ids:     []string{"matchProperty"},
	},
	{
		name:    "invalid-48",
		source:  "(class { x = function x() {}; })",
		options: "[\"never\"]",
		ids:     []string{"notMatchProperty"},
	},
	{
		name:    "invalid-49",
		source:  "var obj = { '\\u1885': function foo() {} };",
		options: "",
		ids:     []string{"matchProperty"},
	},
}

// funcNameMatchingSilentCases are the rows upstream is clean on.
var funcNameMatchingSilentCases = []funcNameMatchingCase{
	{
		name:    "valid-0",
		source:  "var foo;",
		options: "",
	},
	{
		name:    "valid-1",
		source:  "var foo = function foo() {};",
		options: "",
	},
	{
		name:    "valid-2",
		source:  "var foo = function foo() {};",
		options: "[\"always\"]",
	},
	{
		name:    "valid-3",
		source:  "var foo = function bar() {};",
		options: "[\"never\"]",
	},
	{
		name:    "valid-4",
		source:  "var foo = function() {}",
		options: "",
	},
	{
		name:    "valid-5",
		source:  "var foo = () => {}",
		options: "",
	},
	{
		name:    "valid-6",
		source:  "foo = function foo() {};",
		options: "",
	},
	{
		name:    "valid-7",
		source:  "foo = function foo() {};",
		options: "[\"always\"]",
	},
	{
		name:    "valid-8",
		source:  "foo = function bar() {};",
		options: "[\"never\"]",
	},
	{
		name:    "valid-9",
		source:  "foo &&= function foo() {};",
		options: "",
	},
	{
		name:    "valid-10",
		source:  "obj.foo ||= function foo() {};",
		options: "",
	},
	{
		name:    "valid-11",
		source:  "obj['foo'] ??= function foo() {};",
		options: "",
	},
	{
		name:    "valid-12",
		source:  "obj.foo = function foo() {};",
		options: "",
	},
	{
		name:    "valid-13",
		source:  "obj.foo = function foo() {};",
		options: "[\"always\"]",
	},
	{
		name:    "valid-14",
		source:  "obj.foo = function bar() {};",
		options: "[\"never\"]",
	},
	{
		name:    "valid-15",
		source:  "obj.foo = function() {};",
		options: "",
	},
	{
		name:    "valid-16",
		source:  "obj.foo = function() {};",
		options: "[\"always\"]",
	},
	{
		name:    "valid-17",
		source:  "obj.foo = function() {};",
		options: "[\"never\"]",
	},
	{
		name:    "valid-18",
		source:  "obj.bar.foo = function foo() {};",
		options: "",
	},
	{
		name:    "valid-19",
		source:  "obj.bar.foo = function foo() {};",
		options: "[\"always\"]",
	},
	{
		name:    "valid-20",
		source:  "obj.bar.foo = function baz() {};",
		options: "[\"never\"]",
	},
	{
		name:    "valid-21",
		source:  "obj['foo'] = function foo() {};",
		options: "",
	},
	{
		name:    "valid-22",
		source:  "obj['foo'] = function foo() {};",
		options: "[\"always\"]",
	},
	{
		name:    "valid-23",
		source:  "obj['foo'] = function bar() {};",
		options: "[\"never\"]",
	},
	{
		name:    "valid-24",
		source:  "obj['foo//bar'] = function foo() {};",
		options: "",
	},
	{
		name:    "valid-25",
		source:  "obj['foo//bar'] = function foo() {};",
		options: "[\"always\"]",
	},
	{
		name:    "valid-26",
		source:  "obj['foo//bar'] = function foo() {};",
		options: "[\"never\"]",
	},
	{
		name:    "valid-27",
		source:  "obj[foo] = function bar() {};",
		options: "",
	},
	{
		name:    "valid-28",
		source:  "obj[foo] = function bar() {};",
		options: "[\"always\"]",
	},
	{
		name:    "valid-29",
		source:  "obj[foo] = function bar() {};",
		options: "[\"never\"]",
	},
	{
		name:    "valid-30",
		source:  "var obj = {foo: function foo() {}};",
		options: "",
	},
	{
		name:    "valid-31",
		source:  "var obj = {foo: function foo() {}};",
		options: "[\"always\"]",
	},
	{
		name:    "valid-32",
		source:  "var obj = {foo: function bar() {}};",
		options: "[\"never\"]",
	},
	{
		name:    "valid-33",
		source:  "var obj = {'foo': function foo() {}};",
		options: "",
	},
	{
		name:    "valid-34",
		source:  "var obj = {'foo': function foo() {}};",
		options: "[\"always\"]",
	},
	{
		name:    "valid-35",
		source:  "var obj = {'foo': function bar() {}};",
		options: "[\"never\"]",
	},
	{
		name:    "valid-36",
		source:  "var obj = {'foo//bar': function foo() {}};",
		options: "",
	},
	{
		name:    "valid-37",
		source:  "var obj = {'foo//bar': function foo() {}};",
		options: "[\"always\"]",
	},
	{
		name:    "valid-38",
		source:  "var obj = {'foo//bar': function foo() {}};",
		options: "[\"never\"]",
	},
	{
		name:    "valid-39",
		source:  "var obj = {foo: function() {}};",
		options: "",
	},
	{
		name:    "valid-40",
		source:  "var obj = {foo: function() {}};",
		options: "[\"always\"]",
	},
	{
		name:    "valid-41",
		source:  "var obj = {foo: function() {}};",
		options: "[\"never\"]",
	},
	{
		name:    "valid-42",
		source:  "var obj = {[foo]: function bar() {}} ",
		options: "",
	},
	{
		name:    "valid-43",
		source:  "var obj = {['x' + 2]: function bar(){}};",
		options: "",
	},
	{
		name:    "valid-44",
		source:  "obj['x' + 2] = function bar(){};",
		options: "",
	},
	{
		name:    "valid-45",
		source:  "var [ bar ] = [ function bar(){} ];",
		options: "",
	},
	{
		name:    "valid-46",
		source:  "function a(foo = function bar() {}) {}",
		options: "",
	},
	{
		name:    "valid-47",
		source:  "module.exports = function foo(name) {};",
		options: "",
	},
	{
		name:    "valid-48",
		source:  "module['exports'] = function foo(name) {};",
		options: "",
	},
	{
		name:    "valid-49",
		source:  "module.exports = function foo(name) {};",
		options: "[{\"includeCommonJSModuleExports\": false}]",
	},
	{
		name:    "valid-50",
		source:  "module.exports = function foo(name) {};",
		options: "[\"always\", {\"includeCommonJSModuleExports\": false}]",
	},
	{
		name:    "valid-51",
		source:  "module.exports = function foo(name) {};",
		options: "[\"never\", {\"includeCommonJSModuleExports\": false}]",
	},
	{
		name:    "valid-52",
		source:  "module['exports'] = function foo(name) {};",
		options: "[{\"includeCommonJSModuleExports\": false}]",
	},
	{
		name:    "valid-53",
		source:  "module['exports'] = function foo(name) {};",
		options: "[\"always\", {\"includeCommonJSModuleExports\": false}]",
	},
	{
		name:    "valid-54",
		source:  "module['exports'] = function foo(name) {};",
		options: "[\"never\", {\"includeCommonJSModuleExports\": false}]",
	},
	{
		name:    "valid-55",
		source:  "({['foo']: function foo() {}})",
		options: "",
	},
	{
		name:    "valid-56",
		source:  "({['foo']: function foo() {}})",
		options: "[\"always\"]",
	},
	{
		name:    "valid-57",
		source:  "({['foo']: function bar() {}})",
		options: "[\"never\"]",
	},
	{
		name:    "valid-58",
		source:  "({['\u2764']: function foo() {}})",
		options: "",
	},
	{
		name:    "valid-59",
		source:  "({[foo]: function bar() {}})",
		options: "",
	},
	{
		name:    "valid-60",
		source:  "({[null]: function foo() {}})",
		options: "",
	},
	{
		name:    "valid-61",
		source:  "({[1]: function foo() {}})",
		options: "",
	},
	{
		name:    "valid-62",
		source:  "({[true]: function foo() {}})",
		options: "",
	},
	{
		name:    "valid-63",
		source:  "({[`x`]: function foo() {}})",
		options: "",
	},
	{
		name:    "valid-64",
		source:  "({[/abc/]: function foo() {}})",
		options: "",
	},
	{
		name:    "valid-65",
		source:  "({[[1, 2, 3]]: function foo() {}})",
		options: "",
	},
	{
		name:    "valid-66",
		source:  "({[{x: 1}]: function foo() {}})",
		options: "",
	},
	{
		name:    "valid-67",
		source:  "[] = function foo() {}",
		options: "",
	},
	{
		name:    "valid-68",
		source:  "({} = function foo() {})",
		options: "",
	},
	{
		name:    "valid-69",
		source:  "[a] = function foo() {}",
		options: "",
	},
	{
		name:    "valid-70",
		source:  "({a} = function foo() {})",
		options: "",
	},
	{
		name:    "valid-71",
		source:  "var [] = function foo() {}",
		options: "",
	},
	{
		name:    "valid-72",
		source:  "var {} = function foo() {}",
		options: "",
	},
	{
		name:    "valid-73",
		source:  "var [a] = function foo() {}",
		options: "",
	},
	{
		name:    "valid-74",
		source:  "var {a} = function foo() {}",
		options: "",
	},
	{
		name:    "valid-75",
		source:  "({ value: function value() {} })",
		options: "[{\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-76",
		source:  "obj.foo = function foo() {};",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-77",
		source:  "obj.bar.foo = function foo() {};",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-78",
		source:  "var obj = {foo: function foo() {}};",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-79",
		source:  "var obj = {foo: function() {}};",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-80",
		source:  "var obj = { value: function value() {} }",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-81",
		source:  "Object.defineProperty(foo, 'bar', { value: function bar() {} })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-82",
		source:  "Object.defineProperties(foo, { bar: { value: function bar() {} } })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-83",
		source:  "Object.create(proto, { bar: { value: function bar() {} } })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-84",
		source:  "Object.defineProperty(foo, 'b' + 'ar', { value: function bar() {} })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-85",
		source:  "Object.defineProperties(foo, { ['bar']: { value: function bar() {} } })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-86",
		source:  "Object.create(proto, { ['bar']: { value: function bar() {} } })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-87",
		source:  "Object.defineProperty(foo, 'bar', { value() {} })",
		options: "[\"never\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-88",
		source:  "Object.defineProperties(foo, { bar: { value() {} } })",
		options: "[\"never\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-89",
		source:  "Object.create(proto, { bar: { value() {} } })",
		options: "[\"never\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-90",
		source:  "Reflect.defineProperty(foo, 'bar', { value: function bar() {} })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-91",
		source:  "Reflect.defineProperty(foo, 'b' + 'ar', { value: function baz() {} })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-92",
		source:  "Reflect.defineProperty(foo, 'bar', { value() {} })",
		options: "[\"never\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-93",
		source:  "foo({ value: function value() {} })",
		options: "[\"always\", {\"considerPropertyDescriptor\": true}]",
	},
	{
		name:    "valid-94",
		source:  "class C { x = function () {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-95",
		source:  "class C { x = function () {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-96",
		source:  "class C { 'x' = function () {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-97",
		source:  "class C { 'x' = function () {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-98",
		source:  "class C { #x = function () {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-99",
		source:  "class C { #x = function () {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-100",
		source:  "class C { [x] = function () {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-101",
		source:  "class C { [x] = function () {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-102",
		source:  "class C { ['x'] = function () {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-103",
		source:  "class C { ['x'] = function () {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-104",
		source:  "class C { x = function x() {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-105",
		source:  "class C { x = function y() {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-106",
		source:  "class C { 'x' = function x() {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-107",
		source:  "class C { 'x' = function y() {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-108",
		source:  "class C { #x = function x() {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-109",
		source:  "class C { #x = function x() {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-110",
		source:  "class C { #x = function y() {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-111",
		source:  "class C { #x = function y() {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-112",
		source:  "class C { [x] = function x() {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-113",
		source:  "class C { [x] = function x() {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-114",
		source:  "class C { [x] = function y() {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-115",
		source:  "class C { [x] = function y() {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-116",
		source:  "class C { ['x'] = function x() {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-117",
		source:  "class C { ['x'] = function y() {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-118",
		source:  "class C { 'xy ' = function foo() {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-119",
		source:  "class C { 'xy ' = function xy() {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-120",
		source:  "class C { ['xy '] = function foo() {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-121",
		source:  "class C { ['xy '] = function xy() {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-122",
		source:  "class C { 1 = function x0() {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-123",
		source:  "class C { 1 = function x1() {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-124",
		source:  "class C { [1] = function x0() {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-125",
		source:  "class C { [1] = function x1() {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-126",
		source:  "class C { [f()] = function g() {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-127",
		source:  "class C { [f()] = function f() {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-128",
		source:  "class C { static x = function x() {}; }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-129",
		source:  "class C { static x = function y() {}; }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-130",
		source:  "class C { x = (function y() {})(); }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-131",
		source:  "class C { x = (function x() {})(); }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-132",
		source:  "(class { x = function x() {}; })",
		options: "[\"always\"]",
	},
	{
		name:    "valid-133",
		source:  "(class { x = function y() {}; })",
		options: "[\"never\"]",
	},
	{
		name:    "valid-134",
		source:  "class C { #x; foo() { this.#x = function x() {}; } }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-135",
		source:  "class C { #x; foo() { this.#x = function x() {}; } }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-136",
		source:  "class C { #x; foo() { this.#x = function y() {}; } }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-137",
		source:  "class C { #x; foo() { this.#x = function y() {}; } }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-138",
		source:  "class C { #x; foo() { a.b.#x = function x() {}; } }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-139",
		source:  "class C { #x; foo() { a.b.#x = function x() {}; } }",
		options: "[\"never\"]",
	},
	{
		name:    "valid-140",
		source:  "class C { #x; foo() { a.b.#x = function y() {}; } }",
		options: "[\"always\"]",
	},
	{
		name:    "valid-141",
		source:  "class C { #x; foo() { a.b.#x = function y() {}; } }",
		options: "[\"never\"]",
	},
}

// decodedFuncNameMatching routes a row's raw JSON through the rule's own exported decoder.
func decodedFuncNameMatching(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeFuncNameMatchingOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding options %q: %v", raw, err)
	}
	return decoded
}

func TestFuncNameMatchingFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range funcNameMatchingFiresCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, FuncNameMatching, funcNameMatchingFile, testCase.source,
				decodedFuncNameMatching(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

func TestFuncNameMatchingStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range funcNameMatchingSilentCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, FuncNameMatching, funcNameMatchingFile, testCase.source,
				decodedFuncNameMatching(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestFuncNameMatchingResolvesAVersionGateTowardEs6 pins a case held out of the tables above.
//
// Upstream's corpus carries `var obj = { 'ᢅ': function foo() {} };` TWICE, byte for byte
// identical, with opposite verdicts:
//
//	invalid[49]   languageOptions: {ecmaVersion: 6}                      reports
//	valid[142]    languageOptions: {ecmaVersion: 5, sourceType: script}  clean
//
// U+1885 entered `Other_ID_Start` in Unicode 8, so it is an identifier under ES6 and not under ES5,
// and `esutils` answers `isIdentifierES5=false, isIdentifierES6=true` for it. That is a version gate
// rather than a contradiction in the corpus, and it is invisible to any extractor that drops
// `languageOptions`.
//
// `rule.Context` carries no `ecmaVersion`, so this rule must answer one way for both rows. It
// answers ES6, deliberately: nothing cohere lints is pinned below ES2015, and choosing ES5 would
// drop findings on modern source. The ES5 row is therefore a KNOWN divergence, recorded here with
// both verdicts rather than removed.
//
// If an `ecmaVersion` channel is ever added, this test is where the choice is written down and the
// held-out row is waiting in upstream's corpus.
func TestFuncNameMatchingResolvesAVersionGateTowardEs6(t *testing.T) {
	t.Parallel()

	const source = "var obj = { 'ᢅ': function foo() {} };"

	result := rule_testing.RunWithOptions(t, FuncNameMatching, funcNameMatchingFile, source,
		decodedFuncNameMatching(t, ""))

	// The ES6 answer, which is upstream's invalid[49].
	rule_testing.ExpectFindings(t, result, "matchProperty")

	// The premise, asserted so this test fails loudly if the predicate ever narrows back to ASCII
	// rather than passing vacuously.
	if !funcNameMatchingIsIdentifier("ᢅ") {
		t.Error("U+1885 is an ES6 identifier start; the predicate no longer says so, which means " +
			"this rule has silently reverted to the ES5 answer for every non-ASCII key")
	}
	// The control. Without it an always-true predicate would satisfy the check above.
	if funcNameMatchingIsIdentifier("0not-an-identifier") {
		t.Error("the identifier predicate accepts everything, so the assertion above proved nothing")
	}
}

// TestDecodeFuncNameMatchingOptions pins the list shapes the config layer delivers and the ones it
// must refuse. Every row above reaches the decoder with a list this file wrote, so this is where the
// refusals are, and where the second element is proved to be read rather than defaulted.
func TestDecodeFuncNameMatchingOptions(t *testing.T) {
	t.Parallel()

	accepted := []struct {
		raw  string
		want FuncNameMatchingOptions
	}{
		{``, FuncNameMatchingOptions{Direction: "always"}},
		{`["never"]`, FuncNameMatchingOptions{Direction: "never"}},
		{`["never", {"considerPropertyDescriptor": true}]`,
			FuncNameMatchingOptions{Direction: "never", ConsiderPropertyDescriptor: true}},
		{`["always", {"includeCommonJSModuleExports": true}]`,
			FuncNameMatchingOptions{Direction: "always", IncludeCommonJSModuleExports: true}},
		{`[{"considerPropertyDescriptor": true}]`,
			FuncNameMatchingOptions{Direction: "always", ConsiderPropertyDescriptor: true}},
	}
	for _, testCase := range accepted {
		decoded, err := DecodeFuncNameMatchingOptions([]byte(testCase.raw))
		if err != nil {
			t.Errorf("%s was refused: %v", testCase.raw, err)
			continue
		}
		if decoded.(FuncNameMatchingOptions) != testCase.want {
			t.Errorf("%s decoded to %+v, want %+v", testCase.raw, decoded, testCase.want)
		}
	}

	for _, raw := range []string{
		`["sometimes"]`,
		`["never", {"considerPropertyDescriptor": true}, "always"]`,
		`[{"considerPropertyDescriptor": true}, "never"]`,
		`["never", {"considerPropertyDescriptors": true}]`,
		`"never"`,
		`{"considerPropertyDescriptor": true}`,
	} {
		if decoded, err := DecodeFuncNameMatchingOptions([]byte(raw)); err == nil {
			t.Errorf("%s decoded to %+v; it must be refused", raw, decoded)
		}
	}
}
