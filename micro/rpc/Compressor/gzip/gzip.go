package gzip

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
)

type GzipCompressor struct {
}

func (g GzipCompressor) Code() uint8 {
	return 1
}

func (g GzipCompressor) Compress(data []byte) ([]byte, error) {
	res := bytes.NewBuffer(nil)
	gz := gzip.NewWriter(res)
	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}
	err = gz.Flush() // The compressed bytes are not necessarily flushed until the Writer is closed.
	if err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func (g GzipCompressor) Decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	defer func() {
		_ = r.Close()
	}()
	if err != nil {
		return nil, err
	}
	res, err := io.ReadAll(r)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return res, nil
}
