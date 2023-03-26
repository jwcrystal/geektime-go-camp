package compressor

type Compressor interface {
	Code() uint8
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
}

type DefaultCompressor struct {
}

func (d DefaultCompressor) Code() uint8 {
	return 0
}

func (d DefaultCompressor) Compress(data []byte) ([]byte, error) {
	return data, nil
}

func (d DefaultCompressor) Decompress(data []byte) ([]byte, error) {
	return data, nil
}
