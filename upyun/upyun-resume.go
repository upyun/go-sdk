package upyun

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	// resumePartSize is the size of each part for resume upload
	resumePartSize = 1024 * 1024
	// resumeFileSizeLowerLimit is the lowest file size limit for resume upload
	resumeFileSizeLowerLimit int64 = resumePartSize * 10
)

var (
	// ResumeRetry is the number of retries for resume upload
	ResumeRetryCount = 3
	// ResumeWaitSeconds is the number of time to wait when net error occurs
	ResumeWaitTime = time.Second * 5
)

// ResumeReporter
type ResumeReporter func(int, int)

// ResumeReporterPrintln is a simple ResumeReporter for test
func ResumeReporterPrintln(partID int, maxPartID int) {
	fmt.Printf("resume test reporter: %v / %v\n", partID, maxPartID)
}

// FragmentFile is like os.File, but only a part of file can be Read().
// return io.EOF when cursor fetch the limit.
type FragmentFile struct {
	offset int64
	limit  int
	cursor int
	*os.File
}

// NewFragmentFile returns a new FragmentFile.
func NewFragmentFile(file *os.File, offset int64, limit int) (*FragmentFile, error) {
	sizedfile := &FragmentFile{
		offset: offset,
		limit:  limit,
		File:   file,
	}
	_, err := sizedfile.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	return sizedfile, nil
}

// Seek likes os.File.Seek()
func (f *FragmentFile) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	case 0:
		f.cursor = int(offset)
		return f.File.Seek(f.offset+offset, 0)
	default:
		return 0, errors.New("whence must be 0")
	}
}

// Read is just like os.File.Read but return io.EOF when catch sizedfile's limit
// or the end of file
func (f *FragmentFile) Read(b []byte) (n int, err error) {
	if f.cursor >= f.limit {
		return 0, io.EOF
	}
	n, err = f.File.Read(b)
	if int(f.cursor)+n > f.limit {
		n = f.limit - f.cursor
	}
	f.cursor += n
	return n, err
}

// Close will not actually close FragmentFile
func (f *FragmentFile) Close() error {
	return nil
}

// MD5 returns md5 of the FragmentFile.
func (f *FragmentFile) MD5() (string, error) {
	cursor := f.cursor
	f.Seek(0, 0)
	md5, _, err := md5sum(f)
	f.Seek(int64(cursor), 0)
	return md5, err
}
