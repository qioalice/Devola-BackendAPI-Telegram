// Copyright © 2019. All rights reserved.
// Author: Alice Qio.
// Contacts: <qioalice@gmail.com>.
// License: https://opensource.org/licenses/MIT

package ikba

import (
	"unsafe"

	"github.com/qioalice/devola/core/chat"
	"github.com/qioalice/devola/core/view"
)

// TODO: Add negative index GetArg support.
// TODO: Add named arguments.

// Encoded is the internal type that represents encoded
// a Telegram Inline Keyboard Button (IKB) action.
//
// Action must have an ID of action, a SSID this button links with and
// may contains arguments.
//
// At this moment (April, 2019) Telegram API (v4.1) allow represent
// action of Inline Keyboard Button as some string that must have length
// not more than 64 byte.
//
// The encode/decode algorithm described below.
//
// Encoded view of Encoded:
//
// < Action ID : sizeof(Encoded) (now 4 byte) >
// < Session ID : sizeof(tSessionID) (now 4 byte) >
// < Args count : 1 byte >
// < Index over last encoded argument : 1 byte >
// < Arg 1 Type : 1 byte >
// < Arg 1 Value : N bytes (depends by Arg 1 Type) > ...
//
// ATTENTION!
// DO NOT FORGET CALL init METHOD OF EACH NEW Encoded INSTANCE!
type Encoded [64]byte

// Predefined position constants that helps to perform encode/decode operations.
const (

	// Position in Encoded where View ID starts from.
	cPosViewID byte = 0

	// Position in Encoded where Session ID starts from.
	cPosSessionID byte = 4

	// Position in Encoded where encoded arguments' part starts from.
	cPosArgs byte = 8

	// Position in Encoded where encoded arguments' counter is.
	cPosArgsCount = cPosArgs + 0

	// Position in Encoded where saved the position
	// starts from a next encoded argument can be placed.
	cPosArgsFree = cPosArgs + 1

	// Position in Encoded where encoded arguments' data starts from.
	cPosArgsContent = cPosArgs + 2

	// Max allowable position in Encoded.
	cPosMax byte = 63

	// Error position value.
	// Returned from some methods.
	cPosErr byte = ^byte(0)
)

// Predefined index constants that means a special cases of encode/decode operations.
const (

	// Bad argument's index.
	// Each encoded argument in Encoded has its own index.
	// Some methods should return that index.
	// If any error occurred, this value indicates it.
	cBadIndex int = -1
)

// Predefined argument's type constants that helps represents (encode/decode) arguments.
//
// ATTENTION!
// DO NOT FORGET ADD ALL NEW CONSTANTS TO THE argNextFromPos METHOD'S SWITCH!
// DO NOT FORGET ADD ALL NEW CONSTANTS TO THE argType2S METHOD'S SWITCH!
// DO NOT FORGET ADD BEHAVIOUR FOR NEW TYPES TO THE dump METHOD!
//
// ATTENTION!
// DO NOT OVERFLOW INT8 (1<<7) -1 (127). BECAUSE!
const (

	// Header of int8 argument
	cArgTypeInt8 byte = 10

	// Header of int16 argument
	cArgTypeInt16 byte = 11

	// Header of int32 argument
	cArgTypeInt32 byte = 12

	// Header of int64 argument
	cArgTypeInt64 byte = 13

	// Header of uint8 argument
	cArgTypeUint8 byte = 14

	// Header of uint16 argument
	cArgTypeUint16 byte = 15

	// Header of uint32 argument
	cArgTypeUint32 byte = 16

	// Header of uint64 argument
	cArgTypeUint64 byte = 17

	// Header of float32 argument
	cArgTypeFloat32 byte = 18

	// Header of float64 argument
	cArgTypeFloat64 byte = 19

	// Header of string argument
	cArgTypeString byte = 20
)

// ext1byte extracts 1 byte from encoded IKB action d starts from startPos
// and returns it.
func (d *Encoded) ext1byte(startPos byte) (v int8) {
	return int8(d[startPos])
}

// ext2bytes extracts 2 bytes from encoded IKB action d starts from startPos
// and returns it.
func (d *Encoded) ext2bytes(startPos byte) (v int16) {
	v |= int16(d[startPos+0]) << 0
	v |= int16(d[startPos+1]) << 8
	return v
}

