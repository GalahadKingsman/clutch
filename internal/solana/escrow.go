package solana

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// Devnet USDC mint (Circle faucet).
const DevnetUSDCMint = "4zMMC9srt5Ri5X14GAgXhaHii3GnPEhPSZeSRmuqXnE"

var (
	discCreateDuel = anchorDisc("global:create_duel")
	discAcceptDuel = anchorDisc("global:accept_duel")
)

type Client struct {
	rpc     *rpc.Client
	program solana.PublicKey
}

func NewClient(rpcURL, programID string) (*Client, error) {
	c := &Client{rpc: rpc.New(rpcURL)}
	if programID != "" {
		prog, err := solana.PublicKeyFromBase58(programID)
		if err != nil {
			return nil, fmt.Errorf("invalid program id: %w", err)
		}
		c.program = prog
	}
	return c, nil
}

func (c *Client) HasProgram() bool {
	return c.program != (solana.PublicKey{})
}

func (c *Client) requireProgram() error {
	if !c.HasProgram() {
		return fmt.Errorf("CLUTCH_PROGRAM_ID is not set")
	}
	return nil
}

func anchorDisc(name string) []byte {
	sum := sha256.Sum256([]byte(name))
	return sum[:8]
}

func (c *Client) DuelPDA(creator solana.PublicKey) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{[]byte("duel"), creator.Bytes()},
		c.program,
	)
}

func (c *Client) RecentBlockhash(ctx context.Context) (solana.Hash, error) {
	out, err := c.rpc.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return solana.Hash{}, err
	}
	return out.Value.Blockhash, nil
}

func (c *Client) BuildCreateDuelTx(
	ctx context.Context,
	creator solana.PublicKey,
	stakeUSD uint64,
	deadline time.Time,
) (serialized string, duelPDA solana.PublicKey, err error) {
	if err := c.requireProgram(); err != nil {
		return "", solana.PublicKey{}, err
	}
	pda, _, err := c.DuelPDA(creator)
	if err != nil {
		return "", solana.PublicKey{}, err
	}

	data := make([]byte, 8+8+8)
	copy(data[:8], discCreateDuel)
	binary.LittleEndian.PutUint64(data[8:16], stakeUSD)
	binary.LittleEndian.PutUint64(data[16:24], uint64(deadline.Unix()))

	ix := solana.NewInstruction(
		c.program,
		solana.AccountMetaSlice{
			solana.Meta(creator).WRITE().SIGNER(),
			solana.Meta(pda).WRITE(),
			solana.Meta(solana.SystemProgramID),
		},
		data,
	)

	bh, err := c.RecentBlockhash(ctx)
	if err != nil {
		return "", solana.PublicKey{}, err
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{ix},
		bh,
		solana.TransactionPayer(creator),
	)
	if err != nil {
		return "", solana.PublicKey{}, err
	}

	raw, err := tx.MarshalBinary()
	if err != nil {
		return "", solana.PublicKey{}, err
	}
	return base64.StdEncoding.EncodeToString(raw), pda, nil
}

func (c *Client) BuildAcceptDuelTx(
	ctx context.Context,
	opponent, creator solana.PublicKey,
) (serialized string, err error) {
	if err := c.requireProgram(); err != nil {
		return "", err
	}
	pda, _, err := c.DuelPDA(creator)
	if err != nil {
		return "", err
	}

	ix := solana.NewInstruction(
		c.program,
		solana.AccountMetaSlice{
			solana.Meta(opponent).SIGNER(),
			solana.Meta(pda).WRITE(),
			solana.Meta(creator),
		},
		discAcceptDuel,
	)

	bh, err := c.RecentBlockhash(ctx)
	if err != nil {
		return "", err
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{ix},
		bh,
		solana.TransactionPayer(opponent),
	)
	if err != nil {
		return "", err
	}

	raw, err := tx.MarshalBinary()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}

func (c *Client) VerifyConfirmedTx(ctx context.Context, signature string) error {
	sig, err := solana.SignatureFromBase58(signature)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}
	out, err := c.rpc.GetSignatureStatuses(ctx, true, sig)
	if err != nil {
		return err
	}
	if len(out.Value) == 0 || out.Value[0] == nil {
		return fmt.Errorf("transaction not found")
	}
	st := out.Value[0]
	if st.Err != nil {
		return fmt.Errorf("transaction failed on-chain")
	}
	if st.ConfirmationStatus != rpc.ConfirmationStatusFinalized &&
		st.ConfirmationStatus != rpc.ConfirmationStatusConfirmed {
		return fmt.Errorf("transaction not confirmed yet")
	}
	return nil
}

func (c *Client) SOLBalance(ctx context.Context, wallet solana.PublicKey) (float64, error) {
	out, err := c.rpc.GetBalance(ctx, wallet, rpc.CommitmentConfirmed)
	if err != nil {
		return 0, err
	}
	return float64(out.Value) / float64(solana.LAMPORTS_PER_SOL), nil
}
