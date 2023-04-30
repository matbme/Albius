package albius

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type GrubConfig map[string]string

const (
	BIOS = "i386-pc"
	EFI  = "x86_64-efi"
)

type FirmwareType string

func GetGrubConfig(targetRoot string) (GrubConfig, error) {
	targetRootGrubFile := filepath.Join(targetRoot, "/etc/default/grub")

	// If grub config file doesn't exist yet, return an empty map
	if _, err := os.Stat(targetRootGrubFile); os.IsNotExist(err) {
		return GrubConfig{}, nil
	}

	content, err := os.ReadFile(targetRootGrubFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read GRUB config file: %s", err)
	}

	config := GrubConfig{}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		kv := strings.SplitN(line, "=", 2)
		config[kv[0]] = kv[1]
	}

	return config, nil
}

func WriteGrubConfig(targetRoot string, config GrubConfig) error {
	fileContents := []byte{}
	for k, v := range config {
		line := fmt.Sprintf("%s=%s\n", k, v)
		fileContents = append(fileContents, []byte(line)...)
	}

	targetRootGrubFile := filepath.Join(targetRoot, "/etc/default/grub")
	err := os.WriteFile(targetRootGrubFile, fileContents, 0644)
	if err != nil {
		return fmt.Errorf("Failed to write GRUB config file: %s", err)
	}

	return nil
}

func AddGrubScript(targetRoot, scriptPath string) error {
	// Ensure script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("Error adding GRUB script: %s does not exist", scriptPath)
	}

	contents, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("Failed to read GRUB script at %s: %s", scriptPath, err)
	}

	targetRootPath := filepath.Join(targetRoot, scriptPath)
	err = os.WriteFile(targetRootPath, contents, 0755) // Grub expects script to be executable
	if err != nil {
		return fmt.Errorf("Failed to writing GRUB script to %s: %s", targetRootPath, err)
	}

	return nil
}

func RemoveGrubScript(targetRoot, scriptName string) error {
	targetRootPath := filepath.Join(targetRoot, "/etc/grub.d", scriptName)

	// Ensure script exists
	if _, err := os.Stat(targetRootPath); os.IsNotExist(err) {
		return fmt.Errorf("Error removing GRUB script: %s does not exist", targetRootPath)
	}

	err := os.Remove(targetRootPath)
	if err != nil {
		return fmt.Errorf("Error removing GRUB script: %s", err)
	}

	return nil
}

func RunGrubInstall(targetRoot, bootDirectory, diskPath string, target FirmwareType) error {
	grubInstallCmd := "grub-install --boot-directory %s --target=%s %s"

	err := RunInChroot(targetRoot, fmt.Sprintf(grubInstallCmd, bootDirectory, target, diskPath))
	if err != nil {
		return fmt.Errorf("Failed to run grub-install: %s", err)
	}

	return nil
}

func RunGrubMkconfig(targetRoot, output string) error {
	grubMkconfigCmd := "grub-mkconfig -o %s"

	err := RunInChroot(targetRoot, fmt.Sprintf(grubMkconfigCmd, output))
	if err != nil {
		return err
	}

	return nil
}