// ext4bytes extracts 4 bytes from encoded IKB action d starts from startPos
// and returns it.
func (d *Encoded) ext4bytes(startPos byte) (v int32) {
	v |= int32(d[startPos+0]) << 0
	v |= int32(d[startPos+1]) << 8
	v |= int32(d[startPos+2]) << 16
	v |= int32(d[startPos+3]) << 24
	return v
}

// ext8bytes extracts 8 bytes from encoded IKB action d starts from startPos
// and returns it.
func (d *Encoded) ext8bytes(startPos byte) (v int64) {
	v |= int64(d[startPos+0]) << 0
	v |= int64(d[startPos+1]) << 8
	v |= int64(d[startPos+2]) << 16
	v |= int64(d[startPos+3]) << 24
	v |= int64(d[startPos+4]) << 32
	v |= int64(d[startPos+5]) << 40
	v |= int64(d[startPos+6]) << 48
	v |= int64(d[startPos+7]) << 56
	return v
}

// extNbytes extracts N bytes from encoded IKB action d starts from startPos
// and returns it.
func (d *Encoded) extNbytes(startPos, bytes byte) []byte {
	v := make([]byte, bytes)
	for i := byte(0); i < bytes; i++ {
		v[i] = d[startPos+i]
	}
	return v
}

// put1byte puts 1 byte v to the encoded IKB action d starts from startPos.
func (d *Encoded) put1byte(startPos byte, v int8) {
	d[startPos] = byte(v)
}

// put2bytes puts 2 bytes v to the encoded IKB action d starts from startPos.
func (d *Encoded) put2bytes(startPos byte, v int16) {
	d[startPos+0] = byte(v >> 0)
	d[startPos+1] = byte(v >> 8)
}

// put4bytes puts 4 bytes v to the encoded IKB action d starts from startPos.
func (d *Encoded) put4bytes(startPos byte, v int32) {
	d[startPos+0] = byte(v >> 0)
	d[startPos+1] = byte(v >> 8)
	d[startPos+2] = byte(v >> 16)
	d[startPos+3] = byte(v >> 24)
}

// put8bytes puts 8 bytes v to the encoded IKB action d starts from startPos.
func (d *Encoded) put8bytes(startPos byte, v int64) {
	d[startPos+0] = byte(v >> 0)
	d[startPos+1] = byte(v >> 8)
	d[startPos+2] = byte(v >> 16)
	d[startPos+3] = byte(v >> 24)
	d[startPos+4] = byte(v >> 32)
	d[startPos+5] = byte(v >> 40)
	d[startPos+6] = byte(v >> 48)
	d[startPos+7] = byte(v >> 56)
}

// putNbytes puts N bytes v to the encoded IKB action d starts from startPos.
func (d *Encoded) putNbytes(startPos byte, v []byte) {
	for i, n := byte(0), byte(len(v)); i < n; i++ {
		d[startPos+i] = v[i]
	}
}

// PutViewID puts the encoded IKB action ID to the current IKB action d.
func (d *Encoded) PutViewID(id view.IDEnc) {
	d.put4bytes(cPosViewID, int32(id))
}

// GetViewID extracts the encoded IKB action ID from the current IKB action d
// and returns it.
func (d *Encoded) GetViewID() view.IDEnc {
	return view.IDEnc(d.ext4bytes(cPosViewID))
}

// PutSessionID puts the session ID IKB linked with to the current IKB action d.
func (d *Encoded) PutSessionID(ssid chat.SessionID) {
	d.put4bytes(cPosSessionID, int32(ssid))
	return
}

// GetSessionID extract the encoded IKB session ID from the current IKB action d
// and returns it.
func (d *Encoded) GetSessionID() chat.SessionID {
	return chat.SessionID(d.ext4bytes(cPosSessionID))
}

