# The Gitea release pipeline

`.gitea/workflows/release.yml` is not a second copy of
[ci-and-release](ci-and-release.md)'s seventh workflow. It is what is left of
that pipeline after the parts that only work on GitHub are removed, and it
exists because this fork is published on Gitea, where the upstream file cannot
run at all: it calls `gh api`, pushes to GHCR, Docker Hub and Quay, posts to
Discord, asks for `ubuntu-26.04` runners, and gates on the tagged commit being
on `main`. This branch is not on `main`.

## What a tag produces

Push a tag that matches the version the tree declares, and one job builds and
publishes everything:

| Asset | Architecture | Built by |
|---|---|---|
| `easywall_<version>_linux_amd64.tar.gz` | amd64 | `go build`, cross-compiled, layout as `.goreleaser.yaml`'s archives |
| `easywall_<version>_linux_arm64.tar.gz` | arm64 | the same, `GOARCH=arm64` |
| `easywall_<version>_amd64.deb` | amd64 | `dpkg-buildpackage -b -d`, i.e. `debian/` unmodified |
| `checksums.txt` | — | `sha256sum` over the three above |
| `<registry>/<owner>/<repo>:<version>` and `:zh` | amd64 | `docker build` against the host's daemon, when one is reachable |

The release body is the `CHANGELOG.md` section for that version, extracted by
`awk`. Nothing is typed into the release twice, so the page and the repository
cannot disagree.

## The tag is the version, or nothing ships

The first step compares the tag against three files — `debian/changelog`'s first
line, `var CurrentVersion` in `internal/shared/version.go`, and the newest `## [x]`
heading in `CHANGELOG.md` — and fails if any of them disagrees. Three places, one
number, because the three are read by different things: dpkg reads the changelog,
the linker flag writes `version.go`, and the release page reads the changelog
heading.

Two of those three are also enforced by the unit tests
(`TestThePackageVersionIsTheReleaseVersion`,
`TestDocsVersionMatchesRelease`), so a mismatch normally fails before a tag ever
exists. The workflow check exists for the case where the tag disagrees with a
tree that is internally consistent — tagging an older commit, or a typo in the
tag name.

## What the probe changed

The first version of this pipeline was written from the upstream file and
assumptions about a self-hosted runner. A throwaway workflow triggered by a
`probe-*` tag measured what the runner actually does, and three of those
assumptions were wrong:

| Assumed | Measured | Consequence |
|---|---|---|
| `actions/checkout@v4` works | `Cannot find: node in PATH` — the runner executes jobs on the host as root, with bash, and no node | the checkout is `git init` + `git fetch --depth=1` + `git checkout`, using the release token |
| act's default shell is bash | `sh -e {0}`, and Debian's `sh` is dash: `set: Illegal option -o pipefail` | the job sets `defaults.run.shell: bash` |
| the runner can build an arm64 image | no `qemu-aarch64` in `/proc/sys/fs/binfmt_misc` | the image is amd64 only; the tarballs and the `.deb` still cover both architectures |

What the probe confirmed: `container:` works and mounts both the workspace and
`/var/run/docker.sock`, `docker pull` reaches Docker Hub, and both secrets are
present. The host itself has no `docker` CLI — only the socket — so the image
step installs the CLI inside the job container and talks to the host's daemon
through the mount.

## Cutting a release

```bash
# 1. the three version files agree, and the tests agree with them
make test

# 2. push the branch, then the tag — the tag is what triggers the run.
#    `gitea` is what this working copy calls the Gitea remote, because `origin`
#    is still the upstream repository it was forked from.
git push gitea zh_locales
git tag -a v2.29.2-zh.1 -m "easywall 2.29.2-zh.1"
git push gitea refs/tags/v2.29.2-zh.1
```

The run appears under the repository's **Actions** tab. Everything it needs is
two secrets, set under **Settings → Actions → Secrets**:

| Secret | What it is |
|---|---|
| `RELEASE_TOKEN` | a personal access token with `repository` read/write and `package` read/write. Used to clone, to call the releases API, and to push the image |
| `REGISTRY_USER` | the account that token belongs to; the token itself is the password |

## Re-running a release

Delete the tag on the remote, push it again:

```bash
git push origin :refs/tags/v2.29.2-zh.1
git push origin refs/tags/v2.29.2-zh.1
```

The publish step is written for this: it looks the release up by tag, updates it
if it exists, creates it if it does not, deletes an asset with the same name
before uploading, and then uploads. A re-run replaces assets instead of failing
with a 409 on the first one that is already there.

## When it fails

The step names say where. The three that will actually bite:

- **`the tag is the version the tree declares`** — the tag is not the version in
  the tree. Decide which one is wrong; do not "fix" the workflow.
- **`Debian package`** — the `.deb` was built but did not contain what it should:
  no `easywall-core`/`easywall-web`, a `Version` that is not the tag, a missing
  `locales/zh.json`. The step refuses to publish it, which is the point: the
  packaging trap in [packaging](packaging.md) shipped a package with no binaries,
  and every assertion here is one more line that would have caught it.
- **`container image`** — marked `continue-on-error: true`, so the tarballs and
  the `.deb` still publish, and the release body says the image is missing. A
  registry problem should not cost the assets that run without a container
  runtime.

Everything else — the version assertion, the unit tests, the tarball contents,
the `.deb` — runs before the release exists. A defect stops the run with nothing
published, rather than after a page has been announced with a broken asset.

## Three ways an assertion lies about a good build

All three were found by running the pipeline, not by reading it. Each reported a
missing artefact on a tree that was correct, so each is a trap worth keeping out.

- **`grep -q` under `set -o pipefail`.** `tar -tzf … | grep -qx …` reports
  failure on a match: `grep -q` exits at the first hit, the producer takes
  `SIGPIPE`, and `pipefail` makes the pipeline non-zero. The listing is captured
  once and grepped without `-q`. The same shape was in the `.deb` checks and the
  `file` architecture checks.
- **`tar` spells a directory with a trailing slash.** `…/web` does not match the
  entry `…/web/`, so a contents check on a directory can only pass by accident.
  Both spellings are accepted.
- **`ls -1 easywall_*` answers a directory with its contents.** The build
  directories in `out/` matched the glob, so `sha256sum $(ls -1 easywall_*)`
  received `easywall-core`, `web`, `locales` and the rest, and exited 1. Both the
  manifest and the upload list now walk the glob and keep what `[ -f ]` accepts.

The general rule: an assertion about an artefact is itself code, and it is
verified by running it against a good build *and* against a broken one. The
negative case matters as much — a contents check that cannot fail is not a
check, which is why each of these was tested with an entry deliberately removed.

## Limits worth knowing

- **The `.deb` is amd64 only.** Building arm64 needs `dpkg-buildpackage` under
  emulation, which this host does not have. Upstream builds it on a native
  `ubuntu-24.04-arm` runner.
- **`-d` skips Debian's build-dependency check.** `debian/control` asks for
  `golang-go (>= 2:1.27~)`, which no Debian suite ships yet; the Go toolchain
  comes from the job image instead, the same substitution upstream makes with
  `actions/setup-go`.
- **This pipeline does not run on `main`.** `.github/workflows/` still holds the
  GitHub release, untouched, for the upstream repository.