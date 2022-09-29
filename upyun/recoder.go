package upyun

type Recoder interface {
	Set(breakpoint *BreakPointConfig) error

	Get(uploadID string) (*BreakPointConfig, error)

	Delete(uploadID string) error
}

var resumeRecode = make(map[string]*BreakPointConfig)

type ResumeRecoder struct {
	UploadID string
}

func (recoder *ResumeRecoder) Get(uploadID string) (*BreakPointConfig, error) {
	return resumeRecode[uploadID], nil
}

func (recoder *ResumeRecoder) Set(breakpoint *BreakPointConfig) error {
	resumeRecode[breakpoint.UploadID] = breakpoint
	return nil
}

func (recoder *ResumeRecoder) Delete(uploadID string) error {
	delete(resumeRecode, uploadID)
	return nil
}