// needForType returns the number of bytes that required to store
// an argument's value with type argType.
//
// WARNING!
// Only for any integer or float type.
// Calls with other type's constant will return a very big value.
func (*Encoded) argNeedForType(argType byte) (numBytes byte) {

	switch argType {

	case cArgTypeInt8,
		cArgTypeUint8:
		return 2

	case cArgTypeInt16,
		cArgTypeUint16:
		return 3

	case cArgTypeInt32,
		cArgTypeUint32,
		cArgTypeFloat32:
		return 5

	case cArgTypeInt64,
		cArgTypeUint64,
		cArgTypeFloat64:
		return 9

	default:
		return cPosMax
	}
}

// argHaveFreeBytes returns true only if numBytes bytes of some argument
// can be saved into current encoded action. Otherwise false is returned.
func (d *Encoded) argHaveFreeBytes(numBytes byte) bool {
	return d[cPosArgsFree]+numBytes <=
		cPosMax
}

// argReserveForType reserves the number of bytes for argument with type argType
// (increases an internal free index position counter), and returns
// a position you can write bytes starting from which.
//
// WARNING!
// If argument with required type can't be stored (no more space),
// cPosErr is returned!
//
// WARNING!
// Only for any integer or float type.
// Otherwise cPosErr is returned!
func (d *Encoded) argReserveForType(argType byte) (startPos byte) {

	// Check whether d has as many free bytes as argType is required.
	requiredBytes := d.argNeedForType(argType)
	if requiredBytes >= cPosMax {
		return cPosErr
	}

	// Extract current start pos and next if it will be correct
	startPos = d[cPosArgsFree]
	nextStartPos := startPos + requiredBytes

	// Check whether nextStartPos <= max allowable position
	if nextStartPos > cPosMax {
		return cPosErr
	}

	// Save arg type, inc start pos counter
	d[startPos] = argType
	d[cPosArgsFree] = nextStartPos
	return startPos + 1
}

// argGet returns a position where argument's content with type argType
// starts from. The search begins from idx argument index.
//
// If index is too long, argument not exists or something wrong else,
// cPosErr is returned.
//
// Example:
//
// d has encoded arguments in that order:
// 0:int32, 1:int16, 2:string, 3:int8, 4:int8
//
// Calls:
//
// argGet(0, int32) == pos of content of 0 arg.
// argGet(0, string) == pos of content of 2 arg.
// argGet(4, int8) == pos of content of 4 arg.
//
// All types presented above are constants, of course.
func (d *Encoded) argGet(argIdx int, argType byte) (startPos byte) {

	// Check index is valid
	if d.ArgCount() <= argIdx {
		return cPosErr
	}

	// Skip unnecessary arguments (argIdx -1)
	startPos = cPosArgsContent
	for argIdx--; argIdx > 0; argIdx-- {
		startPos = d.argNextFromPos(startPos)
	}

	// Try to find required argument
	nextFreeIndex := d[cPosArgsFree]
	for startPos != cPosErr && startPos < nextFreeIndex {
		if d[startPos] == argType {
			// Found, return argument's content position
			return startPos + 1
		}
		// Go to next arg
		startPos = d.argNextFromPos(startPos)
	}

	// Not found
	return cPosErr
}

// argNextFromPos returns the next argument's position in d if pos is
// position of some argument.
//
// WARNING! Do not have any check!
func (d *Encoded) argNextFromPos(pos byte) (nextArgPos byte) {

	switch d[pos] {

	case cArgTypeInt8,
		cArgTypeUint8:
		return pos + 2

	case cArgTypeInt16,
		cArgTypeUint16:
		return pos + 3

	case cArgTypeInt32,
		cArgTypeUint32,
		cArgTypeFloat32:
		return pos + 5

	case cArgTypeInt64,
		cArgTypeUint64,
		cArgTypeFloat64:
		return pos + 9

	case cArgTypeString:
		// d[pos] - arg type string, d[pos+1] - len of string
		return pos + 2 + d[pos+1]

	default:
		// THIS IS ERROR SWITCH BRANCH!
		// DO NOT "PUT" ANY CASES TO THIS BRANCH!
		//
		// it should never happen, but if it will, let it be safe
		// not just pos, because it may cause infinity loop in caller's
		// not pos + too big C, because it may cause seg fault
		return cPosErr
	}
}

