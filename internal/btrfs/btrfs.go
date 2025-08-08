// Package btrfs provides utilities for interacting with BTRFS snapshots,
// particularly for reading and parsing snapshot metadata stored in XML format.
package btrfs

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/xeyossr/btrfs-rollback/internal/config"
)

// Global dryRun flag controlled via cmd package
var dryRun bool

// SetDryRun sets the dry-run mode for this package
func SetDryRun(enabled bool) {
	dryRun = enabled
}

// Snapshot represents the metadata of a BTRFS snapshot.
type Snapshot struct {
	ID             string `xml:"num"`
	Method         string `xml:"type"`
	Date           string `xml:"date"`
	DescriptionStr string `xml:"description"`
}

// realIsBtrfsSubvolume returns true if the path is a Btrfs subvolume.
// In dry-run mode, always returns true for simulation.
func realIsBtrfsSubvolume(path string) bool {
	cmd := exec.Command("btrfs", "subvolume", "show", path)
	err := cmd.Run()
	return err == nil
}

// listSubvolumes returns all subvolumes inside the snapshot directory defined in cfg.
func listSubvolumes(cfg config.Config) ([]string, error) {
	var subvolumes []string
	snapshotDir := filepath.Join(cfg.Mountpoint, cfg.SubvolSnapshots)

	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(snapshotDir, entry.Name(), "snapshot")
		if entry.IsDir() && realIsBtrfsSubvolume(fullPath) {
			subvolumes = append(subvolumes, fullPath)
		}
	}

	return subvolumes, nil
}

// ReadSnapshotInfo reads the info.xml file inside snapshotPath and returns a Snapshot
func ReadSnapshotInfo(cfg config.Config, snapshotPath string) (Snapshot, error) {
	infoXMLPath := filepath.Join(filepath.Dir(snapshotPath), "info.xml")

	xmlData, err := os.ReadFile(infoXMLPath)
	if err != nil {
		return Snapshot{}, fmt.Errorf("could not read XML file: %v", err)
	}

	var snapshot Snapshot
	err = xml.Unmarshal(xmlData, &snapshot)
	if err != nil {
		return Snapshot{}, fmt.Errorf("could not unmarshal XML: %v", err)
	}

	return snapshot, nil
}

// GetSnapshots returns all snapshots found in the snapshot subvolume directory
// Sorts them by numeric ID in ascending order
func GetSnapshots(cfg config.Config) ([]Snapshot, error) {
	subvolPaths, err := listSubvolumes(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshot subvolumes: %v", err)
	}

	var snapshots []Snapshot
	for _, subvol := range subvolPaths {
		snapshot, err := ReadSnapshotInfo(cfg, subvol)
		if err != nil {
			return nil, fmt.Errorf("could not read snapshot info for %s: %v", subvol, err)
		}
		snapshots = append(snapshots, snapshot)
	}

	sort.Slice(snapshots, func(i, j int) bool {
		idI, errI := strconv.Atoi(snapshots[i].ID)
		idJ, errJ := strconv.Atoi(snapshots[j].ID)
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}
		return idI > idJ
	})

	return snapshots, nil
}

// MountSubvol performs the mount operation for the BTRFS subvolume if not already mounted
func MountSubvol(cfg config.Config) error {
	// Check if the mountpoint is already mounted by inspecting the mount status
	cmd := exec.Command("mount", "--grep", cfg.Mountpoint)
	output, err := cmd.CombinedOutput()
	if err != nil || len(output) == 0 {
		// Mount the device with the appropriate subvolid if it's not mounted
		if cfg.Dev != nil && *cfg.Dev != "" {
			// Mount the btrfs subvolume if it's not mounted
			cmd := exec.Command("mount", "-o", "subvolid=5", *cfg.Dev, cfg.Mountpoint)
			if dryRun {
				fmt.Printf("[dry-run] Would execute: %v\n", cmd.Args)
			}
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to mount %s: %v", *cfg.Dev, err)
			}
		} else {
			return fmt.Errorf("device not specified, unable to mount %s", cfg.Mountpoint)
		}
	} else {
		fmt.Println("Mountpoint is already mounted.")
	}
	return nil
}

// Rollback performs a rollback by:
// 1. Renaming the current SubvolMain to SubvolMain.old
// 2. Creating a snapshot from the specified snapshot subvolume into SubvolMain
// 3. Setting the default subvolume to SubvolMain
// Honors dry-run mode by printing commands instead of executing them
func Rollback(cfg config.Config, snapshotID string) error {
	subvolMainPath := filepath.Join(cfg.Mountpoint, cfg.SubvolMain)
	oldSubvolMainPath := subvolMainPath + ".old"
	snapshotPath := filepath.Join(cfg.Mountpoint, cfg.SubvolSnapshots, snapshotID, "snapshot")

	// Remove old subvolume if it exists
	if _, err := os.Stat(oldSubvolMainPath); err == nil {
		if dryRun {
			fmt.Printf("[dry-run] Would delete old subvolume: %s\n", oldSubvolMainPath)
		} else {
			fmt.Printf("Removing old subvolume: %s\n", oldSubvolMainPath)
			err := exec.Command("btrfs", "subvolume", "delete", oldSubvolMainPath).Run()
			if err != nil {
				return fmt.Errorf("failed to delete old subvolume: %v", err)
			}
		}
	}

	// Move the current subvolume
	if dryRun {
		fmt.Printf("[dry-run] Would move %s to %s\n", subvolMainPath, oldSubvolMainPath)
	} else {
		err := exec.Command("mv", subvolMainPath, oldSubvolMainPath).Run()
		if err != nil {
			return fmt.Errorf("failed to move %s to %s: %v", subvolMainPath, oldSubvolMainPath, err)
		}
	}

	// Create the snapshot
	if dryRun {
		fmt.Printf("[dry-run] Would create snapshot from %s to %s\n", snapshotPath, subvolMainPath)
	} else {
		err := exec.Command("btrfs", "subvolume", "snapshot", snapshotPath, subvolMainPath).Run()
		if err != nil {
			return fmt.Errorf("failed to create snapshot %s: %v", snapshotID, err)
		}
	}

	// Set default subvolume
	if dryRun {
		fmt.Printf("[dry-run] Would set default subvolume to: %s\n", subvolMainPath)
	} else {
		err := exec.Command("btrfs", "subvolume", "set-default", subvolMainPath).Run()
		if err != nil {
			return fmt.Errorf("failed to set default subvolume: %v", err)
		}
	}

	fmt.Printf("Rollback completed using snapshot %s\n", snapshotID)
	return nil
}
