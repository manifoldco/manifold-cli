package session

import (
	"bytes"
	"crypto/rand"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/scrypt"

	"github.com/manifoldco/go-base64"
)

const n = 32768
const r = 8
const p = 1
const edSeedSize = 32

func newKeyMaterial(password string) (*string, *string, *string, error) {
	saltBytes := make([]byte, 16)
	_, err := rand.Read(saltBytes)
	if err != nil {
		return nil, nil, nil, err
	}
	salt := base64.New(saltBytes)
	saltStr := salt.String()

	pubKey, _, err := deriveKeypair(password, salt)
	if err != nil {
		return nil, nil, nil, err
	}

	alg := "eddsa"
	pubKeyStr := base64.New(pubKey).String()
	return &alg, &saltStr, &pubKeyStr, nil
}

func deriveKeypair(password string, salt *base64.Value) (
	ed25519.PublicKey, ed25519.PrivateKey, error) {

	// Stretch the password + salt using scrypt
	dk, err := scrypt.Key([]byte(password), []byte(*salt), n, r, p, edSeedSize)
	if err != nil {
		return nil, nil, err
	}

	// Derive ed25519 signing key, using the scrypt output as seed
	return ed25519.GenerateKey(bytes.NewBuffer(dk))
}

func sign(privkey ed25519.PrivateKey, token string) *base64.Value {
	b := ed25519.Sign(privkey, []byte(token))
	return base64.New(b)
}
