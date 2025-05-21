package frame

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"slices"
)

// Format for the preanalzyed frames cache (.vld-cache)
// Version 1
//
// +---------------+-------------------+------------------------+
// | Name          | Bytes             | Offset                 |
// +---------------+-------------------+------------------------+
// | Magic         | 4                 | 0                      |
// | Version       | 1                 | 4                      |
// | Magic         | 4                 | 5                      |
// | Checksum      | 20                | 9                      |
// | Magic         | 4                 | 29                     |
// | Compression   | 1                 | 33                     |
// | Magic         | 4                 | 34                     |
// | Length        | 4                 | 38                     |
// | Data          | variable <Length> | 42                     |
// +---------------+-------------------+------------------------+
//

const (
	chceksumDecodedLength   int = 20
	plainDataEntryThreshold int = 5000
)

var (
	magicSequence []uint8 = []uint8{0x56, 0x4C, 0x44, 0x21}
	formatVersion []uint8 = []uint8{0x01}
	plainData     uint8   = 0xF0
	flateData     uint8   = 0xF1
)

func ExportCachedFrameCollection(f io.Writer, fc FrameCollection, checksum string) error {
	if _, err := f.Write(magicSequence); err != nil {
		return fmt.Errorf("frame: failed to encode the magic sequence: %w", err)
	}

	if _, err := f.Write(formatVersion); err != nil {
		return fmt.Errorf("frame: failed to encode the format version: %w", err)
	}

	if _, err := f.Write(magicSequence); err != nil {
		return fmt.Errorf("frame: failed to encode the magic sequence: %w", err)
	}

	checksumDecoded, err := hex.DecodeString(checksum)
	if err != nil {
		return fmt.Errorf("frame: failed to decode the checksum hex: %w", err)
	}

	if _, err := f.Write(checksumDecoded); err != nil {
		return fmt.Errorf("frame: failed to encode decoded hex checksum: %w", err)
	}

	if _, err := f.Write(magicSequence); err != nil {
		return fmt.Errorf("frame: failed to encode magic sequence: %w", err)
	}

	var (
		dataBuffer      *bytes.Buffer = &bytes.Buffer{}
		dataCompression uint8
	)

	// NOTE: The compression feature is disabled for now. The specific handling of EOF in flate needs to be investigated.
	if false && fc.Count() > plainDataEntryThreshold {
		dataCompression = flateData

		flateWriter, err := flate.NewWriter(dataBuffer, flate.BestCompression)
		if err != nil {
			return fmt.Errorf("frame: failed to create the flate compress writer: %w", err)
		}

		defer flateWriter.Close()

		if err := encodeFrameCollectionPlain(flateWriter, fc); err != nil {
			return fmt.Errorf("frame: failed to encode the compressed frames data to buffer: %w", err)
		}

		if err := flateWriter.Flush(); err != nil {
			return fmt.Errorf("frame: failed to flush the encoded compressed frame data: %w", err)
		}
	} else {
		dataCompression = plainData

		if err := encodeFrameCollectionPlain(dataBuffer, fc); err != nil {
			return fmt.Errorf("frame: failed to encode the plain frame data to buffer: %w", err)
		}
	}

	if _, err := f.Write([]uint8{dataCompression}); err != nil {
		return fmt.Errorf("frame: failed to encode data compression: %w", err)
	}

	if _, err := f.Write(magicSequence); err != nil {
		return fmt.Errorf("frame: failed to encode magic sequence: %w", err)
	}

	if err := binary.Write(f, binary.LittleEndian, uint32(dataBuffer.Len())); err != nil {
		return fmt.Errorf("frame: failed to encode the data length: %w", err)
	}

	if _, err := f.Write(magicSequence); err != nil {
		return fmt.Errorf("frame: failed to encode magic sequence: %w", err)
	}

	if _, err := f.Write(dataBuffer.Bytes()); err != nil {
		return fmt.Errorf("frame: failed to encode the data: %w", err)
	}

	return nil
}

func ImportCachedFrameCollection(f io.Reader) (FrameCollection, string, error) {
	var (
		magicBuffer           []uint8 = make([]uint8, len(magicSequence))
		versionBuffer         []uint8 = make([]uint8, 1)
		checksumBuffer        []uint8 = make([]uint8, chceksumDecodedLength)
		dataCompressionBuffer []uint8 = make([]uint8, 1)
		dataBuffer            []uint8
	)

	if _, err := io.ReadFull(f, magicBuffer); err != nil || !slices.Equal(magicBuffer, magicSequence) {
		return nil, "", fmt.Errorf("frame: failed to decode and check the magic sequence: %w", err)
	}

	if _, err := io.ReadFull(f, versionBuffer); err != nil {
		return nil, "", fmt.Errorf("frame: failed to decode the version: %w", err)
	}

	if _, err := io.ReadFull(f, magicBuffer); err != nil || !slices.Equal(magicBuffer, magicSequence) {
		return nil, "", fmt.Errorf("frame: failed to decode and check the magic sequence: %w", err)
	}

	if _, err := io.ReadFull(f, checksumBuffer); err != nil {
		return nil, "", fmt.Errorf("frame: failed to decode the checksum: %w", err)
	}

	if _, err := io.ReadFull(f, magicBuffer); err != nil || !slices.Equal(magicBuffer, magicSequence) {
		return nil, "", fmt.Errorf("frame: failed to decode and check the magic sequence: %w", err)
	}

	if _, err := io.ReadFull(f, dataCompressionBuffer); err != nil {
		return nil, "", fmt.Errorf("frame: failed to decode data compression: %w", err)
	}

	var compression bool
	switch dataCompressionBuffer[0] {
	case plainData:
		compression = false
	case flateData:
		compression = true
	default:
		return nil, "", fmt.Errorf("frame: invalid or corrupted data compression value")
	}

	if _, err := io.ReadFull(f, magicBuffer); err != nil || !slices.Equal(magicBuffer, magicSequence) {
		return nil, "", fmt.Errorf("frame: failed to decode and check the magic sequence: %w", err)
	}

	var length uint32
	if err := binary.Read(f, binary.LittleEndian, &length); err != nil {
		return nil, "", fmt.Errorf("frame: failed to decode the data length: %w", err)
	}

	if _, err := io.ReadFull(f, magicBuffer); err != nil || !slices.Equal(magicBuffer, magicSequence) {
		return nil, "", fmt.Errorf("frame: failed to decode and check the magic sequence: %w", err)
	}

	dataBuffer = make([]uint8, length)
	if _, err := io.ReadFull(f, dataBuffer); err != nil {
		return nil, "", fmt.Errorf("frame: failed to decode the data: %w", err)
	}

	var (
		dataBufferReader *bytes.Reader = bytes.NewReader(dataBuffer)
		frames           FrameCollection
	)

	if compression {
		flateReader := flate.NewReader(dataBufferReader)
		defer flateReader.Close()

		if fc, err := decodeFrameCollectionPlain(flateReader); err != nil {
			return nil, "", fmt.Errorf("frame: failed to decode the compressed frame data: %w", err)
		} else {
			frames = fc
		}
	} else {
		if fc, err := decodeFrameCollectionPlain(dataBufferReader); err != nil {
			return nil, "", fmt.Errorf("frame: failed to decode the plain frame data: %w", err)
		} else {
			frames = fc
		}
	}

	return frames, hex.EncodeToString(checksumBuffer), nil
}

