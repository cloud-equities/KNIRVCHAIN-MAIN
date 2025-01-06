package chain

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// VectorData represents a vector of float64 values
type VectorData struct {
	Data       []float64 `json:"data"`
	Dimensions int       `json:"dimensions"`
}

// SerializeVector serializes VectorData to a byte array
func SerializeVector(data *VectorData) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.LittleEndian, int32(len(data.Data))); err != nil {
		return nil, fmt.Errorf("failed to write vector data length: %w", err)
	}

	if err := binary.Write(buf, binary.LittleEndian, int32(data.Dimensions)); err != nil {
		return nil, fmt.Errorf("failed to write vector dimensions: %w", err)
	}

	for _, v := range data.Data {
		if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
			return nil, fmt.Errorf("failed to write vector value: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// DeserializeVector deserializes a byte array into VectorData
func DeserializeVector(data []byte) (*VectorData, error) {
	buf := bytes.NewReader(data)

	var dataLen int32
	if err := binary.Read(buf, binary.LittleEndian, &dataLen); err != nil {
		return nil, fmt.Errorf("failed to read vector data length: %w", err)
	}

	var dimensions int32
	if err := binary.Read(buf, binary.LittleEndian, &dimensions); err != nil {
		return nil, fmt.Errorf("failed to read vector dimensions: %w", err)
	}

	vectorData := &VectorData{
		Dimensions: int(dimensions),
		Data:       make([]float64, dataLen),
	}

	for i := 0; i < int(dataLen); i++ {
		if err := binary.Read(buf, binary.LittleEndian, &vectorData.Data[i]); err != nil {
			return nil, fmt.Errorf("failed to read vector value: %w", err)
		}
	}

	return vectorData, nil
}
