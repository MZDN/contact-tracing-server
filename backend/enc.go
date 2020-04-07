package backend

import (
	crypto_rand "crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/box"
)

type CENReportV4 struct {
	HashedPK   []byte `json:"hashedPK"`
	EncodedMsg []byte `json:"encodedMsg"` // enc( “1”, PK_A)
	// protobuf of ints symptoms, diseases, signature
	// just to prevent attack, verified against EncMsg
	// NO user timestamp
}

func Encrypt(symptoms []byte, recipientPublicKey *[32]byte, senderPrivateKey *[32]byte) (encrypted []byte, err error) {
	var nonce [24]byte
	if _, err := io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		return nil, err
	}
	encrypted = box.Seal(nonce[:], symptoms, &nonce, recipientPublicKey, senderPrivateKey)
	return encrypted, nil
}

func Decrypt(encrypted []byte, senderPublicKey *[32]byte, recipientPrivateKey *[32]byte) (decrypted []byte, err error) {
	var decryptNonce [24]byte
	copy(decryptNonce[:], encrypted[:24])
	decrypted, ok := box.Open(nil, encrypted[24:], &decryptNonce, senderPublicKey, recipientPrivateKey)
	if !ok {
		return nil, fmt.Errorf("decryption error")
	}
	fmt.Println(string(decrypted))
	return decrypted, nil
}

func MakeCENReport(symptoms []byte, recipientPublicKey *[32]byte, senderPrivateKey *[32]byte) (report CENReport, err error) {
	encrypted, err := Encrypt(symptoms, recipientPublicKey, senderPrivateKey)
	if err != nil {
		return report, err
	}
	var rPub []byte
	copy(rPub[:], recipientPublicKey[:32])
	hashedPK := Computehash(rPub)
	report = CENReport{HashedPK: hashedPK, EncodedMsg: encrypted}
	return report, nil
}
