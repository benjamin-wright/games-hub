//go:build integration

package controller_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	v1alpha1 "github.com/benjamin-wright/games-hub/apps/platform/db-operator/internal/api/v1alpha1"
)

var (
	k8sClient client.Client
	ctx       context.Context
	cancel    context.CancelFunc
	scheme    = runtime.NewScheme()
)

const (
	timeout  = 60 * time.Second
	interval = 250 * time.Millisecond
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.Background())

	// Register schemes.
	Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
	Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

	// Resolve kubeconfig: prefer KUBECONFIG env, fall back to ~/.scratch/games-hub.yaml.
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		Expect(err).NotTo(HaveOccurred())
		kubeconfigPath = filepath.Join(home, ".scratch", "games-hub.yaml")
	}
	_, err := os.Stat(kubeconfigPath)
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("kubeconfig not found at %s — is the k3d cluster running?", kubeconfigPath))

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	Expect(err).NotTo(HaveOccurred())

	// Apply CRDs from the helm/crds directory so the test is self-contained.
	crdDir := filepath.Join("..", "..", "helm", "crds")
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", crdDir, "--kubeconfig", kubeconfigPath)
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	Expect(cmd.Run()).To(Succeed(), "failed to apply CRDs — is kubectl available?")

	// Create a direct client for test assertions.
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	cancel()
})
