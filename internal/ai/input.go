package ai

type FileInput struct {
	Filename    string `json:"filename"`
	Data        []byte `json:"data"`
	ContentType string `json:"contentType"`
}

type SendMessageInput struct {
	Text string     `json:"text"`
	File *FileInput `json:"file,omitempty"`
}
