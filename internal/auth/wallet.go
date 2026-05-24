package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gagliardetto/solana-go"
)

// BuildSignMessage returns the message users sign to link a Solana wallet.
func BuildSignMessage(nonce string) string {
	return fmt.Sprintf("CLUTCH wants you to sign in with your Solana account.\n\nNonce: %s", nonce)
}

// VerifySolanaSignature checks ed25519 signature for a base58 Solana pubkey.
func VerifySolanaSignature(walletAddress, message, signatureB64 string) error {
	pubkey, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		return fmt.Errorf("invalid wallet address: %w", err)
	}

	sig, err := decodeSignature(signatureB64)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	if !ed25519.Verify(ed25519.PublicKey(pubkey[:]), []byte(message), sig) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}

func decodeSignature(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if raw, err := base64.StdEncoding.DecodeString(s); err == nil {
		return raw, nil
	}
	if raw, err := base64.RawStdEncoding.DecodeString(s); err == nil {
		return raw, nil
	}
	return nil, fmt.Errorf("unsupported signature encoding")
}
