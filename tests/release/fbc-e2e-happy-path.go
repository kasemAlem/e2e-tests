package release

import (
	// "fmt"
	"time"

	// "github.com/devfile/library/v2/pkg/util"
	// ecp "github.com/enterprise-contract/enterprise-contract-controller/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appservice "github.com/redhat-appstudio/application-api/api/v1alpha1"

	// "github.com/redhat-appstudio/e2e-tests/pkg/constants"
	"github.com/redhat-appstudio/e2e-tests/pkg/framework"
	"github.com/redhat-appstudio/e2e-tests/pkg/utils"
	// "github.com/redhat-appstudio/e2e-tests/pkg/utils/release"
	// "github.com/redhat-appstudio/e2e-tests/pkg/utils/tekton"
	// releaseApi "github.com/redhat-appstudio/release-service/api/v1alpha1"
	// "gopkg.in/yaml.v2"
	// corev1 "k8s.io/api/core/v1"
)

var _ = framework.ReleaseSuiteDescribe("[HACBS-2385] e2e test for fbc happy path.", Label("release", "fbcHappyPath", "HACBS"), func() {
	defer GinkgoRecover()

	var fw *framework.Framework
	var err error

	var component *appservice.Component
	// var releaseCR *releaseApi.Release
	var stageOfflineToken = ""
	var devNamespace = "dev-release-team"
	// var managedNamespace = "managed-release-team-tenant"

	stageOptions := utils.Options{
		ToolchainApiUrl: "https://api-toolchain-host-operator.apps.stone-stg-host.qc0p.p1.openshiftapps.com" + "/workspaces/" + devNamespace,
		KeycloakUrl:     "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token",
		OfflineToken:    stageOfflineToken,
	}
	GinkgoWriter.Printf("\nToolchainApiUrl : %s\n", stageOptions.ToolchainApiUrl)
	BeforeAll(func() {
		// Initialize the tests controllers

		fw, err = framework.NewFrameworkWithTimeout(
			devNamespace,
			time.Minute*60,
			stageOptions,
		)

		// TODO Kasem
		_, err = fw.AsKubeDeveloper.HasController.CreateApplication(fbcApplicationName, fw.UserNamespace)
		Expect(err).NotTo(HaveOccurred())

		component, err = fw.AsKubeDeveloper.HasController.CreateComponentWithDockerSource(fbcApplicationName, fbcComponentName, fw.UserNamespace, fbcSourceGitUrl, "", "", "", 50051)
		//

	})

	// AfterAll(func() {
	// 	err = fw.AsKubeAdmin.CommonController.Github.DeleteRef(constants.StrategyConfigsRepo, scGitRevision)
	// 	if err != nil {
	// 		Expect(err.Error()).To(ContainSubstring("Reference does not exist"))
	// 	}
	// 	if !CurrentSpecReport().Failed() {
	// 		Expect(fw.SandboxController.DeleteUserSignup(fw.UserName)).To(BeTrue())
	// 		Expect(fw.AsKubeAdmin.CommonController.DeleteNamespace(managedNamespace)).NotTo(HaveOccurred())
	// 	}
	// })

	var _ = Describe("Post-release verification", func() {

		It("verify app and component are created.", func() {
			component, err = fw.AsKubeAdmin.HasController.GetComponent(fbcComponentName, fw.UserNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(fw.AsKubeAdmin.HasController.WaitForComponentPipelineToBeFinished(component, "", 2)).To(Succeed())
		})

		// It("verifies that a Release CR should have been created in the dev namespace", func() {
		// 	Eventually(func() error {
		// 		releaseCR, err = fw.AsKubeAdmin.ReleaseController.GetFirstReleaseInNamespace(devNamespace)
		// 		return err
		// 	}, releaseCreationTimeout, defaultInterval).Should(Succeed())
		// })

		// It("verifies that Release PipelineRun is triggered", func() {
		// 	Eventually(func() error {
		// 		pr, err := fw.AsKubeAdmin.ReleaseController.GetPipelineRunInNamespace(managedNamespace, releaseCR.GetName(), releaseCR.GetNamespace())
		// 		if err != nil {
		// 			GinkgoWriter.Printf("release pipelineRun for release '%s' in namespace '%s' not created yet: %+v\n", releaseCR.GetName(), releaseCR.GetNamespace(), err)
		// 			return err
		// 		}
		// 		if !pr.HasStarted() {
		// 			return fmt.Errorf("pipelinerun %s/%s hasn't started yet", pr.GetNamespace(), pr.GetName())
		// 		}
		// 		return nil
		// 	}, releasePipelineRunCreationTimeout, defaultInterval).Should(Succeed(), fmt.Sprintf("timed out waiting for a pipelinerun to start for a release %s/%s", releaseCR.GetName(), releaseCR.GetNamespace()))
		// })

		// It("verifies that Release PipelineRun should eventually succeed", func() {
		// 	Eventually(func() error {
		// 		pr, err := fw.AsKubeAdmin.ReleaseController.GetPipelineRunInNamespace(managedNamespace, releaseCR.GetName(), releaseCR.GetNamespace())
		// 		Expect(err).ShouldNot(HaveOccurred())
		// 		if !pr.IsDone() {
		// 			return fmt.Errorf("release pipelinerun %s/%s did not finish yet", pr.GetNamespace(), pr.GetName())
		// 		}
		// 		Expect(utils.HasPipelineRunSucceeded(pr)).To(BeTrue(), fmt.Sprintf("release pipelinerun %s/%s did not succeed", pr.GetNamespace(), pr.GetName()))
		// 		return nil
		// 	}, releasePipelineRunCompletionTimeout, defaultInterval).Should(Succeed())
		// })

		// It("verifies that a Release is marked as succeeded.", func() {
		// 	Eventually(func() error {
		// 		releaseCR, err = fw.AsKubeAdmin.ReleaseController.GetFirstReleaseInNamespace(devNamespace)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		if !releaseCR.IsReleased() {
		// 			return fmt.Errorf("release %s/%s is not marked as finished yet", releaseCR.GetNamespace(), releaseCR.GetName())
		// 		}
		// 		return nil
		// 	}, releaseCreationTimeout, defaultInterval).Should(Succeed())
		// })
	})
})
