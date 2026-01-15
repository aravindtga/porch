// Copyright 2026 The Nephio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package kind

import (
	"github.com/nephio-project/porch/test/disaster/api/environment/shell"
	"github.com/nephio-project/porch/test/e2e/suiteutils"
)

const (
	dbCacheCluster = "porch-disaster-test-dbcache"
	crCacheCluster = "porch-disaster-test-crcache"
)

var (
	dbCacheKubeconfigFile     = "/test/disaster/deployment/kubeconfigs/porch_dbcache.conf"
	crCacheKubeconfigFile     = "/test/disaster/deployment/kubeconfigs/porch_crcache.conf"
	dataClusterKubeconfigFile = "/test/disaster/deployment/kubeconfigs/data_cluster.conf"
)

func Wipe(t *suiteutils.MultiClusterTestSuite) {
	if err := deleteCluster(t, dbCacheCluster); err != nil {
		t.Fatalf("error wiping Porch cluster: error deleting Kind clusters: %w", err)
	}
	t.DropCachedClients(dbCacheKubeconfigFile)
	// if err := deleteCluster(crCacheCluster); err != nil {
	// 	t.Fatalf("error wiping Porch cluster: error deleting Kind clusters: %w", err)
	// }
}

func Reinstall(t *suiteutils.MultiClusterTestSuite) {
	if err := createCluster(t, dbCacheCluster, t.PorchRoot+dbCacheKubeconfigFile); err != nil {
		t.Fatalf("error reinstalling Porch cluster %q: error creating Kind cluster: %w", dbCacheCluster, err)
	}
	// if err := createCluster(crCacheCluster, crCacheKubeconfigFile); err != nil {
	// 	t.Fatalf("error reinstalling Porch cluster %q: error creating Kind cluster: %w", crCacheCluster, err)
	// }

	// recreated cluster has different kubeconfig - re-do client setup
	UseDBCacheCluster(t)
	if err := installPorchDBCache(t); err != nil {
		t.Fatalf("error reinstalling Porch cluster %q: error deploying Porch: %w", dbCacheCluster, err.Error())
	}
}

func UseDataCluster(t *suiteutils.MultiClusterTestSuite) {
	t.UseKubeconfigFile(dataClusterKubeconfigFile)
}
func UseDBCacheCluster(t *suiteutils.MultiClusterTestSuite) {
	t.UseKubeconfigFile(dbCacheKubeconfigFile)
}
func UseCRCacheCluster(t *suiteutils.MultiClusterTestSuite) {
	t.UseKubeconfigFile(crCacheKubeconfigFile)
}

func deleteCluster(t *suiteutils.MultiClusterTestSuite, clusterName string) error {
	return (shell.ShellRunner{PorchRoot: t.PorchRoot}).RunCommandLine("kind", "delete", "cluster", "--name", clusterName)
}

func createCluster(t *suiteutils.MultiClusterTestSuite, clusterName string, kubeconfigPath string) error {
	return (shell.ShellRunner{PorchRoot: t.PorchRoot}).RunCommandLine("kind", "create", "cluster", "--name", clusterName, "--kubeconfig", kubeconfigPath)
}

func installPorchDBCache(t *suiteutils.MultiClusterTestSuite) error {
	return (shell.ShellRunner{PorchRoot: t.PorchRoot}).RunCommandLine("make",
		"-C", t.PorchRoot, "load-images-to-kind",
		"deploy-current-config", "IMAGE_REPO=porch-kind", "IMAGE_TAG=test",
		"KIND_CONTEXT_NAME="+dbCacheCluster, "DEPLOYPORCHCONFIGDIR=\""+t.PorchRoot+"/.build/disaster-test/dbcache\"", "KUBECONFIG="+t.PorchRoot+dbCacheKubeconfigFile)
}
