// Package generate streams clang's Objective-C JSON AST into checked bindings.
package generate

import (
	"encoding/json"
	"fmt"
	"io"
)

type writtenType struct {
	Qual      string `json:"qualType"`
	Desugared string `json:"desugaredQualType"`
}
type reference struct {
	Name string `json:"name"`
}
type location struct {
	File           string
	Offset, Length int
	Valid          bool
}
type node struct {
	Kind, Name, Value, Tag                                          string
	Location, Begin, End                                            location
	Type, Result, Underlying                                        writtenType
	Super, Interface, Getter, Setter                                reference
	Instance, Implicit, Variadic, Readonly, ClassProperty, Complete bool
	Children                                                        []*node
	Framework                                                       string
	// KeepNeedlessWords is the generator's, not clang's: a method whose shortened name would
	// collide with a sibling's keeps its full one (see emitClass).
	KeepNeedlessWords bool
	// SetterOnly is the generator's too: a property's setter offered as a method, since its
	// getter's result can't cross (see emitClass).
	SetterOnly bool
	// Unowned is a property whose value the object doesn't hold: weak, assign or unsafe_unretained
	// (NSView's window, its superview). Leaves read it (leaves.go).
	Unowned bool
}
type astStream struct {
	decoder *json.Decoder
	file    string
}