func ChecksumEqualPeek(f io.Reader, checksum string) (bool, error) {
	var (
		magicBuffer    = make([]uint8, len(magicSequence))
		versionBuffer  = make([]uint8, 1)
		checksumBuffer = make([]uint8, chceksumDecodedLength)
	)

	if _, err := io.ReadFull(f, magicBuffer); err != nil || !slices.Equal(magicBuffer, magicSequence) {
		return false, fmt.Errorf("frame: failed to decode and check the magic sequence: %w", err)
	}

	if _, err := io.ReadFull(f, versionBuffer); err != nil {
		return false, fmt.Errorf("frame: failed to decode the version: %w", err)
	}

	if _, err := io.ReadFull(f, magicBuffer); err != nil || !slices.Equal(magicBuffer, magicSequence) {
		return false, fmt.Errorf("frame: failed to decode and check the magic sequence: %w", err)
	}

	if _, err := io.ReadFull(f, checksumBuffer); err != nil {
		return false, fmt.Errorf("frame: failed to decode the checksum: %w", err)
	}

	if _, err := io.ReadFull(f, magicBuffer); err != nil || !slices.Equal(magicBuffer, magicSequence) {
		return false, fmt.Errorf("frame: failed to decode and check the magic sequence: %w", err)
	}

	targetChecksum, err := hex.DecodeString(checksum)
	if err != nil {
		return false, fmt.Errorf("frame: failed to decode the target checksum: %w", err)
	}

	return slices.Equal(targetChecksum, checksumBuffer), nil
}

func encodeFrameCollectionPlain(f io.Writer, fc FrameCollection) error {
	for _, frame := range fc.GetAll() {
		if err := binary.Write(f, binary.LittleEndian, uint32(frame.OrdinalNumber)); err != nil {
			return fmt.Errorf("frame: failed to binary encode the frame ordinal number: %w", err)
		}

		if err := binary.Write(f, binary.LittleEndian, frame.Brightness); err != nil {
			return fmt.Errorf("frame: failed to binary encode the frame brightness: %w", err)
		}

		if err := binary.Write(f, binary.LittleEndian, frame.ColorDifference); err != nil {
			return fmt.Errorf("frame: failed to binary encode the frame color difference: %w", err)
		}

		if err := binary.Write(f, binary.LittleEndian, frame.BinaryThresholdDifference); err != nil {
			return fmt.Errorf("frame: failed to binary encode the frame binary threshold difference: %w", err)
		}
	}

	return nil
}

func decodeFrameCollectionPlain(r io.Reader) (FrameCollection, error) {
	var (
		frames                    []*Frame = make([]*Frame, 0, plainDataEntryThreshold)
		ordinalNumber             uint32
		brightness                float64
		colorDifference           float64
		binaryThresholdDifference float64
	)

	for {
		if err := binary.Read(r, binary.LittleEndian, &ordinalNumber); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("frame: failed to decode the frame ordinal number: %w", err)
		}

		if err := binary.Read(r, binary.LittleEndian, &brightness); err != nil {
			return nil, fmt.Errorf("frame: failed to decode the frame brightness: %w", err)
		}

		if err := binary.Read(r, binary.LittleEndian, &colorDifference); err != nil {
			return nil, fmt.Errorf("frame: failed to decode the frame color difference: %w", err)
		}

		if err := binary.Read(r, binary.LittleEndian, &binaryThresholdDifference); err != nil {
			return nil, fmt.Errorf("frame: failed to decode the frame bianry threshold difference: %w", err)
		}

		frames = append(frames, &Frame{
			OrdinalNumber:             int(ordinalNumber),
			ColorDifference:           colorDifference,
			BinaryThresholdDifference: binaryThresholdDifference,
			Brightness:                brightness,
		})
	}

	fc := NewFrameCollection(len(frames))
	defer fc.Lock()

	for _, frame := range frames {
		if err := fc.Push(frame); err != nil {
			return nil, fmt.Errorf("frame: failed to push the decoded frame to collection: %w", err)
		}
	}

	return fc, nil
}
