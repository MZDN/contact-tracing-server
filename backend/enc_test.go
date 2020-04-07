package backend

import (
	crypto_rand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"golang.org/x/crypto/nacl/box"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

func TestEncryption(t *testing.T) {
	senderPublicKey, senderPrivateKey, err := box.GenerateKey(crypto_rand.Reader)
	if err != nil {
		panic(err)
	}

	recipientPublicKey, recipientPrivateKey, err := box.GenerateKey(crypto_rand.Reader)
	if err != nil {
		panic(err)
	}

	msg := []byte("severe fever,coughing,hard to breathe")
	encrypted, err := Encrypt(msg, recipientPublicKey, senderPrivateKey)
	if err != nil {
		t.Fatalf("err:%v", err)
	}
	fmt.Printf("Encrypted:%x len(RawMsg)=%d, len(EncMsg)=%d\n", encrypted, len(msg), len(encrypted))

	decrypted, err := Decrypt(encrypted, senderPublicKey, recipientPrivateKey)
	if err != nil {
		panic("decryption error")
	}
	fmt.Printf("Decrypted:%v len(RawMsg)=%d, len(EncMsg)=%d\n", string(decrypted), len(msg), len(encrypted))
}

func TestCENReport(t *testing.T) {
	_, senderPrivateKey, err := box.GenerateKey(crypto_rand.Reader)
	if err != nil {
		panic(err)
	}

	recipientPublicKey, _, err := box.GenerateKey(crypto_rand.Reader)
	if err != nil {
		panic(err)
	}

	symptoms := []byte("high fever, dry cough,hard to breathe")
	report, err := MakeCENReport(symptoms, recipientPublicKey, senderPrivateKey)
	if err != nil {
		panic(err)
	}
	r, err := json.Marshal(report)
	if err != nil {
		panic(err)
	}
	fmt.Printf("msg: %v\n", string(r))
}

func TestGCM(t *testing.T) {
	plaintext := []byte("severe fever,coughing,hard to breathe")

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(err.Error())
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	fmt.Printf("Key:\t\t%x", key)
	fmt.Printf("\nCiphertext:\t%x", ciphertext)

	unsealedtext, err := aesgcm.Open(nil, nonce, []byte(ciphertext), nil)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("\nunsealedtext:\t%v", string(unsealedtext))
}
