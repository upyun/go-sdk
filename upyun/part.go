package upyun

import (
	"context"
	"errors"
	"io"
	"sync"
)

type chunk struct {
	Id   int
	data []byte
}

type multipartUploader struct {
	reader   io.ReadSeeker
	config   *PutObjectConfig
	skiper   *skiper
	partSize int64

	// 多线程上传任务队列
	queue chan *chunk

	// 记录上传时的错误
	errout chan error

	wg sync.WaitGroup
}

func createMultipartUploader(
	reader io.ReadSeeker,
	config *PutObjectConfig,
	skiper *skiper,
	partSize int64,
) *multipartUploader {
	return &multipartUploader{
		reader:   reader,
		config:   config,
		skiper:   skiper,
		partSize: partSize,
	}
}

func errorJoin(c chan error) error {
	if len(c) == 0 {
		return nil
	}
	s := make([]error, 0)
	for i := range c {
		s = append(s, i)
	}
	return errors.Join(s...)
}

func (p *multipartUploader) product(ctx context.Context) {
	defer close(p.queue)

	// 跳过已经上传的文件前一部分
	var partId int64
	if p.skiper != nil {
		partId = p.skiper.FirstMissPartId()
		if _, err := p.reader.Seek(partId*p.partSize, io.SeekCurrent); err != nil {
			p.errout <- err
			return
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			buffer := make([]byte, p.partSize)
			n, err := p.reader.Read(buffer)
			if err != nil {
				if err != io.EOF {
					p.errout <- err
				}
				return
			}

			// 如果分片已经存在，则跳过
			if p.skiper != nil && p.skiper.IsSkip(partId) {
				partId++
				continue
			}

			p.queue <- &chunk{
				Id:   int(partId),
				data: buffer[:n],
			}
			partId++
		}
	}
}

func (p *multipartUploader) work(ctx context.Context, cancel context.CancelFunc, fn func(id int, data []byte) error) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case ch, ok := <-p.queue:
			if !ok {
				return
			}
			if err := fn(ch.Id, ch.data); err != nil {
				p.errout <- err
				cancel()
				return
			}
		}
	}
}

func (p *multipartUploader) Go(fn func(id int, data []byte) error) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 由于需要包含product的错误所以需要+1
	p.errout = make(chan error, p.config.MultipartUploadWorkers+1)
	p.queue = make(chan *chunk, p.config.MultipartUploadWorkers)

	// 生成任务
	go p.product(ctx)

	// 消费任务
	for i := 0; i < p.config.MultipartUploadWorkers; i++ {
		p.wg.Add(1)
		go p.work(ctx, cancel, fn)
	}

	p.wg.Wait()
	close(p.errout)

	return errorJoin(p.errout)
}
