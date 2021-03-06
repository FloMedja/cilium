// Copyright 2017, 2018 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package envoy

import (
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cilium/cilium/pkg/flowdebug"
	"github.com/cilium/cilium/pkg/logging"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log = logging.DefaultLogger

var (
	// envoyLevelMap maps logrus.Level values to Envoy (spdlog) log levels.
	envoyLevelMap = map[logrus.Level]string{
		logrus.PanicLevel: "off",
		logrus.FatalLevel: "critical",
		logrus.ErrorLevel: "error",
		logrus.WarnLevel:  "warning",
		logrus.InfoLevel:  "info",
		logrus.DebugLevel: "debug",
		// spdlog "trace" not mapped
	}

	tracing = false
)

// EnableTracing changes Envoy log level to "trace", producing the most logs.
func EnableTracing() {
	tracing = true
}

func mapLogLevel(level logrus.Level) string {
	if tracing {
		return "trace"
	}

	// Suppress the debug level if not debugging at flow level.
	if level == logrus.DebugLevel && !flowdebug.Enabled() {
		level = logrus.InfoLevel
	}
	return envoyLevelMap[level]
}

type admin struct {
	adminURL string
	level    string
}

func (a *admin) transact(query string) error {
	resp, err := http.Get(a.adminURL + query)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	ret := strings.Replace(string(body), "\r", "", -1)
	log.Debugf("Envoy: Admin response to %s: %s", query, ret)
	return nil
}

func (a *admin) changeLogLevel(level logrus.Level) error {
	envoyLevel := mapLogLevel(level)

	if envoyLevel == a.level {
		log.Debugf("Envoy: Log level is already set as: %v", envoyLevel)
		return nil
	}

	err := a.transact("logging?level=" + envoyLevel)
	if err != nil {
		log.WithError(err).Warnf("Envoy: Failed to set log level to: %v", envoyLevel)
	} else {
		a.level = envoyLevel
	}
	return err
}

func (a *admin) quit() error {
	return a.transact("quitquitquit")
}

// Envoy manages a running Envoy proxy instance via the
// ListenerDiscoveryService and RouteDiscoveryService gRPC APIs.
type Envoy struct {
	stopCh chan struct{}
	errCh  chan error
	admin  *admin
}

// GetEnvoyVersion returns the envoy binary version string
func GetEnvoyVersion() string {
	out, err := exec.Command("cilium-envoy", "--version").Output()
	if err != nil {
		log.WithError(err).Fatal(`Envoy: Binary "cilium-envoy" cannot be executed`)
	}
	return strings.TrimSpace(string(out))
}

// StartEnvoy starts an Envoy proxy instance.
func StartEnvoy(adminPort uint32, stateDir, logPath string, baseID uint64) *Envoy {
	bootstrapPath := filepath.Join(stateDir, "bootstrap.pb")
	adminAddress := "127.0.0.1:" + strconv.FormatUint(uint64(adminPort), 10)
	xdsPath := getXDSPath(stateDir)

	e := &Envoy{
		stopCh: make(chan struct{}),
		errCh:  make(chan error, 1),
		admin:  &admin{adminURL: "http://" + adminAddress + "/"},
	}

	// Use the same structure as Istio's pilot-agent for the node ID:
	// nodeType~ipAddress~proxyId~domain
	nodeId := "host~127.0.0.1~no-id~localdomain"

	// Create static configuration
	createBootstrap(bootstrapPath, nodeId, "cluster1", "version1",
		xdsPath, "cluster1", adminPort)

	log.Debugf("Envoy: Starting: %v", *e)

	// make it a buffered channel so we can not only
	// read the written value but also skip it in
	// case no one reader reads it.
	started := make(chan bool, 1)
	go func() {
		logger := &lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    100, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   //days
			Compress:   true, // disabled by default
		}
		defer logger.Close()
		for {
			cmd := exec.Command("cilium-envoy", "-l", mapLogLevel(log.Level), "-c", bootstrapPath, "--base-id", strconv.FormatUint(baseID, 10))
			cmd.Stderr = logger
			cmd.Stdout = logger

			if err := cmd.Start(); err != nil {
				log.WithError(err).Warn("Envoy: Failed to start proxy")
				select {
				case started <- false:
				default:
				}
				return
			}
			log.Debugf("Envoy: Started proxy")
			select {
			case started <- true:
			default:
			}

			log.Infof("Envoy: Proxy started with pid %d", cmd.Process.Pid)

			// We do not return after a successful start, but watch the Envoy process
			// and restart it if it crashes.
			// Waiting for the process execution is done in the goroutime.
			// The purpose of the "crash channel" is to inform the loop about their
			// Envoy process crash - after closing that channel by the goroutime,
			// the loop continues, the channel is recreated and the new process
			// is watched again.
			crashCh := make(chan struct{})
			go func() {
				if err := cmd.Wait(); err != nil {
					log.WithError(err).Warn("Envoy: Proxy crashed")
				}
				close(crashCh)
			}()

			// start again after a short wait. If Cilium exits this should be enough
			// time to not start Envoy again in that case.
			log.Info("Envoy: Sleeping for 100ms before restarting proxy")
			time.Sleep(100 * time.Millisecond)

			select {
			case <-crashCh:
				// Start Envoy again
				continue
			case <-e.stopCh:
				log.Infof("Envoy: Stopping proxy with pid %d", cmd.Process.Pid)
				if err := e.admin.quit(); err != nil {
					log.WithError(err).Fatalf("Envoy: Envoy admin quit failed, killing process with pid %d", cmd.Process.Pid)

					if err := cmd.Process.Kill(); err != nil {
						log.WithError(err).Fatal("Envoy: Stopping Envoy failed")
						e.errCh <- err
					}
				}
				close(e.errCh)
				return
			}
		}
	}()

	if <-started {
		return e
	}

	return nil
}

// isEOF returns true if the error message ends in "EOF". ReadMsgUnix returns extra info in the beginning.
func isEOF(err error) bool {
	strerr := err.Error()
	errlen := len(strerr)
	return errlen >= 3 && strerr[errlen-3:] == io.EOF.Error()
}

// StopEnvoy kills the Envoy process started with StartEnvoy. The gRPC API streams are terminated
// first.
func (e *Envoy) StopEnvoy() error {
	close(e.stopCh)
	err, ok := <-e.errCh
	if ok {
		return err
	}
	return nil
}

// ChangeLogLevel changes Envoy log level to correspond to the logrus log level 'level'.
func (e *Envoy) ChangeLogLevel(level logrus.Level) {
	e.admin.changeLogLevel(level)
}
