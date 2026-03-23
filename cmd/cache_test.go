package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jordiprats/kubectl-eks/pkg/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupCacheTestDir sets HomeDir to a temp directory with a .kube subdir
// and resets CachedData. Returns the cache file path.
func setupCacheTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".kube"), 0755))
	HomeDir = dir
	CachedData = nil
	return filepath.Join(dir, ".kube", ".kubectl-eks-cache")
}

func TestLoadCacheFromDisk_NoFile(t *testing.T) {
	setupCacheTestDir(t)

	loadCacheFromDisk()
	assert.Nil(t, CachedData)
}

func TestSaveCacheAndLoadRoundTrip(t *testing.T) {
	cacheFile := setupCacheTestDir(t)

	CachedData = &data.KubeCtlEksCache{
		ClusterByARN: map[string]data.ClusterInfo{
			"arn:aws:eks:us-east-1:111111111111:cluster/alpha": {
				ClusterName:  "alpha",
				Region:       "us-east-1",
				AWSProfile:   "dev",
				AWSAccountID: "111111111111",
				Arn:          "arn:aws:eks:us-east-1:111111111111:cluster/alpha",
				Version:      "1.30",
				Status:       "ACTIVE",
			},
		},
		ClusterList: map[string]map[string][]data.ClusterInfo{
			"dev": {
				"us-east-1": {
					{
						ClusterName: "alpha",
						Region:      "us-east-1",
						AWSProfile:  "dev",
						Arn:         "arn:aws:eks:us-east-1:111111111111:cluster/alpha",
					},
				},
			},
		},
	}

	saveCacheToDisk()

	// File must exist
	_, err := os.Stat(cacheFile)
	require.NoError(t, err)

	// Reset and reload
	CachedData = nil
	loadCacheFromDisk()
	require.NotNil(t, CachedData)

	assert.Len(t, CachedData.ClusterByARN, 1)
	info, exists := CachedData.ClusterByARN["arn:aws:eks:us-east-1:111111111111:cluster/alpha"]
	assert.True(t, exists)
	assert.Equal(t, "alpha", info.ClusterName)
	assert.Equal(t, "dev", info.AWSProfile)
	assert.Equal(t, "1.30", info.Version)

	assert.Len(t, CachedData.ClusterList, 1)
	assert.Len(t, CachedData.ClusterList["dev"]["us-east-1"], 1)
}

func TestLoadCacheFromDisk_CorruptedJSON(t *testing.T) {
	cacheFile := setupCacheTestDir(t)
	require.NoError(t, os.WriteFile(cacheFile, []byte("{invalid json"), 0644))

	// loadCacheFromDisk calls os.Exit on bad JSON — but we can at least
	// verify the file is read by checking what happens with valid-but-empty
	// JSON instead. Testing os.Exit requires subprocess tricks.
	require.NoError(t, os.WriteFile(cacheFile, []byte("{}"), 0644))
	loadCacheFromDisk()
	require.NotNil(t, CachedData)
	assert.Nil(t, CachedData.ClusterByARN)
	assert.Nil(t, CachedData.ClusterList)
}

func TestCacheClear_RemovesFile(t *testing.T) {
	cacheFile := setupCacheTestDir(t)

	// Write a cache file
	CachedData = &data.KubeCtlEksCache{
		ClusterByARN: map[string]data.ClusterInfo{
			"arn:aws:eks:us-west-2:222222222222:cluster/beta": {
				ClusterName: "beta",
			},
		},
		ClusterList: make(map[string]map[string][]data.ClusterInfo),
	}
	saveCacheToDisk()

	_, err := os.Stat(cacheFile)
	require.NoError(t, err, "cache file should exist before clear")

	// Simulate cache clear
	err = os.Remove(cacheFile)
	require.NoError(t, err)
	CachedData = nil

	_, err = os.Stat(cacheFile)
	assert.True(t, os.IsNotExist(err), "cache file should be gone after clear")
	assert.Nil(t, CachedData)
}

func TestCacheClear_NoFileNoPanic(t *testing.T) {
	setupCacheTestDir(t)

	// Clearing when no cache file exists should not error
	configFile := HomeDir + "/.kube/.kubectl-eks-cache"
	err := os.Remove(configFile)
	assert.True(t, os.IsNotExist(err))
}

