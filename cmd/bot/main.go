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
	UpdateID int64 `json:"update_id"`
	Message  *struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		From struct {
			ID int64 `json:"id"`
		} `json:"from"`
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
			if item.Message == nil || item.Message.Text == "" {
				continue
			}
			chatID := item.Message.Chat.ID
			telegramID := strconv.FormatInt(item.Message.From.ID, 10)

			reply := handler.Handle(ctx, chatID, telegramID, item.Message.Text)
			if reply.Text == "" {
				continue
			}
			if err := client.sendMessage(ctx, chatID, reply.Text); err != nil {
				log.Printf("kirim pesan gagal: %v", err)
			}
		}
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

func (client *telegramClient) sendMessage(ctx context.Context, chatID int64, text string) error {
	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(chatID, 10))
	form.Set("text", text)
	form.Set("disable_web_page_preview", "true")

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", client.token)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}
	request.URL.RawQuery = form.Encode()

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
