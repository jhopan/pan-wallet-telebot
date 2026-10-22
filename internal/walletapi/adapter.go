package walletapi

import (
	"context"

	"github.com/jhopan/pan-wallet-telegram-bot/internal/bot"
)

// Adapter menghubungkan Client ke antarmuka yang dipakai bot.
type Adapter struct {
	client *Client
}

// NewAdapter membungkus client agar sesuai kontrak bot.WalletAPI.
func NewAdapter(client *Client) *Adapter {
	return &Adapter{client: client}
}

func (adapter *Adapter) Resolve(ctx context.Context, telegramID string) (bot.Wallet, error) {
	wallet, err := adapter.client.Resolve(ctx, telegramID)
	if err != nil {
		return bot.Wallet{}, err
	}
	return bot.Wallet{
		PublicID: wallet.PublicID, Balance: wallet.Balance, Held: wallet.Held,
		Available: wallet.Available, Status: wallet.Status, IsNew: wallet.IsNew,
	}, nil
}

func (adapter *Adapter) Balance(ctx context.Context, publicID string) (bot.Wallet, error) {
	wallet, err := adapter.client.Balance(ctx, publicID)
	if err != nil {
		return bot.Wallet{}, err
	}
	return bot.Wallet{
		PublicID: wallet.PublicID, Balance: wallet.Balance, Held: wallet.Held,
		Available: wallet.Available, Status: wallet.Status,
	}, nil
}

func (adapter *Adapter) CreateTopup(ctx context.Context, publicID string, amount int64) (bot.Topup, error) {
	topup, err := adapter.client.CreateTopup(ctx, publicID, amount)
	if err != nil {
		return bot.Topup{}, err
	}
	return bot.Topup{Total: topup.Total, PayURL: adapter.client.PayURL(topup.PayURL)}, nil
}

func (adapter *Adapter) LinkConfirm(ctx context.Context, code, telegramID string) (bot.Wallet, error) {
	wallet, err := adapter.client.LinkConfirm(ctx, code, telegramID)
	if err != nil {
		return bot.Wallet{}, err
	}
	return bot.Wallet{PublicID: wallet.PublicID, Balance: wallet.Balance,
		Held: wallet.Held, Available: wallet.Available, Status: wallet.Status}, nil
}
