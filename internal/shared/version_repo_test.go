package shared

import (
	"regexp"
	"strings"
	"testing"
)

// The update check names one repository and the package metadata names
// another: `debian/control` was moved to leunghope/EasyFireWall with the rest
// of the migration, `githubReleasesURL` was not, and every 2.29.2-zh.N build
// advertised upstream's 2.29.2 as an update — a release with no Chinese in it.
// Both are claims about where this fork publishes, so they are compared
// instead of trusted to stay in step.
func TestTheUpdateCheckNamesTheRepositoryThePackagePointsAt(t *testing.T) {
	release := regexp.MustCompile(
		`githubReleasesURL\s*=\s*"https://api\.github\.com/repos/([^/"]+)/([^/"]+)/releases/latest"`)
	m := release.FindStringSubmatch(repoFile(t, "internal", "shared", "version.go"))
	if m == nil {
		t.Fatal("internal/shared/version.go: githubReleasesURL is no longer a " +
			"https://api.github.com/repos/<owner>/<repo>/releases/latest URL; " +
			"update this test's pattern together with it")
	}

	browser := regexp.MustCompile(`(?m)^Vcs-Browser:\s*https://github\.com/([^/\s]+)/(\S+)\s*$`)
	v := browser.FindStringSubmatch(repoFile(t, "debian", "control"))
	if v == nil {
		t.Fatal("debian/control: no `Vcs-Browser: https://github.com/<owner>/<repo>` " +
			"line to compare the update check against")
	}

	if !strings.EqualFold(m[1], v[1]) || !strings.EqualFold(m[2], v[2]) {
		t.Errorf("the update check asks %s/%s while the package points at %s/%s — "+
			"the banner would offer a release from the wrong repository",
			m[1], m[2], v[1], v[2])
	}
}
