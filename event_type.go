// Copyright © 2019. All rights reserved.
// Author: Alice Qio.
// Contacts: <qioalice@gmail.com>.
// License: https://opensource.org/licenses/MIT

package tgbotapi

import (
	"github.com/qioalice/devola-core/event"
)

// Constants of Type.
// Use these constants to figure out what kind of event is occurred
// (by comparing Type).
const (

	// Marker of invalid type.
	CTypeInvalid event.Type = 0 + iota

	// Chat text command.
	// tEvent's Data field represents a lowercase command without arguments.
	CTypeCommand event.Type = 100

	// Pressed keyboard button.
	// Technically this is a text (in a chat), but a text sent by
	// pressing to the keyboard button.
	// tEvent's Data field represents this keyboard button data.
	CTypeKeyboardButton event.Type = 101

	// Typed text.
	// tEvent's Data field represents the whole text but with trimmed
	// leading and trailing spaces.
	CTypeText event.Type = 102

	// Pressed inline keyboard button.
	// tEvent's Data field stored TAction value, representing the your
	// action causes occurred event.
	CTypeInlineKeyboardButton event.Type = 200
)
