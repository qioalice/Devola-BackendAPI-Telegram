package api_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/qioalice/devola-backend-telegram/api"
)

func TestNewInlineQueryResultArticle(t *testing.T) {
	result := api.NewInlineQueryResultArticle("id", "title", "message")

	require.False(t, result.Type != "article")
	require.False(t, result.ID != "id")
	require.False(t, result.Title != "title")
	require.False(t, result.InputMessageContent.(api.InputTextMessageContent).Text != "message")
}

func TestNewInlineQueryResultArticleMarkdown(t *testing.T) {
	result := api.NewInlineQueryResultArticleMarkdown("id", "title", "*message*")

	require.False(t, result.Type != "article")
	require.False(t, result.ID != "id")
	require.False(t, result.Title != "title")
	require.False(t, result.InputMessageContent.(api.InputTextMessageContent).Text != "*message*")
	require.False(t, result.InputMessageContent.(api.InputTextMessageContent).ParseMode != "Markdown")
}

func TestNewInlineQueryResultArticleHTML(t *testing.T) {
	result := api.NewInlineQueryResultArticleHTML("id", "title", "<b>message</b>")

	require.False(t, result.Type != "article")
	require.False(t, result.ID != "id")
	require.False(t, result.Title != "title")
	require.False(t, result.InputMessageContent.(api.InputTextMessageContent).Text != "<b>message</b>")
	require.False(t, result.InputMessageContent.(api.InputTextMessageContent).ParseMode != "HTML")
}

func TestNewInlineQueryResultGIF(t *testing.T) {
	result := api.NewInlineQueryResultGIF("id", "google.com")

	require.False(t, result.Type != "gif")
	require.False(t, result.ID != "id")
	require.False(t, result.URL != "google.com")
}

func TestNewInlineQueryResultMPEG4GIF(t *testing.T) {
	result := api.NewInlineQueryResultMPEG4GIF("id", "google.com")

	require.False(t, result.Type != "mpeg4_gif")
	require.False(t, result.ID != "id")
	require.False(t, result.URL != "google.com")
}

func TestNewInlineQueryResultPhoto(t *testing.T) {
	result := api.NewInlineQueryResultPhoto("id", "google.com")

	require.False(t, result.Type != "photo")
	require.False(t, result.ID != "id")
	require.False(t, result.URL != "google.com")
}

func TestNewInlineQueryResultPhotoWithThumb(t *testing.T) {
	result := api.NewInlineQueryResultPhotoWithThumb("id", "google.com", "thumb.com")

	require.False(t, result.Type != "photo")
	require.False(t, result.ID != "id")
	require.False(t, result.URL != "google.com")
	require.False(t, result.ThumbURL != "thumb.com")
}

func TestNewInlineQueryResultVideo(t *testing.T) {
	result := api.NewInlineQueryResultVideo("id", "google.com")

	require.False(t, result.Type != "video")
	require.False(t, result.ID != "id")
	require.False(t, result.URL != "google.com")
}

func TestNewInlineQueryResultAudio(t *testing.T) {
	result := api.NewInlineQueryResultAudio("id", "google.com", "title")

	require.False(t, result.Type != "audio")
	require.False(t, result.ID != "id")
	require.False(t, result.URL != "google.com")
	require.False(t, result.Title != "title")
}

func TestNewInlineQueryResultVoice(t *testing.T) {
	result := api.NewInlineQueryResultVoice("id", "google.com", "title")

	require.False(t, result.Type != "voice")
	require.False(t, result.ID != "id")
	require.False(t, result.URL != "google.com")
	require.False(t, result.Title != "title")
}

func TestNewInlineQueryResultDocument(t *testing.T) {
	result := api.NewInlineQueryResultDocument("id", "google.com", "title", "mime/type")

	require.False(t, result.Type != "document")
	require.False(t, result.ID != "id")
	require.False(t, result.URL != "google.com")
	require.False(t, result.Title != "title")
	require.False(t, result.MimeType != "mime/type")
}

func TestNewInlineQueryResultLocation(t *testing.T) {
	result := api.NewInlineQueryResultLocation("id", "name", 40, 50)

	require.False(t, result.Type != "location")
	require.False(t, result.ID != "id")
	require.False(t, result.Title != "name")
	require.False(t, result.Latitude != 40)
	require.False(t, result.Longitude != 50)
}

func TestNewEditMessageText(t *testing.T) {
	edit := api.NewEditMessageText(ChatID, ReplyToMessageID, "new text")

	require.False(t, edit.Text != "new text")
	require.False(t, edit.BaseEdit.ChatID != ChatID)
	require.False(t, edit.BaseEdit.MessageID != ReplyToMessageID)
}

func TestNewEditMessageCaption(t *testing.T) {
	edit := api.NewEditMessageCaption(ChatID, ReplyToMessageID, "new caption")

	require.False(t, edit.Caption != "new caption")
	require.False(t, edit.BaseEdit.ChatID != ChatID)
	require.False(t, edit.BaseEdit.MessageID != ReplyToMessageID)
}

func TestNewEditMessageReplyMarkup(t *testing.T) {
	markup := api.InlineKeyboardMarkup{
		InlineKeyboard: [][]api.InlineKeyboardButton{
			{
				{Text: "test"},
			},
		},
	}

	edit := api.NewEditMessageReplyMarkup(ChatID, ReplyToMessageID, markup)

	require.False(t, edit.ReplyMarkup.InlineKeyboard[0][0].Text != "test")
	require.False(t, edit.BaseEdit.ChatID != ChatID)
	require.False(t, edit.BaseEdit.MessageID != ReplyToMessageID)
}
