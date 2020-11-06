package kubernetes

import (
	"context"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"

	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/service"
	"github.com/openservicemesh/osm/pkg/tests"
)

var (
	testMeshName = "mesh"
)

const (
	nsInformerSyncTimeout = 3 * time.Second
)

var _ = Describe("Test Namespace KubeController Methods", func() {
	Context("Testing namespace controller", func() {
		It("should return a new namespace controller", func() {
			kubeClient := testclient.NewSimpleClientset()
			stop := make(chan struct{})
			kubeController, err := NewKubernetesController(kubeClient, testMeshName, stop)
			Expect(err).ToNot(HaveOccurred())
			Expect(kubeController).ToNot(BeNil())
		})
	})

	Context("Testing ListMonitoredNamespaces", func() {
		It("should return monitored namespaces", func() {
			// Create namespace controller
			kubeClient := testclient.NewSimpleClientset()
			stop := make(chan struct{})
			kubeController, err := NewKubernetesController(kubeClient, testMeshName, stop)
			Expect(err).ToNot(HaveOccurred())
			Expect(kubeController).ToNot(BeNil())

			// Create a test namespace that is monitored
			testNamespaceName := fmt.Sprintf("%s-1", tests.Namespace)
			testNamespace := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   testNamespaceName,
					Labels: map[string]string{constants.OSMKubeResourceMonitorAnnotation: testMeshName},
				},
			}
			_, err = kubeClient.CoreV1().Namespaces().Create(context.TODO(), &testNamespace, metav1.CreateOptions{})
			Expect(err).To(BeNil())

			// Eventually asserts that all return values apart from the first value are nil or zero-valued,
			// so asserting that an error is nil is implicit.
			Eventually(func() ([]string, error) {
				return kubeController.ListMonitoredNamespaces()
			}, nsInformerSyncTimeout).Should(Equal([]string{testNamespaceName}))
		})
	})

	Context("Testing GetNamespace", func() {
		It("should return existing namespace if it exists", func() {
			// Create namespace controller
			kubeClient := testclient.NewSimpleClientset()
			stop := make(chan struct{})
			kubeController, err := NewKubernetesController(kubeClient, testMeshName, stop)
			Expect(err).ToNot(HaveOccurred())
			Expect(kubeController).ToNot(BeNil())

			// Create a test namespace that is monitored
			testNamespaceName := fmt.Sprintf("%s-1", tests.Namespace)
			testNamespace := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   testNamespaceName,
					Labels: map[string]string{constants.OSMKubeResourceMonitorAnnotation: testMeshName},
				},
			}

			// Create it
			nsCreate, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), &testNamespace, metav1.CreateOptions{})
			Expect(err).To(BeNil())

			// Check it is present
			Eventually(func() *corev1.Namespace {
				return kubeController.GetNamespace(testNamespaceName)
			}, nsInformerSyncTimeout).Should(Equal(nsCreate))

			// Delete it
			err = kubeClient.CoreV1().Namespaces().Delete(context.TODO(), testNamespaceName, metav1.DeleteOptions{})
			Expect(err).To(BeNil())

			// Check it is gone
			Eventually(func() *corev1.Namespace {
				return kubeController.GetNamespace(testNamespaceName)
			}, nsInformerSyncTimeout).Should(BeNil())
		})
	})

	Context("Testing IsMonitoredNamespace", func() {
		It("should work as expected", func() {
			// Create namespace controller
			kubeClient := testclient.NewSimpleClientset()
			stop := make(chan struct{})
			kubeController, err := NewKubernetesController(kubeClient, testMeshName, stop)
			Expect(err).ToNot(HaveOccurred())
			Expect(kubeController).ToNot(BeNil())

			// Create a test namespace that is monitored
			testNamespaceName := fmt.Sprintf("%s-1", tests.Namespace)
			testNamespace := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   testNamespaceName,
					Labels: map[string]string{constants.OSMKubeResourceMonitorAnnotation: testMeshName},
				},
			}

			_, err = kubeClient.CoreV1().Namespaces().Create(context.TODO(), &testNamespace, metav1.CreateOptions{})
			Expect(err).To(BeNil())

			Eventually(func() bool {
				return kubeController.IsMonitoredNamespace(testNamespaceName)
			}, nsInformerSyncTimeout).Should(BeTrue())

			fakeNamespaceIsMonitored := kubeController.IsMonitoredNamespace("fake")
			Expect(fakeNamespaceIsMonitored).ToNot(BeTrue())
		})
	})

	Context("service controller", func() {
		var kubeClient *testclient.Clientset
		var kubeController Controller
		var err error

		BeforeEach(func() {
			kubeClient = testclient.NewSimpleClientset()
			kubeController, err = NewKubernetesController(kubeClient, testMeshName, make(chan struct{}))
			Expect(err).ToNot(HaveOccurred())
			Expect(kubeController).ToNot(BeNil())
		})

		It("should create and delete services, and be detected if NS is monitored", func() {
			meshSvc := tests.BookbuyerService

			// Create monitored namespace for this service
			testNamespace := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   tests.BookbuyerService.Namespace,
					Labels: map[string]string{constants.OSMKubeResourceMonitorAnnotation: testMeshName},
				},
			}
			_, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), &testNamespace, metav1.CreateOptions{})
			Expect(err).To(BeNil())

			svc := tests.NewServiceFixture(meshSvc.Name, meshSvc.Namespace, nil)
			_, err = kubeClient.CoreV1().Services(meshSvc.Namespace).Create(context.TODO(), svc, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())
			<-kubeController.GetAnnouncementsChannel(Services)

			svcIncache := kubeController.GetService(meshSvc)
			Expect(svcIncache).To(Equal(svc))

			err = kubeClient.CoreV1().Services(meshSvc.Namespace).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{})
			Expect(err).ToNot(HaveOccurred())
			<-kubeController.GetAnnouncementsChannel(Services)

			svcIncache = kubeController.GetService(meshSvc)
			Expect(svcIncache).To(BeNil())
		})

		It("should return nil when the given MeshService is not found", func() {
			meshSvc := tests.BookbuyerService

			svcIncache := kubeController.GetService(meshSvc)
			Expect(svcIncache).To(BeNil())
		})

		It("should return an empty list when no services are found", func() {
			services := kubeController.ListServices()
			Expect(len(services)).To(Equal(0))
		})

		It("should return a list of Services", func() {
			// Define services to test with
			testSvcs := []service.MeshService{
				tests.BookbuyerService,
				tests.BookstoreV1Service,
				tests.BookstoreV2Service,
				tests.BookwarehouseService,
			}

			// Test services could belong to the same namespace, so ensure we create a list of unique namespaces
			testNamespaces := mapset.NewSet()
			for _, svc := range testSvcs {
				testNamespaces.Add(svc.Namespace)
			}

			// Add namespace if doesn't exist
			for ns := range testNamespaces.Iter() {
				namespace := ns.(string)

				if kubeController.IsMonitoredNamespace(namespace) {
					continue
				}

				testNamespace := corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name:   namespace,
						Labels: map[string]string{constants.OSMKubeResourceMonitorAnnotation: testMeshName},
					},
				}
				_, err = kubeClient.CoreV1().Namespaces().Create(context.TODO(), &testNamespace, metav1.CreateOptions{})
				Expect(err).To(BeNil())
			}

			// Add services
			for _, svcAdd := range testSvcs {
				svcSpec := tests.NewServiceFixture(svcAdd.Name, svcAdd.Namespace, nil)
				_, err := kubeClient.CoreV1().Services(svcAdd.Namespace).Create(context.TODO(), svcSpec, metav1.CreateOptions{})
				Expect(err).ToNot(HaveOccurred())
				<-kubeController.GetAnnouncementsChannel(Services)
			}

			services := kubeController.ListServices()
			Expect(len(testSvcs)).To(Equal(len(services)))

			// Remove services one by one, check each iteration
			for len(testSvcs) > 0 {
				svcRem := testSvcs[0]
				err := kubeClient.CoreV1().Services(svcRem.Namespace).Delete(context.TODO(), svcRem.Name, metav1.DeleteOptions{})
				Expect(err).ToNot(HaveOccurred())
				<-kubeController.GetAnnouncementsChannel(Services)

				testSvcs = testSvcs[1:]

				services := kubeController.ListServices()
				Expect(len(testSvcs)).To(Equal(len(services)))
			}
		})
	})
})
