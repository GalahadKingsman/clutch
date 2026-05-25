package solana

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func (c *Client) USDCBalance(ctx context.Context, owner solana.PublicKey, mint string) (float64, error) {
	mintPK, err := solana.PublicKeyFromBase58(mint)
	if err != nil {
		return 0, err
	}
	ata, _, err := solana.FindAssociatedTokenAddress(owner, mintPK)
	if err != nil {
		return 0, err
	}
	out, err := c.rpc.GetTokenAccountBalance(ctx, ata, rpc.CommitmentConfirmed)
	if err != nil {
		// No ATA yet → 0 balance
		return 0, nil
	}
	if out.Value == nil || out.Value.UiAmount == nil {
		return 0, nil
	}
	return *out.Value.UiAmount, nil
}