// argType2S returns a string name of type argType.
func (d *Encoded) argType2S(argType byte) string {

	switch argType {

	case cArgTypeInt8:
		return "int8"

	case cArgTypeInt16:
		return "int16"

	case cArgTypeInt32:
		return "int32"

	case cArgTypeInt64:
		return "int64"

	case cArgTypeUint8:
		return "uint8"

	case cArgTypeUint16:
		return "uint16"

	case cArgTypeUint32:
		return "uint32"

	case cArgTypeUint64:
		return "uint64"

	case cArgTypeFloat32:
		return "float32"

	case cArgTypeFloat64:
		return "float64"

	case cArgTypeString:
		return "string"

	default:
		return "UNKNOWN"
	}
}

// ArgCount returns the number of stored arguments in encoded IKB action d.
func (d *Encoded) ArgCount() (num int) {
	return int(d[cPosArgsCount])
}

// argCountIncPostfix increases the number of stored arguments in encoded
// IKB action d and returns the value before increasing.
//
// (postfix increase operator that Golang don't have).
func (d *Encoded) argCountIncPostfix() (oldValue int) {
	oldValue = int(d[cPosArgsCount])
	d[cPosArgsCount]++
	return oldValue
}

// PutArgInt puts int argument v to the encoded IKB action d.
// If it was successfully, returns the index of that argument.
// Otherwise -1 is returned (argument has not been added).
func (d *Encoded) PutArgInt(v int) (argIdx int) {
	return d.PutArgInt32(int32(v))
}

// PutArgInt8 puts int8 argument v to the encoded IKB action d.
//
// If it was successfully, returns the index of that argument.
// Otherwise -1 is returned (argument has not been added).
func (d *Encoded) PutArgInt8(v int8) (argIdx int) {

	startPos := d.argReserveForType(cArgTypeInt8)
	if startPos == cPosErr {
		return cBadIndex
	}

	d.put1byte(startPos, v)
	return d.argCountIncPostfix()
}

// PutArgInt16 puts int16 argument v to the encoded IKB action d.
//
// If it was successfully, returns the index of that argument.
// Otherwise -1 is returned (argument has not been added).
func (d *Encoded) PutArgInt16(v int16) (argIdx int) {

	startPos := d.argReserveForType(cArgTypeInt16)
	if startPos == cPosErr {
		return cBadIndex
	}

	d.put2bytes(startPos, v)
	return d.argCountIncPostfix()
}

// PutArgInt32 puts int32 argument v to the encoded IKB action d.
//
// If it was successfully, returns the index of that argument.
// Otherwise -1 is returned (argument has not been added).
func (d *Encoded) PutArgInt32(v int32) (argIdx int) {

	startPos := d.argReserveForType(cArgTypeInt32)
	if startPos == cPosErr {
		return cBadIndex
	}

	d.put4bytes(startPos, v)
	return d.argCountIncPostfix()
}

// PutArgInt64 puts int64 argument v to the encoded IKB action d.
//
// If it was successfully, returns the index of that argument.
// Otherwise -1 is returned (argument has not been added).
func (d *Encoded) PutArgInt64(v int64) (argIdx int) {

	startPos := d.argReserveForType(cArgTypeInt64)
	if startPos == cPosErr {
		return cBadIndex
	}

	d.put8bytes(startPos, v)
	return d.argCountIncPostfix()
}

// PutArgUint puts uint argument v to the encoded IKB action d.
//
// If it was successfully, returns the index of that argument.
// Otherwise -1 is returned (argument has not been added).
func (d *Encoded) PutArgUint(v uint) (argIdx int) {

	return d.PutArgUint32(uint32(v))
}

// PutArgUint8 puts uint8 argument v to the encoded IKB action d.
//
// If it was successfully, returns the index of that argument.
// Otherwise -1 is returned (argument has not been added).
func (d *Encoded) PutArgUint8(v uint8) (argIdx int) {

	startPos := d.argReserveForType(cArgTypeUint8)
	if startPos == cPosErr {
		return cBadIndex
	}

	d.put1byte(startPos, int8(v))
	return d.argCountIncPostfix()
}

