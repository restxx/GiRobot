package stream

import "errors"

const MOD = 8

type CryptBufIO struct {
	Buffer
	deIdx  int
	deFunc func([]byte)
}

func NewCryptBuffIO(fn func([]byte)) *CryptBufIO {

	return &CryptBufIO{deFunc: fn}
}

func (b *CryptBufIO) Reset() {
	b.buf = b.buf[:0]
	b.rIdx = 0
	b.deIdx = 0
	b.lastRead = opInvalid
}

// 所有己解密的字节
func (b *CryptBufIO) DecBytes() []byte { return b.buf[b.rIdx:b.deIdx] }

// 己解密的字节但未读取的长度
func (b *CryptBufIO) Len() int { return b.deIdx - b.rIdx }

func (b *CryptBufIO) Write(p []byte) (n int, err error) {
	err = nil
	b.lastRead = opInvalid
	m, ok := b.tryGrowByReslice(len(p))
	if !ok {
		m = b.grow(len(p))
	}
	n = copy(b.buf[m:], p)

	// 得到未解密长度 _len
	_len := len(b.buf) - b.deIdx
	dl := int(_len - _len%MOD)
	b.deFunc(b.buf[b.deIdx : b.deIdx+dl])
	b.deIdx += dl

	return
}

func (b *CryptBufIO) Next(n int) (data []byte, err error) {
	b.lastRead = opInvalid
	err = nil
	if n > b.Len() {
		return data, errors.New("Read Next too long to Next ")
	}
	data = b.buf[b.rIdx : b.rIdx+n]
	b.rIdx += n
	if n > 0 {
		b.lastRead = opRead
	}
	return
}

func (b *CryptBufIO) Peek(n int) (data []byte, err error) {
	b.lastRead = opInvalid
	err = nil
	if n > b.Len() {
		return data, errors.New("Read Next too long to Peek ")
	}
	data = b.buf[b.rIdx : b.rIdx+n]

	return
}
