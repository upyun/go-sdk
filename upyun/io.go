package upyun

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
)

type UpYunPutReader interface {
	Len() (n int)
	MD5() (ret string)
	Read([]byte) (n int, err error)
}

type Chunk struct {
	buf  io.Reader
	buf2 *bytes.Buffer
	id   int
	n    int
}

func (c *Chunk) Read(b []byte) (n int, err error) {
	if c.buf2 != nil {
		return c.buf2.Read(b)
	}
	return c.buf.Read(b)
}

func (c *Chunk) Len() int {
	return c.n
}
func (c *Chunk) ID() int {
	return c.id
}

func (c *Chunk) MD5() string {
	c.buf2 = bytes.NewBuffer(nil)
	reader := io.TeeReader(c.buf, c.buf2)
	hash := md5.New()
	_, _ = io.Copy(hash, reader)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func GetReadChunk(input io.Reader, size, partSize int64, ch chan *Chunk) {
	id := 0
	bytesLeft := size
	for bytesLeft > 0 {
		n := partSize
		if bytesLeft <= partSize {
			n = bytesLeft
		}
		reader := io.LimitReader(input, n)
		ch <- &Chunk{
			buf: reader,
			id:  id,
			n:   int(n),
		}
		id++
		bytesLeft -= n
	}
	close(ch)
}
