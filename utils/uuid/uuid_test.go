package uuid

import (
	"fmt"
	"strconv"
	"testing"
)

func TestUUID(t *testing.T) {
	fmt.Println(UUID())
}

func TestInt32UUID(t *testing.T) {
	fmt.Println(Int32UUID())
}

func TestUUIDWithLen(t *testing.T) {
	fmt.Printf("%x\n", genMsgID(0x01, 0x02))
}

func genMsgID(lNode, rNode byte) uint64 {
	id := UUIDWithLen(12)
	fmt.Println(id)
	id64, _ := strconv.ParseInt(id, 16, 64)
	fmt.Printf("%x\n", id64)
	return uint64(int64(lNode)<<56 | int64(rNode)<<48 | id64)
}

/*

Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^BenchmarkUUIDWithLen$ myftp/utils/uuid

goos: linux
goarch: amd64
pkg: myftp/utils/uuid
cpu: Intel(R) Core(TM) i5-10500 CPU @ 3.10GHz
BenchmarkUUIDWithLen
BenchmarkUUIDWithLen-2
 2513677               480.6 ns/op            96 B/op          3 allocs/op
PASS
ok      myftp/utils/uuid        1.705s

*/
func BenchmarkUUIDWithLen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		UUIDWithLen(32)
	}
}
