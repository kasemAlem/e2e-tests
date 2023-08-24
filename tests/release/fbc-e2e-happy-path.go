package release

import (
	"fmt"
	"time"

	ecp "github.com/enterprise-contract/enterprise-contract-controller/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appservice "github.com/redhat-appstudio/application-api/api/v1alpha1"
	"github.com/redhat-appstudio/e2e-tests/pkg/framework"
	"github.com/redhat-appstudio/e2e-tests/pkg/utils"
	releaseApi "github.com/redhat-appstudio/release-service/api/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

var _ = framework.ReleaseSuiteDescribe("[HACBS-2385] e2e test for fbc happy path.", Label("release", "fbcHappyPath", "HACBS"), func() {
	defer GinkgoRecover()

	var err error
	var devWorkspace = "dev-release-team"
	var managedWorkspace = "managed-release-team"
	var devNamespace = devWorkspace + "-tenant"
	var managedNamespace = managedWorkspace + "-tenant"

	var fbcApplicationName = "fbc-kas-app"
	var fbcComponentName = "fbc-kas-comp"
	var fbcReleasePlanName = "fbc-kas-releaseplan"
	var fbcReleasePlanAdmissionName = "fbc-kas-releaseplanadmission"
	var fbcReleaseStrategyName = "fbc-kas-strategy"
	var fbcEnterpriseContractPolicyName = "fbc-kas-policy"
	var fbcServiceAccountName = "release-service-account"
	var fbcSourceGitUrl = "https://github.com/redhat-appstudio-qe/fbc-sample-repo"

	var component *appservice.Component
	var releaseCR *releaseApi.Release
	var buildPr *v1beta1.PipelineRun
	var releasePr *v1beta1.PipelineRun
	var snapshot *appservice.Snapshot

	// We can put it in json file as performance team do

	stageOptions := utils.Options{
		ToolchainApiUrl: utils.GetEnv("RHTAP_TOOLCHAIN_API_URL", ""),
		KeycloakUrl:     utils.GetEnv("RHTAP_KEYLOAK_URL", ""),
		OfflineToken:    utils.GetEnv("RHTAP_OFFLINE_TOKEN", ""),
	}

	dev_fw, err := framework.NewFrameworkWithTimeout(
		devWorkspace,
		time.Minute*2,
		stageOptions,
	)
	Expect(err).NotTo(HaveOccurred())

	managed_fw, err := framework.NewFrameworkWithTimeout(
		managedWorkspace,
		time.Minute*2,
		stageOptions,
	)
	Expect(err).NotTo(HaveOccurred())

	BeforeAll(func() {

		_, err = dev_fw.AsKubeDeveloper.HasController.CreateApplication(fbcApplicationName, dev_fw.UserNamespace)
		Expect(err).NotTo(HaveOccurred())

		componentObj := appservice.ComponentSpec{
			ComponentName: fbcComponentName,
			Application:   fbcApplicationName,
			Source: appservice.ComponentSource{
				ComponentSourceUnion: appservice.ComponentSourceUnion{
					GitSource: &appservice.GitSource{
						URL: fbcSourceGitUrl,
					},
				},
			},
			TargetPort: 50051,
		}
		component, err = dev_fw.AsKubeDeveloper.HasController.CreateComponent(componentObj, dev_fw.UserNamespace, "", "", fbcApplicationName, false, map[string]string{})
		Expect(err).ShouldNot(HaveOccurred())

		_, err = managed_fw.AsKubeDeveloper.ReleaseController.CreateReleaseStrategy(fbcReleaseStrategyName, managedNamespace, "fbc-release", "quay.io/hacbs-release/pipeline-fbc-release:main", fbcEnterpriseContractPolicyName, fbcServiceAccountName, []releaseApi.Params{
			{Name: "fromIndex", Value: "quay.io/scoheb/fbc-index-testing:latest"},
			{Name: "targetIndex", Value: "quay.io/scoheb/fbc-target-index-testing:latest"},
			{Name: "binaryImage", Value: "registry.redhat.io/openshift4/ose-operator-registry:v4.12"},
			{Name: "requestUpdateTimeout", Value: "420"},
			{Name: "buildTimeoutSeconds", Value: "480"},
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = dev_fw.AsKubeDeveloper.ReleaseController.CreateReleasePlan(fbcReleasePlanName, devNamespace, fbcApplicationName, managedNamespace, "true")
		Expect(err).NotTo(HaveOccurred())

		_, err = managed_fw.AsKubeDeveloper.ReleaseController.CreateReleasePlanAdmission(fbcReleasePlanAdmissionName, devNamespace, fbcApplicationName, managedNamespace, "", "", fbcReleaseStrategyName)
		Expect(err).NotTo(HaveOccurred())

		defaultEcPolicySpec := ecp.EnterpriseContractPolicySpec{
			Description: "Red Hat's enterprise requirements",
			PublicKey:   "k8s://openshift-pipelines/public-key",
			Sources: []ecp.Source{{
				Name:   "Default",
				Policy: []string{"github.com/enterprise-contract/ec-policies//policy/lib", "github.com/enterprise-contract/ec-policies//policy/release"},
				Data:   []string{"github.com/enterprise-contract/ec-policies//data"},
			}},
			Configuration: &ecp.EnterpriseContractPolicyConfiguration{
				Collections: []string{"minimal"},
				Exclude:     []string{"cve", "step_image_registries"},
				Include:     []string{"@slsa1", "@slsa2", "@slsa3"},
			},
		}

		_, err = managed_fw.AsKubeDeveloper.TektonController.CreateEnterpriseContractPolicy(fbcEnterpriseContractPolicyName, managedNamespace, defaultEcPolicySpec)
		Expect(err).NotTo(HaveOccurred())

	})

	AfterAll(func() {
		if !CurrentSpecReport().Failed() {
			Expect(dev_fw.AsKubeDeveloper.HasController.DeleteApplication(fbcApplicationName, devNamespace, false)).NotTo(HaveOccurred())
			Expect(managed_fw.AsKubeDeveloper.TektonController.DeleteEnterpriseContractPolicy(fbcEnterpriseContractPolicyName, managedNamespace, false)).NotTo(HaveOccurred())
			Expect(managed_fw.AsKubeDeveloper.ReleaseController.DeleteReleaseStrategy(fbcReleaseStrategyName, managedNamespace, false)).NotTo(HaveOccurred())
			Expect(managed_fw.AsKubeDeveloper.ReleaseController.DeleteReleasePlanAdmission(fbcReleasePlanAdmissionName, managedNamespace, false)).NotTo(HaveOccurred())
		}
	})

	var _ = Describe("Post-release verification", func() {

		It("verifies that a build PipelineRun is created in dev namespace and succeeds", func() {
			Expect(dev_fw.AsKubeDeveloper.HasController.WaitForComponentPipelineToBeFinished(component, "", 2, dev_fw.AsKubeDeveloper.TektonController)).To(Succeed())
			buildPr, err = dev_fw.AsKubeDeveloper.HasController.GetComponentPipelineRun(component.Name, fbcApplicationName, devNamespace, "")
			Expect(err).ShouldNot(HaveOccurred())
			snapshot, err = dev_fw.AsKubeDeveloper.IntegrationController.GetSnapshot("", buildPr.Name, component.Name, devNamespace)
			Expect(err).ShouldNot(HaveOccurred())
			releaseCR, err = dev_fw.AsKubeDeveloper.ReleaseController.GetRelease("", snapshot.Name, devNamespace)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("verifies the fbc release pipelinerun is running and succeeds", func() {
			Eventually(func() error {
				releasePr, err = managed_fw.AsKubeAdmin.ReleaseController.GetPipelineRunInNamespace(managed_fw.UserNamespace, releaseCR.GetName(), releaseCR.GetNamespace())
				Expect(err).ShouldNot(HaveOccurred())
				if !releasePr.IsDone() {
					return fmt.Errorf("release pipelinerun %s in namespace %s did not finish yet", releasePr.Name, releasePr.Namespace)
				}
				Expect(utils.HasPipelineRunSucceeded(releasePr)).To(BeTrue(), fmt.Sprintf("release pipelinerun %s/%s did not succeed", releasePr.GetNamespace(), releasePr.GetName()))
				return nil
			}, releasePipelineRunCompletionTimeout, defaultInterval).Should(Succeed(), fmt.Sprint("timed out when waiting for release pipelinerun to succeed"))
		})

		It("verifies release CR completed and set succeeded.", func() {
			Eventually(func() error {
				releaseCR, err = dev_fw.AsKubeDeveloper.ReleaseController.GetFirstReleaseInNamespace(dev_fw.UserNamespace)
				if err != nil {
					return err
				}
				if !releaseCR.IsReleased() {
					return fmt.Errorf("release %s/%s is not marked as finished yet", releaseCR.GetNamespace(), releaseCR.GetName())
				}
				return nil
			}, releaseCreationTimeout, defaultInterval).Should(Succeed())
		})

	})
})
