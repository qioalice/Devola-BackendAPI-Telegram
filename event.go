// Copyright © 2019. All rights reserved.
// Author: Alice Qio.
// Contacts: <qioalice@gmail.com>.
// License: https://opensource.org/licenses/MIT

package tgbotapi

import (
	"reflect"
	"unsafe"

	"github.com/qioalice/devola-core/event"

	"github.com/qioalice/devola-backend-telegram/ikba"
)

// Event represents Telegram Bot Update type.
//
// When user somehow interact with Telegram Bot, it sends an api.Update
// object to the Telegram Bot.
// The Event object creates using api.Update object and answers to the
// following questions:
// - What kind of event is occurred?
// - What data has been received with occurred event?
//
// This is a part of Ctx.
//
// Thus you can always to get info about what event you're handling
// inside your handler.
//
// More info: devola-core/event.Type, devola-core/event.Data, ikba.Encoded, Ctx.
type Event struct {
	event.Event

	// TODO: Add arguments supporting (for example for commands)

	// Encoded IKB action.
	// It's a pointer to avoid reallocate memory for ikba.Encoded object
	// and it is a pointer to the Data field with casted type.
	//
	// Not nil only if Type == CTypeInlineKeyboardButton.
	ikbae *ikba.Encoded `json:"-"`
}

// MakeEvent creates a new Event object with passed event type and event data,
// but also initializes IKB encoded action pointer if it is IKB event.
func MakeEvent(typ event.Type, data event.Data) *Event {
	var e Event
	e.Type, e.Data = typ, data

	// ikbae field will point to the Data field but with right type
	if typ == CTypeInlineKeyboardButton {
		dataHeader := (*reflect.StringHeader)(unsafe.Pointer(&e.Data))
		e.ikbae = (*ikba.Encoded)(unsafe.Pointer(dataHeader.Data))
	}

	return &e
}
