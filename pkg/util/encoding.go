package util

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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

func LoadFromFile(e interface{}, filePath string) error {
	logrus.Debugf("Loading file %s", filePath)
	n, err := ioutil.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "error reading file")
	}
	gzipBuf := bytes.NewBuffer(n)
	zr, err := gzip.NewReader(gzipBuf)
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

func LoadFromHttpServer(e interface{}, url string) error {
	logrus.Debugf("Loading from %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "error on get of %s", url)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("error retrieving resource from server: %s", resp.Status)
	}
	defer resp.Body.Close()
	n, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading body content")
	}
	gzipBuf := bytes.NewBuffer(n)
	zr, err := gzip.NewReader(gzipBuf)
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
