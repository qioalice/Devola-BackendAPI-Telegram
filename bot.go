// Package tgbotapi has functions and types used for interacting with
// the Telegram Bot API.
package tgbotapi

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	// Official router for "github.com/valyala/fasthttp".
	"github.com/fasthttp/router"

	// 2x to 3x faster than encoding/json.
	// Used pregenerated code in *_ffjson.go files.
	"github.com/pquerna/ffjson/ffjson"

	// Generate POST request with file.
	// TODO: Eliminate dependence (fasthttp provides these features).
	"github.com/technoweenie/multipartstreamer"

	// Reduce alloc/GC operations for RAW JSON data.
	// Already used in "github.com/valyala/fasthttp"
	"github.com/valyala/bytebufferpool"

	// Faster and better than net/http.
	"github.com/valyala/fasthttp"
)

// BotAPI allows you to interact with the Telegram Bot API.
type BotAPI struct {
	Token  string `json:"token"`
	Debug  bool   `json:"debug"`
	Buffer int    `json:"buffer"`

	Self   *User            `json:"-"`
	Client *fasthttp.Client `json:"-"`

	chUpdates chan Update
	status    int32

	// Here stored buffers for RAW JSON Telegram API responses
	rawResponses bytebufferpool.Pool
}

// Predefined constants of internal bot object status
// (BotAPI.status field bitmasks).
const (

	// BotAPI object is currently stopped.
	cStStopped int32 = 0x00000000

	// BotAPI object is currently running using long poll.
	cStServedLongPoll int32 = 0x00000001

	// BotAPI object is currently running using webhook.
	cStServedWebhook int32 = 0x00000002

	// BotAPI object will be stopped soon but not yet.
	cStStopRequested int32 = 0x00000004
)

// Predefined errors.
var (
	eNilBotObject = errors.New("nil bot object")
)

// NewBotAPI creates a new BotAPI instance.
//
// It requires a token, provided by @BotFather on Telegram.
func NewBotAPI(token string) (*BotAPI, error) {
	return NewBotAPIWithClient(token, &fasthttp.Client{})
}

// NewBotAPIWithClient creates a new BotAPI instance
// and allows you to pass a fasthttp.Client.
//
// It requires a token, provided by @BotFather on Telegram.
func NewBotAPIWithClient(token string, client *fasthttp.Client) (*BotAPI, error) {
	bot := &BotAPI{
		Token:  token,
		Client: client,
		Buffer: 100,
	}

	self, err := bot.GetMe()
	if err != nil {
		return nil, err
	}

	bot.Self = self

	return bot, nil
}

// IsServedLongPoll returns true if the bot is now connected to the Telegram API
// using long poll.
func (bot *BotAPI) IsServedLongPoll() bool {
	return bot.getStatus()&
		(cStServedLongPoll|cStStopRequested) != 0
}

// IsServedWebhook returns true if the bot is now connected to the Telegram API
// using webhook.
func (bot *BotAPI) IsServedWebhook() bool {
	return bot.getStatus()&
		(cStServedWebhook|cStStopRequested) != 0
}

// IsServed returns true if the bot is now connected to the Telegram API.
func (bot *BotAPI) IsServed() bool {
	return bot.getStatus()&
		(cStServedLongPoll|cStServedWebhook|cStStopRequested) != 0
}

// IsStopped returns true if the bot is NOT connected to the Telegram API.
func (bot *BotAPI) IsStopped() bool {
	return bot.getStatus() == cStStopped
}

// MakeRequest makes a request to a specific endpoint with our token.
func (bot *BotAPI) MakeRequest(endpoint string, params *fasthttp.Args) (*APIResponse, error) {

	switch {
	case bot.Debug && params != nil:
		log.Println("MakeRequest", endpoint, params)
	case bot.Debug:
		log.Println("MakeRequest", endpoint)
	default:
	}

	var (
		respRAW = bot.rawResponses.Get()
		err     error
	)

	_, respRAW.B, err = bot.Client.Post(respRAW.B, endpoint, params)
	if err != nil {
		return nil, err
	}

	resp := new(APIResponse)
	resp.RAW = respRAW

	if err = ffjson.Unmarshal(respRAW.B, resp); err != nil {
		return nil, err
	}

	bot.debugLog(endpoint, params, hex.EncodeToString(respRAW.B))

	if !resp.Ok {
		err = &Error{Code: resp.ErrorCode, Message: resp.Description}
		if resp.Parameters != nil {
			err.(*Error).ResponseParameters = *resp.Parameters
		}
		return nil, err
	}

	return resp, nil
}

//
func (bot *BotAPI) Dealloc(buf *bytebufferpool.ByteBuffer) {
	bot.rawResponses.Put(buf)
}

