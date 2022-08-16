package util

import (
	"fmt"
	"net"
	"runtime/debug"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func IsAddressReachable(address string) bool {
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		logrus.Infof("Couldn't connect to %s: %s", address, err)
		return false
	}
	conn.Close()
	return true
}

// FatalIf exits if the error is not nil
func FatalIf(err error) {
	if err != nil {
		if stackErr, ok := err.(stackTracer); ok {
			logrus.WithField("stacktrace", fmt.Sprintf("%+v", stackErr.StackTrace()))
		} else {
			debug.PrintStack()
		}
		logrus.Fatalf("Fatal error: %s\n", err)
	}
}

// TimeToAge converts a time to a age string
func TimeToAge(t time.Time) string {
	duration := time.Since(t)
	duration = duration.Round(time.Minute)
	if duration.Hours() > 30 {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
	hour := duration / time.Hour
	duration -= hour * time.Hour
	minute := duration / time.Minute
	return fmt.Sprintf("%02d:%02d", hour, minute)
}
