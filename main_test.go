package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func randID() int64 {
	t := time.Now()
	rand.NewSource(t.UnixNano())
	return rand.Int63()
}

// BotAPIInterface is an interface that matches the methods used from tgbotapi.BotAPI
type BotAPIInterface interface {
	GetUpdatesChan(tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}

// MockBotAPI is a mock implementation of the BotAPIInterface
type MockBotAPI struct {
	mock.Mock
}

func (m *MockBotAPI) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	args := m.Called(config)
	return args.Get(0).(tgbotapi.UpdatesChannel)
}

func TestTelebot(t *testing.T) {
	mockBot := new(MockBotAPI)

	// updates - 0xc00018ede0
	// u - {Offset:0 Limit:0 Timeout:60 AllowedUpdates:[]}
	mockBot.On("GetUpdatesChan", mock.Anything).Return("0xc00018ede0")

	updates := telebot(&tgbotapi.BotAPI{})

	assert.NotNil(t, updates, "UpdatesChannel should not be nil")

	mockBot.AssertExpectations(t)
}

func TestCritical(t *testing.T) {
	critChance := 0.1
	isCrit := rand.Float64() <= critChance
	t.Run("param", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			t.Logf("isCrit = %v\n rand.Float64() = %.f\n", isCrit, rand.Float64())
		}
	})
}

func Test_lootMobs(t *testing.T) {
	tests := []struct {
		name   string
		chatID int64
	}{
		{"1", 1},
		{"2", -1},
		{"3", 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lootMobs(tt.chatID)
		})
	}
}

func Test_nextRoom(t *testing.T) {
	type args struct {
		bot     *tgbotapi.BotAPI
		message *tgbotapi.Message
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{
				bot:     &tgbotapi.BotAPI{},
				message: &tgbotapi.Message{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextRoom(tt.args.bot, tt.args.message)
		})
	}
}

func TestRoom(t *testing.T) {
	texts := []string{"Ну что, монстр, как тебе мой уровень мастерства?",
		"Еще один враг на моем пути! Неужели они не понимают, что я — герой?",
		"Похоже, этот монстр не читал мой гайд по победе!",
		"Снова победа! Как же скучно побеждать таких слабаков!",
		"Монстр, ты был великолепен... в своих мечтах!",
		"Я думал, будет сложнее. Может, в следующий раз выбери кого-то посильнее?",
		"Еще один монстр в списке моих жертв. У кого-то явно неудачный день!",
		"Этот монстр не знал, с кем связался. Теперь он знает!",
		"Победа! Не забудьте оставить отзыв о моем мастерстве!",
		"Монстр, ты был хорош, но, увы, я — лучше!",
		"Тень повержена, но страх остается.",
		"Каждая победа — это лишь шаг к новой тьме.",
		"Монстр мертв, но его крики еще звучат в моей голове.",
		"Смерть одного — это начало страха для других.",
		"Я победил, но цена была высока.",
		"Кровь на моих руках, и это лишь начало.",
		"Победа — это иллюзия, скрывающая настоящую тьму.",
		"Каждый враг, которого я убиваю, делает меня немного более бездушным.",
		"Монстр пал, но его тень навсегда останется со мной.",
		"Я победил, но в этом мире нет места для истинного триумфа."}
	fmt.Println(len(texts))
	for i := 0; i < 1000; i++ {
		r := rand.Intn(len(texts) + 20)
		if r >= 20 {
			fmt.Printf("%d: %d из 20\n", i, r)
		} else {
			fmt.Printf("%d: %d из 20\n%s\n", i, r, texts[r])
		}
	}
}
