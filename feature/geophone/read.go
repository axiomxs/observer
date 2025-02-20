package geophone

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/bclswl0827/observer/config"
	"github.com/bclswl0827/observer/driver/serial"
)

func (g *Geophone) Read(port io.ReadWriteCloser, conf *config.Conf, packet *Packet, packetLen int) error {
	// Filter frame header
	_, err := serial.Filter(port, SYNC_WORD[:], 128)
	if err != nil {
		return err
	}

	// Read data frame
	checksumLen := len(packet.Checksum)
	buf := make([]byte, g.getSize(packetLen, checksumLen))
	n, err := serial.Read(port, buf, TIMEOUT_THRESHOLD)
	if err != nil {
		return err
	}

	// Allocate memory for data frame
	packet.EHZ = make([]int32, packetLen)
	packet.EHE = make([]int32, packetLen)
	packet.EHN = make([]int32, packetLen)

	// Create reader for data frame
	reader := bytes.NewReader(buf[:n])

	// Parse EHZ channel
	err = binary.Read(reader, binary.LittleEndian, packet.EHZ)
	if err != nil {
		return err
	}

	// Parse EHE channel
	err = binary.Read(reader, binary.LittleEndian, packet.EHE)
	if err != nil {
		return err
	}

	// Parse EHN channel
	err = binary.Read(reader, binary.LittleEndian, packet.EHN)
	if err != nil {
		return err
	}

	// Parse checksum
	for i := 0; i < checksumLen; i++ {
		err = binary.Read(reader, binary.LittleEndian, &packet.Checksum[i])
		if err != nil {
			return err
		}
	}

	// Compare checksum
	err = g.isChecksumCorrect(packet)
	if err != nil {
		return err
	}

	// Get channel counts by offsetting
	packet.EHZ = g.getCounts(packet.EHZ)
	packet.EHE = g.getCounts(packet.EHE)
	packet.EHN = g.getCounts(packet.EHN)

	return nil
}
