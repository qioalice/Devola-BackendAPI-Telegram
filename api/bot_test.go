package api_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/qioalice/devola-backend-telegram/api"
)

const (
	TestToken               = "153667468:AAHlSHlMqSt1f_uFmVRJbm5gntu2HI4WW8I"
	ChatID                  = 76918703
	SupergroupChatID        = -1001120141283
	ReplyToMessageID        = 35
	ExistingPhotoFileID     = "AgADAgADw6cxG4zHKAkr42N7RwEN3IFShCoABHQwXEtVks4EH2wBAAEC"
	ExistingDocumentFileID  = "BQADAgADOQADjMcoCcioX1GrDvp3Ag"
	ExistingAudioFileID     = "BQADAgADRgADjMcoCdXg3lSIN49lAg"
	ExistingVoiceFileID     = "AwADAgADWQADjMcoCeul6r_q52IyAg"
	ExistingVideoFileID     = "BAADAgADZgADjMcoCav432kYe0FRAg"
	ExistingVideoNoteFileID = "DQADAgADdQAD70cQSUK41dLsRMqfAg"
	ExistingStickerFileID   = "BQADAgADcwADjMcoCbdl-6eB--YPAg"
)

func getBot(t *testing.T) (*api.BotAPI, error) {
	bot, err := api.NewBotAPI(TestToken)
	if bot != nil {
		bot.Debug = true
	}
	require.NoError(t, err, fmt.Sprintf("%T", err))
	return bot, err
}

func TestNewBotAPI_notoken(t *testing.T) {
	_, err := api.NewBotAPI("")
	require.Error(t, err)
}

func TestGetUpdates(t *testing.T) {
	bot, _ := getBot(t)
	u := api.NewUpdate(0)
	_, err := bot.GetUpdates(u)
	require.NoError(t, err)
}

