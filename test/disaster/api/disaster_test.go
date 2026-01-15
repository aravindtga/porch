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
package disaster

import (
	"os"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/helper/chroot"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/nephio-project/porch/test/disaster/api/environment/gitea"
	"github.com/nephio-project/porch/test/disaster/api/environment/kind"
	"github.com/nephio-project/porch/test/disaster/api/environment/packagevariants"
	"github.com/nephio-project/porch/test/disaster/api/environment/packagevariantsets"
	repositories "github.com/nephio-project/porch/test/disaster/api/environment/porchrepositories"
	"github.com/nephio-project/porch/test/disaster/api/environment/postgres"
	"github.com/nephio-project/porch/test/e2e/suiteutils"
	"github.com/stretchr/testify/suite"
)

type PorchDisasterSuite struct {
	suiteutils.MultiClusterTestSuite

	skipVariants    bool
	skipVariantSets bool
}

func (t *PorchDisasterSuite) SetupSuite() {
	suite := &t.MultiClusterTestSuite

	t.PorchRoot = func() string {

		repo, _ := git.PlainOpen("../../../")

		// Try to grab the repository Storer
		s := repo.Storer.(*filesystem.Storage)
		// Try to get the underlying billy.Filesystem
		fs := s.Filesystem().(*chroot.ChrootHelper)

		return strings.Replace(fs.Root(), "/.git", "", 1)
	}()

	if t.skipVariantSets {
		os.Setenv("SKIP_VARIANT_SETS", "true")
	}
	if t.skipVariants {
		os.Setenv("SKIP_VARIANTS", "true")
	}

	if os.Getenv("SETUP_ENV") == "true" {
		setupEnv(suite)
	}
	if os.Getenv("SETUP_ENV") == "reset" {
		resetEnv(suite)
	}

	suite.SetupSuite()
}

func TestDisasterRecovery(t *testing.T) {
	// Skip if not running E2E tests
	if os.Getenv("E2E") == "" {
		t.Skip("Skipping disaster-recovery tests in non-E2E environment")
	}

	disSuite := PorchDisasterSuite{
		skipVariants:    true,
		skipVariantSets: true,
	}
	suite.Run(t, &disSuite)
}

func (t *PorchDisasterSuite) TestCompleteDisaster() {
	s := &t.MultiClusterTestSuite

	// expectedCountsBefore := &suiteutils.PackageRevisionStatusCounts{
	// 	Total:            1100,
	// 	Draft:            5,
	// 	Proposed:         5,
	// 	Published:        1065,
	// 	DeletionProposed: 5,
	// }
	expectedCountsBefore := &suiteutils.PackageRevisionStatusCounts{
		Total:            1044,
		Draft:            5,
		Proposed:         5,
		Published:        1029,
		DeletionProposed: 5,
	}
	t.PackageRevisionCountsMustMatch(expectedCountsBefore)

	kind.UseDataCluster(s)
	gitea.Backup(s)
	postgres.Backup(s)

	kind.UseDBCacheCluster(s)
	repositoriesToReconcile := repositories.Backup(s)
	variantsToReconcile := packagevariants.Backup(s)
	variantSetsToReconcile := packagevariantsets.Backup(s)

	kind.Wipe(s)
	kind.UseDataCluster(s)
	gitea.Wipe(s)
	postgres.Wipe(s)

	kind.Reinstall(s)
	kind.UseDataCluster(s)
	gitea.Restore(s)
	postgres.Restore(s)

	kind.UseDBCacheCluster(s)
	repositories.Reconcile(s, repositoriesToReconcile, 20)
	packagevariants.Reconcile(s, variantsToReconcile, 20)
	packagevariantsets.Reconcile(s, variantSetsToReconcile, 20)

	t.PackageRevisionCountsMustMatch(expectedCountsBefore)
}