// readAST consumes the translation unit's inner array individually. No raw JSON
// translation unit or complete syntax tree is retained. Even discarded nodes
// are walked because their locations affect the next declaration's file delta.
func readAST(input io.Reader, visit func(*node) error) error {
	s := &astStream{decoder: json.NewDecoder(input)}
	if err := s.expect('{'); err != nil {
		return err
	}
	kind := ""
	for s.decoder.More() {
		token, err := s.decoder.Token()
		if err != nil {
			return err
		}
		key := token.(string)
		switch key {
		case "loc":
			if _, err := s.readLocation(); err != nil {
				return err
			}
		case "range":
			if err := s.readRange(&node{}); err != nil {
				return err
			}
		case "kind":
			if err := s.decoder.Decode(&kind); err != nil {
				return err
			}
		case "inner":
			if kind != "TranslationUnitDecl" {
				return fmt.Errorf("expected TranslationUnitDecl, got %q", kind)
			}
			if err := s.expect('['); err != nil {
				return err
			}
			for s.decoder.More() {
				n, err := s.readNode()
				if err != nil {
					return err
				}
				if err := visit(n); err != nil {
					return err
				}
			}
			if err := s.expect(']'); err != nil {
				return err
			}
		default:
			if err := s.discard(); err != nil {
				return err
			}
		}
	}
	if err := s.expect('}'); err != nil {
		return err
	}
	if kind != "TranslationUnitDecl" {
		return fmt.Errorf("expected translation unit")
	}
	if _, err := s.decoder.Token(); err != io.EOF {
		return fmt.Errorf("trailing JSON: %v", err)
	}
	return nil
}
func (s *astStream) expect(want rune) error {
	token, err := s.decoder.Token()
	if err != nil {
		return err
	}
	if token != json.Delim(want) {
		return fmt.Errorf("JSON: expected %c, got %v", want, token)
	}
	return nil
}
func (s *astStream) discard() error {
	token, err := s.decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); ok {
		for s.decoder.More() {
			if delimiter == '{' {
				key, err := s.decoder.Token()
				if err != nil {
					return err
				}
				if key == "loc" {
					if _, err := s.readLocation(); err != nil {
						return err
					}
					continue
				}
				if key == "range" {
					if err := s.readRange(&node{}); err != nil {
						return err
					}
					continue
				}
			}
			if err := s.discard(); err != nil {
				return err
			}
		}
		_, err = s.decoder.Token()
	}
	return err
}
func (s *astStream) readLocation() (location, error) {
	result := location{File: s.file}
	if err := s.expect('{'); err != nil {
		return result, err
	}
	for s.decoder.More() {
		token, err := s.decoder.Token()
		if err != nil {
			return result, err
		}
		switch token.(string) {
		case "file":
			if err := s.decoder.Decode(&s.file); err != nil {
				return result, err
			}
			result.File = s.file
		case "offset":
			if err := s.decoder.Decode(&result.Offset); err != nil {
				return result, err
			}
			result.Valid = true
		case "tokLen":
			if err := s.decoder.Decode(&result.Length); err != nil {
				return result, err
			}
		case "spellingLoc":
			if _, err := s.readLocation(); err != nil {
				return result, err
			}
		case "expansionLoc":
			expansion, err := s.readLocation()
			if err != nil {
				return result, err
			}
			result = expansion
		// includedFrom is the including file, not the location's own file. It
		// neither consumes nor supplies clang's source-location filename delta.
		default:
			if err := s.discard(); err != nil {
				return result, err
			}
		}
	}
	return result, s.expect('}')
}
func (s *astStream) readRange(n *node) error {
	if err := s.expect('{'); err != nil {
		return err
	}
	for s.decoder.More() {
		token, err := s.decoder.Token()
		if err != nil {
			return err
		}
		switch token.(string) {
		case "begin":
			n.Begin, err = s.readLocation()
		case "end":
			n.End, err = s.readLocation()
		default:
			err = s.discard()
		}
		if err != nil {
			return err
		}
	}
	return s.expect('}')
}
func (s *astStream) readNode() (*node, error) {
	n := &node{}
	if err := s.expect('{'); err != nil {
		return nil, err
	}
	for s.decoder.More() {
		token, err := s.decoder.Token()
		if err != nil {
			return nil, err
		}
		switch token.(string) {
		case "kind":
			err = s.decoder.Decode(&n.Kind)
		case "name":
			err = s.decoder.Decode(&n.Name)
		case "value":
			var value json.Token
			value, err = s.decoder.Token()
			if err == nil {
				switch v := value.(type) {
				case string:
					n.Value = v
				case bool:
					n.Value = fmt.Sprint(v)
				case float64:
					n.Value = fmt.Sprint(v)
				default:
					err = fmt.Errorf("unexpected AST value %v", value)
				}
			}
		case "tagUsed":
			err = s.decoder.Decode(&n.Tag)
		case "type":
			err = s.decoder.Decode(&n.Type)
		case "returnType":
			err = s.decoder.Decode(&n.Result)
		case "fixedUnderlyingType":
			err = s.decoder.Decode(&n.Underlying)
		case "super":
			err = s.decoder.Decode(&n.Super)
		case "interface":
			err = s.decoder.Decode(&n.Interface)
		case "getter":
			err = s.decoder.Decode(&n.Getter)
		case "setter":
			err = s.decoder.Decode(&n.Setter)
		case "instance":
			err = s.decoder.Decode(&n.Instance)
		case "isImplicit":
			err = s.decoder.Decode(&n.Implicit)
		case "variadic":
			err = s.decoder.Decode(&n.Variadic)
		case "readonly":
			err = s.decoder.Decode(&n.Readonly)
		case "weak", "assign", "unsafe_unretained":
			var set bool
			err = s.decoder.Decode(&set)
			n.Unowned = n.Unowned || set
		case "class":
			err = s.decoder.Decode(&n.ClassProperty)
		case "completeDefinition":
			err = s.decoder.Decode(&n.Complete)
		case "loc":
			n.Location, err = s.readLocation()
		case "range":
			err = s.readRange(n)
		case "inner":
			if err = s.expect('['); err == nil {
				for s.decoder.More() {
					var child *node
					child, err = s.readNode()
					if err != nil {
						break
					}
					if keepChildren(n.Kind) {
						n.Children = append(n.Children, child)
					}
				}
				if err == nil {
					err = s.expect(']')
				}
			}
		default:
			err = s.discard()
		}
		if err != nil {
			return nil, err
		}
	}
	if err := s.expect('}'); err != nil {
		return nil, err
	}
	return n, nil
}

// Keep only declaration facts and evaluated enum expressions. Function bodies,
// typedef type trees and arbitrary expressions are visited then discarded.
func keepChildren(kind string) bool {
	switch kind {
	case "ObjCInterfaceDecl", "ObjCProtocolDecl", "ObjCCategoryDecl", "ObjCMethodDecl", "ObjCPropertyDecl", "ParmVarDecl", "FunctionDecl", "EnumDecl", "EnumConstantDecl", "RecordDecl", "TypedefDecl", "FieldDecl", "ConstantExpr", "ImplicitCastExpr":
		return true
	}
	return false
}
