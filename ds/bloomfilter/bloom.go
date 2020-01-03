package bloom

import (
	"bytes"
	"encoding/binary"
	"github.com/liyue201/gostl/algorithm/hash"
	"github.com/liyue201/gostl/ds/bitmap"
	"github.com/liyue201/gostl/utils/sync"
	"math"
	gosync "sync"
)

const Salt = "g9hmj2fhgr"

var defaultLocker sync.FakeLocker

// BloomFilter's option
type Option struct {
	locker sync.Locker
}

type Options func(option *Option)

// WithThreadSave use to config BloomFilter with thread safety
func WithThreadSave() Options {
	return func(option *Option) {
		option.locker = &gosync.RWMutex{}
	}
}

// BloomFilter is an implementation of bloom filter
type BloomFilter struct {
	m      uint64
	k      uint64
	b      *bitmap.Bitmap
	locker sync.Locker
}

// New new a BloomFilter with m bits and k hash functions
func New(m, k uint64, opts ...Options) *BloomFilter {
	option := Option{
		locker: defaultLocker,
	}
	for _, opt := range opts {
		opt(&option)
	}
	return &BloomFilter{
		m:      m,
		k:      k,
		b:      bitmap.New(m),
		locker: option.locker,
	}
}

// New new a BloomFilter with n and fp.
// n is the capacity of the BloomFilter
// fp is the tolerated error rate of the BloomFilter
func NewWithEstimates(n uint64, fp float64) *BloomFilter {
	m, k := EstimateParameters(n, fp)
	return New(m, k)
}

//NewFromData new a BloomFilter by data passed, the data was generated by function 'Data()'
func NewFromData(data []byte, opts ...Options) *BloomFilter {
	option := Option{
		locker: defaultLocker,
	}
	for _, opt := range opts {
		opt(&option)
	}
	b := &BloomFilter{
		locker: option.locker,
	}
	reader := bytes.NewReader(data)
	binary.Read(reader, binary.LittleEndian, &b.m)
	binary.Read(reader, binary.LittleEndian, &b.k)
	b.b = bitmap.NewFromData(data[8+8:])
	return b
}

// EstimateParameters estimates m and k from n and p
func EstimateParameters(n uint64, p float64) (m uint64, k uint64) {
	m = uint64(math.Ceil(-1 * float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)))
	k = uint64(math.Ceil(math.Ln2 * float64(m) / float64(n)))
	return
}

// Add add a value to the BloomFilter
func (bf *BloomFilter) Add(val string) {
	bf.locker.Lock()
	defer bf.locker.Unlock()

	hashs := hash.GenHashInts([]byte(Salt+val), int(bf.k))
	for i := uint64(0); i < bf.k; i++ {
		bf.b.Set(hashs[i] % bf.m)
	}
}

// Contains returns true if value passed is (high probability) in the BloomFilter, or false if not.
func (bf *BloomFilter) Contains(val string) bool {
	bf.locker.RLock()
	defer bf.locker.RUnlock()

	hashs := hash.GenHashInts([]byte(Salt+val), int(bf.k))
	for i := uint64(0); i < bf.k; i++ {
		if !bf.b.IsSet(hashs[i] % bf.m) {
			return false
		}
	}
	return true
}

// Contains returns the data of BloomFilter, it can bee used to new a BloomFilter by using function 'NewFromData' .
func (bf *BloomFilter) Data() []byte {
	bf.locker.Lock()
	defer bf.locker.Unlock()

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, bf.m)
	binary.Write(buf, binary.LittleEndian, bf.k)
	buf.Write(bf.b.Data())
	return buf.Bytes()
}
