package upyun

import (
	"io"
)

// ProgressEventType defines transfer progress event type
type ProgressEventType int

const (
	// TransferStartedEvent transfer started, set TotalBytes
	TransferStartedEvent ProgressEventType = 1 + iota
	// TransferDataEvent transfer data, set ConsumedBytes anmd TotalBytes
	TransferDataEvent
	// TransferCompletedEvent transfer completed
	TransferCompletedEvent
	// TransferFailedEvent transfer encounters an error
	TransferFailedEvent
)

// ProgressEvent defines progress event
type ProgressEvent struct {
	ConsumedBytes int64
	TotalBytes    int64
	EventType     ProgressEventType
}

// ProgressListener listens progress change
type ProgressListener interface {
	ProgressChanged(event *ProgressEvent)
}

// -------------------- Private --------------------

func newProgressEvent(eventType ProgressEventType, consumed, total int64) *ProgressEvent {
	return &ProgressEvent{
		ConsumedBytes: consumed,
		TotalBytes:    total,
		EventType:     eventType}
}

// publishProgress
func publishProgress(listener ProgressListener, event *ProgressEvent) {
	if listener != nil && event != nil {
		listener.ProgressChanged(event)
	}
}

type teeReader struct {
	reader        io.Reader
	listener      ProgressListener
	consumedBytes int64
	totalBytes    int64
}

func TeeReader(reader io.Reader, totalBytes int64, listener ProgressListener) *teeReader {
	return &teeReader{
		reader:        reader,
		listener:      listener,
		consumedBytes: 0,
		totalBytes:    totalBytes,
	}
}

func (t *teeReader) Read(p []byte) (n int, err error) {
	n, err = t.reader.Read(p)

	// Read encountered error
	if err != nil && err != io.EOF {
		event := newProgressEvent(TransferFailedEvent, t.consumedBytes, t.totalBytes)
		publishProgress(t.listener, event)
	}

	if n > 0 {
		t.consumedBytes += int64(n)
		// Progress
		if t.listener != nil {
			event := newProgressEvent(TransferDataEvent, t.consumedBytes, t.totalBytes)
			publishProgress(t.listener, event)
		}
	}

	return
}

func (t *teeReader) Close() error {
	if rc, ok := t.reader.(io.ReadCloser); ok {
		return rc.Close()
	}
	return nil
}
