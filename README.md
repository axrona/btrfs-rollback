# btrfs-rollback

Btrfs rollback tool written in Go for Arch-based distributions.

> [!WARNING]
> This tool is designed to be used in a live environment, such as a live USB or a booted snapshot. **Do not use this tool on your primary system without understanding the consequences.** Running it on the wrong system or environment may result in critical data loss or system instability.

## ⚙️ Configuration

By default, the config file is located at `/etc/btrfs-rollback.toml`:

```toml
# Name of your root subvolume (usually "@")
subvol_main = "@"

# Name of the subvolume where snapshots are stored
subvol_snapshots = "@snapshots"

# Temporary directory where the Btrfs root will be mounted
mountpoint = "/btrfs"

# Path to the Btrfs device. Required only if auto-mounting is needed
dev = "/dev/sda2"
```

Make sure to adjust the configuration to reflect your system's Btrfs setup and the live environment you are working in.

## 🛠️ Installation

### Install via `yay` (AUR):
```bash
yay -S btrfs-rollback
```
This will install the tool from the AUR. Alternatively, you can manually build and install it by cloning the repository and running the following commands:
```bash
git clone https://github.com/xeyossr/btrfs-rollback.git
cd btrfs-rollback
go build .
sudo install -Dm755 btrfs-rollback /usr/bin/btrfs-rollback
```

## 🚀 Usage
1. **Run the tool:** Once installed, you can run the btrfs-rollback tool directly from the command line.
```bash
btrfs-rollback
```
2. **Follow the prompts:** The tool will show a list of available snapshots and allow you to select one for rollback.

## 📜 License

This project is licensed under the GNU General Public License v3.0 (GPL-3.0).

For more details, see the [LICENSE](LICENSE) file.