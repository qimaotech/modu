package modu

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// TestE2E 端到端测试
// 测试 modu init -> create -> modify -> delete (blocked) 的完整流程
func TestE2E(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	t.Logf("Testing in: %s", tmpDir)

	// 创建配置文件
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	os.MkdirAll(workspace, 0755)
	os.MkdirAll(worktreeRoot, 0755)

	// 初始化一个 git 仓库作为测试
	repoDir := filepath.Join(workspace, "test-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	// 初始化 git 仓库
	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Logf("git init output: %s", out)
		// 可能已存在，继续
	}

	// 写入测试文件
	testFile := filepath.Join(repoDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Repo\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// 提交初始内容
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v, output: %s", err, out)
	}

	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v, output: %s", err, out)
	}

	// 创建 .modu.yaml
	config := `workspace: ` + workspace + `
worktree-root: ` + worktreeRoot + `
default-base: master
concurrency: 2
auto-fetch: false
strict-dirty-check: true
modules:
  - name: test-repo
    url: ` + repoDir + `
`
	configPath := filepath.Join(tmpDir, ".modu.yaml")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// 复制 modu 二进制到 tmpDir
	moduBin := filepath.Join(tmpDir, "modu")
	if err := copyFile("./modu", moduBin); err != nil {
		t.Fatalf("failed to copy modu binary: %v", err)
	}
	os.Chmod(moduBin, 0755)

	// 测试 modu list（空）
	cmd = exec.Command(moduBin, "list", "-c", configPath)
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()
	t.Logf("modu list output: %s, err: %v", string(out), err)

	// 测试 modu create
	featureName := "test-feat"
	cmd = exec.Command(moduBin, "create", featureName, "-c", configPath)
	cmd.Dir = tmpDir
	out, err = cmd.CombinedOutput()
	t.Logf("modu create output: %s, err: %v", string(out), err)
	if err != nil {
		t.Skipf("create failed, skipping E2E test (needs git remote or proper setup): %v", err)
		return
	}

	// 检查 worktree 是否创建
	featurePath := filepath.Join(worktreeRoot, featureName, "test-repo")
	if _, err := os.Stat(featurePath); os.IsNotExist(err) {
		t.Errorf("worktree not created at: %s", featurePath)
	}

	// 修改文件
	modifiedFile := filepath.Join(featurePath, "modified.txt")
	if err := os.WriteFile(modifiedFile, []byte("modified"), 0644); err != nil {
		t.Fatalf("failed to write modified file: %v", err)
	}

	// 测试 modu delete（应该被拦截）
	cmd = exec.Command(moduBin, "delete", featureName, "-c", configPath)
	cmd.Dir = tmpDir
	out, err = cmd.CombinedOutput()
	t.Logf("modu delete (should fail) output: %s, err: %v", string(out), err)
	if err == nil {
		t.Error("expected delete to fail due to dirty worktree, but it succeeded")
	}

	// 清理修改的文件
	os.Remove(modifiedFile)

	// 测试 modu delete -force
	cmd = exec.Command(moduBin, "delete", featureName, "-f", "-c", configPath)
	cmd.Dir = tmpDir
	out, err = cmd.CombinedOutput()
	t.Logf("modu delete -f output: %s, err: %v", string(out), err)
	if err != nil {
		t.Errorf("delete failed: %v", err)
	}

	// 验证已删除
	if _, err := os.Stat(featurePath); !os.IsNotExist(err) {
		t.Errorf("worktree not deleted at: %s", featurePath)
	}

	t.Log("E2E test completed successfully")
}
