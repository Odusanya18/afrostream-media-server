package ts

import (

)

type Packet struct {
	Header
	AdaptationField
	Payload
}

// Structures
type Header struct {
	Bytes
	TransportErrorIndicator byte;
	PayloadUnitStartIndicator byte;
	TransportPriority byte;
	PID uint16; // 13b
	TransportScramblingControl byte; // 2b
	AdaptationFieldControl byte; // 2b
	ContinuityCounter byte; // 4b
}

type AdaptationField struct {
	Bytes
	AdaptationFieldLength byte; // 1Byte
	DiscontinuityIndicator byte;
	RandomAccessIndicator byte;
	ElementaryStreamPriorityIndicator byte;
	PCR_Flag byte;
	OPCR_Flag byte;
	SplicingPointFlag byte;
	TransportPrivateDataFlag byte;
	AdaptationFieldExtensionFlag byte;
	PCR PCR;
}

type PCR struct {
	Bytes
	ProgramClockReferenceBase uint64; // 33b;
	Reserved byte; // 6b;
	ProgramClockReferenceExtension uint16; // 9b
}

// Bytes
func (packet Packet) ToBytes() (data Data) {
	data = *NewData(188)

	// Get header
	data.PushBytes(packet.Header)

	// Get adaptation field
	if packet.HasAdaptationField() {
		data.PushBytes(packet.AdaptationField)
	}

	return
}

func (header Header) ToBytes() (data Data) {
	data = *NewData(4)
	data.Write(0x47)

	data.PushObj(header.TransportErrorIndicator, 1)
	data.PushObj(header.PayloadUnitStartIndicator, 1)
	data.PushObj(header.TransportPriority, 1)
	data.PushObj(header.PID, 13)
	data.PushObj(header.TransportScramblingControl, 2)
	data.PushObj(header.AdaptationFieldControl, 2)
	data.PushObj(header.ContinuityCounter, 4)
	return
}

func (pcr PCR) ToBytes() (data Data) {
	data = *NewData(4)
	data.PushObj(pcr.ProgramClockReferenceBase, 33)
	data.PushObj(pcr.Reserved, 6)
	data.PushObj(pcr.ProgramClockReferenceExtension, 9)
	return
}

func (field AdaptationField) ToBytes() (data Data) {
	// Compute Adaptation length adding the first byte length
	adaptationLength := int(field.AdaptationFieldLength + 1)

	data = *NewData(int(adaptationLength));

	data.PushObj(field.AdaptationFieldLength, 8)
	data.PushObj(field.DiscontinuityIndicator, 1)
	data.PushObj(field.RandomAccessIndicator, 1)
	data.PushObj(field.ElementaryStreamPriorityIndicator, 1)
	data.PushObj(field.PCR_Flag, 1)
	data.PushObj(field.OPCR_Flag, 1)
	data.PushObj(field.SplicingPointFlag, 1)
	data.PushObj(field.TransportPrivateDataFlag, 1)

	if field.PCR_Flag == 1 {
		data.PushBytes(field.PCR)
	}

	// Stuffing bytes
	for data.Offset < adaptationLength {

		data.PushObj(0xff, 8)
	}

	return
}

// Check if the current packet has adaptation field flag activated
func (packet Packet) HasAdaptationField() bool {
	return packet.Header.AdaptationFieldControl & 2 != 0
}

// Check if the current packet has payload field flag activated
func (packet Packet) HasPayload() bool {
	return packet.Header.AdaptationFieldControl & 1 != 0
}