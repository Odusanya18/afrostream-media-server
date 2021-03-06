package ts

type PMT struct {
	Packet
	PointField byte // 1Byte
	Section    ProgramMapSection
}

type ProgramMapSection struct {
	Bytes
	TableID                byte // 1Byte
	SectionSyntaxIndicator byte
	SectionLength          uint16 // 12b
	ProgramNumber          uint16 // 2Bytes
	VersionNumber          byte   // 5b
	CurrentNextIndicator   byte
	SectionNumber          byte   // 1Byte
	LastSectionNumber      byte   // 1Byte
	PCR_PID                uint16 // 13b
	ProgramInfoLength      uint16 // 12b
	Sections               []ProgramMapSubSection
}

type ProgramMapSubSection struct {
	StreamType    byte   // 1Byte
	ElementaryPID uint16 // 13b
	ESInfoLength  uint16 // 12b
	Descriptor Bytes // If needed
}

type Descriptor struct {
	Bytes
	DescriptorTag byte // 1Byte
	DescriptorLength byte // 1Byte
}

type DescriptorData struct {
	Descriptor
	DescriptorData []byte
}

// To bytes
func (pmt PMT) ToBytes() (data Data) {
	data = pmt.Packet.ToBytes()
	// Set PointField
	data.PushObj(pmt.PointField, 8)

	// Program association section
	data.PushBytes(pmt.Section)

	if pmt.HasPayload() {
		// Push payload
		data.PushBytes(pmt.Payload)
	}

	// Fill remaining bytes with 0xff
	data.FillRemaining(0xff)

	return
}

func (section ProgramMapSection) ToBytes() (data Data) {
	// In general, 13 bytes after sectionLength and 3 bytes before
	sectionLength := section.GetSectionLength()
	data = *NewData(section.Size())

	data.PushObj(section.TableID, 8)
	data.PushObj(section.SectionSyntaxIndicator, 1)
	data.PushUInt(0, 1)    // Private
	data.PushUInt(0x03, 2) // Reserved
	data.PushUInt(uint32(sectionLength), 12)
	data.PushObj(section.ProgramNumber, 16)
	data.PushUInt(0x03, 2) // Reserved
	data.PushObj(section.VersionNumber, 5)
	data.PushObj(section.CurrentNextIndicator, 1)
	data.PushObj(section.SectionNumber, 8)
	data.PushObj(section.LastSectionNumber, 8)
	data.PushUInt(0x07, 3) // Reserved
	data.PushObj(section.PCR_PID, 13)
	data.PushUInt(0x0f, 4) // Reserved
	data.PushObj(section.ProgramInfoLength, 12)

	// 5 Bytes per sub section
	for programIndex := 0; programIndex < len(section.Sections); programIndex++ {
		data.PushBytes(section.Sections[programIndex])
	}

	crc32 := data.GenerateCRC32ToOffset()
	data.PushObj(crc32, 32)
	data.FillRemaining(0xff)

	return
}

func (section ProgramMapSection) Size() (size int) {
	sectionLength := section.GetSectionLength()
	size = sectionLength + 3  // Bytes before section length + CRC32

	for i := 0; i < len(section.Sections); i++ {
		size += section.Sections[i].Size()
	}

	return
}

func (section ProgramMapSubSection) ToBytes() (data Data) {
	data = *NewData(section.Size())

	data.PushObj(section.StreamType, 8)
	data.PushUInt(0x07, 3) // Reserved
	data.PushObj(section.ElementaryPID, 13)
	data.PushUInt(0x0f, 4) // Reserved
	data.PushObj(section.ESInfoLength, 12)

	if section.Descriptor != nil {
		data.PushBytes(section.Descriptor)
	}
	return
}

func (section ProgramMapSubSection) Size() (int) {
	size := 5
	if section.Descriptor != nil {
		size += section.Descriptor.Size()
	}
	return size
}

func (section ProgramMapSection) GetSectionLength() (int) {
	// Compute the section Length
	sectionLength := 13

	for _, subSection := range section.Sections {
		sectionLength += subSection.Size()
	}
	return sectionLength
}

func (descriptor DescriptorData) ToBytes() (data Data) {
	data = *NewData(descriptor.Size())

	data.PushObj(descriptor.DescriptorTag, 8)
	data.PushObj(descriptor.DescriptorLength, 8)
	data.PushObj(descriptor.DescriptorData, len(descriptor.DescriptorData) * 8)

	return
}

func (descriptor DescriptorData) Size() (int) {
	return descriptor.Descriptor.Size() + len(descriptor.DescriptorData)
}

func (descriptor Descriptor) Size() (int) {
	return 2
}

// Constructor
func NewPMT(PCR_PID uint16) (pmt *PMT) {
	pmt = new(PMT)

	pmt.PID = 4096
	pmt.PayloadUnitStartIndicator = 1
	pmt.AdaptationFieldControl = 1

	pmt.Section.TableID = 2
	pmt.Section.ProgramNumber = 1
	pmt.Section.SectionSyntaxIndicator = 1
	pmt.Section.SectionLength = 13
	pmt.Section.CurrentNextIndicator = 1
	pmt.Section.PCR_PID = PCR_PID

	return
}

func NewDebugPMT() (pmt *PMT) {
	pmt = new(PMT)

	usingAudio := false

	pmt.PID = 4096
	pmt.PayloadUnitStartIndicator = 1
	pmt.AdaptationFieldControl = 1

	pmt.Section.TableID = 2
	pmt.Section.ProgramNumber = 1
	pmt.Section.SectionSyntaxIndicator = 1
	pmt.Section.CurrentNextIndicator = 1
	pmt.Section.PCR_PID = 256

	if usingAudio {
		pmt.Section.Sections = make([]ProgramMapSubSection, 2)
	} else {
		pmt.Section.Sections = make([]ProgramMapSubSection, 1)
	}

	// Register video stream
	pmt.Section.Sections[0].StreamType = 27
	pmt.Section.Sections[0].ElementaryPID = 256
	pmt.Section.Sections[0].ESInfoLength = 0

	// Register audio stream
	if usingAudio {
		pmt.Section.Sections[1].StreamType = 15
		pmt.Section.Sections[1].ElementaryPID = 257
		pmt.Section.Sections[1].ESInfoLength = 6

		descriptor := new(DescriptorData)
		descriptor.DescriptorTag = 10
		descriptor.DescriptorLength = 4
		descriptor.DescriptorData = []byte{0x75, 0x6e, 0x64, 0x00}

		pmt.Section.Sections[1].Descriptor = descriptor
	}
	return
}