// PutArgUint16 puts uint16 argument v to the encoded IKB action d.
//
// If it was successfully, returns the index of that argument.
// Otherwise -1 is returned (argument has not been added).
func (d *Encoded) PutArgUint16(v uint16) (argIdx int) {

	startPos := d.argReserveForType(cArgTypeUint16)
	if startPos == cPosErr {
		return cBadIndex
	}

	d.put2bytes(startPos, int16(v))
	return d.argCountIncPostfix()
}

// PutArgUint32 puts uint32 argument v to the encoded IKB action d.
//
// If it was successfully, returns the index of that argument.
// Otherwise -1 is returned (argument has not been added).
func (d *Encoded) PutArgUint32(v uint32) (argIdx int) {

	startPos := d.argReserveForType(cArgTypeUint32)
	if startPos == cPosErr {
		return cBadIndex
	}

	d.put4bytes(startPos, int32(v))
	return d.argCountIncPostfix()
}

// PutArgUint64 puts uint64 argument v to the encoded IKB action d.
//
// If it was successfully, returns the index of that argument.
// Otherwise -1 is returned (argument has not been added).
func (d *Encoded) PutArgUint64(v uint64) (argIdx int) {

	startPos := d.argReserveForType(cArgTypeUint64)
	if startPos == cPosErr {
		return cBadIndex
	}
	d.put8bytes(startPos, int64(v))

	return d.argCountIncPostfix()
}

// PutArgFloat32 puts float32 argument v to the encoded IKB action d.
//
// If it was successfully, returns the index of that argument.
// Otherwise -1 is returned (argument has not been added).
func (d *Encoded) PutArgFloat32(v float32) (argIdx int) {

	startPos := d.argReserveForType(cArgTypeFloat32)
	if startPos == cPosErr {
		return cBadIndex
	}

	d.put4bytes(startPos, *(*int32)(unsafe.Pointer(&v)))
	return d.argCountIncPostfix()
}

// PutArgFloat64 puts float64 argument v to the encoded IKB action d.
//
// If it was successfully, returns the index of that argument.
// Otherwise -1 is returned (argument has not been added).
func (d *Encoded) PutArgFloat64(v float64) (argIdx int) {

	startPos := d.argReserveForType(cArgTypeFloat64)
	if startPos == cPosErr {
		return cBadIndex
	}

	d.put8bytes(startPos, *(*int64)(unsafe.Pointer(&v)))
	return d.argCountIncPostfix()
}

// PutArgString puts string argument v to the encoded IKB action d.
//
// If it was successfully, returns the index of that argument.
// Otherwise -1 is returned (argument has not been added).
func (d *Encoded) PutArgString(v string) (argIdx int) {

	// String encoding: Arg Type byte, string len byte, string content
	strlen := byte(len(v))
	if !d.argHaveFreeBytes(2 + strlen) {
		return cBadIndex
	}

	// Get start pos, update free index for next argument
	startPos := d[cPosArgsFree]
	d[cPosArgsFree] += strlen + 2

	// Save arg type, save string len
	d[startPos+0] = cArgTypeString
	d[startPos+1] = strlen

	// Save string content
	d.putNbytes(startPos+2, []byte(v))
	return d.argCountIncPostfix()
}

// GetArgInt extracts int argument from encoded IKB action d,
// starting search from startIdx argument's index.
//
// Returns it and true as success if it is, or zero value and false if error.
func (d *Encoded) GetArgInt(startIdx int) (v int, success bool) {

	var vv int32
	vv, success = d.GetArgInt32(startIdx)
	return int(vv), success
}

// GetArgInt8 extracts int8 argument from encoded IKB action d,
// starting search from startIdx argument's index.
//
// Returns it and true as success if it is, or zero value and false if error.
func (d *Encoded) GetArgInt8(startIdx int) (v int8, success bool) {

	startPos := d.argGet(startIdx, cArgTypeInt8)
	if startPos == cPosErr {
		return 0, false
	}
	return d.ext1byte(startPos), true
}

// GetArgInt16 extracts int16 argument from encoded IKB action d,
// starting search from startIdx argument's index.
//
// Returns it and true as success if it is, or zero value and false if error.
func (d *Encoded) GetArgInt16(startIdx int) (v int16, success bool) {

	startPos := d.argGet(startIdx, cArgTypeInt16)
	if startPos == cPosErr {
		return 0, false
	}
	return d.ext2bytes(startPos), true
}

