package backend

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"io"
	"log"
	"math/big"

	"github.com/gogo/protobuf/proto"
)

const (
	PrefixSize      = 3
	PublicKeyPrefix = 0
	RawSigPrefix    = 1
	MemoPrefix      = 2
)

// ECDSASignature storing R,S for DER?
type ECDSASignature struct {
	R, S *big.Int
}

// P256 encapsulates the ECDSA P256 curve
func P256() elliptic.Curve {
	return elliptic.P256()
}

// FromECDSAPub maps a ecdsa PublicKey into raw byte
func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(P256(), pub.X, pub.Y)
}

// ByteToPublicKey maps a public key in byte form into a public key
func ByteToPublicKey(pubByte []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(P256(), pubByte)
	if x == nil {
		return nil, fmt.Errorf("Invalid pubkey")
	}
	pub := ecdsa.PublicKey{Curve: P256(), X: x, Y: y}
	return &pub, nil
}

// Sign uses the private key to sign a memo m
func Sign(priv *ecdsa.PrivateKey, m []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, priv, m)
	if err != nil {
		return nil, err
	}
	signed, err := asn1.Marshal(ECDSASignature{r, s})
	if err != nil {
		return nil, err
	}

	pubkey := FromECDSAPub(&priv.PublicKey) // this is uncompressed 65 bytes key
	prefix := make([]byte, PrefixSize)
	prefix[PublicKeyPrefix] = byte(len(pubkey))
	prefix[RawSigPrefix] = byte(len(signed))
	prefix[MemoPrefix] = byte(len(m))                            // prefix[len(pkey),len(sig),len(m)]
	sigWithPubkeyMemo := append(append(pubkey, signed...), m...) // [PK, sig, memo]
	fullsig := append(prefix, sigWithPubkeyMemo...)              // [prefix,PK, sig, memo]
	fmt.Printf("\n[pk=%d,sig=%d,m=%d]\npk=%x\nsig=%x\nm=%x\n", len(pubkey), len(signed), len(m), pubkey, signed, m)
	return fullsig, nil
}

// VerifySign uses the signature and a message m to verify against a public key
func VerifySign(signature []byte) ([]byte, error) {
	prefix := signature[:PrefixSize]
	pubKeySize := prefix[PublicKeyPrefix]
	rawSigSize := prefix[RawSigPrefix]

	sigWithPubkeyMemo := signature[PrefixSize:]

	pubByte := sigWithPubkeyMemo[:pubKeySize] // PK of [PK, sig, memo]
	pub, err := ByteToPublicKey(pubByte)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\nPublic key %x (%x %x)\n", FromECDSAPub(pub), pub.X, pub.Y)

	signed := sigWithPubkeyMemo[pubKeySize : pubKeySize+rawSigSize] // sig of [PK, sig, memo]
	fmt.Printf("\nrawsig=%x ,len(rawsig)=%d\n", signed, len(signed))
	rawsig := &ECDSASignature{}
	_, err = asn1.Unmarshal(signed, rawsig)
	if err != nil {
		return nil, err
	}

	m := sigWithPubkeyMemo[pubKeySize+rawSigSize:] // m of [PK, sig, memo]
	if ok := ecdsa.Verify(pub, m, rawsig.R, rawsig.S); !ok {
		return nil, fmt.Errorf("signature invalid")
	}
	return m, nil
}

// Encrypt encodes a plaintext msg using session secret and returns ciphertext after
func Encrypt(ss []byte, msg []byte) ([]byte, error) {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	block, err := aes.NewCipher(ss)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	cipherText := aesgcm.Seal(nil, nonce, msg, nil)
	cipherText = append(nonce, cipherText...)
	return cipherText, nil
}

// Decrypt decode a cipherText using session secret and returns plaintext msg
func Decrypt(ss []byte, cipherText []byte) ([]byte, error) {
	decryptNonce := make([]byte, 12)
	copy(decryptNonce[:], cipherText[:12])
	block, err := aes.NewCipher(ss)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plainText, err := aesgcm.Open(nil, decryptNonce, cipherText[12:], nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}

// MakeCENReport generate the cenReport using rPub, sPriv, memoByte
func MakeCENReport(rPub *ecdsa.PublicKey, sPriv *ecdsa.PrivateKey, memoByte []byte) (report CENReport, err error) {
	ss := GenerateSessionSecret(rPub, sPriv)
	encrypted, err := Encrypt(ss[:], memoByte)
	if err != nil {
		return report, err
	}
	hashedPK := Computehash(FromECDSAPub(rPub))
	report = CENReport{HashedPK: hashedPK, EncodedMsg: encrypted}
	return report, nil
}

// GenerateSessionSecret computes session secret. Result should be stored as tuple map[hashedPK] => (pub_K, ss)
func GenerateSessionSecret(rPub *ecdsa.PublicKey, sPriv *ecdsa.PrivateKey) [32]byte {
	a, _ := rPub.Curve.ScalarMult(rPub.X, rPub.Y, sPriv.D.Bytes())
	ss := sha256.Sum256(a.Bytes())
	return ss
}

// DecryptCENReport decrypts CENReport using session secret. Should be invoked only if hashedPK we found
func DecryptCENReport(report CENReport, ss []byte) (memoByte []byte, err error) {
	cipherText := report.EncodedMsg
	if memoByte, err = Decrypt(ss, cipherText); err != nil {
		return memoByte, err
	}
	return memoByte, err
}

func (m *Memo) Bytes() (mbyte []byte) {
	mByte, err := proto.Marshal(m)
	if err != nil {
		log.Fatal("marshaling error: ", err)
		return nil
	}
	return mByte
}

func FromMemoByte(mbyte []byte) (m *Memo) {
	m = &Memo{}
	err := proto.Unmarshal(mbyte, m)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}
	return m
}