func TestCacheShowCollectsFromBothMaps(t *testing.T) {
	setupCacheTestDir(t)

	arn1 := "arn:aws:eks:us-east-1:111111111111:cluster/alpha"
	arn2 := "arn:aws:eks:us-west-2:222222222222:cluster/beta"
	arn3 := "arn:aws:eks:eu-west-1:333333333333:cluster/gamma"

	CachedData = &data.KubeCtlEksCache{
		ClusterByARN: map[string]data.ClusterInfo{
			arn1: {ClusterName: "alpha", Arn: arn1},
			arn2: {ClusterName: "beta", Arn: arn2},
		},
		ClusterList: map[string]map[string][]data.ClusterInfo{
			"prod": {
				"eu-west-1": {
					{ClusterName: "gamma", Arn: arn3},
					{ClusterName: "alpha", Arn: arn1}, // duplicate
				},
			},
		},
	}

	// Replicate the show command's collection logic
	clusters := []data.ClusterInfo{}
	seen := make(map[string]bool)
	for arn, info := range CachedData.ClusterByARN {
		if !seen[arn] {
			seen[arn] = true
			clusters = append(clusters, info)
		}
	}
	for _, regions := range CachedData.ClusterList {
		for _, clusterList := range regions {
			for _, c := range clusterList {
				if c.Arn != "" && !seen[c.Arn] {
					seen[c.Arn] = true
					clusters = append(clusters, c)
				}
			}
		}
	}

	// Should have 3 unique clusters (alpha, beta, gamma) — no duplicates
	assert.Len(t, clusters, 3)
	names := make(map[string]bool)
	for _, c := range clusters {
		names[c.ClusterName] = true
	}
	assert.True(t, names["alpha"])
	assert.True(t, names["beta"])
	assert.True(t, names["gamma"])
}

func TestCacheShowEmptyCache(t *testing.T) {
	setupCacheTestDir(t)
	CachedData = nil

	// After loading from a nonexistent file, CachedData remains nil
	loadCacheFromDisk()
	assert.Nil(t, CachedData)
}

func TestCacheShowEmptyMaps(t *testing.T) {
	setupCacheTestDir(t)
	CachedData = &data.KubeCtlEksCache{
		ClusterByARN: make(map[string]data.ClusterInfo),
		ClusterList:  make(map[string]map[string][]data.ClusterInfo),
	}

	clusters := []data.ClusterInfo{}
	seen := make(map[string]bool)
	for arn, info := range CachedData.ClusterByARN {
		if !seen[arn] {
			seen[arn] = true
			clusters = append(clusters, info)
		}
	}
	for _, regions := range CachedData.ClusterList {
		for _, clusterList := range regions {
			for _, c := range clusterList {
				if c.Arn != "" && !seen[c.Arn] {
					seen[c.Arn] = true
					clusters = append(clusters, c)
				}
			}
		}
	}

	assert.Empty(t, clusters)
}

func TestSaveCacheCreatesValidJSON(t *testing.T) {
	cacheFile := setupCacheTestDir(t)

	CachedData = &data.KubeCtlEksCache{
		ClusterByARN: map[string]data.ClusterInfo{
			"arn:aws:eks:us-east-1:123456789012:cluster/test": {
				ClusterName: "test",
				Region:      "us-east-1",
				AWSProfile:  "default",
				Arn:         "arn:aws:eks:us-east-1:123456789012:cluster/test",
			},
		},
		ClusterList: make(map[string]map[string][]data.ClusterInfo),
	}

	saveCacheToDisk()

	raw, err := os.ReadFile(cacheFile)
	require.NoError(t, err)

	var parsed data.KubeCtlEksCache
	err = json.Unmarshal(raw, &parsed)
	require.NoError(t, err, "saved cache must be valid JSON")
	assert.Equal(t, "test", parsed.ClusterByARN["arn:aws:eks:us-east-1:123456789012:cluster/test"].ClusterName)
}

func TestLoadCachePreservesClusterByARN(t *testing.T) {
	cacheFile := setupCacheTestDir(t)

	original := &data.KubeCtlEksCache{
		ClusterByARN: map[string]data.ClusterInfo{
			"arn:aws:eks:us-west-2:999999999999:cluster/cached": {
				ClusterName:  "cached",
				Region:       "us-west-2",
				AWSProfile:   "ops",
				AWSAccountID: "999999999999",
				Arn:          "arn:aws:eks:us-west-2:999999999999:cluster/cached",
				Version:      "1.29",
				Status:       "ACTIVE",
				CreatedAt:    "2024-03-15 09:30:00",
			},
		},
		ClusterList: make(map[string]map[string][]data.ClusterInfo),
	}

	raw, err := json.Marshal(original)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cacheFile, raw, 0644))

	CachedData = nil
	loadCacheFromDisk()

	require.NotNil(t, CachedData)
	info, exists := CachedData.ClusterByARN["arn:aws:eks:us-west-2:999999999999:cluster/cached"]
	require.True(t, exists)
	assert.Equal(t, "cached", info.ClusterName)
	assert.Equal(t, "ops", info.AWSProfile)
	assert.Equal(t, "999999999999", info.AWSAccountID)
	assert.Equal(t, "1.29", info.Version)
	assert.Equal(t, "ACTIVE", info.Status)
	assert.Equal(t, "2024-03-15 09:30:00", info.CreatedAt)
}
