// Package tgbotapi has functions and types used for interacting with
// the Telegram Bot API.
package tgbotapi

// CHANGELIST.

// VER 1.1, 2019.06.01
//
// Dependencies:
//
// - JSON encode/decode backend switched from internal golang encoding/json package
//   to the github.com/pquerna/ffjson .
//
// - HTTP/S client/server switched from internal golang net/http package
//   to the github.com/valyala/fasthttp .
//
// - Added reusing []byte buffer for part of RAW JSON response from Telegram API
//   by github.com/valyala/bytebufferpool package (is already dependence for fasthttp).
//
// - Implemented setting webhook handler for fasthttp by github.com/fasthttp/router .
//
// - Switched tests backend to the github.com/stretchr/testify 's set of packages.
//
// API changing:
//
// - Added methods BotAPI.IsServedLongPoll, BotAPI.IsServedWebhook,
//   BotAPI.IsServed, BotAPI.IsStopped to figure out current status of BotAPI.
//
// - Added BotAPI.Dealloc method to mark as unused bytebufferpool.ByteBuffer object
//   ([]byte), that stored RAW JSON response in itself
//   and has a part of each APIResponse object.
//
// - BotAPI.GetUpdatesChan now does not starts long poll receiving Telegram Bot API
//   updates, but returns a channel in which updates will be passed
//   (for both webhook and long polling).
//
// - Merged BotAPI.SetWebhook and BotAPI.ListenForWebhook to the BotAPI.ServeWebhook.
//   Now this method creates fasthttp server object, initializes it,
//   makes http request to the Telegram API about webhook registration.
//
// - Added BotAPI.ServeLongPoll method to start long poll receiving Telegram Bot API
//   updates (in a separate goroutine).
//
// - Merged BotAPI.StopReceivingUpdates, BotAPI.DeleteWebhook to the BotAPI.Stop.
//   Now this method determines what type of receiving updates is used now,
//   and then stops one of them (webhook or long poll) automatically.
//
// - Added BotAPI.SendChatAction because BotAPI.Send is bad and don't work
//   to send chat actions.
//
// - Added BotAPI.SendMediaGroup because BotAPI.Send is bad and don't work
//   to send media group.
//
// Optimizations:
//
// - Changed way of generating API URL used to connect to the Telegram servers.
//
// - Switched net/http.Args to the fasthttp.Args as GET/POST requests args.
//   Added reusing these objects (to decrease GC calls).
//
// - Switched net/http.Client to the fasthttp.Client as http GET/POST client.
//   Added reusing request/response objects (to decrease GC calls).
//
// - Switched net/http.Server to the fasthttp.Server as webhook server.
//   Added reusing request/response objects (to decrease GC calls).
//
// - Implemented using only one []Update buffer to first storing decoded from JSON
//   Telegram Bot API incoming events (to decrease GC calls).

// VER 1.0
// Forked from github.com/go-telegram-bot-api/telegram-bot-api
// Commit: cea05bfc443c89b43d5fa0ddd28eeb15dfe20639
