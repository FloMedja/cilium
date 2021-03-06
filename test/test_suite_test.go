// Copyright 2017 Authors of Cilium
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

package ciliumTest

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/cilium/cilium/pkg/logging"
	"github.com/cilium/cilium/test/config"
	. "github.com/cilium/cilium/test/ginkgo-ext"
	ginkgoext "github.com/cilium/cilium/test/ginkgo-ext"
	"github.com/cilium/cilium/test/helpers"

	gops "github.com/google/gops/agent"
	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var (
	log             = logging.DefaultLogger
	DefaultSettings = map[string]string{
		"K8S_VERSION": "1.9",
	}
	k8sNodesEnv         = "K8S_NODES"
	commandsLogFileName = "cmds.log"
)

func init() {

	// Open socket for using gops to get stacktraces in case the tests deadlock.
	if err := gops.Listen(gops.Options{}); err != nil {
		errorString := fmt.Sprintf("unable to start gops: %s", err)
		fmt.Println(errorString)
		os.Exit(-1)
	}

	for k, v := range DefaultSettings {
		getOrSetEnvVar(k, v)
	}

	config.CiliumTestConfig.ParseFlags()

	os.RemoveAll(helpers.TestResultsPath)
}

func configLogsOutput() {
	log.SetLevel(logrus.DebugLevel)
	log.Out = &config.TestLogWriter
	logrus.SetFormatter(&config.Formatter)
	log.Formatter = &config.Formatter
	log.Hooks.Add(&config.LogHook{})
}

func ShowCommands() {
	if !config.CiliumTestConfig.ShowCommands {
		return
	}

	helpers.SSHMetaLogs = helpers.NewWriter(os.Stdout)
}

func TestTest(t *testing.T) {
	configLogsOutput()
	ShowCommands()

	if config.CiliumTestConfig.HoldEnvironment {
		RegisterFailHandler(helpers.Fail)
	} else {
		RegisterFailHandler(Fail)
	}
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf(
		"%s.xml", ginkgoext.GetScopeWithVersion()))
	RunSpecsWithDefaultAndCustomReporters(
		t, ginkgoext.GetScopeWithVersion(), []ginkgo.Reporter{junitReporter})
}

func goReportVagrantStatus() chan bool {
	if ginkgoconfig.DefaultReporterConfig.Verbose ||
		ginkgoconfig.DefaultReporterConfig.Succinct {
		// Dev told us they want more/less information than default. Skip.
		return nil
	}

	exit := make(chan bool)
	go func() {
		done := false
		iter := 0
		for {
			var out string
			select {
			case ok := <-exit:
				if ok {
					out = "●\n"
				} else {
					out = "◌\n"
				}
				done = true
			default:
				out = string(rune(int('◜') + iter%4))
			}
			fmt.Printf("\rSpinning up vagrant VMs... %s", out)
			if done {
				return
			}
			time.Sleep(250 * time.Millisecond)
			iter++
		}
	}()
	return exit
}

var _ = BeforeAll(func() {
	var err error

	if !config.CiliumTestConfig.Reprovision {
		// The developer has explicitly told us that they don't care
		// about updating Cilium inside the guest, so skip setup below.
		return
	}

	if config.CiliumTestConfig.SSHConfig != "" {
		// If we set a different VM that it's not in our test environment
		// ginkgo cannot provision it, so skip setup below.
		return
	}

	if progressChan := goReportVagrantStatus(); progressChan != nil {
		defer func() { progressChan <- err == nil }()
	}

	switch ginkgoext.GetScope() {
	case helpers.Runtime:
		err = helpers.CreateVM(helpers.Runtime)
		if err != nil {
			Fail(fmt.Sprintf("error starting VM %q: %s", helpers.Runtime, err))
		}

		vm := helpers.InitRuntimeHelper(helpers.Runtime, log.WithFields(
			logrus.Fields{"testName": "BeforeSuite"}))
		err = vm.SetUpCilium()

		if err != nil {
			Fail(fmt.Sprintf("cilium was unable to be set up correctly: %s", err))
		}

	case helpers.K8s:
		//FIXME: This should be:
		// Start k8s1 and provision kubernetes.
		// When finish, start to build cilium in background
		// Start k8s2
		// Wait until compilation finished, and pull cilium image on k8s2

		// Name for K8s VMs depends on K8s version that is running.

		err = helpers.CreateVM(helpers.K8s1VMName())
		if err != nil {
			Fail(fmt.Sprintf("error starting VM %q: %s", helpers.K8s1VMName(), err))
		}

		err = helpers.CreateVM(helpers.K8s2VMName())

		if err != nil {
			Fail(fmt.Sprintf("error starting VM %q: %s", helpers.K8s2VMName(), err))
		}
		// For Nightly test we need to have more than two kubernetes nodes. If
		// the env variable K8S_NODES is present, more nodes will be created.
		if nodes := os.Getenv(k8sNodesEnv); nodes != "" {
			nodesInt, err := strconv.Atoi(nodes)
			if err != nil {
				Fail(fmt.Sprintf("%s value is not a number %q", k8sNodesEnv, nodes))
			}
			for i := 3; i <= nodesInt; i++ {
				vmName := fmt.Sprintf("%s%d-%s", helpers.K8s, i, helpers.GetCurrentK8SEnv())
				err = helpers.CreateVM(vmName)
				if err != nil {
					Fail(fmt.Sprintf("error starting VM %q: %s", vmName, err))
				}
			}
		}
	}
	return
})

var _ = AfterAll(func() {
	if !helpers.IsRunningOnJenkins() {
		log.Infof("AfterSuite: not running on Jenkins; leaving VMs running for debugging")
		return
	}

	scope := ginkgoext.GetScope()
	log.Infof("cleaning up VMs started for %s tests", scope)
	switch scope {
	case helpers.Runtime:
		helpers.DestroyVM(helpers.Runtime)
	case helpers.K8s:
		helpers.DestroyVM(helpers.K8s1VMName())
		helpers.DestroyVM(helpers.K8s2VMName())
	}
	return
})

func getOrSetEnvVar(key, value string) {
	if val := os.Getenv(key); val == "" {
		log.Infof("environment variable %q was not set; setting to default value %q", key, value)
		os.Setenv(key, value)
	}
}

var _ = AfterEach(func() {

	defer config.TestLogWriterReset()
	err := helpers.CreateLogFile(config.TestLogFileName, config.TestLogWriter.Bytes())
	if err != nil {
		log.WithError(err).Errorf("cannot create log file '%s'", config.TestLogFileName)
		return
	}

	defer helpers.SSHMetaLogs.Reset()
	err = helpers.CreateLogFile(commandsLogFileName, helpers.SSHMetaLogs.Bytes())
	if err != nil {
		log.WithError(err).Errorf("cannot create log file '%s'", commandsLogFileName)
		return
	}
})
