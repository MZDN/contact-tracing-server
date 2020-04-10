package backend

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

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

func TestCTReport(t *testing.T) {

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
	fmt.Printf("\nPublic key A (sender) %x (%x %x)\n", FromECDSAPub(sPub), sPub.X, sPub.Y)
	fmt.Printf("\nPublic key B (recipient) %x (%x,%x)\n", FromECDSAPub(rPub), rPub.X, rPub.Y)

	synctime := time.Now().Unix() // The time both A and B seen each other
	synctimeByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(synctimeByte, uint64(synctime))

	sss := GenerateSessionSecret(rPub, sPriv) //sender generated ss
	rss := GenerateSessionSecret(sPub, rPriv) //recipient generated ss
	fmt.Printf("\nsession secret: \n[sss=%x]\n[rss=%x]\nSync Time:%d\n", sss, rss, synctime)
	if !bytes.Equal(sss[:], rss[:]) {
		panic("session secret mismatch!")
	}

	//Step 1.a - A sends msgA with these bytes: [PK_A(t), siga, ts] to B
	//memoA := &FindMyPKMemo{ReportType: 0, DiseaseID: 0, SymptomID: []int32{7}}
	fmt.Printf("\nA Signed msgA with synctimeByte=[%x](%d)\n", synctimeByte, synctime)

	msgA, err := Sign(sPriv, synctimeByte)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nmsgA=(prefix,pk,sig,ts)=%x\n", msgA)
	DecryptedCiphertextA, err := VerifySign(msgA)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nB Verified msgA: DecryptedCiphertextA=[%x](%d)\n", DecryptedCiphertextA, int64(binary.LittleEndian.Uint64(DecryptedCiphertextA)))

	//Step 1.b - B sends msgB with these bytes: [PK_B(t), sigb, ts] to A
	fmt.Printf("\nB Signed msgB with synctimeByte=[%x](%d)\n", synctimeByte, synctime)

	msgB, err := Sign(rPriv, synctimeByte)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nmsgB=(prefix,pk,sig,ts)=%x\n", msgB)
	DecryptedCiphertextB, err := VerifySign(msgB)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nA Verified msgB: DecryptedCiphertextB=[%x](%d)\n", DecryptedCiphertextB, int64(binary.LittleEndian.Uint64(DecryptedCiphertextB)))

	//Step2 - Encryption. A is sick and computes EncodedMsg following a protobuf serialization scheme within a CTReport R using (1b)'s PK_B(t)
	memoS := &ContactTracingMemo{ReportType: 1, DiseaseID: 2, SymptomID: []int32{5, 7, 123}}
	memoByteS, err := proto.Marshal(memoS)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	fmt.Printf("\nA Made report for B(pub=%x)\nencoded memo:%x\n", FromECDSAPub(sPub), memoByteS)
	report, err := MakeCTReport(rPub, sPriv, memoByteS)
	if err != nil {
		panic(err)
	}
	r, err := json.Marshal(report)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nA submitted report for B: %v\n", string(r))

	//A's report, which is included in reports []CTReport sent to ProcessReport
	//TODO: backend.ProcessReport([]CTReport{report})

	//Step3.a - Decryption. B queries a report by hitting ProcessQuery(query []byte, timestamp uint64) and getting back reports []CTReport one of which contains R of Step2
	//TODO: backend.ProcessQuery(query []byte, timestamp uint64)

	//Step3.b - B filter query with seen map[hashedPk](pub,rss).
	//TODO: B filter query with seen map[hashedPk](pub,rss)

	//Step3.c - B Then uses recipient session secret generate by (1a)'s PK_A(t) to decode to the memo, following a protobuf deserialization scheme
	fmt.Printf("\nB decrypting report from A using session secret(rss=%x)\n", rss[:])

	dmemoByteS, err := DecryptCTReport(report, rss[:])
	if err != nil {
		panic(err)
	}
	vMemoS := &ContactTracingMemo{}
	err = proto.Unmarshal(dmemoByteS, vMemoS)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}
	fmt.Printf("\nB Decrypted Memo from A: {reportType=%v, vMemoS=[%s]} \nDecrypted memoByte=[%x])\n", vMemoS.GetReportType(), vMemoS.String(), dmemoByteS)
}
