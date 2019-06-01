package api_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/qioalice/devola-backend-telegram/api"
)

func TestUserStringWith(t *testing.T) {
	user := api.User{
		ID:           0,
		FirstName:    "Test",
		LastName:     "Test",
		UserName:     "",
		LanguageCode: "en",
		IsBot:        false,
	}
	require.Equal(t, "Test Test", user.String())
}

func TestUserStringWithUserName(t *testing.T) {
	user := api.User{
		ID:           0,
		FirstName:    "Test",
		LastName:     "Test",
		UserName:     "@test",
		LanguageCode: "en",
	}
	require.Equal(t, "@test", user.String())
}

func TestMessageTime(t *testing.T) {
	message := api.Message{Date: 0}
	require.Equal(t, time.Unix(0, 0), message.Time())
}

func TestMessageIsCommandWithCommand(t *testing.T) {
	message := api.Message{Text: "/command"}
	message.Entities = []api.MessageEntity{{Type: "bot_command", Offset: 0, Length: 8}}
	require.True(t, message.IsCommand())
}

func TestIsCommandWithText(t *testing.T) {
	message := api.Message{Text: "some text"}
	require.False(t, message.IsCommand())
}

func TestIsCommandWithEmptyText(t *testing.T) {
	message := api.Message{Text: ""}
	require.False(t, message.IsCommand())
}

func TestCommandWithCommand(t *testing.T) {
	message := api.Message{Text: "/command"}
	message.Entities = []api.MessageEntity{{Type: "bot_command", Offset: 0, Length: 8}}
	require.Equal(t, "command", message.Command())
}

func TestCommandWithEmptyText(t *testing.T) {
	message := api.Message{Text: ""}
	require.Empty(t, message.Command())
}

func TestCommandWithNonCommand(t *testing.T) {
	message := api.Message{Text: "test text"}
	require.Empty(t, message.Command())
}

func TestCommandWithBotName(t *testing.T) {
	message := api.Message{Text: "/command@testbot"}
	message.Entities = []api.MessageEntity{{Type: "bot_command", Offset: 0, Length: 16}}
	require.Equal(t, "command", message.Command())
}

func TestCommandWithAtWithBotName(t *testing.T) {
	message := api.Message{Text: "/command@testbot"}
	message.Entities = []api.MessageEntity{{Type: "bot_command", Offset: 0, Length: 16}}
	require.Equal(t, "command@testbot", message.CommandWithAt())
}

func TestMessageCommandArgumentsWithArguments(t *testing.T) {
	message := api.Message{Text: "/command with arguments"}
	message.Entities = []api.MessageEntity{{Type: "bot_command", Offset: 0, Length: 8}}
	require.Equal(t, "with arguments", message.CommandArguments())
}

func TestMessageCommandArgumentsWithMalformedArguments(t *testing.T) {
	message := api.Message{Text: "/command-without argument space"}
	message.Entities = []api.MessageEntity{{Type: "bot_command", Offset: 0, Length: 8}}
	require.Equal(t, "without argument space", message.CommandArguments())
}

func TestMessageCommandArgumentsWithoutArguments(t *testing.T) {
	message := api.Message{Text: "/command"}
	require.Empty(t, message.CommandArguments())
}

func TestMessageCommandArgumentsForNonCommand(t *testing.T) {
	message := api.Message{Text: "test text"}
	require.Empty(t, message.CommandArguments())
}

func TestMessageEntityParseURLGood(t *testing.T) {
	entity := api.MessageEntity{URL: "https://www.google.com"}
	_, err := entity.ParseURL()
	require.NoError(t, err)
}

func TestMessageEntityParseURLBad(t *testing.T) {
	entity := api.MessageEntity{URL: ""}
	_, err := entity.ParseURL()
	require.Error(t, err)
}

func TestChatIsPrivate(t *testing.T) {
	chat := api.Chat{ID: 10, Type: "private"}
	require.True(t, chat.IsPrivate())
}

func TestChatIsGroup(t *testing.T) {
	chat := api.Chat{ID: 10, Type: "group"}
	require.True(t, chat.IsGroup())
}

func TestChatIsChannel(t *testing.T) {
	chat := api.Chat{ID: 10, Type: "channel"}
	require.True(t, chat.IsChannel())
}

func TestChatIsSuperGroup(t *testing.T) {
	chat := api.Chat{ID: 10, Type: "supergroup"}
	require.True(t, chat.IsSuperGroup())
}

func TestMessageEntityIsMention(t *testing.T) {
	entity := api.MessageEntity{Type: "mention"}
	require.True(t, entity.IsMention())
}

func TestMessageEntityIsHashtag(t *testing.T) {
	entity := api.MessageEntity{Type: "hashtag"}
	require.True(t, entity.IsHashtag())
}

func TestMessageEntityIsBotCommand(t *testing.T) {
	entity := api.MessageEntity{Type: "bot_command"}
	require.True(t, entity.IsCommand())
}

func TestMessageEntityIsUrl(t *testing.T) {
	entity := api.MessageEntity{Type: "url"}
	require.True(t, entity.IsUrl())
}

func TestMessageEntityIsEmail(t *testing.T) {
	entity := api.MessageEntity{Type: "email"}
	require.True(t, entity.IsEmail())
}

func TestMessageEntityIsBold(t *testing.T) {
	entity := api.MessageEntity{Type: "bold"}
	require.True(t, entity.IsBold())
}

func TestMessageEntityIsItalic(t *testing.T) {
	entity := api.MessageEntity{Type: "italic"}
	require.True(t, entity.IsItalic())
}

func TestMessageEntityIsCode(t *testing.T) {
	entity := api.MessageEntity{Type: "code"}
	require.True(t, entity.IsCode())
}

func TestMessageEntityIsPre(t *testing.T) {
	entity := api.MessageEntity{Type: "pre"}
	require.True(t, entity.IsPre())
}

func TestMessageEntityIsTextLink(t *testing.T) {
	entity := api.MessageEntity{Type: "text_link"}
	require.True(t, entity.IsTextLink())
}

func TestFileLink(t *testing.T) {
	file := api.File{FilePath: "test/test.txt"}
	require.Equal(t, "https://api.telegram.org/file/bottoken/test/test.txt", file.Link("token"))
}
