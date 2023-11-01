package signal

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/hkdf"
)

const (
	labelImageKeys    = "WhatsApp Image Keys"
	labelAudioKeys    = "WhatsApp Audio Keys"
	labelMutationKeys = "WhatsApp Mutation Keys"
	labelHistoryKeys  = "WhatsApp History Keys"
	labelDocumentKeys = "WhatsApp Document Keys"
	labelVideoKeys    = "WhatsApp Video Keys"
)

func decodeHex(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}

func decodeBase64(s string) []byte {
	b, _ := base64.StdEncoding.DecodeString(s)
	return b
}

func zlibCompress(b []byte) []byte {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write(b)
	w.Close()
	return buf.Bytes()
}

func AesCbcPkcs7Encrypt(key, iv []byte, plaintext []byte) (ciphertext []byte, err error) {
	// create aes-128-cbc block cipher
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCEncrypter(c, iv)

	// add pkcs7 padding
	paddingLength := c.BlockSize() - len(plaintext)%c.BlockSize()
	padding := bytes.Repeat([]byte{byte(paddingLength)}, paddingLength)
	plaintext = append(plaintext, padding...)
	ciphertext = make([]byte, len(plaintext))

	for i := 0; i < len(ciphertext); i += c.BlockSize() {
		out := ciphertext[i:][:c.BlockSize()]
		in := plaintext[i:][:c.BlockSize()]
		cbc.CryptBlocks(out, in)
	}

	return ciphertext, nil
}

func AesCbcPkcs7Decrypt(key, iv []byte, ciphertext []byte) (plaintext []byte, err error) {
	// create aes-128-cbc block cipher
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCDecrypter(c, iv)
	plaintext = make([]byte, len(ciphertext))

	for i := 0; i < len(plaintext); i += c.BlockSize() {
		out := plaintext[i:][:c.BlockSize()]
		in := ciphertext[i:][:c.BlockSize()]
		cbc.CryptBlocks(out, in)
	}

	// remove pkcs7 padding
	plaintext = plaintext[:len(plaintext)-int(plaintext[len(plaintext)-1])]
	return plaintext, nil
}

func encryptMediaBytes(key, iv, hmacKey []byte, plaintext []byte) (ciphertext []byte, mac []byte, err error) {
	// create aes-128-cbc block cipher
	ciphertext, err = AesCbcPkcs7Encrypt(key, iv, plaintext)
	if err != nil {
		return nil, nil, err
	}

	if hmacKey != nil {
		h := hmac.New(sha256.New, hmacKey)
		h.Write(iv)
		h.Write(ciphertext)
		mac = h.Sum(nil)[:0xA]
	}

	return ciphertext, mac, nil
}

func decryptMediaBytes(key, iv, hmacKey []byte, ciphertext []byte) (plaintext []byte, mac []byte, err error) {
	i := len(ciphertext) - len(ciphertext)%aes.BlockSize
	ciphertext, mac = ciphertext[:i], ciphertext[i:]

	plaintext, err = AesCbcPkcs7Decrypt(key, iv, ciphertext)
	if err != nil {
		return nil, nil, err
	}

	if hmacKey != nil {
		h := hmac.New(sha256.New, hmacKey)
		h.Write(iv)
		h.Write(ciphertext)

		if bytes.Compare(mac, h.Sum(nil)[:len(mac)]) != 0 {
			return nil, nil, errors.New("message authentication failed")
		}
	}

	return plaintext, mac, nil
}

func expandMediaKeys(key, label []byte) ([]byte, []byte, []byte) {
	b := hkdf.Extract(sha256.New, key, bytes.Repeat([]byte("\x00"), 32))
	expander := hkdf.Expand(sha256.New, b, label)

	bb := make([]byte, 32*4)

	expander.Read(bb[:32])
	expander.Read(bb[32:][:32])
	expander.Read(bb[64:][:32])
	expander.Read(bb[96:][:32])

	aesKey, iv := bb[16:][:32], bb[:16]
	hmacKey := bb[48:][:32]

	return aesKey, iv, hmacKey
}

