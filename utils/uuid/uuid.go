package uuid

import (
	"hash/crc32"
	"strings"

	uuid "github.com/satori/go.uuid"
)

const UIDLen = 12 // default len

// UUIDWithLen 指定长度UUID
func UUIDWithLen(length int) string {
	uuid := uuid.NewV4().String()
	uuid = strings.Replace(uuid, "-", "", -1)
	return uuid[:length]
}

// UUID
func UUID() string {
	return UUIDWithLen(UIDLen)
}

// int32 UUID
func Int32UUID() int32 {
	uuid := uuid.NewV4()
	uuidHash := int32(crc32.ChecksumIEEE([]byte(uuid.String())))
	if uuidHash < 0 {
		return uuidHash * -1
	}
	return uuidHash
}
