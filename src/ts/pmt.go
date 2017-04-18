package ts

type PMT struct {
	Packet
	PointField byte // 1Byte
	Section ProgramMapSection
}

type ProgramMapSection struct {
	Bytes
	TableID                byte; // 1Byte
	SectionSyntaxIndicator byte;
	SectionLength          uint16; // 12b
	ProgramNumber          uint16; // 2Bytes
	VersionNumber          byte; // 5b
	CurrentNextIndicator   byte;
	SectionNumber          byte; // 1Byte
	LastSectionNumber      byte; // 1Byte
	PCR_PID                uint16; // 13b
	ProgramInfoLength      uint16; // 12b
	Sections               []ProgramMapSubSection;
}

type ProgramMapSubSection struct {
	StreamType byte; // 1Byte
	ElementaryPID uint16; // 13b
	ESInfoLength uint16; // 12b
}

// To bytes
func (pmt PMT) Bytes() (data Data) {
	data = pmt.Packet.ToBytes()
	// Set PointField
	data.PushObj(pmt.PointField, 8)

	// Program association section
	data.PushBytes(pmt.Section)

	if (pmt.HasPayload()) {
		// Push payload
		data.PushBytes(pmt.Payload)
	}

	// Fill remaining bytes with 0xff
	data.FillRemaining(0xff)

	return
}

func (section ProgramMapSection) ToBytes() (data Data) {
	data = *NewData(4)

	data.PushObj(section.TableID, 8)
	data.PushObj(section.SectionSyntaxIndicator, 1)
	data.PushObj(0, 1) // Private
	data.PushObj(0x03, 2) // Reserved
	data.PushObj(section.SectionLength, 12)
	data.PushObj(section.ProgramNumber, 16)
	data.PushObj(0x03, 2) // Reserved
	data.PushObj(section.VersionNumber, 5)
	data.PushObj(section.CurrentNextIndicator, 1)
	data.PushObj(section.SectionNumber, 8)
	data.PushObj(section.LastSectionNumber, 8)
	data.PushObj(0x07, 3) // Reserved
	data.PushObj(section.PCR_PID, 13)
	data.PushObj(0x0f, 4) // Reserved
	data.PushObj(section.ProgramInfoLength, 12)

	for programIndex := 0; programIndex < len(section.Sections); programIndex++ {
		data.PushObj(section.Sections[programIndex].StreamType, 8)
		data.PushObj(0x07, 3) // Reserved
		data.PushObj(section.Sections[programIndex].ElementaryPID, 13)
		data.PushObj(0x0f, 4) // Reserved
		data.PushObj(section.Sections[programIndex].ESInfoLength, 12)
	}

	data.PushObj(data.GenerateCRC32(), 32)

	return
}

// Constructor
func NewPMT() (pmt *PMT) {
	pmt = new(PMT)

	pmt.PID = 4096
	pmt.PayloadUnitStartIndicator = 1
	pmt.AdaptationFieldControl = 1

	pmt.Section.TableID = 2
	pmt.Section.SectionSyntaxIndicator = 1
	pmt.Section.SectionLength = 13
	pmt.Section.CurrentNextIndicator = 1
	pmt.Section.PCR_PID = 256

	pmt.Section.Sections = make([]ProgramMapSubSection, 2)

	// Register video stream
	pmt.Section.Sections[0].StreamType = 27
	pmt.Section.Sections[0].ElementaryPID = 256
	pmt.Section.Sections[0].ESInfoLength = 0

	// Register audio stream
	pmt.Section.Sections[0].StreamType = 15
	pmt.Section.Sections[0].ElementaryPID = 257
	pmt.Section.Sections[0].ESInfoLength = 0 // 6b

	return
}