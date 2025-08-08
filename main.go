package main

import (
	"fmt"
	"log"

	utils "github.com/xeyossr/btrfs-rollback/internal"
	"github.com/xeyossr/btrfs-rollback/internal/btrfs"
	"github.com/xeyossr/btrfs-rollback/internal/cmd"
	"github.com/xeyossr/btrfs-rollback/internal/ui"
)

var (
	Must = utils.Must
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

	cfg := cmd.Config()
	btrfs.SetDryRun(cmd.DryRun())

	// Root?
	utils.CheckIfRoot()
	
	// Mount the BTRFS subvolume first
	err := btrfs.MountSubvol(cfg)
	Must(err)

	snapshots, err := btrfs.GetSnapshots(cfg)
	Must(err)

	snapshotID, err := ui.RunUI(snapshots)
	Must(err)

	ui.ClearScreen()

	if !cmd.DryRun() {
		confirmed := ui.Confirm("Are you sure you want to rollback to snapshot " + snapshotID + "?")
		fmt.Println()
		if !confirmed {
			return
		}
	}

	btrfs.Rollback(cfg, snapshotID)
}
