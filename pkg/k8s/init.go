// Copyright 2016-2017 Authors of Cilium
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

// Package k8s abstracts all Kubernetes specific behaviour
package k8s

import (
	"fmt"
	"os"

	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/node"

	go_version "github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
)

// Init initializes the Kubernetes package. It is required to call Configure()
// beforehand.
func Init() error {
	compatibleVersions, err := go_version.NewConstraint(compatibleK8sVersions)
	if err != nil {
		return fmt.Errorf("unable to parse compatible k8s verions: %s", err)
	}

	if err := createDefaultClient(); err != nil {
		return fmt.Errorf("unable to create k8s client: %s", err)
	}

	sv, err := GetServerVersion()
	if err != nil {
		return fmt.Errorf("k8s client failed to talk to k8s api-server: %s", err)
	}
	if !compatibleVersions.Check(sv) {
		return fmt.Errorf("k8s version (%v) is not in compatible range (%v)", sv, compatibleK8sVersions)
	}

	if nodeName := os.Getenv(EnvNodeNameSpec); nodeName != "" {
		// Use of the environment variable overwrites the node-name
		// automatically derived
		node.SetName(nodeName)

		k8sNode, err := GetNode(Client(), nodeName)
		if err != nil {
			log.WithError(err).Warning("Unable to retrieve k8s node information, skipping...")
			return nil
		}

		n := ParseNode(k8sNode)
		log.WithField(logfields.NodeName, n.Name).Info("Retrieved node information from kubernetes")

		log.WithFields(logrus.Fields{
			logfields.NodeName:         n.Name,
			logfields.IPAddr + ".ipv4": n.GetNodeIP(false),
			logfields.IPAddr + ".ipv6": n.GetNodeIP(true),
		}).Info("Received own node information from API server")

		if err := node.UseNodeCIDR(n); err != nil {
			return fmt.Errorf("unable to retrieve k8s node CIDR: %s", err)
		}

		if err := node.UseNodeAddresses(n); err != nil {
			return fmt.Errorf("unable to use k8s node addresses: %s", err)
		}

		// Annotate addresses will occur later since the user might
		// want to specify them manually
	}

	return nil
}
