package backend

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"testing"

	cloudkms "cloud.google.com/go/kms/apiv1"
	"google.golang.org/api/iterator"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

const (
	ProjectID          = "wolk-1307"
	LocationID         = "global" // Location of the key rings.
	RingID             = "contactTracing"
	KeyID              = "mkkey"
	CryptoKeyVersionID = "1"
)

func TestKMS(t *testing.T) {

	// Create the KMS client.
	ctx := context.Background()
	client, err := cloudkms.NewKeyManagementClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	keyRingParent := fmt.Sprintf("projects/%s/locations/%s", ProjectID, LocationID) // The resource name of the key rings.
	fmt.Printf("keyRingParent: %v\n", keyRingParent)
	//var b bytes.Buffer

	// Build the CreateKeyRing request.
	req := &kmspb.CreateKeyRingRequest{
		Parent:    keyRingParent,
		KeyRingId: RingID,
	}
	// Call the CreateKeyRing API.
	//keyRingName := projects/PROJECT_ID/locations/global/keyRings

	//create KeyRing
	var keyring *kmspb.KeyRing
	keyring, err = client.CreateKeyRing(ctx, req)
	if err != nil {
		fmt.Printf("Failed to Created key rings err: %v\n", err)
	} else {
		fmt.Printf("Created key ring: %s\n", keyring)
	}

	// Build the listkeyRing request.
	req1 := &kmspb.ListKeyRingsRequest{
		Parent: keyRingParent,
	}
	// Query the listkeyRing API.
	it := client.ListKeyRings(ctx, req1)

	// Iterate and print results.
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to list key rings: %v\n", err)
		}
		fmt.Printf("Existed KeyRing: %q\n", resp.Name)
	}
	//var b bytes.Buffer

	// Build the CreateCryptoKey(ASYMMETRIC) request.
	cryptoKeyParent := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", ProjectID, LocationID, RingID) // The resource name of the key rings.

	req2 := &kmspb.CreateCryptoKeyRequest{
		Parent:      cryptoKeyParent,
		CryptoKeyId: KeyID,
		CryptoKey: &kmspb.CryptoKey{
			Purpose: kmspb.CryptoKey_ASYMMETRIC_SIGN,
			VersionTemplate: &kmspb.CryptoKeyVersionTemplate{
				Algorithm: kmspb.CryptoKeyVersion_EC_SIGN_P256_SHA256,
			},
		},
	}
	// Call the CreateCryptoKey API.
	var ckey *kmspb.CryptoKey
	ckey, err = client.CreateCryptoKey(ctx, req2)
	if err != nil {
		fmt.Printf("Failed to CreateCryptoKey: %v\n", err)
	} else {
		fmt.Printf("Created crypto key. %s", ckey)
	}
}

func TestKMSSignAndVerify(t *testing.T) {
	ctx := context.Background()
	client, err := cloudkms.NewKeyManagementClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	//Sign with P256
	// Find the digest of the message? hashed of List<DailyTracingKey>||TS?
	//projects/PROJECT_ID/locations/global/keyRings/RING_ID/cryptoKeys/KEY_ID/cryptoKeyVersions/1
	asymName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s/cryptoKeyVersions/%s", ProjectID, LocationID, RingID, KeyID, CryptoKeyVersionID) // The resource name of the key rings.
	message := make([]byte, 512)
	if _, err := io.ReadFull(rand.Reader, message); err != nil {
		panic(err.Error())
	}

	// digest := sha256.New()
	// digest.Write(message)
	digest := Computehash(message)
	// Build the signing request.
	req3 := &kmspb.AsymmetricSignRequest{
		Name: asymName,
		Digest: &kmspb.Digest{
			Digest: &kmspb.Digest_Sha256{
				Sha256: digest,
			},
		},
	}

	// Call the AsymmetricSign API.
	var AsymResp *kmspb.AsymmetricSignResponse
	var sig []byte
	AsymResp, err = client.AsymmetricSign(ctx, req3)
	if err != nil {
		fmt.Printf("AsymmetricSign err %s", err)
	} else {
		sig = AsymResp.Signature
		// Return the signature bytes.
		fmt.Printf("AsymmetricSign: sig %x, len(sig)=%d\n", sig, len(sig))
	}

	//Verify:
	// Retrieve the public key from KMS.
	pubKeyresp, err := client.GetPublicKey(ctx, &kmspb.GetPublicKeyRequest{Name: asymName})
	if err != nil {
		fmt.Printf("GetPublicKey err %s", err)
	}
	block, _ := pem.Decode([]byte(pubKeyresp.Pem))
	abstractKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("x509.ParsePKIXPublicKey err %s", err)
	}
	ecKey, ok := abstractKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatalf("key at '%s' is not EC", asymName)
	}
	// Verify Elliptic Curve signature.
	var parsedSig struct{ R, S *big.Int }
	if _, err = asn1.Unmarshal(sig, &parsedSig); err != nil {
		t.Fatalf("asn1.Unmarshal: %v", err)
	}
	msgHash := Computehash(message) //hashed of List<DailyTracingKey>||TS?
	if !ecdsa.Verify(ecKey, msgHash, parsedSig.R, parsedSig.S) {
		t.Fatalf("ecdsa.Verify failed on key: %s", asymName)
	} else {
		fmt.Printf("ecdsa.Verify success on key %s", asymName)
	}

}
