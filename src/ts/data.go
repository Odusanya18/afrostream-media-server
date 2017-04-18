package ts

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

type Data struct {
	Data   []byte
	Offset int
}

// Push object on Data
//		object 			object to write
//		objectSize		part to write of this object
// Push the object on Data
func (data *Data) PushObj(object interface{}, objectSize int) {
	// Transform object in bytes
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, object)
	objectBytes := buf.Bytes()
	data.PushObj(objectBytes, objectSize)
	return
}

func (data *Data) PushAll(bytes []byte) {
	data.Push(bytes, len(bytes))
}

func (data *Data) PushData(dataPushed Data) {
	data.Push(dataPushed.Data, len(dataPushed.Data))
}

// Write bytes on Data
//		bytes 			bytes to write
//		bytesSize		part to write of these bytes
// Push the object on Data
func (data *Data) Push(bytes []byte, sizeToPush int) {
	// While there are resulting bits to push
	bitIndex := 0
	for bitIndex < sizeToPush {
		// Number of bits in actual Byte
		residualBits := data.GetResidualBits()

		// Number of pushed bits
		pushedBits := Min(residualBits, sizeToPush-bitIndex)

		// Get the corresponding byte to write
		writtenByte := GetByte(bytes,
			bitIndex,
			bitIndex+pushedBits) << byte(residualBits-pushedBits)

		// Write byte on Data
		data.WriteOR(writtenByte)

		// Update Data offset
		data.Offset += pushedBits

		// Update bit index
		bitIndex += pushedBits
	}

}

func (data *Data) PushBytes(bytes Bytes) {
	toBytes := bytes.ToBytes()
	data.PushData(toBytes)
}

// Apply or operation to current byte and given byte
func (data *Data) WriteOR(byte byte) {
	*data.GetCurrentByte() |= byte
}

// Write byte on Data
func (data *Data) Write(byte byte) {
	*data.GetCurrentByte() = byte
	data.Offset += 8
}

// Fill rest of Data with byte
func (data *Data) FillRemaining(byte byte) {
	lenData := len(data.Data)

	for data.Offset != lenData {
		data.Write(byte)
	}
}

// Create byte from start and end indices
//
// 		startIndex		start index read in bits
//		endIndex		end index to read in bits,
// 						startIndex - endIndex < 8
//
// 		shift			number of shift to add to the resulting byte
//
func GetByte(data []byte, startIndex int, endIndex int) byte {
	var startIndexByte int = startIndex / 8
	var endIndexByte int = endIndex / 8
	var startIndexInByte = uint8(startIndex % 8)
	var endIndexInByte = uint8(endIndex % 8)

	// If start and end are on the same byte
	if startIndexByte == endIndexByte {
		return SelectByte(data[startIndexByte], startIndexInByte, endIndexInByte)
	} else {

		// Compute shift from start
		shift := 8 - startIndexInByte

		// Get first part
		startByte := SelectByte(data[startIndexByte], startIndexInByte, 7)

		// Get second part in the next byte
		endByte := SelectByte(data[endIndexByte], 0, endIndexInByte-1)

		// Create byte by coupling start and endByte
		return startByte | (endByte >> shift)
	}
}

// Select part of a byte
func SelectByte(src byte, start, end uint8) byte {
	tmp := src << (8 - end)
	return tmp >> (8 - end + start)
}

// Get the current byte
func (data *Data) GetCurrentByte() *byte {
	return &data.Data[data.GetByteIndex()]
}

// Set the current byte
func (data *Data) SetCurrentByte(byte byte) {
	data.Data[data.GetByteIndex()] = byte
}

// Get the index of the current starting byte in bits
func (data *Data) GetStartBitsIndex() int {
	return data.GetByteIndex() * 8
}

// Get the index of the current starting byte in byte
func (data *Data) GetByteIndex() int {
	return data.Offset / 8
}

// Get number of residual bits in the current byte
func (data *Data) GetResidualBits() int {
	return 8 - data.Offset%8
}

// Create Data with offset support
func NewData(length int) *Data {
	data := new(Data)
	data.Data = make([]byte, length)
	return data
}

func (data *Data) PrintBinary() {
	fmt.Printf("%08b\n", data.Data)
}

func (data *Data) PrintHex() {
	fmt.Printf("% 8X\n", data.Data)
}

func (data *Data) GenerateCRC32() uint32 {
	return crc32.Checksum(data.Data, crc32.IEEETable)
}
