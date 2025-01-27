// Copyright (c) 2021 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package tests

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/open-cluster-management/observability-e2e-test/pkg/utils"
)

const (
	MCO_OPERATOR_NAMESPACE = "open-cluster-management"
	MCO_CR_NAME            = "observability"
	MCO_NAMESPACE          = "open-cluster-management-observability"
	MCO_ADDON_NAMESPACE    = "open-cluster-management-addon-observability"
	MCO_LABEL              = "name=multicluster-observability-operator"
	MCO_LABEL_OWNER        = "owner=multicluster-observability-operator"
)

var (
	EventuallyTimeoutMinute  time.Duration = 60 * time.Second
	EventuallyIntervalSecond time.Duration = 1 * time.Second

	hubClient kubernetes.Interface
	dynClient dynamic.Interface
	err       error
)

var _ = Describe("Observability:", func() {

	BeforeEach(func() {
		hubClient = utils.NewKubeClient(
			testOptions.HubCluster.MasterURL,
			testOptions.KubeConfig,
			testOptions.HubCluster.KubeContext)

		dynClient = utils.NewKubeClientDynamic(
			testOptions.HubCluster.MasterURL,
			testOptions.KubeConfig,
			testOptions.HubCluster.KubeContext)
	})

	It("[P2][Sev2][Observability][Stable] Modifying MCO CR for reconciling (reconcile/g0)", func() {
		By("Modifying MCO CR for reconciling")
		err := utils.ModifyMCOCR(testOptions)
		Expect(err).ToNot(HaveOccurred())

		By("Waiting for MCO retentionResolutionRaw filed to take effect")
		Eventually(func() error {
			name := MCO_CR_NAME + "-thanos-compact"
			compact, err := hubClient.AppsV1().StatefulSets(MCO_NAMESPACE).Get(name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			argList := compact.Spec.Template.Spec.Containers[0].Args
			for _, arg := range argList {
				if arg == "--retention.resolution-raw=3d" {
					return nil
				}
			}
			return fmt.Errorf("Failed to find modified retention field")
		}, EventuallyTimeoutMinute*5, EventuallyIntervalSecond*5).Should(Succeed())

		By("Wait for thanos compact pods are ready")
		// ensure the thanos rule pods are restarted successfully before processing
		Eventually(func() error {
			err = utils.CheckStatefulSetPodReady(testOptions, MCO_CR_NAME+"-thanos-compact", 1)
			if err != nil {
				return err
			}
			return nil
		}, EventuallyTimeoutMinute*10, EventuallyIntervalSecond*5).Should(Succeed())

		By("Wait for alertmanager pods are ready")
		// ensure the thanos rule pods are restarted successfully before processing
		Eventually(func() error {
			err = utils.CheckStatefulSetPodReady(testOptions, MCO_CR_NAME+"-alertmanager", 3)
			if err != nil {
				return err
			}
			return nil
		}, EventuallyTimeoutMinute*10, EventuallyIntervalSecond*5).Should(Succeed())
	})

	It("[P2][Sev2][Observability][Stable] Checking node selector for all pods (reconcile/g0)", func() {
		By("Checking node selector for all pods")
		Eventually(func() error {
			err = utils.CheckAllPodNodeSelector(testOptions)
			if err != nil {
				return err
			}
			return nil
		}, EventuallyTimeoutMinute*5, EventuallyIntervalSecond*5).Should(Succeed())
	})

	It("[P2][Sev2][Observability][Stable] Checking podAntiAffinity for all pods (reconcile/g0)", func() {
		By("Checking podAntiAffinity for all pods")
		Eventually(func() error {
			err := utils.CheckAllPodsAffinity(testOptions)
			if err != nil {
				return err
			}
			return nil
		}, EventuallyTimeoutMinute*5, EventuallyIntervalSecond*5).Should(Succeed())
	})

	It("[P2][Sev2][Observability][Stable] Checking alertmanager storage resize (reconcile/g0)", func() {
		By("Resizing alertmanager storage")
		Eventually(func() error {
			err := utils.CheckStorageResize(testOptions, MCO_CR_NAME+"-alertmanager", "2Gi")
			if err != nil {
				return err
			}
			return nil
		}, EventuallyTimeoutMinute*5, EventuallyIntervalSecond*5).Should(Succeed())
	})

	It("[P2][Sev2][Observability][Stable] Customize the replicas for thanos query (reconcile/g0)", func() {
		Eventually(func() error {
			err := utils.UpdateDeploymentReplicas(testOptions, MCO_CR_NAME+"-thanos-query", "query", 3, 3)
			if err != nil {
				return err
			}
			return nil
		}, EventuallyTimeoutMinute*5, EventuallyIntervalSecond*5).Should(Succeed())

		Eventually(func() error {
			err := utils.UpdateDeploymentReplicas(testOptions, MCO_CR_NAME+"-thanos-query", "query", 0, 3)
			if err != nil {
				return err
			}
			return nil
		}, EventuallyTimeoutMinute*5, EventuallyIntervalSecond*5).Should(Succeed())

		Eventually(func() error {
			err := utils.UpdateDeploymentReplicas(testOptions, MCO_CR_NAME+"-thanos-query", "query", 2, 2)
			if err != nil {
				return err
			}
			return nil
		}, EventuallyTimeoutMinute*5, EventuallyIntervalSecond*5).Should(Succeed())

		By("Wait for thanos query pods are ready")
		Eventually(func() error {
			err = utils.CheckDeploymentPodReady(testOptions, MCO_CR_NAME+"-thanos-query", 2)
			if err != nil {
				return err
			}
			return nil
		}, EventuallyTimeoutMinute*10, EventuallyIntervalSecond*5).Should(Succeed())

	})

	It("[P2][Sev2][Observability][Stable] Revert MCO CR changes (reconcile/g0)", func() {

		By("Revert MCO CR changes")
		err := utils.RevertMCOCRModification(testOptions)
		Expect(err).ToNot(HaveOccurred())

		By("Waiting for MCO retentionResolutionRaw filed to take effect")
		Eventually(func() error {
			name := MCO_CR_NAME + "-thanos-compact"
			compact, err := hubClient.AppsV1().StatefulSets(MCO_NAMESPACE).Get(name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			argList := compact.Spec.Template.Spec.Containers[0].Args
			for _, arg := range argList {
				if arg == "--retention.resolution-raw=5d" {
					return nil
				}
			}
			return fmt.Errorf("Failed to find modified retention field")
		}, EventuallyTimeoutMinute*5, EventuallyIntervalSecond*5).Should(Succeed())

		By("Wait for thanos compact pods are ready")
		// ensure the thanos rule pods are restarted successfully before processing
		Eventually(func() error {
			err = utils.CheckStatefulSetPodReady(testOptions, MCO_CR_NAME+"-thanos-compact", 1)
			if err != nil {
				return err
			}
			return nil
		}, EventuallyTimeoutMinute*10, EventuallyIntervalSecond*5).Should(Succeed())

		By("Checking MCO components in default HA mode")
		Eventually(func() error {
			err = utils.CheckMCOComponentsInHighMode(testOptions)
			if err != nil {
				return err
			}
			return nil
		}, EventuallyTimeoutMinute*15, EventuallyIntervalSecond*5).Should(Succeed())
	})

	AfterEach(func() {
		testFailed = testFailed || CurrentGinkgoTestDescription().Failed
		if testFailed {
			utils.PrintMCOObject(testOptions)
			utils.PrintAllMCOPodsStatus(testOptions)
			utils.PrintAllOBAPodsStatus(testOptions)
		} else {
			Expect(utils.IntegrityChecking(testOptions)).NotTo(HaveOccurred())
		}
	})
})
