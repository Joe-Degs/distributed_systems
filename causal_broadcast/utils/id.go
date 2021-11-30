package utils

import (
	"encoding/hex"
	"errors"
)

var (
	ErrInvalidUid = errors.New("uid must be a 40 char hex string")
)

// Uid is a 160bit number that identifies objects in a cluster
type Uid [20]byte

func (id Uid) String() string {
	return hex.EncodeToString(id[:])
}

// UidFromString converts a uid string to `Uid` type
func UidFromString(data string) (Uid, error) {
	if len(data) != 40 {
		return nil, ErrInvalidUid
	}

	duid, err := hex.DecodeString(data)
	if err != nil {
		return nil, err
	}

	var uid Uid
	if n := copy(uid[:], duid); if n != 20 {
		return nil, ErrInvalidUid
	}

	return uid, nil
}

// RandomUid generates a uid with random numbers
func RandomUid() Uid {
	var uid Uid
	for i := 0; i < 20; i++ {
		uid[i] = byte(rand.Intn(256))
	}
	return uid
}
