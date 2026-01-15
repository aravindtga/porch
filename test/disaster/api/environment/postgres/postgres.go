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
package postgres

import (
	"os"

	"github.com/nephio-project/porch/test/disaster/api/environment/kind"
	"github.com/nephio-project/porch/test/disaster/api/environment/shell"
	"github.com/nephio-project/porch/test/e2e/suiteutils"
)

const (
	namespace  = "porch-system"
	podName    = "porch-postgresql-0"
	pvcName    = "data-porch-postgresql-0"
	dumpScript = `export PGPASSFILE="/tmp/.pgpass"; echo "localhost:5432:porch:porch:porch" > "$PGPASSFILE"; chmod 600 "$PGPASSFILE"; pg_dump -U porch -h localhost -d porch`
	psqlScript = `export PGPASSFILE="/tmp/.pgpass"; echo "localhost:5432:porch:porch:porch" > "$PGPASSFILE"; chmod 600 "$PGPASSFILE"; psql -U porch -h localhost -d porch`

	dumpFile = "./dumped_db.sql"
)

var (
	porchRoot    = os.Getenv("PORCHDIR")
	wipeDBScript = porchRoot + "/api/sql/porch-db-cleardown.sql"
)

func Backup(t *suiteutils.MultiClusterTestSuite) {
	if err := os.RemoveAll(dumpFile); err != nil {
		t.Fatalf("error backing up Porstres: error deleting previous backup: %w", err)
	}

	kind.UseDataCluster(t)
	args := append([]string{"exec", "-n", namespace, "-i", podName, "--", "bash", "-c"}, dumpScript)
	(shell.ShellRunner{PorchRoot: t.PorchRoot}).RunCommandLineIntoFile(dumpFile, "kubectl", args...)
}

func Wipe(t *suiteutils.MultiClusterTestSuite) {
	kind.UseDataCluster(t)
	args := append([]string{"exec", "-n", namespace, "-i", podName, "--", "bash", "-c"}, psqlScript)
	(shell.ShellRunner{PorchRoot: t.PorchRoot}).RunCommandLineFedFromFile(wipeDBScript, "kubectl", args...)
}

func Restore(t *suiteutils.MultiClusterTestSuite) {
	kind.UseDataCluster(t)
	Wipe(t)
	args := append([]string{"exec", "-n", namespace, "-i", podName, "--", "bash", "-c"}, psqlScript)
	(shell.ShellRunner{PorchRoot: t.PorchRoot}).RunCommandLineFedFromFile(dumpFile, "kubectl", args...)
}
