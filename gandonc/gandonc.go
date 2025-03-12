package gandonc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

const (
	MagicNumberHeaderCAR = 0x72616301
	MagicNumberHeaderGAN = 0x676e6101
	MagicNumberIndex     = 1
	MagicNumberData      = 2
	MagicNumberEnd       = 0xFFFFFFFF
	KeyString            = "7f13a9cf-2f55-4898-8294-b6b0655d59f1"
)

type GanDecryptor struct {
	inputFile string
	outputDir string
	keyBytes  []byte
	keyLength int
	fileSize  int64
	index     map[uint32]string
	isGan     bool
	byteIndex uint32
}

func NewGanDecryptor(inputFile, outputDir string) (*GanDecryptor, error) {
	fileInfo, err := os.Stat(inputFile)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	return &GanDecryptor{
		inputFile: inputFile,
		outputDir: outputDir,
		keyBytes:  []byte(KeyString),
		keyLength: len(KeyString),
		fileSize:  fileInfo.Size(),
		index:     make(map[uint32]string),
	}, nil
}

func (g *GanDecryptor) decryptContent(data []byte) []byte {
	decrypted := make([]byte, len(data))
	for i, byteVal := range data {
		counter := i + 2
		keyIndex := (counter - 36*(counter/36)) % g.keyLength
		decrypted[i] = byteVal ^ g.keyBytes[keyIndex]
	}
	return decrypted
}

func (g *GanDecryptor) readPadding(f *os.File, length uint32, sectionType string) error {
	paddingLength := (4 - (length % 4)) % 4
	if paddingLength > 0 {
		_, err := f.Seek(int64(paddingLength), io.SeekCurrent)
		g.byteIndex += paddingLength
		return err
	}
	return nil
}

func (g *GanDecryptor) Process() error {
	f, err := os.Open(g.inputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	var header [16]byte
	if _, err := io.ReadFull(f, header[:]); err != nil {
		return errors.New("file too short")
	}
	fmt.Println("Нeader length: " + strconv.Itoa(len(header)))
	g.byteIndex += 16

	magicNumber := binary.LittleEndian.Uint32(header[0:4])
	if magicNumber == MagicNumberHeaderGAN {
		g.isGan = true
	}

	indexLength := binary.LittleEndian.Uint32(header[12:16])

	for i := uint32(0); i < indexLength; i++ {
		var entry [12]byte
		if _, err := io.ReadFull(f, entry[:]); err != nil {
			return errors.New("index sect too short")
		}
		g.byteIndex += 12

		entryType := binary.LittleEndian.Uint32(entry[0:4])
		if entryType != MagicNumberIndex {
			return errors.New("bad entry type")
		}

		dataOffset := binary.LittleEndian.Uint32(entry[4:8])
		filenameLength := binary.LittleEndian.Uint32(entry[8:12])

		filenameBytes := make([]byte, filenameLength+1)
		if _, err := io.ReadFull(f, filenameBytes); err != nil {
			return errors.New("bad filename")
		}
		g.byteIndex += filenameLength + 1

		filename := string(filenameBytes[:filenameLength])
		g.index[dataOffset] = filename

		if err := g.readPadding(f, filenameLength+1, "index"); err != nil {
			return err
		}
	}

	for g.byteIndex < uint32(g.fileSize) {
		var entryType uint32
		if err := binary.Read(f, binary.LittleEndian, &entryType); err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("FILE EOF")
				break
			}
			return err
		}
		g.byteIndex += 4

		if entryType == MagicNumberEnd {
			_, err := f.Seek(4, io.SeekCurrent)
			g.byteIndex += 4
			return err
		} else if entryType == MagicNumberData {
			var nextOffset, fileSize uint32
			if err := binary.Read(f, binary.LittleEndian, &nextOffset); err != nil {
				return err
			}
			if err := binary.Read(f, binary.LittleEndian, &fileSize); err != nil {
				return err
			}
			g.byteIndex += 8

			fileContent := make([]byte, fileSize)
			if _, err := io.ReadFull(f, fileContent); err != nil {
				return errors.New("неполный file_content")
			}
			g.byteIndex += fileSize

			offset := g.byteIndex - fileSize - 12
			filename, exists := g.index[offset]
			if !exists {
				filename = fmt.Sprintf("file-%d.extracted", offset)
			}

			if g.isGan {
				fmt.Printf("decoding  %s\n", filename)
				fileContent = g.decryptContent(fileContent)
			}

			outputPath := filepath.Join(g.outputDir, filename)
			if err := os.WriteFile(outputPath, fileContent, 0644); err != nil {
				return err
			}

			if err := g.readPadding(f, fileSize, "data"); err != nil {
				return err
			}
		} else {
			return errors.New("unknown entry type")
		}
	}
	return nil
}
