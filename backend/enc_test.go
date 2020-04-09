package backend

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"crypto/rand"

	"github.com/golang/protobuf/proto"
)

func TestEncryption(t *testing.T) {
	rPriv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	sPriv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	rPub := &rPriv.PublicKey
	sPub := &sPriv.PublicKey

	fmt.Printf("\nPrivate key (sender) %x", sPriv.D)
	fmt.Printf("\nPublic key (sender) (%x %x)\n", sPub.X, sPub.Y)
	fmt.Printf("\nPublic key (recipient) (%x,%x)\n", rPub.X, rPub.Y)

	sss := GenerateSessionSecret(rPub, sPriv) //sender generated ss
	rss := GenerateSessionSecret(sPub, rPriv) //recipient generated ss
	fmt.Printf("\nsession secret: [sss=%x] [rss=%x] \n", sss, rss)
	if !bytes.Equal(sss[:], rss[:]) {
		panic("session secret mismatch!")
	}

	msg := []byte("severe fever,coughing,hard to breathe")
	encrypted, err := Encrypt(sss[:], msg)
	if err != nil {
		t.Fatalf("err:%v", err)
	}
	fmt.Printf("\nEncrypted:%x len(RawMsg)=%d, len(EncMsg)=%d\n", encrypted, len(msg), len(encrypted))

	decrypted, err := Decrypt(rss[:], encrypted)
	if err != nil {
		panic("decryption error")
	}
	fmt.Printf("Decrypted:%v len(RawMsg)=%d, len(EncMsg)=%d\n", string(decrypted), len(msg), len(encrypted))
}

func TestCENReport(t *testing.T) {

	fmt.Printf("--ECC Parameters--\n")
	fmt.Printf(" Name: %s\n", elliptic.P256().Params().Name)
	fmt.Printf(" N: %x\n", elliptic.P256().Params().N)
	fmt.Printf(" P: %x\n", elliptic.P256().Params().P)
	fmt.Printf(" Gx: %x\n", elliptic.P256().Params().Gx)
	fmt.Printf(" Gy: %x\n", elliptic.P256().Params().Gy)
	fmt.Printf(" Bitsize: %x\n\n", elliptic.P256().Params().BitSize)

	rPriv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	sPriv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	rPub := &rPriv.PublicKey
	sPub := &sPriv.PublicKey

	fmt.Printf("\nPrivate key (sender) %x", sPriv.D)
	fmt.Printf("\nPublic key (sender) %x (%x %x)\n", FromECDSAPub(sPub), sPub.X, sPub.Y)
	fmt.Printf("\nPublic key (recipient) %x (%x,%x)\n", FromECDSAPub(rPub), rPub.X, rPub.Y)

	sss := GenerateSessionSecret(rPub, sPriv) //sender generated ss
	rss := GenerateSessionSecret(sPub, rPriv) //recipient generated ss
	fmt.Printf("\nsession secret: [sss=%x] [rss=%x] \n", sss, rss)
	if !bytes.Equal(sss[:], rss[:]) {
		panic("session secret mismatch!")
	}

	//memo := []byte("high fever, dry cough,hard to breathe")
	//memo := &Memo{ReportType: 1, DiseaseID: 2, SymptomID: []int32{5, 7, 123}}

	//Step 1.a - A sends these bytes: [PK_A(t), siga, memoA] to B
	memoA := &Memo{ReportType: 0, DiseaseID: 0, SymptomID: []int32{7}}
	memoByteA := memoA.Bytes()
	sigA, err := Sign(sPriv, memoByteA)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nSigned: memoByteA=[%x]\nsigA=%x\n", memoByteA, sigA)
	vmemoByteA, err := VerifySign(sigA)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nVerified: vmemoByteA=[%x]\n", vmemoByteA)
	vMemoA := FromMemoByte(vmemoByteA)
	fmt.Printf("\nDecrypted Msg: {reportType=%v, vmemoByteA=[%s]} (msg=[%x])\n", vMemoA.GetReportType(), vMemoA.String(), vmemoByteA)

	//Step 1.b - B sends these bytes: [PK_B(t), sigb, memoB] to A
	memoB := &Memo{ReportType: 0, DiseaseID: 0}
	memoByteB := memoB.Bytes()
	sigB, err := Sign(sPriv, memoByteB)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nSigned: memoByteB=[%x]\nsigB=%x\n", memoByteB, sigB)
	vmemoByteB, err := VerifySign(sigB)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nVerified: vmemoByteB=[%x]\n", vmemoByteB)
	vMemoB := FromMemoByte(vmemoByteB)
	fmt.Printf("\nDecrypted Msg: {reportType=%v, vmemoByteB=[%s]} (msg=[%x])\n", vMemoB.GetReportType(), vMemoB.String(), vmemoByteB)

	//Step2 - Encryption. A is sick and computes EncodedMsg following a protobuf serialization scheme within a CENReport R using (1b)'s PK_B(t)
	memoS := &Memo{ReportType: 1, DiseaseID: 2, SymptomID: []int32{5, 7, 123}}
	memoByteS, err := proto.Marshal(memoS)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	report, err := MakeCENReport(rPub, sPriv, memoByteS)
	if err != nil {
		panic(err)
	}
	r, err := json.Marshal(report)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nreport: %v\n", string(r))

	//A's report, which is included in reports []CENReport sent to ProcessReport
	//TODO: backend.ProcessReport([]CENReport{report})

	//Step3.a - Decryption. B queries a report by hitting ProcessQuery(query []byte, timestamp uint64) and getting back reports []CENReport one of which contains R of Step2
	//TODO: backend.ProcessQuery(query []byte, timestamp uint64)

	//Step3.b - B filter query with seen map[hashedPk](pub,rss).
	//TODO: B filter query with seen map[hashedPk](pub,rss)

	//Step3.c - B Then uses recipient session secret generate by (1a)'s PK_A(t) to decode to the memo, following a protobuf deserialization scheme
	dmemoByteS, err := DecryptCENReport(report, rss[:])
	if err != nil {
		panic(err)
	}
	vMemoS := &Memo{}
	err = proto.Unmarshal(dmemoByteS, vMemoS)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}
	fmt.Printf("\nDecrypted Memo: {reportType=%v, vMemoS=[%s]} (msg=[%x])\n", vMemoS.GetReportType(), vMemoS.String(), dmemoByteS)
}
