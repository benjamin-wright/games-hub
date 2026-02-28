//go:build integration

package controller_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/benjamin-wright/games-hub/apps/platform/db-operator/internal/api/v1alpha1"
)

var _ = Describe("PostgresDatabaseReconciler", func() {
	var (
		ns     *corev1.Namespace
		pgdb   *v1alpha1.PostgresDatabase
		lookup types.NamespacedName
	)

	BeforeEach(func() {
		// Create a unique namespace for each test.
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pgdb-",
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed())

		pgdb = &v1alpha1.PostgresDatabase{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-db",
				Namespace: ns.Name,
			},
			Spec: v1alpha1.PostgresDatabaseSpec{
				DatabaseName:    "mydb",
				PostgresVersion: "16",
				StorageSize:     resource.MustParse("256Mi"),
			},
		}
		lookup = types.NamespacedName{Name: pgdb.Name, Namespace: ns.Name}
	})

	AfterEach(func() {
		// Cleanup: delete the namespace (cascades all resources).
		_ = k8sClient.Delete(ctx, ns)
	})

	It("should create a StatefulSet and headless Service for a new PostgresDatabase", func() {
		Expect(k8sClient.Create(ctx, pgdb)).To(Succeed())

		// Verify the headless Service is created.
		Eventually(func(g Gomega) {
			var svc corev1.Service
			g.Expect(k8sClient.Get(ctx, lookup, &svc)).To(Succeed())
			g.Expect(svc.Spec.ClusterIP).To(Equal(corev1.ClusterIPNone))
			g.Expect(svc.Spec.Ports).To(HaveLen(1))
			g.Expect(svc.Spec.Ports[0].Port).To(Equal(int32(5432)))
		}, timeout, interval).Should(Succeed())

		// Verify the StatefulSet is created with the right image and PVC.
		Eventually(func(g Gomega) {
			var sts appsv1.StatefulSet
			g.Expect(k8sClient.Get(ctx, lookup, &sts)).To(Succeed())
			g.Expect(sts.Spec.Template.Spec.Containers).To(HaveLen(1))
			g.Expect(sts.Spec.Template.Spec.Containers[0].Image).To(Equal("postgres:16"))
			g.Expect(sts.Spec.VolumeClaimTemplates).To(HaveLen(1))
			g.Expect(*sts.Spec.Replicas).To(Equal(int32(1)))
		}, timeout, interval).Should(Succeed())

		// Verify owner references are set on both resources.
		Eventually(func(g Gomega) {
			var svc corev1.Service
			g.Expect(k8sClient.Get(ctx, lookup, &svc)).To(Succeed())
			g.Expect(svc.OwnerReferences).To(HaveLen(1))
			g.Expect(svc.OwnerReferences[0].Name).To(Equal(pgdb.Name))

			var sts appsv1.StatefulSet
			g.Expect(k8sClient.Get(ctx, lookup, &sts)).To(Succeed())
			g.Expect(sts.OwnerReferences).To(HaveLen(1))
			g.Expect(sts.OwnerReferences[0].Name).To(Equal(pgdb.Name))
		}, timeout, interval).Should(Succeed())

		// The finalizer must be present on the CR.
		Eventually(func(g Gomega) {
			var fetched v1alpha1.PostgresDatabase
			g.Expect(k8sClient.Get(ctx, lookup, &fetched)).To(Succeed())
			g.Expect(fetched.Finalizers).To(ContainElement("games-hub.io/postgres-database"))
		}, timeout, interval).Should(Succeed())
	})

	It("should initially set status phase to Pending before the StatefulSet is ready", func() {
		Expect(k8sClient.Create(ctx, pgdb)).To(Succeed())

		// The reconciler should report Pending immediately after creating the StatefulSet,
		// since the pod won't be ready yet.
		Eventually(func(g Gomega) {
			var fetched v1alpha1.PostgresDatabase
			g.Expect(k8sClient.Get(ctx, lookup, &fetched)).To(Succeed())
			g.Expect(fetched.Status.Phase).To(Equal(v1alpha1.DatabasePhasePending))
		}, timeout, interval).Should(Succeed())
	})

	It("should transition to Ready when the StatefulSet has ready replicas", func() {
		Expect(k8sClient.Create(ctx, pgdb)).To(Succeed())

		// On a real cluster the StatefulSet controller will schedule the pod and
		// it will become ready once Postgres starts. Wait for the phase to reflect that.
		Eventually(func(g Gomega) {
			var fetched v1alpha1.PostgresDatabase
			g.Expect(k8sClient.Get(ctx, lookup, &fetched)).To(Succeed())
			g.Expect(fetched.Status.Phase).To(Equal(v1alpha1.DatabasePhaseReady))
		}, timeout, interval).Should(Succeed())
	})

	It("should cascade-delete StatefulSet and Service when the CR is deleted", func() {
		Expect(k8sClient.Create(ctx, pgdb)).To(Succeed())

		// Wait for the owned resources to exist.
		Eventually(func(g Gomega) {
			var sts appsv1.StatefulSet
			g.Expect(k8sClient.Get(ctx, lookup, &sts)).To(Succeed())
			var svc corev1.Service
			g.Expect(k8sClient.Get(ctx, lookup, &svc)).To(Succeed())
		}, timeout, interval).Should(Succeed())

		// Delete the PostgresDatabase CR.
		Expect(k8sClient.Delete(ctx, pgdb)).To(Succeed())

		// The CR should be fully removed (finalizer handled).
		Eventually(func(g Gomega) {
			var fetched v1alpha1.PostgresDatabase
			err := k8sClient.Get(ctx, lookup, &fetched)
			g.Expect(err).To(HaveOccurred())
			g.Expect(client.IgnoreNotFound(err)).To(Succeed())
		}, timeout, interval).Should(Succeed())

		// StatefulSet should be gone.
		Eventually(func(g Gomega) {
			var sts appsv1.StatefulSet
			err := k8sClient.Get(ctx, lookup, &sts)
			g.Expect(err).To(HaveOccurred())
			g.Expect(client.IgnoreNotFound(err)).To(Succeed())
		}, timeout, interval).Should(Succeed())

		// Service should be gone.
		Eventually(func(g Gomega) {
			var svc corev1.Service
			err := k8sClient.Get(ctx, lookup, &svc)
			g.Expect(err).To(HaveOccurred())
			g.Expect(client.IgnoreNotFound(err)).To(Succeed())
		}, timeout, interval).Should(Succeed())
	})

	It("should leave no orphaned resources after deletion", func() {
		Expect(k8sClient.Create(ctx, pgdb)).To(Succeed())

		// Wait for owned resources.
		Eventually(func(g Gomega) {
			var sts appsv1.StatefulSet
			g.Expect(k8sClient.Get(ctx, lookup, &sts)).To(Succeed())
		}, timeout, interval).Should(Succeed())

		// Delete the CR.
		Expect(k8sClient.Delete(ctx, pgdb)).To(Succeed())

		// Wait for CR deletion to complete.
		Eventually(func(g Gomega) {
			var fetched v1alpha1.PostgresDatabase
			err := k8sClient.Get(ctx, lookup, &fetched)
			g.Expect(err).To(HaveOccurred())
			g.Expect(client.IgnoreNotFound(err)).To(Succeed())
		}, timeout, interval).Should(Succeed())

		// Verify no StatefulSets or Services with the operator-managed label remain.
		labels := client.MatchingLabels{
			"app.kubernetes.io/managed-by": "db-operator",
			"app.kubernetes.io/instance":   pgdb.Name,
		}

		var stsList appsv1.StatefulSetList
		Expect(k8sClient.List(ctx, &stsList, client.InNamespace(ns.Name), labels)).To(Succeed())
		Expect(stsList.Items).To(BeEmpty(), fmt.Sprintf("orphaned StatefulSets: %v", stsList.Items))

		var svcList corev1.ServiceList
		Expect(k8sClient.List(ctx, &svcList, client.InNamespace(ns.Name), labels)).To(Succeed())
		Expect(svcList.Items).To(BeEmpty(), fmt.Sprintf("orphaned Services: %v", svcList.Items))
	})

	It("should set the correct environment variables on the Postgres container", func() {
		Expect(k8sClient.Create(ctx, pgdb)).To(Succeed())

		Eventually(func(g Gomega) {
			var sts appsv1.StatefulSet
			g.Expect(k8sClient.Get(ctx, lookup, &sts)).To(Succeed())

			container := sts.Spec.Template.Spec.Containers[0]
			envMap := make(map[string]string)
			for _, e := range container.Env {
				envMap[e.Name] = e.Value
			}

			g.Expect(envMap["POSTGRES_DB"]).To(Equal("mydb"))
			g.Expect(envMap["POSTGRES_USER"]).To(Equal("postgres"))
			g.Expect(envMap["POSTGRES_PASSWORD"]).NotTo(BeEmpty())
		}, timeout, interval).Should(Succeed())
	})
})
