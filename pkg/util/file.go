package util

import (
	"os"

	"github.com/sirupsen/logrus"
)

func RemoveTempDir(tempDir string) {
	err := os.RemoveAll(tempDir)
	logrus.Warnf("Couldn't remove tempdir %s: %s", tempDir, err)
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
