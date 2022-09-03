package daemon

import (
	"os"
	"path"
	"syscall"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/kubectlfzfserver"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"

	"github.com/sevlyar/go-daemon"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func SetDaemonFlags(fs *pflag.FlagSet) {
	fs.String("daemon", "", `Send signal to the daemon:
  start - run as a daemon
  stop — fast shutdown`)
	defaultName := path.Base(os.Args[0])
	fs.String("daemon-name", defaultName, "Daemon name")
	defaultPidPath := path.Join("/tmp/kubectl-fzf-server", defaultName+".pid")
	defaultLogPath := path.Join("/tmp/kubectl-fzf-server", defaultName+".log")
	fs.String("daemon-pid-file", defaultPidPath, "Daemon's PID file path")
	fs.String("daemon-log-file", defaultLogPath, "Daemon's log file path")
}

func termHandler(sig os.Signal) error {
	logrus.Infoln("Terminating daemon...")
	return daemon.ErrStop
}

func StartDaemon() {
	daemonCmd := viper.GetString("daemon")
	daemonPidFilePath := viper.GetString("daemon-pid-file")
	daemonLogFilePath := viper.GetString("daemon-log-file")
	daemonName := viper.GetString("daemon-name")

	pidDir := path.Dir(daemonPidFilePath)
	logDir := path.Dir(daemonLogFilePath)
	err := os.MkdirAll(pidDir, 0755)
	util.FatalIf(err)
	err = os.MkdirAll(logDir, 0755)
	util.FatalIf(err)

	daemon.AddCommand(daemon.StringFlag(&daemonCmd, "stop"), syscall.SIGTERM, termHandler)
	cntxt := &daemon.Context{
		PidFileName: daemonPidFilePath,
		PidFilePerm: 0644,
		LogFileName: daemonLogFilePath,
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{daemonName},
	}

	if len(daemon.ActiveFlags()) > 0 {
		logrus.Infof("Stopping daemon...")
		d, err := cntxt.Search()
		if err != nil {
			logrus.Fatalf("Unable send signal to the daemon: %s", err)
		}
		err = daemon.SendCommands(d)
		if err != nil {
			logrus.Fatalf("Unable send command to the daemon: %s", err)
		}
		return
	}
	logrus.Infof("Starting daemon...")

	d, err := cntxt.Reborn()
	if err != nil {
		logrus.Fatalln(err)
	}
	if d != nil {
		return
	}
	defer util.FatalIf(cntxt.Release())

	logrus.Infoln("- - - - - - - - - - - - - - -")
	logrus.Infoln("daemon started")

	go kubectlfzfserver.StartKubectlFzfServer()

	err = daemon.ServeSignals()
	if err != nil {
		logrus.Infof("Error: %s", err.Error())
	}
	logrus.Infoln("daemon terminated")
}
