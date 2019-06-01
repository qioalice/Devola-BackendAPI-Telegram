// Copyright © 2019. All rights reserved.
// Author: Alice Qio.
// Contacts: <qioalice@gmail.com>.
// License: https://opensource.org/licenses/MIT

package ikba

// EncodedDumpNode is type for method Encoded.dump.
//
// This type represents one node of encoded IKB action Encoded.
// All fields has JSON tags and it's easy to JSON dump output.
//
// Object of this type fills by Encoded.dump method.
//
// More info: Encoded, Encoded.dump.
type EncodedDumpNode struct {

	// Type is a description of IKB encoded node.
	// If node represents encoded argument, Type is a string description
	// of argument's type.
	Type string `json:"type"`

	// Position in encoded IKB action where this node byte view starts.
	Pos byte `json:"pos"`

	// Position in encoded IKB action where encoded argument's header is placed.
	// If node is not about encoded argument, this field is zero.
	PosType byte `json:"pos_type,omitempty"`

	// Position in encoded IKB action where encoded argument's content is placed.
	// If node is not about encoded argument, this field is zero or the same
	// as Pos field.
	// Anyway better use Pos field for non-argument's nodes.
	PosContent byte `json:"pos_value,omitempty"`

	// This is RAW type header of encoded argument if node is about it.
	// Otherwise it is zero.
	TypeHeader byte `json:"type_encoded,omitempty"`

	// Typed RAW value stored in interface{} variable that represents by node.
	// Thus, for example, if node about encoded int8 argument, Value is this
	// int8 argument.
	Value interface{} `json:"value"`
}