// GetArgInt32 extracts int32 argument from encoded IKB action d,
// starting search from startIdx argument's index.
//
// Returns it and true as success if it is, or zero value and false if error.
func (d *Encoded) GetArgInt32(startIdx int) (v int32, success bool) {

	startPos := d.argGet(startIdx, cArgTypeInt32)
	if startPos == cPosErr {
		return 0, false
	}
	return d.ext4bytes(startPos), true
}

// GetArgInt64 extracts int64 argument from encoded IKB action d,
// starting search from startIdx argument's index.
//
// Returns it and true as success if it is, or zero value and false if error.
func (d *Encoded) GetArgInt64(startIdx int) (v int64, success bool) {

	startPos := d.argGet(startIdx, cArgTypeInt64)
	if startPos == cPosErr {
		return 0, false
	}
	return d.ext8bytes(startPos), true
}

// GetArgUint extracts uint argument from encoded IKB action d,
// starting search from startIdx argument's index.
//
// Returns it and true as success if it is, or zero value and false if error.
func (d *Encoded) GetArgUint(startIdx int) (v uint, success bool) {

	var vv uint32
	vv, success = d.GetArgUint32(startIdx)
	return uint(vv), success
}

// GetArgUint8 extracts uint8 argument from encoded IKB action d,
// starting search from startIdx argument's index.
//
// Returns it and true as success if it is, or zero value and false if error.
func (d *Encoded) GetArgUint8(startIdx int) (v uint8, success bool) {

	startPos := d.argGet(startIdx, cArgTypeUint8)
	if startPos == cPosErr {
		return 0, false
	}
	v = uint8(d.ext1byte(startPos))
	return v, true
}

// GetArgUint16 extracts uint16 argument from encoded IKB action d,
// starting search from startIdx argument's index.
//
// Returns it and true as success if it is, or zero value and false if error.
func (d *Encoded) GetArgUint16(startIdx int) (v uint16, success bool) {

	startPos := d.argGet(startIdx, cArgTypeUint16)
	if startPos == cPosErr {
		return 0, false
	}
	v = uint16(d.ext2bytes(startPos))
	return v, true
}

// GetArgUint32 extracts uint32 argument from encoded IKB action d,
// starting search from startIdx argument's index.
//
// Returns it and true as success if it is, or zero value and false if error.
func (d *Encoded) GetArgUint32(startIdx int) (v uint32, success bool) {

	startPos := d.argGet(startIdx, cArgTypeUint32)
	if startPos == cPosErr {
		return 0, false
	}
	v = uint32(d.ext4bytes(startPos))
	return v, true
}

// GetArgUint64 extracts uint64 argument from encoded IKB action d,
// starting search from startIdx argument's index.
//
// Returns it and true as success if it is, or zero value and false if error.
func (d *Encoded) GetArgUint64(startIdx int) (v uint64, success bool) {

	startPos := d.argGet(startIdx, cArgTypeUint64)
	if startPos == cPosErr {
		return 0, false
	}
	v = uint64(d.ext8bytes(startPos))
	return v, true
}

// GetArgFloat32 extracts float32 argument from encoded IKB action d,
// starting search from startIdx argument's index.
//
// Returns it and true as success if it is, or zero value and false if error.
func (d *Encoded) GetArgFloat32(startIdx int) (v float32, success bool) {

	startPos := d.argGet(startIdx, cArgTypeFloat32)
	if startPos == cPosErr {
		return 0, false
	}
	vv := d.ext4bytes(startPos)
	return *(*float32)(unsafe.Pointer(&vv)), true
}

// GetArgFloat64 extracts float64 argument from encoded IKB action d,
// starting search from startIdx argument's index.
//
// Returns it and true as success if it is, or zero value and false if error.
func (d *Encoded) GetArgFloat64(startIdx int) (v float64, success bool) {

	startPos := d.argGet(startIdx, cArgTypeFloat64)
	if startPos == cPosErr {
		return 0, false
	}
	vv := d.ext8bytes(startPos)
	return *(*float64)(unsafe.Pointer(&vv)), true
}

