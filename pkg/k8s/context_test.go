package k8s

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// writeKubeconfig writes a clientcmdapi.Config to a temp file and sets the
// KUBECONFIG env var so client-go picks it up.
func writeKubeconfig(t *testing.T, cfg *clientcmdapi.Config) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	require.NoError(t, clientcmd.WriteToFile(*cfg, path))
	t.Setenv("KUBECONFIG", path)
	return path
}

func TestFindContextForCluster_NotFound(t *testing.T) {
	cfg := clientcmdapi.NewConfig()
	cfg.Clusters["other"] = &clientcmdapi.Cluster{Server: "https://other.example.com"}
	cfg.Contexts["other-ctx"] = &clientcmdapi.Context{Cluster: "other"}
	cfg.CurrentContext = "other-ctx"
	writeKubeconfig(t, cfg)

	name, ok := FindContextForCluster("arn:aws:eks:us-east-1:123456789012:cluster/my-cluster")
	assert.False(t, ok)
	assert.Empty(t, name)
}

func TestFindContextForCluster_FoundWithValidCreds(t *testing.T) {
	// Start a fake API server that responds to /version
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := version.Info{Major: "1", Minor: "30", GitVersion: "v1.30.0"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	clusterARN := "arn:aws:eks:us-east-1:123456789012:cluster/my-cluster"
	cfg := clientcmdapi.NewConfig()
	cfg.Clusters[clusterARN] = &clientcmdapi.Cluster{
		Server:                   srv.URL,
		CertificateAuthorityData: nil,
		InsecureSkipTLSVerify:    true,
	}
	cfg.Contexts["my-ctx"] = &clientcmdapi.Context{Cluster: clusterARN}
	cfg.CurrentContext = "my-ctx"
	writeKubeconfig(t, cfg)

	name, ok := FindContextForCluster(clusterARN)
	assert.True(t, ok)
	assert.Equal(t, "my-ctx", name)
}

func TestFindContextForCluster_ContextExistsButServerUnreachable(t *testing.T) {
	clusterARN := "arn:aws:eks:us-west-2:111111111111:cluster/dead-cluster"
	cfg := clientcmdapi.NewConfig()
	cfg.Clusters[clusterARN] = &clientcmdapi.Cluster{
		Server:                "https://127.0.0.1:1", // nothing listening
		InsecureSkipTLSVerify: true,
	}
	cfg.Contexts["dead-ctx"] = &clientcmdapi.Context{Cluster: clusterARN}
	cfg.CurrentContext = "dead-ctx"
	writeKubeconfig(t, cfg)

	name, ok := FindContextForCluster(clusterARN)
	assert.False(t, ok)
	// Context was found in config, so name is returned even though creds are invalid
	assert.Equal(t, "dead-ctx", name)
}

func TestUseContext_Switches(t *testing.T) {
	cfg := clientcmdapi.NewConfig()
	cfg.Clusters["c1"] = &clientcmdapi.Cluster{Server: "https://c1.example.com"}
	cfg.Clusters["c2"] = &clientcmdapi.Cluster{Server: "https://c2.example.com"}
	cfg.Contexts["ctx1"] = &clientcmdapi.Context{Cluster: "c1"}
	cfg.Contexts["ctx2"] = &clientcmdapi.Context{Cluster: "c2"}
	cfg.CurrentContext = "ctx1"
	path := writeKubeconfig(t, cfg)

	err := UseContext("ctx2")
	require.NoError(t, err)

	// Read back and verify
	updated, err := clientcmd.LoadFromFile(path)
	require.NoError(t, err)
	assert.Equal(t, "ctx2", updated.CurrentContext)
}

func TestUseContext_NotFound(t *testing.T) {
	cfg := clientcmdapi.NewConfig()
	cfg.Contexts["ctx1"] = &clientcmdapi.Context{Cluster: "c1"}
	cfg.CurrentContext = "ctx1"
	writeKubeconfig(t, cfg)

	err := UseContext("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestUseContext_EmptyConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	require.NoError(t, os.WriteFile(path, []byte(""), 0600))
	t.Setenv("KUBECONFIG", path)

	err := UseContext("anything")
	assert.Error(t, err)
}
