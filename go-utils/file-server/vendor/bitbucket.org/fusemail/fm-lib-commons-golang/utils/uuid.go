package utils

// Provides uuid utilities.
// uuid is based on msgid and recipient
// The algorithm is based on java portal code.
// Details in https://fusemail.atlassian.net/browse/MAIL-961

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

func valueOf(c uint8) int {
	//base62_chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
	base62IndexMap := [...]int{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, // 0-9 is 0 to 9
		0, 0, 0, 0, 0, 0, 0,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, // A-Z is 10 to 35
		0, 0, 0, 0, 0, 0,
		36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61} // a-z is 36 to 61

	if c > 'z' || c < '0' {
		return 0
	}

	return base62IndexMap[c-'0']

}

// DecodeBase62 is used to get a timestamp out of an exim message ID, e.g. 1dQzET-0002Sg-4E
// The first 6 chars are a base62 encoded 32 bit integer
func DecodeBase62(encoded string) int64 {
	var decoded int64
	var round float64
	decoded = 0
	round = 0
	for i := len(encoded) - 1; i >= 0; i-- {
		decoded += int64(float64(valueOf(encoded[i])) * math.Pow(62, float64(round)))
		round++
	}
	return decoded
}

func getUUID(ts int64, hashStr string) []byte {
	var hc int64
	buf := new(bytes.Buffer)

	hash := sha256.New()
	hash.Write([]byte(hashStr))
	md := hash.Sum(nil)

	for i := 20; i < 24; i++ {
		ts = ts<<8 | int64(md[i]&0xff)
	}

	for i := 24; i < 32; i++ {
		hc = hc<<8 | int64(md[i]&0xff)
	}

	binary.Write(buf, binary.BigEndian, ts)
	binary.Write(buf, binary.BigEndian, hc)

	return buf.Bytes()
}

// MsgRecipientUUID generates UUID for msgid and recipient.
// uuid is created based on msigid and recipient
// get timestamp by base62-decoding msgid first 6 character
// get sha256-hash on msgid+lowercase-recipient
// hexdump the shifted timestamp and hash
func MsgRecipientUUID(msgid, recipient string) string {
	var ts int64
	if len(msgid) < 6 { // re MAIL-1220 panic
		ts = time.Now().Unix()
	} else {
		ts = DecodeBase62(msgid[0:6])
	}
	recipient = strings.ToLower(recipient)
	return MsgUUID(ts, msgid+recipient)
}

// MsgUUID generates UUID according to timestamp and string.
func MsgUUID(ts int64, hashStr string) string {
	uid := getUUID(ts, hashStr)
	return fmt.Sprintf("%X", uid)
}
