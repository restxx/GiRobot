package stream

type IBuffIO interface {
	Write(p []byte) (n int, err error)
	Next(n int) (data []byte, err error)
	Peek(n int) (data []byte, err error)
	Empty() bool
	Len() int
	Reset()
}
