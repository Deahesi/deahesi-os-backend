package parsers

import (
	"errors"
	"io"
	"net/http"
)

type File struct {
	Filename    string
	ContentType string
	Data        []byte
}

func ParseFileFromRequest(r *http.Request) (*File, error) {
	file, header, err := r.FormFile("file")
	if errors.Is(err, http.ErrMissingFile) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return &File{
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Data:        data,
	}, nil
}