func TestSendWithMessage(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = "markdown"
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithMessageReply(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	msg.ReplyToMessageID = ReplyToMessageID
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithMessageForward(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewForward(ChatID, ChatID, ReplyToMessageID)
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewPhoto(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewPhotoUpload(ChatID, "tests/image.jpg")
	msg.Caption = "Test"
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewPhotoWithFileBytes(t *testing.T) {
	bot, _ := getBot(t)
	data, _ := ioutil.ReadFile("tests/image.jpg")
	b := api.FileBytes{Name: "image.jpg", Bytes: data}
	msg := api.NewPhotoUpload(ChatID, b)
	msg.Caption = "Test"
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewPhotoWithFileReader(t *testing.T) {
	bot, _ := getBot(t)
	f, _ := os.Open("tests/image.jpg")
	reader := api.FileReader{Name: "image.jpg", Reader: f, Size: -1}
	msg := api.NewPhotoUpload(ChatID, reader)
	msg.Caption = "Test"
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewPhotoReply(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewPhotoUpload(ChatID, "tests/image.jpg")
	msg.ReplyToMessageID = ReplyToMessageID
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithExistingPhoto(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewPhotoShare(ChatID, ExistingPhotoFileID)
	msg.Caption = "Test"
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewDocument(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewDocumentUpload(ChatID, "tests/image.jpg")
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithExistingDocument(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewDocumentShare(ChatID, ExistingDocumentFileID)
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewAudio(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewAudioUpload(ChatID, "tests/audio.mp3")
	msg.Title = "TEST"
	msg.Duration = 10
	msg.Performer = "TEST"
	msg.MimeType = "audio/mpeg"
	msg.FileSize = 688
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithExistingAudio(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewAudioShare(ChatID, ExistingAudioFileID)
	msg.Title = "TEST"
	msg.Duration = 10
	msg.Performer = "TEST"
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewVoice(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewVoiceUpload(ChatID, "tests/voice.ogg")
	msg.Duration = 10
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithExistingVoice(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewVoiceShare(ChatID, ExistingVoiceFileID)
	msg.Duration = 10
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithContact(t *testing.T) {
	bot, _ := getBot(t)
	contact := api.NewContact(ChatID, "5551234567", "Test")
	if _, err := bot.Send(contact); err != nil {
		t.Error(err)
	}
}

func TestSendWithLocation(t *testing.T) {
	bot, _ := getBot(t)
	_, err := bot.Send(api.NewLocation(ChatID, 40, 40))
	require.NoError(t, err)
}

func TestSendWithVenue(t *testing.T) {
	bot, _ := getBot(t)
	venue := api.NewVenue(ChatID, "A Test Location", "123 Test Street", 40, 40)
	if _, err := bot.Send(venue); err != nil {
		t.Error(err)
	}
}

func TestSendWithNewVideo(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewVideoUpload(ChatID, "tests/video.mp4")
	msg.Duration = 10
	msg.Caption = "TEST"
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithExistingVideo(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewVideoShare(ChatID, ExistingVideoFileID)
	msg.Duration = 10
	msg.Caption = "TEST"
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewVideoNote(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewVideoNoteUpload(ChatID, 240, "tests/videonote.mp4")
	msg.Duration = 10
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithExistingVideoNote(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewVideoNoteShare(ChatID, 240, ExistingVideoNoteFileID)
	msg.Duration = 10
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewSticker(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewStickerUpload(ChatID, "tests/image.jpg")
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithExistingSticker(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewStickerShare(ChatID, ExistingStickerFileID)
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithNewStickerAndKeyboardHide(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewStickerUpload(ChatID, "tests/image.jpg")
	msg.ReplyMarkup = api.ReplyKeyboardRemove{
		RemoveKeyboard: true,
		Selective:      false,
	}
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestSendWithExistingStickerAndKeyboardHide(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewStickerShare(ChatID, ExistingStickerFileID)
	msg.ReplyMarkup = api.ReplyKeyboardRemove{
		RemoveKeyboard: true,
		Selective:      false,
	}
	_, err := bot.Send(msg)
	require.NoError(t, err)
}

func TestGetFile(t *testing.T) {
	bot, _ := getBot(t)
	file := api.FileConfig{FileID: ExistingPhotoFileID}
	_, err := bot.GetFile(file)
	require.NoError(t, err)
}

func TestSendChatConfig(t *testing.T) {
	bot, _ := getBot(t)
	_, err := bot.SendChatAction(ChatID, api.ChatTyping)
	require.NoError(t, err)
}

func TestSendEditMessage(t *testing.T) {
	bot, _ := getBot(t)
	msg, err := bot.Send(api.NewMessage(ChatID, "Testing editing."))
	require.NoError(t, err)
	edit := api.EditMessageTextConfig{
		BaseEdit: api.BaseEdit{
			ChatID:    ChatID,
			MessageID: msg.MessageID,
		},
		Text: "Updated text.",
	}
	_, err = bot.Send(edit)
	require.NoError(t, err)
}

func TestGetUserProfilePhotos(t *testing.T) {
	bot, _ := getBot(t)
	_, err := bot.GetUserProfilePhotos(api.NewUserProfilePhotos(ChatID))
	require.NoError(t, err)
}

func TestSetWebhookWithCert(t *testing.T) {
	bot, _ := getBot(t)
	time.Sleep(time.Second * 2)
	err := bot.Stop()
	require.NoError(t, err)
	wh := api.NewWebhookWithCert("https://example.com/tgbotapi-test/"+bot.Token, "tests/cert.pem")
	_, err = bot.ServeWebHook(wh, "/pattern")
	require.NoError(t, err)
	require.True(t, bot.IsServed())
	require.True(t, bot.IsServedWebhook())
	require.False(t, bot.IsServedLongPoll())
	require.False(t, bot.IsStopped())
	_, err = bot.GetWebhookInfo()
	require.NoError(t, err)
	err = bot.Stop()
	require.NoError(t, err)
	require.False(t, bot.IsServed())
	require.False(t, bot.IsServedWebhook())
	require.False(t, bot.IsServedLongPoll())
	require.True(t, bot.IsStopped())
}

func TestSetWebhookWithoutCert(t *testing.T) {
	bot, _ := getBot(t)
	time.Sleep(time.Second * 2)
	err := bot.Stop()
	require.NoError(t, err)
	wh := api.NewWebhook("https://example.com/tgbotapi-test/" + bot.Token)
	_, err = bot.ServeWebHook(wh, "/pattern")
	require.NoError(t, err)
	require.True(t, bot.IsServed())
	require.True(t, bot.IsServedWebhook())
	require.False(t, bot.IsServedLongPoll())
	require.False(t, bot.IsStopped())
	info, err := bot.GetWebhookInfo()
	require.NoError(t, err)
	log.Println(info)
	err = bot.Stop()
	require.NoError(t, err)
	require.False(t, bot.IsServed())
	require.False(t, bot.IsServedWebhook())
	require.False(t, bot.IsServedLongPoll())
	require.True(t, bot.IsStopped())
}

func TestUpdatesChan(t *testing.T) {
	bot, _ := getBot(t)
	if bot.GetUpdatesChan() == nil {
		t.Fail()
	}
}

func TestLongPolling(t *testing.T) {
	bot, _ := getBot(t)
	ucfg := api.NewUpdate(0)
	ucfg.Timeout = 5
	err := bot.ServeLongPoll(ucfg)
	require.NoError(t, err)
	require.True(t, bot.IsServed())
	require.True(t, bot.IsServedLongPoll())
	require.False(t, bot.IsServedWebhook())
	require.False(t, bot.IsStopped())
	time.Sleep(5 * time.Second)
	err = bot.Stop()
	require.NoError(t, err)
	require.False(t, bot.IsServed())
	require.False(t, bot.IsServedLongPoll())
	require.False(t, bot.IsServedWebhook())
	require.True(t, bot.IsStopped())
}

func TestSendWithMediaGroup(t *testing.T) {
	bot, _ := getBot(t)

	cfg := api.NewMediaGroup(ChatID,
		api.NewInputMediaPhoto("https://i.imgur.com/unQLJIb.jpg"),
		api.NewInputMediaPhoto("https://i.imgur.com/J5qweNZ.jpg"),
		api.NewInputMediaVideo("https://i.imgur.com/F6RmI24.mp4"),
	)
	_, err := bot.SendMediaGroup(cfg)
	require.NoError(t, err)
}

func ExampleNewBotAPI() {
	bot, err := api.NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := api.NewUpdate(0)
	u.Timeout = 60
	if err := bot.ServeLongPoll(u); err != nil {
		panic(err)
	}
	updates := bot.GetUpdatesChan()
	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()
	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		msg := api.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID
		if sent, err := bot.Send(msg); err != nil {
			_ = sent
			panic(err)
		}
	}
	if err := bot.Stop(); err != nil {
		panic(err)
	}
}

func ExampleNewWebhook() {
	bot, err := api.NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	whc := api.NewWebhookWithCert("https://www.google.com:8443/"+bot.Token, "cert.pem")
	srv, err := bot.ServeWebHook(whc, "/"+bot.Token)
	if err != nil {
		panic(err)
	}
	info, err := bot.GetWebhookInfo()
	if err != nil {
		panic(err)
	}
	if info.LastErrorDate != 0 {
		panic(fmt.Sprintf("[Telegram callback failed]%s", info.LastErrorMessage))
	}
	go func() {
		if err := srv.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem"); err != nil {
			panic(err)
		}
	}()
	for update := range bot.GetUpdatesChan() {
		log.Printf("%+v\n", update)
	}
}

func ExampleAnswerInlineQuery() {
	bot, err := api.NewBotAPI("MyAwesomeBotToken") // create new bot
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := api.NewUpdate(0)
	u.Timeout = 60
	if err := bot.ServeLongPoll(u); err != nil {
		panic(err)
	}
	updates := bot.GetUpdatesChan()
	for update := range updates {
		if update.InlineQuery == nil { // if no inline query, ignore it
			continue
		}
		article := api.NewInlineQueryResultArticle(update.InlineQuery.ID, "Echo", update.InlineQuery.Query)
		article.Description = update.InlineQuery.Query
		inlineConf := api.InlineConfig{
			InlineQueryID: update.InlineQuery.ID,
			IsPersonal:    true,
			CacheTime:     0,
			Results:       []interface{}{article},
		}
		if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
			log.Println(err)
		}
	}
}

func TestDeleteMessage(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = "markdown"
	message, _ := bot.Send(msg)
	deleteMessageConfig := api.DeleteMessageConfig{
		ChatID:    message.Chat.ID,
		MessageID: message.MessageID,
	}
	_, err := bot.DeleteMessage(deleteMessageConfig)
	require.NoError(t, err)
}

func TestPinChatMessage(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewMessage(SupergroupChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = "markdown"
	message, _ := bot.Send(msg)
	pinChatMessageConfig := api.PinChatMessageConfig{
		ChatID:              message.Chat.ID,
		MessageID:           message.MessageID,
		DisableNotification: false,
	}
	_, err := bot.PinChatMessage(pinChatMessageConfig)
	require.NoError(t, err)
}

func TestUnpinChatMessage(t *testing.T) {
	bot, _ := getBot(t)
	msg := api.NewMessage(SupergroupChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = "markdown"
	message, _ := bot.Send(msg)
	// We need pin message to unpin something
	pinChatMessageConfig := api.PinChatMessageConfig{
		ChatID:              message.Chat.ID,
		MessageID:           message.MessageID,
		DisableNotification: false,
	}
	_, err := bot.PinChatMessage(pinChatMessageConfig)
	unpinChatMessageConfig := api.UnpinChatMessageConfig{
		ChatID: message.Chat.ID,
	}
	_, err = bot.UnpinChatMessage(unpinChatMessageConfig)
	require.NoError(t, err)
}