// makeMessageRequest makes a request to a method that returns a Message.
func (bot *BotAPI) makeMessageRequest(endpoint string, params *fasthttp.Args) (*Message, error) {
	endpoint = bot.gAPIURL(endpoint)

	resp, err := bot.MakeRequest(endpoint, params)
	if err != nil {
		return nil, err
	}

	msg := new(Message)
	if err = ffjson.Unmarshal(resp.Result, msg); err != nil {
		return nil, err
	}

	bot.debugLog(endpoint, params, msg)

	return msg, nil
}

// UploadFile makes a request to the API with a file.
//
// Requires the parameter to hold the file not be in the params.
// File should be a string to a file path, a FileBytes struct,
// a FileReader struct, or a url.URL.
//
// Note that if your FileReader has a size set to -1, it will read
// the file into memory to calculate a size.
//noinspection GoUnhandledErrorResult
func (bot *BotAPI) UploadFile(endpoint string, params map[string]string, fieldname string, file interface{}) (*APIResponse, error) {
	endpoint = bot.gAPIURL(endpoint)
	ms := multipartstreamer.New()

	switch f := file.(type) {

	case string:
		ms.WriteFields(params)

		fileHandle, err := os.Open(f)
		if err != nil {
			return nil, err
		}
		defer fileHandle.Close()

		fi, err := os.Stat(f)
		if err != nil {
			return nil, err
		}

		ms.WriteReader(fieldname, fileHandle.Name(), fi.Size(), fileHandle)

	case FileBytes:
		ms.WriteFields(params)

		buf := bytes.NewBuffer(f.Bytes)
		ms.WriteReader(fieldname, f.Name, int64(len(f.Bytes)), buf)

	case FileReader:
		ms.WriteFields(params)

		if f.Size != -1 {
			ms.WriteReader(fieldname, f.Name, f.Size, f.Reader)

			break
		}

		data, err := ioutil.ReadAll(f.Reader)
		if err != nil {
			return nil, err
		}

		buf := bytes.NewBuffer(data)

		ms.WriteReader(fieldname, f.Name, int64(len(data)), buf)

	case url.URL:
		params[fieldname] = f.String()

		ms.WriteFields(params)

	default:
		return nil, errors.New(ErrBadFileType)
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	respObj := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(respObj)

	// emulate ms.SetupRequest(*http.Request),
	// SetBodyStream call includes setting ContentLength
	req.SetRequestURI(endpoint)
	req.Header.SetMethod("POST")
	// req.Header.SetContentType("application/x-www-form-urlencoded")
	req.Header.Add("Content-Type", ms.ContentType)
	req.SetBodyStream(ms.GetReader(), -1)

	if err := bot.Client.Do(req, respObj); err != nil {
		return nil, err
	}

	resp := new(APIResponse)
	if err := ffjson.Unmarshal(respObj.Body(), resp); err != nil {
		return nil, err
	}

	bot.debugLog(endpoint, []interface{}{params, fieldname, file}, resp)

	if !resp.Ok {
		err := &Error{Code: resp.ErrorCode, Message: resp.Description}
		if resp.Parameters != nil {
			err.ResponseParameters = *resp.Parameters
		}
		return nil, err
	}

	return resp, nil
}

// GetFileDirectURL returns direct URL to file
//
// It requires the FileID.
func (bot *BotAPI) GetFileDirectURL(fileID string) (string, error) {
	return bot.gFileURL(fileID), nil
}

// GetMe fetches the currently authenticated bot.
//
// This method is called upon creation to validate the token,
// and so you may get this data from BotAPI.Self without the need for
// another request.
func (bot *BotAPI) GetMe() (*User, error) {
	endpoint := bot.gAPIURL("getMe")

	resp, err := bot.MakeRequest(endpoint, nil)
	if err != nil {
		return nil, err
	}

	usr := new(User)
	if err := ffjson.Unmarshal(resp.Result, usr); err != nil {
		return nil, err
	}

	bot.debugLog(endpoint, nil, usr)

	return usr, nil
}

// IsMessageToMe returns true if message directed to this bot.
//
// It requires the Message.
func (bot *BotAPI) IsMessageToMe(message *Message) bool {
	return strings.Contains(message.Text, "@"+bot.Self.UserName)
}

// Send will send a Chattable item to Telegram.
//
// It requires the Chattable to send.
func (bot *BotAPI) Send(c Chattable) (*Message, error) {
	switch c.(type) {
	case Fileable:
		return bot.sendFile(c.(Fileable))
	default:
		return bot.sendChattable(c)
	}
}

// debugLog checks if the bot is currently running in debug mode, and if
// so will display information about the request and response in the
// debug log.
func (bot *BotAPI) debugLog(context string, v, message interface{}) {
	if bot.Debug && v != nil {
		log.Printf("%s req : %+v\n", context, v)

	}
	if bot.Debug && message != nil {
		log.Printf("%s resp: %+v\n", context, message)
	}
}

// sendExisting will send a Message with an existing file to Telegram.
func (bot *BotAPI) sendExisting(method string, config Fileable) (*Message, error) {

	v, err := config.values()
	if err != nil {
		return nil, err
	}
	defer fasthttp.ReleaseArgs(v)

	message, err := bot.makeMessageRequest(method, v)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// uploadAndSend will send a Message with a new file to Telegram.
func (bot *BotAPI) uploadAndSend(method string, config Fileable) (*Message, error) {

	params, err := config.params()
	if err != nil {
		return nil, err
	}

	file := config.getFile()

	resp, err := bot.UploadFile(method, params, config.name(), file)
	if err != nil {
		return nil, err
	}

	message := new(Message)
	if err = ffjson.Unmarshal(resp.Result, message); err != nil {
		return nil, err
	}

	bot.debugLog(method, nil, message)

	return message, nil
}

// sendFile determines if the file is using an existing file or uploading
// a new file, then sends it as needed.
func (bot *BotAPI) sendFile(config Fileable) (*Message, error) {
	if config.useExistingFile() {
		return bot.sendExisting(config.method(), config)
	}

	return bot.uploadAndSend(config.method(), config)
}

// sendChattable sends a Chattable.
func (bot *BotAPI) sendChattable(config Chattable) (*Message, error) {

	v, err := config.values()
	if err != nil {
		return nil, err
	}
	defer fasthttp.ReleaseArgs(v)

	message, err := bot.makeMessageRequest(config.method(), v)

	if err != nil {
		return nil, err
	}

	return message, nil
}

// SendChatAction sends the bot action in chat.
//
// Chat ID must be not equal 0 nor -1.
// Action must be one of constants starts from "Chat...".
func (bot *BotAPI) SendChatAction(chatID int64, action string) (bool, error) {
	endpoint := bot.gAPIURL("sendChatAction")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	v.Add("chat_id", strconv.FormatInt(chatID, 10))
	v.Add("action", action)

	resp, err := bot.MakeRequest(endpoint, v)
	if err != nil {
		return false, err
	}

	result := false
	if len(resp.Result) >= 4 &&
		resp.Result[0] == 't' && resp.Result[1] == 'r' &&
		resp.Result[2] == 'u' && resp.Result[3] == 'e' {
		result = true
	}

	bot.debugLog(endpoint, v, result)

	return result, nil
}

// SendMediaGroup sends the group of photos and videos to chat.
func (bot *BotAPI) SendMediaGroup(config MediaGroupConfig) ([]Message, error) {
	endpoint := bot.gAPIURL("sendMediaGroup")

	v, err := config.values()
	if err != nil {
		return nil, err
	}
	defer fasthttp.ReleaseArgs(v)

	resp, err := bot.MakeRequest(endpoint, v)
	if err != nil {
		return nil, err
	}

	msgs := make([]Message, 0, len(config.InputMedia))
	if err = ffjson.Unmarshal(resp.Result, &msgs); err != nil {
		return nil, err
	}

	bot.debugLog(endpoint, v, msgs)

	return msgs, nil
}

// GetUserProfilePhotos gets a user's profile photos.
//
// It requires UserID.
// Offset and Limit are optional.
func (bot *BotAPI) GetUserProfilePhotos(config UserProfilePhotosConfig) (*UserProfilePhotos, error) {
	endpoint := bot.gAPIURL("getUserProfilePhotos")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	v.Add("user_id", strconv.Itoa(config.UserID))
	if config.Offset != 0 {
		v.Add("offset", strconv.Itoa(config.Offset))
	}
	if config.Limit != 0 {
		v.Add("limit", strconv.Itoa(config.Limit))
	}

	resp, err := bot.MakeRequest(endpoint, v)
	if err != nil {
		return nil, err
	}

	profilePhotos := new(UserProfilePhotos)
	if err := ffjson.Unmarshal(resp.Result, profilePhotos); err != nil {
		return nil, err
	}

	bot.debugLog(endpoint, v, profilePhotos)

	return profilePhotos, nil
}

// GetFile returns a File which can download a file from Telegram.
//
// Requires FileID.
func (bot *BotAPI) GetFile(config FileConfig) (*File, error) {
	endpoint := bot.gAPIURL("getFile")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	v.Add("file_id", config.FileID)

	resp, err := bot.MakeRequest(endpoint, v)
	if err != nil {
		return nil, err
	}

	file := new(File)
	if err := ffjson.Unmarshal(resp.Result, file); err != nil {
		return nil, err
	}

	bot.debugLog(endpoint, v, file)

	return file, nil
}

// GetUpdates fetches updates.
// If a WebHook is set, this will not return any data!
//
// Offset, Limit, and Timeout are optional.
// To avoid stale items, set Offset to one higher than the previous item.
// Set Timeout to a large number to reduce requests so you can get updates
// instantly instead of having to wait between requests.
func (bot *BotAPI) GetUpdates(config UpdateConfig) ([]Update, error) {

	if bot.chUpdates != nil {
		return nil, errors.New("already served")
	}

	const defaultBufferLen = 64

	buflen := config.Limit
	if buflen <= 0 {
		buflen = defaultBufferLen
	}

	buf := make([]Update, 0, buflen)
	if err := bot.getUpdates(config, &buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func (bot *BotAPI) getUpdates(config UpdateConfig, writeTo *[]Update) error {
	endpoint := bot.gAPIURL("getUpdates")

	var v *fasthttp.Args

	if config.Offset != 0 || config.Limit > 0 || config.Timeout > 0 {
		v = fasthttp.AcquireArgs()
		defer fasthttp.ReleaseArgs(v)

		if config.Offset != 0 {
			v.Add("offset", strconv.Itoa(config.Offset))
		}
		if config.Limit > 0 {
			v.Add("limit", strconv.Itoa(config.Limit))
		}
		if config.Timeout > 0 {
			v.Add("timeout", strconv.Itoa(config.Timeout))
		}
	}

	resp, err := bot.MakeRequest(endpoint, v)
	if err != nil {
		return err
	}

	// TODO: Add AB test

	if err = json.Unmarshal(resp.Result, writeTo); err != nil {
		return err
	}

	bot.debugLog(endpoint, v, writeTo)

	return nil
}

//
func (bot *BotAPI) initUpdatesChan() *BotAPI {
	if bot.chUpdates == nil {
		bot.chUpdates = make(chan Update, bot.Buffer)
	}
	return bot
}

// GetUpdatesChan returns a channel for getting updates.
func (bot *BotAPI) GetUpdatesChan() UpdatesChannel {
	return bot.initUpdatesChan().chUpdates
}

// serveBegin performs first checks before starts receiving an incoming
// Telegram Bot API events.
// If this method return not nil error, serve reqeust will not be confirmed.
func (bot *BotAPI) serveBegin(serveAt int32) error {
	if bot == nil {
		return eNilBotObject
	}

	// Enable serving only if bot is currently stopped.
	if !atomic.CompareAndSwapInt32(&bot.status, cStStopped, serveAt) {
		switch atomic.LoadInt32(&bot.status) {

		case cStServedLongPoll:
			return errors.New("already long polling")

		case cStServedWebhook:
			return errors.New("already webhooked")

		case cStStopRequested:
			return errors.New("stop is requested, try again later")
		}

		return errors.New("unknown error, try again later")
	}
	return nil
}

// serveLongPoll performs long polling Telegram Bot API updates in
// infinity loop in the current goroutine until it will not stopped by
// StopLongPolling method.
//
// When first try to get updates will be complete, the pDone will set to true
// (if it's not nil) and error object of that operation will be stored to the pErr
// (if it's not nil too).
func (bot *BotAPI) serveLongPoll(config UpdateConfig, updates []Update) {
	for {
		// If stop request received, approve it and end long polling.
		if atomic.CompareAndSwapInt32(&bot.status, cStStopRequested, cStStopped) {
			return
		}

		if err := bot.getUpdates(config, &updates); err != nil {
			log.Println(err)
			log.Println("Failed to get updates, retrying in 3 seconds...")
			time.Sleep(time.Second * 3)
			continue
		}

		for _, update := range updates {
			if update.UpdateID >= config.Offset {
				config.Offset = update.UpdateID + 1
				bot.chUpdates <- update
			}
		}

		(*reflect.SliceHeader)(unsafe.Pointer(&updates)).Len = 0
	}
}

// serveWebhook is fasthttp.Server handler that calls when new Telegram Bot API
// event has been received and should be processsed by this method.
func (bot *BotAPI) serveWebhook(ctx *fasthttp.RequestCtx) {

	var update Update

	if err := ffjson.Unmarshal(ctx.Request.Body(), &update); err == nil {
		bot.chUpdates <- update
	}
}

// ServeLongPoll starts receiving an incoming Telegram Bot API events
// using long polling.
//
// Successfully started only if receiving is not running already
// (neither a long polling, nor a webhook).
func (bot *BotAPI) ServeLongPoll(config UpdateConfig) error {
	if err := bot.serveBegin(cStServedLongPoll); err != nil {
		return err
	}

	bot.initUpdatesChan()

	const defaultBufferLen = 256

	if config.Limit <= 0 {
		config.Limit = defaultBufferLen
	}

	updates := make([]Update, 0, config.Limit)

	// Try to perform test query.
	prevLimit := config.Limit
	config.Limit = 1
	if err := bot.getUpdates(config, &updates); err != nil {
		return err
	}

	// Reset changed update's limit in config and len of slice
	config.Limit = prevLimit
	(*reflect.SliceHeader)(unsafe.Pointer(&updates)).Len = 0

	go bot.serveLongPoll(config, updates)
	return nil
}

// ServeWebHook registers receiving an incoming Telegram Bot API events
// using webhook.
// Returns a fasthttp.Server object with predefined handler on the webhook
// and you should listen it to start receiving updates.
//
// If you do not have a legitimate TLS certificate, you need to include
// your self signed certificate with the config.
//
// Successfully started only if receiving is not running already
// (neither a long polling, nor a webhook).
func (bot *BotAPI) ServeWebHook(config WebhookConfig, pattern string) (*fasthttp.Server, error) {
	if err := bot.serveBegin(cStServedWebhook); err != nil {
		return nil, err
	}

	var (
		r      = router.New()
		s      = new(fasthttp.Server)
		params = make(map[string]string)
		resp   *APIResponse
		err    error
	)

	bot.initUpdatesChan()

	r.POST(pattern, bot.serveWebhook)
	s.Handler = r.Handler

	if config.Certificate == nil {

		v := fasthttp.AcquireArgs()
		defer fasthttp.ReleaseArgs(v)

		v.Add("url", config.URL.String())
		if config.MaxConnections != 0 {
			v.Add("max_connections", strconv.Itoa(config.MaxConnections))
		}

		resp, err = bot.MakeRequest(bot.gAPIURL("setWebhook"), v)
	} else {

		params["url"] = config.URL.String()
		if config.MaxConnections != 0 {
			params["max_connections"] = strconv.Itoa(config.MaxConnections)
		}

		resp, err = bot.UploadFile("setWebhook", params, "certificate", config.Certificate)
	}

	if err != nil {
		s.Handler = nil
		return nil, err
	}

	// TODO: resp handling?
	_ = resp
	return s, err
}

// Stop requests to stop receiving an incoming Telegram Bot API events.
// You can use this method for stopping both of long polling and webhook serving.
// Does nothing (and returns nil as error) if bot is not started.
func (bot *BotAPI) Stop() error {

	var err error

	switch {
	case bot == nil:
		err = eNilBotObject

	// Was running using long poll, should wait until stopping is confirmed
	// by serveLongPoll's iteration.
	case atomic.CompareAndSwapInt32(&bot.status, cStServedLongPoll, cStStopRequested):
		for atomic.LoadInt32(&bot.status) != cStStopped {
			time.Sleep(100 * time.Millisecond)
		}

	// Was running using webhook, will be completely stopped here.
	case atomic.CompareAndSwapInt32(&bot.status, cStServedWebhook, cStStopRequested):
		endpoint := bot.gAPIURL("setWebhook")
		if _, err = bot.MakeRequest(endpoint, nil); err == nil {
			atomic.StoreInt32(&bot.status, cStStopped)
		}
	}

	return err
}

// GetWebhookInfo allows you to fetch information about a webhook and if
// one currently is set, along with pending update count and error messages.
func (bot *BotAPI) GetWebhookInfo() (*WebhookInfo, error) {
	endpoint := bot.gAPIURL("getWebhookInfo")

	resp, err := bot.MakeRequest(endpoint, nil)
	if err != nil {
		return nil, err
	}

	info := new(WebhookInfo)

	if err = ffjson.Unmarshal(resp.Result, info); err != nil {
		return nil, err
	}

	return info, err
}

// AnswerInlineQuery sends a response to an inline query.
//
// Note that you must respond to an inline query within 30 seconds.
func (bot *BotAPI) AnswerInlineQuery(config InlineConfig) (*APIResponse, error) {
	endpoint := bot.gAPIURL("answerInlineQuery")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	v.Add("inline_query_id", config.InlineQueryID)
	v.Add("cache_time", strconv.Itoa(config.CacheTime))
	v.Add("is_personal", strconv.FormatBool(config.IsPersonal))
	v.Add("next_offset", config.NextOffset)
	data, err := ffjson.Marshal(config.Results)
	if err != nil {
		return nil, err
	}
	v.Add("results", string(data))
	v.Add("switch_pm_text", config.SwitchPMText)
	v.Add("switch_pm_parameter", config.SwitchPMParameter)

	bot.debugLog(endpoint, v, nil)

	return bot.MakeRequest(endpoint, v)
}

// AnswerCallbackQuery sends a response to an inline query callback.
func (bot *BotAPI) AnswerCallbackQuery(config CallbackConfig) (*APIResponse, error) {
	endpoint := bot.gAPIURL("answerCallbackQuery")

	v := fasthttp.AcquireArgs()
	fasthttp.ReleaseArgs(v)

	v.Add("callback_query_id", config.CallbackQueryID)
	if config.Text != "" {
		v.Add("text", config.Text)
	}
	v.Add("show_alert", strconv.FormatBool(config.ShowAlert))
	if config.URL != "" {
		v.Add("url", config.URL)
	}
	v.Add("cache_time", strconv.Itoa(config.CacheTime))

	bot.debugLog(endpoint, v, nil)

	return bot.MakeRequest(endpoint, v)
}

// KickChatMember kicks a user from a chat. Note that this only will work
// in supergroups, and requires the bot to be an admin. Also note they
// will be unable to rejoin until they are unbanned.
func (bot *BotAPI) KickChatMember(config KickChatMemberConfig) (*APIResponse, error) {
	endpoint := bot.gAPIURL("kickChatMember")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	if config.SuperGroupUsername == "" {
		v.Add("chat_id", strconv.FormatInt(config.ChatID, 10))
	} else {
		v.Add("chat_id", config.SuperGroupUsername)
	}
	v.Add("user_id", strconv.Itoa(config.UserID))

	if config.UntilDate != 0 {
		v.Add("until_date", strconv.FormatInt(config.UntilDate, 10))
	}

	bot.debugLog(endpoint, v, nil)

	return bot.MakeRequest(endpoint, v)
}

// LeaveChat makes the bot leave the chat.
func (bot *BotAPI) LeaveChat(config ChatConfig) (*APIResponse, error) {
	endpoint := bot.gAPIURL("leaveChat")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	if config.SuperGroupUsername == "" {
		v.Add("chat_id", strconv.FormatInt(config.ChatID, 10))
	} else {
		v.Add("chat_id", config.SuperGroupUsername)
	}

	bot.debugLog(endpoint, v, nil)

	return bot.MakeRequest(endpoint, v)
}

// GetChat gets information about a chat.
func (bot *BotAPI) GetChat(config ChatConfig) (*Chat, error) {
	endpoint := bot.gAPIURL("getChat")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	if config.SuperGroupUsername == "" {
		v.Add("chat_id", strconv.FormatInt(config.ChatID, 10))
	} else {
		v.Add("chat_id", config.SuperGroupUsername)
	}

	resp, err := bot.MakeRequest(endpoint, v)
	if err != nil {
		return nil, err
	}

	chat := new(Chat)
	if err = ffjson.Unmarshal(resp.Result, chat); err != nil {
		return nil, err
	}

	bot.debugLog(endpoint, v, chat)

	return chat, nil
}

// GetChatAdministrators gets a list of administrators in the chat.
//
// If none have been appointed, only the creator will be returned.
// Bots are not shown, even if they are an administrator.
func (bot *BotAPI) GetChatAdministrators(config ChatConfig) ([]ChatMember, error) {
	endpoint := bot.gAPIURL("getChatAdministrators")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	if config.SuperGroupUsername == "" {
		v.Add("chat_id", strconv.FormatInt(config.ChatID, 10))
	} else {
		v.Add("chat_id", config.SuperGroupUsername)
	}

	resp, err := bot.MakeRequest(endpoint, v)
	if err != nil {
		return nil, err
	}

	var members []ChatMember
	if err = ffjson.Unmarshal(resp.Result, &members); err != nil {
		return nil, err
	}

	bot.debugLog(endpoint, v, members)

	return members, err
}

// GetChatMembersCount gets the number of users in a chat.
func (bot *BotAPI) GetChatMembersCount(config ChatConfig) (int, error) {
	endpoint := bot.gAPIURL("getChatMembersCount")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	if config.SuperGroupUsername == "" {
		v.Add("chat_id", strconv.FormatInt(config.ChatID, 10))
	} else {
		v.Add("chat_id", config.SuperGroupUsername)
	}

	resp, err := bot.MakeRequest(endpoint, v)
	if err != nil {
		return -1, err
	}

	var count int
	if err = ffjson.Unmarshal(resp.Result, &count); err != nil {
		return -1, err
	}

	bot.debugLog(endpoint, v, count)

	return count, err
}

// GetChatMember gets a specific chat member.
func (bot *BotAPI) GetChatMember(config ChatConfigWithUser) (*ChatMember, error) {
	endpoint := bot.gAPIURL("getChatMember")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	if config.SuperGroupUsername == "" {
		v.Add("chat_id", strconv.FormatInt(config.ChatID, 10))
	} else {
		v.Add("chat_id", config.SuperGroupUsername)
	}
	v.Add("user_id", strconv.Itoa(config.UserID))

	resp, err := bot.MakeRequest(endpoint, v)
	if err != nil {
		return nil, err
	}

	member := new(ChatMember)
	if err = ffjson.Unmarshal(resp.Result, member); err != nil {
		return nil, err
	}

	bot.debugLog(endpoint, v, member)

	return member, err
}

// UnbanChatMember unbans a user from a chat. Note that this only will work
// in supergroups and channels, and requires the bot to be an admin.
func (bot *BotAPI) UnbanChatMember(config ChatMemberConfig) (*APIResponse, error) {
	endpoint := bot.gAPIURL("unbanChatMember")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	if config.SuperGroupUsername != "" {
		v.Add("chat_id", config.SuperGroupUsername)
	} else if config.ChannelUsername != "" {
		v.Add("chat_id", config.ChannelUsername)
	} else {
		v.Add("chat_id", strconv.FormatInt(config.ChatID, 10))
	}
	v.Add("user_id", strconv.Itoa(config.UserID))

	bot.debugLog(endpoint, v, nil)

	return bot.MakeRequest(endpoint, v)
}

// RestrictChatMember to restrict a user in a supergroup. The bot must be an
// administrator in the supergroup for this to work and must have the
// appropriate admin rights. Pass True for all boolean parameters to lift
// restrictions from a user. Returns True on success.
func (bot *BotAPI) RestrictChatMember(config RestrictChatMemberConfig) (*APIResponse, error) {
	endpoint := bot.gAPIURL("restrictChatMember")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	if config.SuperGroupUsername != "" {
		v.Add("chat_id", config.SuperGroupUsername)
	} else if config.ChannelUsername != "" {
		v.Add("chat_id", config.ChannelUsername)
	} else {
		v.Add("chat_id", strconv.FormatInt(config.ChatID, 10))
	}
	v.Add("user_id", strconv.Itoa(config.UserID))

	if config.CanSendMessages != nil {
		v.Add("can_send_messages", strconv.FormatBool(*config.CanSendMessages))
	}
	if config.CanSendMediaMessages != nil {
		v.Add("can_send_media_messages", strconv.FormatBool(*config.CanSendMediaMessages))
	}
	if config.CanSendOtherMessages != nil {
		v.Add("can_send_other_messages", strconv.FormatBool(*config.CanSendOtherMessages))
	}
	if config.CanAddWebPagePreviews != nil {
		v.Add("can_add_web_page_previews", strconv.FormatBool(*config.CanAddWebPagePreviews))
	}
	if config.UntilDate != 0 {
		v.Add("until_date", strconv.FormatInt(config.UntilDate, 10))
	}

	bot.debugLog(endpoint, v, nil)

	return bot.MakeRequest(endpoint, v)
}

// PromoteChatMember add admin rights to user
func (bot *BotAPI) PromoteChatMember(config PromoteChatMemberConfig) (*APIResponse, error) {
	endpoint := bot.gAPIURL("promoteChatMember")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	if config.SuperGroupUsername != "" {
		v.Add("chat_id", config.SuperGroupUsername)
	} else if config.ChannelUsername != "" {
		v.Add("chat_id", config.ChannelUsername)
	} else {
		v.Add("chat_id", strconv.FormatInt(config.ChatID, 10))
	}
	v.Add("user_id", strconv.Itoa(config.UserID))

	if config.CanChangeInfo != nil {
		v.Add("can_change_info", strconv.FormatBool(*config.CanChangeInfo))
	}
	if config.CanPostMessages != nil {
		v.Add("can_post_messages", strconv.FormatBool(*config.CanPostMessages))
	}
	if config.CanEditMessages != nil {
		v.Add("can_edit_messages", strconv.FormatBool(*config.CanEditMessages))
	}
	if config.CanDeleteMessages != nil {
		v.Add("can_delete_messages", strconv.FormatBool(*config.CanDeleteMessages))
	}
	if config.CanInviteUsers != nil {
		v.Add("can_invite_users", strconv.FormatBool(*config.CanInviteUsers))
	}
	if config.CanRestrictMembers != nil {
		v.Add("can_restrict_members", strconv.FormatBool(*config.CanRestrictMembers))
	}
	if config.CanPinMessages != nil {
		v.Add("can_pin_messages", strconv.FormatBool(*config.CanPinMessages))
	}
	if config.CanPromoteMembers != nil {
		v.Add("can_promote_members", strconv.FormatBool(*config.CanPromoteMembers))
	}

	bot.debugLog(endpoint, v, nil)

	return bot.MakeRequest(endpoint, v)
}

// GetGameHighScores allows you to get the high scores for a game.
func (bot *BotAPI) GetGameHighScores(config GetGameHighScoresConfig) ([]GameHighScore, error) {

	v, _ := config.values()
	defer fasthttp.ReleaseArgs(v)

	resp, err := bot.MakeRequest(bot.gAPIURL(config.method()), v)
	if err != nil {
		return []GameHighScore{}, err
	}

	var highScores []GameHighScore
	if err = ffjson.Unmarshal(resp.Result, &highScores); err != nil {
		return nil, err
	}

	return highScores, nil
}

// AnswerShippingQuery allows you to reply to Update with shipping_query parameter.
func (bot *BotAPI) AnswerShippingQuery(config ShippingConfig) (*APIResponse, error) {
	endpoint := bot.gAPIURL("answerShippingQuery")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	v.Add("shipping_query_id", config.ShippingQueryID)
	v.Add("ok", strconv.FormatBool(config.OK))
	if config.OK == true {
		data, err := json.Marshal(config.ShippingOptions)
		if err != nil {
			return nil, err
		}
		v.Add("shipping_options", string(data))
	} else {
		v.Add("error_message", config.ErrorMessage)
	}

	bot.debugLog(endpoint, v, nil)

	return bot.MakeRequest(endpoint, v)
}

// AnswerPreCheckoutQuery allows you to reply to Update with pre_checkout_query.
func (bot *BotAPI) AnswerPreCheckoutQuery(config PreCheckoutConfig) (*APIResponse, error) {
	endpoint := bot.gAPIURL("answerPreCheckoutQuery")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	v.Add("pre_checkout_query_id", config.PreCheckoutQueryID)
	v.Add("ok", strconv.FormatBool(config.OK))
	if config.OK != true {
		v.Add("error", config.ErrorMessage)
	}

	bot.debugLog(endpoint, v, nil)

	return bot.MakeRequest(endpoint, v)
}

// DeleteMessage deletes a message in a chat
func (bot *BotAPI) DeleteMessage(config DeleteMessageConfig) (*APIResponse, error) {

	v, _ := config.values()
	defer fasthttp.ReleaseArgs(v)

	bot.debugLog(config.method(), v, nil)

	return bot.MakeRequest(bot.gAPIURL(config.method()), v)
}

// GetInviteLink get InviteLink for a chat
func (bot *BotAPI) GetInviteLink(config ChatConfig) (string, error) {
	endpoint := bot.gAPIURL("exportChatInviteLink")

	v := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(v)

	if config.SuperGroupUsername == "" {
		v.Add("chat_id", strconv.FormatInt(config.ChatID, 10))
	} else {
		v.Add("chat_id", config.SuperGroupUsername)
	}

	resp, err := bot.MakeRequest(endpoint, v)
	if err != nil {
		return "", err
	}

	var inviteLink string
	if err := ffjson.Unmarshal(resp.Result, &inviteLink); err != nil {
		return "", err
	}

	return inviteLink, nil
}

// PinChatMessage pin message in supergroup
func (bot *BotAPI) PinChatMessage(config PinChatMessageConfig) (*APIResponse, error) {

	v, _ := config.values()
	defer fasthttp.ReleaseArgs(v)

	bot.debugLog(config.method(), v, nil)

	return bot.MakeRequest(bot.gAPIURL(config.method()), v)
}

// UnpinChatMessage unpin message in supergroup
func (bot *BotAPI) UnpinChatMessage(config UnpinChatMessageConfig) (*APIResponse, error) {

	v, _ := config.values()
	defer fasthttp.ReleaseArgs(v)

	bot.debugLog(config.method(), v, nil)

	return bot.MakeRequest(bot.gAPIURL(config.method()), v)
}

// SetChatTitle change title of chat.
func (bot *BotAPI) SetChatTitle(config SetChatTitleConfig) (*APIResponse, error) {

	v, _ := config.values()
	defer fasthttp.ReleaseArgs(v)

	bot.debugLog(config.method(), v, nil)

	return bot.MakeRequest(bot.gAPIURL(config.method()), v)
}

// SetChatDescription change description of chat.
func (bot *BotAPI) SetChatDescription(config SetChatDescriptionConfig) (*APIResponse, error) {

	v, _ := config.values()
	defer fasthttp.ReleaseArgs(v)

	bot.debugLog(config.method(), v, nil)

	return bot.MakeRequest(bot.gAPIURL(config.method()), v)
}

// SetChatPhoto change photo of chat.
func (bot *BotAPI) SetChatPhoto(config SetChatPhotoConfig) (*APIResponse, error) {

	params, err := config.params()
	if err != nil {
		return nil, err
	}

	file := config.getFile()

	return bot.UploadFile(config.method(), params, config.name(), file)
}

// DeleteChatPhoto delete photo of chat.
func (bot *BotAPI) DeleteChatPhoto(config DeleteChatPhotoConfig) (*APIResponse, error) {

	v, _ := config.values()
	defer fasthttp.ReleaseArgs(v)

	bot.debugLog(config.method(), v, nil)

	return bot.MakeRequest(bot.gAPIURL(config.method()), v)
}

// gAPIURL generates and returns a full Telegram API URL with passed method name.
// Full URL includes bot token, Telegram API URI and method name.
func (bot *BotAPI) gAPIURL(endpoint string) string {
	return makeURL(APIEndpoint, bot.Token, endpoint)
}

// gFileURL generates and returns a full direct link to download some file
// from Telegram Bot API using passed file ID.
func (bot *BotAPI) gFileURL(fileID string) string {
	return makeURL(FileEndpoint, bot.Token, fileID)
}

// getStatus returns a value of BotAPI.status field value as atomic operation,
// and returns 0 if bot is nil.
func (bot *BotAPI) getStatus() int32 {
	if bot == nil {
		return cStStopped
	}
	return atomic.LoadInt32(&bot.status)
}
