package upyun

import (
	"bytes"
	"testing"
)

func TestMultipartUploader(t *testing.T) {
	uploader := createMultipartUploader(
		bytes.NewReader([]byte("hello world")),
		&PutObjectConfig{
			MultipartUploadWorkers: 4,
		},
		nil,
		1,
	)

	var result = make([]byte, 11)
	uploader.Go(func(id int, data []byte) error {
		Equal(t, 1, len(data))
		result[id] = data[0]
		return nil
	})
	Equal(t, string(result), "hello world")
}

func TestMultipartUploaderSeek(t *testing.T) {
	payload := []byte("hello world")
	uploader := createMultipartUploader(
		bytes.NewReader(payload),
		&PutObjectConfig{
			MultipartUploadWorkers: 4,
		},
		createSkiper(
			2,
			[]*DisorderPart{
				&DisorderPart{ID: 0, Size: 1},
				&DisorderPart{ID: 1, Size: 1},
				&DisorderPart{ID: 3, Size: 1},
				&DisorderPart{ID: 5, Size: 1},
				&DisorderPart{ID: 9, Size: 1},
				&DisorderPart{ID: 10, Size: 1},
			},
		),
		1,
	)

	var result = []byte{'h', 'e', 0, 'l', 0, ' ', 0, 0, 0, 'l', 'd'}
	uploader.Go(func(id int, data []byte) error {
		Equal(t, 1, len(data))
		result[id] = data[0]
		return nil
	})

	Equal(t, string(result), "hello world")
}
