//go:build integration
// +build integration

package executor_test

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/blessnetwork/b7s/executor"
	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/telemetry"
	"github.com/blessnetwork/b7s/testing/helpers"
	"github.com/blessnetwork/b7s/testing/mocks"
)

const (
	runtimeDirEnv     = "B7S_INTEG_RUNTIME_DIR"
	cleanupDisableEnv = "B7S_INTEG_CLEANUP_DISABLE"
)

func TestExecutor_Execute(t *testing.T) {

	const (
		dirPattern = "b7s-executor-integration-test-"

		testFunction = "./testdata/md5sum/md5sum.wasm"

		functionID = "function-id"
		requestID  = "dummy-request-id"

		chunkSize = 128
		fileSize  = 256
	)

	// Setup directories.
	workspace, err := os.MkdirTemp("", dirPattern)
	require.NoError(t, err)
	if !cleanupDisabled() {
		defer os.RemoveAll(workspace)
	}

	var (
		workdir     = filepath.Join(workspace, "t", requestID) // request work directory
		fsRoot      = filepath.Join(workdir, "fs")             // function FS root
		functiondir = filepath.Join(workspace, functionID)     // function location
		registry    = prometheus.NewRegistry()
	)

	sink, err := telemetry.CreateMetricSink(registry, telemetry.MetricsConfig{Counters: executor.Counters})
	require.NoError(t, err)

	metrics, err := telemetry.CreateMetrics(sink, false)
	require.NoError(t, err)

	t.Logf("working directory: %v", workspace)
	createDirs(t, workdir, fsRoot, functiondir)

	// Stage executable to working directory.
	copyFunction(t, testFunction, functiondir)

	// Create a random testfile.
	testfile, hash := createTestFile(t, fsRoot, fileSize)

	// Create executor.
	executor, err := executor.New(
		mocks.NoopLogger,
		executor.WithWorkDir(workspace),
		executor.WithRuntimeDir(os.Getenv(runtimeDirEnv)),
		executor.WithMetrics(metrics),
	)
	require.NoError(t, err)

	// Execute the function.
	req := execute.Request{
		FunctionID: functionID,
		Method:     path.Base(testFunction),
		Arguments: []string{
			"--chunk",
			fmt.Sprintf("%v", chunkSize),
			"--file",
			filepath.Base(testfile), // Specify name only because the path is relative to FS root.
		},
	}

	res, err := executor.ExecuteFunction(context.Background(), requestID, req)
	require.NoError(t, err)

	// Verify the execution result.
	require.Equal(t, codes.OK, res.Code)
	require.Equal(t, hash, res.Result.Stdout)

	// Verify usage info - for now, only that they are non-zero.
	cpuTimeTotal := res.Usage.CPUSysTime + res.Usage.CPUUserTime
	require.NotZero(t, cpuTimeTotal)
	require.NotZero(t, res.Usage.WallClockTime)
	require.NotZero(t, res.Usage.MemoryMaxKB)

	verifyExecutionMetrics(t, registry, req, res)
}

func createTestFile(t *testing.T, dir string, size int) (string, string) {
	t.Helper()

	const (
		filePattern = "testfile-"
	)

	f, err := os.CreateTemp(dir, filePattern)
	require.NoError(t, err)
	defer f.Close()

	buf := make([]byte, size)

	_, err = rand.Read(buf)
	require.NoError(t, err)

	n, err := f.Write(buf)
	require.NoError(t, err)
	require.Equal(t, size, n)

	// Calculate the hash of the file payload.
	hash := md5.New()
	_, err = hash.Write(buf)
	require.NoError(t, err)

	md5sum := fmt.Sprintf("%x", hash.Sum(nil))

	return f.Name(), md5sum
}

func createDirs(t *testing.T, dirs ...string) {
	t.Helper()

	// Create directory structure.
	for _, dir := range dirs {
		err := os.MkdirAll(dir, os.ModePerm)
		require.NoError(t, err)
	}

	return
}

func copyFunction(t *testing.T, filePath string, target string) {
	t.Helper()

	payload, err := os.ReadFile(filePath)
	require.NoError(t, err)

	_, name := path.Split(filePath)
	targetPath := filepath.Join(target, name)

	err = os.WriteFile(targetPath, payload, os.ModePerm)
	require.NoError(t, err)
}

func cleanupDisabled() bool {
	return os.Getenv(cleanupDisableEnv) == "yes"
}

func verifyExecutionMetrics(t *testing.T, gatherer prometheus.Gatherer, req execute.Request, res execute.Result) {
	t.Helper()

	metrics := helpers.MetricMap(t, gatherer)

	helpers.CounterCmp(t, metrics, float64(1), "b7s_executor_function_executions", "function", req.FunctionID)
	helpers.CounterCmp(t, metrics, float64(res.Usage.CPUSysTime.Milliseconds()), "b7s_executor_function_executions_cpu_sys_time_milliseconds")
	helpers.CounterCmp(t, metrics, float64(res.Usage.CPUUserTime.Milliseconds()), "b7s_executor_function_executions_cpu_user_time_milliseconds")
	helpers.CounterCmp(t, metrics, float64(0), "b7s_executor_function_executions_err")
	helpers.CounterCmp(t, metrics, float64(1), "b7s_executor_function_executions_ok", "function", req.FunctionID)
}
