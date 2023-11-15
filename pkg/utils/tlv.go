package utils

import (
	"encoding/binary"
	"errors"
	"io"
)

func ReadUint32(reader io.Reader) (uint32, error) {
	tagBuffer := make([]byte, 4)

	read, err := reader.Read(tagBuffer)
	if err != nil && err != io.EOF {
		return 0, err
	}

	if read != 4 {
		return 0, errors.New("failed to read valid tag")
	}

	return binary.LittleEndian.Uint32(tagBuffer), nil
}

func Write(writer io.Writer, tag uint32, buffer []byte) error {
	_, err := writer.Write(binary.LittleEndian.AppendUint32([]byte{}, tag))
	if err != nil {
		return err
	}
	_, err = writer.Write(binary.LittleEndian.AppendUint32([]byte{}, uint32(len(buffer))))
	if err != nil {
		return err
	}
	_, err = writer.Write(buffer)
	if err != nil {
		return err
	}

	return nil
}

func Read(reader io.Reader) (uint32, []byte, error) {
	tag, err := ReadUint32(reader)
	if err != nil {
		return 0, nil, err
	}
	length, err := ReadUint32(reader)
	if err != nil {
		return 0, nil, err
	}

	if length > 2048 {
		return 0, nil, errors.New("unsupported length read")
	}

	readbuffer := make([]byte, length)
	_, err = reader.Read(readbuffer)
	if err != nil && err != io.EOF {
		return 0, nil, err
	}
	return tag, readbuffer, nil

	/*
		bytesRead := 0

		readbuffer := make([]byte, 0, 1024)

		logrus.Debugf("Starting to read buffer, need %d", length)

		buffer := []byte{}
		for {
			n, err := reader.Read(readbuffer)
			if n == 0 {
				if err != nil && err != io.EOF {
					return 0, nil, err
				}
				if err == io.EOF {
					break
				}
				continue
			}
			logrus.Debugf("Read %d bytes from stream %d vs %d \n", n, bytesRead, length)
			bytesRead += n
			buffer = append(buffer, readbuffer[:n]...)
			if bytesRead >= int(length) {
				break
			}
		}

		return tag, buffer, nil
	*/
}
