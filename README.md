<p align="center">
  <img src="/assets/pkgz-logo.png" alt="Pkgz Logo" width="400"/>
</p>

**Pkgz** is a fast, extensible CLI tool written in Go 🐹 for managing multiple package types on Linux and BSD distributions.


---

## ✨ Features

- ✅ Install, remove, update, refresh and search apps  
- 🔍 Interactive source selection if app is available in multiple sources  
- 🔐 Automatically uses `doas` or `sudo` for privilege elevation  
- 📦 Supports:
  - Apt / Nala (Debian/Ubuntu)
  - Flatpak
  - Pacman (Arch)
  - Paru (AUR helper)
  - Yay (AUR helper)
  - DNF (Fedora/RHEL)
  - APK (Alpine)
  - Pacstall
  - Zypper (openSUSE)
  - XBPS (Void)
  - Nix (`Untested`)
  - FreeBSD & FreeBSD Ports (`Untested`)
  - OpenBSD & OpenBSD Ports (`Untested`)
- ⚙️ Configurable via `~/.config/pkgz/config.toml`  
- 🌱 Extensible to support other package managers  

---

## 📦 Requirements

To use **pkgz**, you’ll need the following:

- **Privilege elevation:**  
  Either `sudo` or `doas` must be installed.

- **At least one supported package manager:**  
  Linux: `apt`, `nala`, `flatpak`, `pacman`, `paru`, `yay`, `dnf`, `zypper`, `apk`, `xbps`, `nix`  or `pacstall`  
  BSD: `FreeBSD pkg`, `FreeBSD Ports`, `OpenBSD pkg`, `OpenBSD Ports`

- **Go compiler:**  
  Only needed if you're building from source.  
  Required for building from source.

---

## ⚙️ Configuration

Create or edit `~/.config/pkgz/config.toml`:

```toml
# Enable/disable package manager sources
[sources]
apt = false
nala = false
flatpak = false
paru = false
yay = false
pacman = false
dnf = false
pacstall = false
zypper = false
xbps = false
nix = false
freebsd = false
freebsd_ports = false
openbsd = false
openbsd_ports = false

# Privilege escalation method (required)
[elevator]
command = "sudo"  # or "doas"
```

**Configuration Notes:**
- Only enable sources you actually use by setting them to `true`
- You **must** have an elevator configured (`sudo` or `doas`)
- The config file must be created manually before first run
- The program will show a template if the config file is missing

---

## 🛠 Installation

### 🔨 Build from Source

**Prerequisites:**
- Go compiler (required for building from source)
- git (for cloning the repository)

#### 🔨 Interactive Build (Recommended)
**Purpose**: Build for your specific target platform interactively
```bash
git clone https://github.com/roguehashrate/pkgz
cd pkgz
chmod +x build.sh
./build.sh
```
The script will ask you to select:
1. Operating system (linux, darwin, freebsd, openbsd)
2. Architecture (amd64, 386, arm64, arm)

**Output**: Creates a compressed binary in `build/{OS}/{ARCH}/`
- Binary: `build/{OS}/{ARCH}/pkgz.gz` (extracts to 'pkgz')

---

### 📦 Install the Binary

After building completes, install the binary to make it available system-wide:

**If you have a compressed binary (.gz):**
```bash
cp build/{OS}/{ARCH}/pkgz.gz ~/.local/bin/
cd ~/.local/bin
gunzip pkgz.gz
chmod +x pkgz
```

**Verify installation:**
```bash
pkgz --version
```

**Note:** Make sure `~/.local/bin` is in your PATH. If not, add:
```bash
echo 'export PATH="$PATH:~/.local/bin"' >> ~/.bashrc
source ~/.bashrc
```
or if you are a zsh user
```bash
echo 'export PATH="$PATH:~/.local/bin"' >> ~/.zshrc
source ~/.zshrc
```

---

## 🚀 Usage

---

Examples:

```bash
pkgz install gimp
pkgz remove gimp
pkgz clean
pkgz info          # Show package counts per source
pkgz info gimp      # Show specific package status
pkgz refresh        # Check for available updates without installing them
pkgz update         # Apply all available updates
pkgz --version
```

Sample output:

```
$ pkgz install gimp
🔍 Searching for 'gimp' in sources...
📦 Found 'gimp' in multiple sources:
1. APT
2. Flatpak
Which one would you like to use? [1-2]: 2
🚀 Installing with Flatpak...
```

---

## 🔐 Privilege Elevation

- Automatically detects and uses `doas` or `sudo`.
- Privileged commands are run with the configured elevator command.

---

## 🧩 Extending Pkgz

To add support for a new package manager:

1. Implement the `Source` interface  
2. Implement the interface methods:
   - `Name()`  
   - `Available(app string)`  
   - `Install(app string)`  
   - `Remove(app string)`  
   - `Update()`
   - `Search(app string)`
3. Add your source to the enabled sources list and config.

---

## 🔄 Migration from Crystal

This project was originally written in Crystal and has moved to Go mostly for compatablity reasons and also because a user is more likely to have Go installed on their system than Crystal so it reduces friction a little.

---

## 📄 License

This project is licensed under the **BSD 2-Clause License**.

See the [LICENSE](LICENSE) file for the full license text.

---

Created by [roguehashrate](https://github.com/roguehashrate)