// GetArgString extracts string argument from encoded IKB action d,
// starting search from startIdx argument's index.
//
// Returns it and true as success if it is, or zero value and false if error.
func (d *Encoded) GetArgString(startIdx int) (v string, success bool) {

	startPos := d.argGet(startIdx, cArgTypeString)
	if startPos == cPosErr {
		v, success = "", false
		return
	}

	// startPos - arg type,
	// startPos + 1 - strlen, startPos + 2,... - string content
	return string(d.extNbytes(startPos+2, startPos+1)), true
}

// copy returns a copy of the current encoded IKB action d.
func (d *Encoded) copy() (copy *Encoded) {
	copied := *d
	return &copied
}

// dump returns a complete debug information about encoded IKB action d.
// Each slice element represent one entity of encoded IKB action d.
func (d *Encoded) dump() []EncodedDumpNode {

	// make result slice with length == len of encoded args +
	// id + ssid + args counter + args free index
	argCount := d.ArgCount()
	dumpRes := make([]EncodedDumpNode, argCount+4)

	// Reflect ID
	dumpRes[0].Type = "Encoded View ID"
	dumpRes[0].Pos = cPosViewID
	dumpRes[0].Value = d.GetViewID()

	// Reflect SSID
	dumpRes[1].Type = "Session ID (SSID)"
	dumpRes[1].Pos = cPosSessionID
	dumpRes[1].Value = d.GetSessionID()

	// Reflect args counter
	dumpRes[2].Type = "Arguments counter"
	dumpRes[2].Pos = cPosArgsCount
	dumpRes[2].Value = argCount

	// Reflect args free index
	dumpRes[3].Type = "Arguments next free position"
	dumpRes[3].Pos = cPosArgsFree
	dumpRes[3].Value = d[cPosArgsFree]

	// Save info about arguments
	pos := cPosArgsContent
	for i := 0; i < argCount; i++ {

		dumpRes[4+i].Type = "Argument (" + d.argType2S(d[pos]) + ")"
		dumpRes[4+i].Pos = pos
		dumpRes[4+i].PosType = pos
		dumpRes[4+i].TypeHeader = d[pos]

		// Save content position
		// By default content starts with offset 1
		// Exceptions: strings
		switch d[pos] {

		case cArgTypeString:
			// pos+0 - arg type, pos+1 - strlen, pos+2,... - content
			dumpRes[4+i].PosContent = pos + 2

		default:
			dumpRes[4+i].PosContent = pos + 1
		}

		// Save values
		// By default content starts with offset 1
		// Exceptions: strings
		switch d[pos] {

		case cArgTypeInt8:
			dumpRes[4+i].Value = d.ext1byte(pos + 1)

		case cArgTypeInt16:
			dumpRes[4+i].Value = d.ext2bytes(pos + 1)

		case cArgTypeInt32:
			dumpRes[4+i].Value = d.ext4bytes(pos + 1)

		case cArgTypeInt64:
			dumpRes[4+i].Value = d.ext8bytes(pos + 1)

		case cArgTypeUint8:
			dumpRes[4+i].Value = uint8(d.ext1byte(pos + 1))

		case cArgTypeUint16:
			dumpRes[4+i].Value = uint16(d.ext2bytes(pos + 1))

		case cArgTypeUint32:
			dumpRes[4+i].Value = uint32(d.ext4bytes(pos + 1))

		case cArgTypeUint64:
			dumpRes[4+i].Value = uint64(d.ext8bytes(pos + 1))

		case cArgTypeFloat32:
			v := int32(d.ext4bytes(pos + 1))
			dumpRes[4+i].Value = *(*float32)(unsafe.Pointer(&v))

		case cArgTypeFloat64:
			v := int64(d.ext8bytes(pos + 1))
			dumpRes[4+i].Value = *(*float64)(unsafe.Pointer(&v))

		case cArgTypeString:
			// pos+1 - strlen, pos+2,... - string content
			dumpRes[4+i].Value = string(d.extNbytes(pos+2, pos+1))

		default:
			dumpRes[4+i].Value = nil
		}
	}

	// Dump completed
	return dumpRes
}

// init initializes the current encoded IKB action object d.
func (d *Encoded) init() {
	d[cPosArgsFree] = cPosArgsContent
}
