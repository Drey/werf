package e2e_build_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type simpleTestOptions struct {
	setupEnvOptions
}

var _ = Describe("Simple build", Label("e2e", "build", "simple"), func() {
	DescribeTable("should succeed and produce expected image",
		func(testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By(fmt.Sprintf("%s: starting", testOpts.State))
			{
				repoDirname := "repo0"
				fixtureRelPath := fmt.Sprintf("simple/%s", testOpts.State)
				buildReportName := "report0.json"

				By(fmt.Sprintf("%s: preparing test repo", testOpts.State))
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

				By(fmt.Sprintf("%s: building images", testOpts.State))
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				buildOut, buildReport := werfProject.BuildWithReport(SuiteData.GetBuildReportPath(buildReportName), nil)
				Expect(buildOut).To(ContainSubstring("Building stage"))
				Expect(buildOut).NotTo(ContainSubstring("Use previously built image"))

				By(fmt.Sprintf("%s: rebuilding same images", testOpts.State))
				Expect(werfProject.Build(nil)).To(And(
					ContainSubstring("Use previously built image"),
					Not(ContainSubstring("Building stage")),
				))

				By(fmt.Sprintf(`%s: checking "dockerfile" image content`, testOpts.State))
				contRuntime.ExpectCmdsToSucceed(
					buildReport.Images["dockerfile"].DockerImageName,
					"test -f /file",
					"echo 'filecontent' | diff /file -",

					"test -f /created-by-run",
				)

				By(fmt.Sprintf(`%s: checking "stapel-shell" image content`, testOpts.State))
				contRuntime.ExpectCmdsToSucceed(
					buildReport.Images["stapel-shell"].DockerImageName,
					"test -f /file",
					"stat -c %u:%g /file | diff <(echo 0:0) -",
					"echo 'filecontent' | diff /file -",

					"test -f /created-by-setup",
				)
			}
		},
		Entry("without repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
			State:                       "state0",
		}}),
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
			State:                       "state0",
		}}),
		Entry("without repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
			State:                       "state1",
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
			State:                       "state1",
		}}),
		Entry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
			State:                       "state0", // TODO(iapershin): change after buildah version upgrade
		}}),
		Entry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
			State:                       "state1",
		}}),
		// TODO(ilya-lesikov): uncomment after Staged Dockerfile builder finished
		// // TODO(1.3): after Full Dockerfile Builder removed and Staged Dockerfile Builder enabled by default this test no longer needed
		// Entry("with local repo using Native Buildah and Staged Dockerfile Builder with rootless isolation", simpleTestOptions{setupEnvOptions{
		// 	ContainerBackendMode:                 "native-rootless",
		// 	WithLocalRepo:               true,
		// 	WithStagedDockerfileBuilder: true,
		// }),
		// TODO(ilya-lesikov): uncomment after Staged Dockerfile builder finished
		// // TODO(1.3): after Full Dockerfile Builder removed and Staged Dockerfile Builder enabled by default this test no longer needed
		// Entry("with local repo using Native Buildah and Staged Dockerfile Builder with chroot isolation", simpleTestOptions{setupEnvOptions{
		// 	ContainerBackendMode:                 "native-chroot",
		// 	WithLocalRepo:               true,
		// 	WithStagedDockerfileBuilder: true,
		// }),
	)
})
