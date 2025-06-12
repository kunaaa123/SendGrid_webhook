package verify

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
)

// ecdsaSignature ใช้เก็บค่า R และ S ของ signature
type ecdsaSignature struct {
	R *big.Int
	S *big.Int
}

// VerifySignature ตรวจสอบความถูกต้องของ signature
func VerifySignature(payload []byte, signature string, timestamp string, publicKey string) (bool, error) {
	// แปลง Public Key จาก PEM format
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return false, fmt.Errorf("failed to parse PEM block")
	}

	// แปลง Public Key
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse public key: %v", err)
	}

	// ตรวจสอบว่าเป็น ECDSA Public Key
	ecdsaKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("key is not an ECDSA public key")
	}

	// ถอดรหัส signature จาก base64
	decodedSig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %v", err)
	}

	// แปลง signature เป็น R,S values
	var sig ecdsaSignature
	if _, err := asn1.Unmarshal(decodedSig, &sig); err != nil {
		return false, fmt.Errorf("failed to unmarshal signature: %v", err)
	}

	// สร้าง hash จาก timestamp + payload
	h := sha256.New()
	h.Write([]byte(timestamp))
	h.Write(payload)
	hash := h.Sum(nil)

	// เพิ่ม logging details
	details := fmt.Sprintf("Payload length: %d, Timestamp: %s", len(payload), timestamp)
	fmt.Printf("Verification Details: %s\n", details)

	// ตรวจสอบ signature
	isValid := ecdsa.Verify(ecdsaKey, hash, sig.R, sig.S)

	if isValid {
		fmt.Printf("Signature verification successful\n")
	} else {
		fmt.Printf("Signature verification failed\n")
	}

	return isValid, nil
}
