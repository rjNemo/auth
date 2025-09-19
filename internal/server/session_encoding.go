package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
)

func encodeSession(state SessionState, secret []byte) (string, error) {
	payload, err := json.Marshal(state)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	sig := mac.Sum(nil)

	combined := append(payload, sig...)
	return base64.RawURLEncoding.EncodeToString(combined), nil
}

func decodeSession(raw string, secret []byte) (SessionState, error) {
	var state SessionState

	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return state, err
	}

	if len(decoded) <= sha256.Size {
		return state, errors.New("session payload too small")
	}

	payload := decoded[:len(decoded)-sha256.Size]
	providedSig := decoded[len(decoded)-sha256.Size:]

	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	expectedSig := mac.Sum(nil)

	if !hmac.Equal(providedSig, expectedSig) {
		return state, errors.New("session signature mismatch")
	}

	if err := json.Unmarshal(payload, &state); err != nil {
		return state, err
	}

	return state, nil
}
