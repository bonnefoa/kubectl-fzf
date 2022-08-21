package util

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func EncodeToFile(data interface{}, filePath string) error {
	logrus.Debugf("Writing encoded data in %s", filePath)

	var gobBuf bytes.Buffer
	enc := gob.NewEncoder(&gobBuf)
	err := enc.Encode(data)
	if err != nil {
		return errors.Wrap(err, "error encoding gob data")
	}

	writer, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, "error creating target file")
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = filePath
	defer archiver.Close()

	_, err = io.Copy(archiver, &gobBuf)
	return err
}

func LoadGobFromFile(e interface{}, filePath string) error {
	logrus.Debugf("Loading file %s", filePath)
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "error reading file")
	}
	return DecodeGob(e, b)
}

func DecodeGob(e interface{}, b []byte) error {
	bbuffer := bytes.NewBuffer(b)
	zr, err := gzip.NewReader(bbuffer)
	if err != nil {
		return errors.Wrap(err, "error creating new gzip reader")
	}
	dec := gob.NewDecoder(zr)
	err = dec.Decode(e)
	if err := zr.Close(); err != nil {
		return errors.Wrap(err, "error closing gzip reader")
	}
	return err
}
