package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jhopan/pan-wallet-telegram-bot/internal/bot"
	"github.com/jhopan/pan-wallet-telegram-bot/internal/config"
	"github.com/jhopan/pan-wallet-telegram-bot/internal/store"
	"github.com/jhopan/pan-wallet-telegram-bot/internal/walletapi"
)

func main() {
	settings, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	client := walletapi.New(settings.PanWalletURL)
	handler := &bot.Handler{
		API:      walletapi.NewAdapter(client),
		Sessions: store.New(),
	}
	telegram := &telegramClient{token: settings.BotToken, http: &http.Client{Timeout: 60 * time.Second}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("sinyal berhenti diterima, menutup bot...")
		cancel()
	}()

	log.Printf("bot jalan. Pan-Wallet: %s", settings.PanWalletURL)
	if err := telegram.poll(ctx, handler); err != nil && ctx.Err() == nil {
		log.Fatal(err)
	}
	log.Println("bot berhenti.")
}

// telegramClient adalah klien Bot API minimal (tanpa library eksternal).
type telegramClient struct {
	token  string
	http   *http.Client
	offset int64
}

type update struct {
	UpdateID      int64            `json:"update_id"`
	Message       *incomingMessage `json:"message"`
	CallbackQuery *callbackQuery   `json:"callback_query"`
}

type incomingMessage struct {
	Text string `json:"text"`
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	From struct {
		ID int64 `json:"id"`
	} `json:"from"`
}

type callbackQuery struct {
	ID   string `json:"id"`
	Data string `json:"data"`
	From struct {
		ID int64 `json:"id"`
	} `json:"from"`
	Message struct {
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

// poll mengambil pesan baru terus-menerus (long polling).
func (client *telegramClient) poll(ctx context.Context, handler *bot.Handler) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		updates, err := client.getUpdates(ctx)
		if err != nil {
			log.Printf("ambil pesan gagal: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}
		for _, item := range updates {
			client.offset = item.UpdateID + 1
			client.dispatch(ctx, handler, item)
		}
	}
}

// dispatch menangani pesan teks maupun tekan tombol.
func (client *telegramClient) dispatch(ctx context.Context, handler *bot.Handler, item update) {
	if item.CallbackQuery != nil {
		query := item.CallbackQuery
		chatID := query.Message.Chat.ID
		telegramID := strconv.FormatInt(query.From.ID, 10)

		reply := handler.HandleCallback(ctx, chatID, telegramID, query.Data)
		// Telegram wajib dijawab agar spinner tombol berhenti.
		_ = client.answerCallback(ctx, query.ID)
		if reply.Text != "" {
			if err := client.sendMessage(ctx, chatID, reply.Text, reply.Keyboard); err != nil {
				log.Printf("kirim pesan gagal: %v", err)
			}
		}
		return
	}

	if item.Message == nil || item.Message.Text == "" {
		return
	}
	chatID := item.Message.Chat.ID
	telegramID := strconv.FormatInt(item.Message.From.ID, 10)

	reply := handler.Handle(ctx, chatID, telegramID, item.Message.Text)
	if reply.Text == "" {
		return
	}
	if err := client.sendMessage(ctx, chatID, reply.Text, reply.Keyboard); err != nil {
		log.Printf("kirim pesan gagal: %v", err)
	}
}

func (client *telegramClient) getUpdates(ctx context.Context) ([]update, error) {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?timeout=30&offset=%d",
		client.token, client.offset)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var payload struct {
		OK     bool     `json:"ok"`
		Result []update `json:"result"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if !payload.OK {
		return nil, fmt.Errorf("Telegram menolak getUpdates (HTTP %d)", response.StatusCode)
	}
	return payload.Result, nil
}

func (client *telegramClient) sendMessage(ctx context.Context, chatID int64, text string, keyboard *bot.Keyboard) error {
	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(chatID, 10))
	form.Set("text", text)
	form.Set("disable_web_page_preview", "true")

	if keyboard != nil && len(keyboard.Rows) > 0 {
		if encoded, err := encodeKeyboard(keyboard); err == nil {
			form.Set("reply_markup", encoded)
		}
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", client.token)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		return fmt.Errorf("sendMessage HTTP %d", response.StatusCode)
	}
	return nil
}

// answerCallback menutup spinner pada tombol yang ditekan.
func (client *telegramClient) answerCallback(ctx context.Context, callbackID string) error {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", client.token)
	form := url.Values{}
	form.Set("callback_query_id", callbackID)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return nil
}

// encodeKeyboard mengubah Keyboard menjadi JSON inline_keyboard Telegram.
func encodeKeyboard(keyboard *bot.Keyboard) (string, error) {
	type button struct {
		Text         string `json:"text"`
		CallbackData string `json:"callback_data,omitempty"`
		URL          string `json:"url,omitempty"`
	}
	var rows [][]button
	for _, row := range keyboard.Rows {
		var encoded []button
		for _, item := range row {
			encoded = append(encoded, button{Text: item.Text, CallbackData: item.Callback, URL: item.URL})
		}
		rows = append(rows, encoded)
	}
	payload, err := json.Marshal(map[string]any{"inline_keyboard": rows})
	return string(payload), err
}
