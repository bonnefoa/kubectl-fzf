package util

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
)

func EncodeToFile(data interface{}, filePath string) error {
	m := new(bytes.Buffer)
	enc := gob.NewEncoder(m)
	err := enc.Encode(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, m.Bytes(), 0600)
	return err
}

func LoadFromFile(e interface{}, filePath string) error {
	n, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	p := bytes.NewBuffer(n)
	dec := gob.NewDecoder(p)
	err = dec.Decode(e)
	return err
}