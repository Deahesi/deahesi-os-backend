package ai

type FileInput struct {
	Filename    string
	Data        []byte
	ContentType string
}

type SendMessageInput struct {
	Text string
	File *FileInput
}
