package frame

import (
	"bytes"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExportCachedFrameCollectionShouldExportAndImport(t *testing.T) {
	cases := []int{
		1, 2, 3,
		11, 22, 33,
		1111, 2222, 3333,
		11111, 22222, 33333,
	}

	var (
		file     *bytes.Buffer = &bytes.Buffer{}
		filePeek *bytes.Buffer = &bytes.Buffer{}
	)

	for _, c := range cases {
		file.Reset()
		filePeek.Reset()

		var (
			collection       FrameCollection = mockFrameCollection(c)
			checksum         string          = "abcdef12345678900987654321abcdef12345678"
			frames           []*Frame        = collection.GetAll()
			importCollection FrameCollection
			importChecksum   string
			importFrames     []*Frame
		)

		err := ExportCachedFrameCollection(file, collection, checksum)
		assert.Nil(t, err)

		if _, err := filePeek.Write(file.Bytes()); err != nil {
			panic(err)
		}

		checksumEqual, err := ChecksumEqualPeek(filePeek, checksum)
		assert.Nil(t, err)
		assert.True(t, checksumEqual)

		importCollection, importChecksum, err = ImportCachedFrameCollection(file)
		assert.Nil(t, err)
		assert.Equal(t, checksum, importChecksum)
		assert.Equal(t, collection.Count(), importCollection.Count())

		importFrames = importCollection.GetAll()

		for index := 0; index < c; index += 1 {
			var (
				frame       *Frame = frames[index]
				importFrame *Frame = importFrames[index]
			)

			assert.Equal(t, frame.OrdinalNumber, importFrame.OrdinalNumber)
			assert.Equal(t, frame.Brightness, importFrame.Brightness)
			assert.Equal(t, frame.ColorDifference, importFrame.ColorDifference)
			assert.Equal(t, frame.BinaryThresholdDifference, importFrame.BinaryThresholdDifference)
		}
	}
}

func mockFrameCollection(capacity int) FrameCollection {
	fc := NewFrameCollection(capacity)
	for index := 0; index < capacity; index += 1 {
		fc.Push(CreateNewFrame(mockImage(color.White), mockImage(color.White), index+1, BinaryThresholdParam))
	}

	return fc
}