func DecryptMediaBytes(ciphertext, key []byte, fileSha256, fileEncSha256 []byte, label []byte) ([]byte, error) {
	aesKey, iv, hmacKey := expandMediaKeys(key, label)
	plaintext, _, err := decryptMediaBytes(aesKey, iv, hmacKey, ciphertext)
	if err != nil {
		return nil, err
	}

	fileEncHasher := sha256.New()
	fileEncHasher.Write(ciphertext)

	fileHasher := sha256.New()
	fileHasher.Write(plaintext)

	if bytes.Compare(fileEncSha256, fileEncHasher.Sum(nil)) != 0 {
		return nil, errors.New("invalid fileEncSha256")
	}
	if bytes.Compare(fileSha256, fileHasher.Sum(nil)) != 0 {
		return nil, errors.New("invalid fileSha256")
	}

	return plaintext, nil
}

func EncryptMediaBytes(plaintext, key []byte, label []byte) (ciphertext, fileSha256, fileEncSha256 []byte, err error) {
	aesKey, iv, hmacKey := expandMediaKeys(key, label)
	ciphertext, mac, err := encryptMediaBytes(aesKey, iv, hmacKey, plaintext)
	if err != nil {
		return nil, nil, nil, err
	}

	fileEncHasher := sha256.New()
	fileEncHasher.Write(ciphertext)
	fileEncHasher.Write(mac)
	fileEncSha256 = fileEncHasher.Sum(nil)

	fileHasher := sha256.New()
	fileHasher.Write(plaintext)
	fileSha256 = fileHasher.Sum(nil)

	ciphertext = append(ciphertext, mac...)
	return
}

//
//func EncryptSyncPatch(patchType string, syncKeyId, syncKey []byte, nonce []byte, data ...*AppStateSyncActionData) (*SyncdPatch, error) {
//	b := hkdf.Extract(sha256.New, syncKey, bytes.Repeat([]byte("\x00"), 32))
//	expander := hkdf.Expand(sha256.New, b, []byte(labelMutationKeys))
//
//	bb := make([]byte, 32*5)
//
//	expander.Read(bb[:32])
//	expander.Read(bb[32:][:32])
//	expander.Read(bb[64:][:32])
//	expander.Read(bb[96:][:32])
//	expander.Read(bb[128:][:32])
//
//	indexHmacKey := bb[:32]
//	aesKey := bb[32:][:32]
//	hmacKey := bb[64:][:32]
//	snapshotHmacKey := bb[96:][:32]
//	patchHmacKey := bb[128:][:32]
//
//	var patch SyncdPatch
//
//	for _, d := range data {
//		bbb, _ := proto.Marshal(d)
//		ciphertext, err := AesCbcPkcs7Encrypt(aesKey, nonce, bbb)
//		if err != nil {
//			return nil, err
//		}
//
//		h := hmac.New(sha512.New, hmacKey)
//
//		h.Write([]byte("\x01"))
//		h.Write(syncKeyId)
//		h.Write(nonce)
//		h.Write(ciphertext)
//		h.Write([]byte("\x00\x00\x00\x00\x00\x00\x00\x07"))
//
//		hh := hmac.New(sha256.New, indexHmacKey)
//		hh.Write(d.Index)
//		hh.Sum(nil)
//
//		patch.Mutations = append(patch.Mutations, &SyncdMutation{
//			Operation: SyncdMutation_SET.Enum(),
//			Record: &SyncdRecord{
//				Index: &SyncdIndex{Blob: hh.Sum(nil)},
//				Value: &SyncdValue{Blob: append(append(nonce, ciphertext...), h.Sum(nil)[:32]...)},
//				KeyId: &KeyId{Id: syncKeyId},
//			},
//		})
//	}
//
//	h1 := hmac.New(sha256.New, snapshotHmacKey)
//	h2 := hmac.New(sha256.New, patchHmacKey)
//
//	h2.Write(h1.Sum(nil))
//
//	for _, mut := range patch.Mutations {
//		h2.Write(mut.Record.Value.Blob[len(mut.Record.Value.Blob)-32:])
//	}
//
//	h2.Write([]byte("\x00\x00\x00\x00\x00\x00\x00\x01"))
//	h2.Write([]byte(patchType))
//
//	patch.SnapshotMac = h1.Sum(nil)
//	patch.PatchMac = h2.Sum(nil)
//	patch.KeyId = &KeyId{Id: syncKeyId}
//	patch.DeviceIndex = proto.Uint32(0)
//
//	return &patch, nil
//}
