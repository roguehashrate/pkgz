<p align="center">
  <img src="/assets/pkgz-logo.png" alt="Pkgz Logo" width="400"/>
</p>

**Pkgz** is a fast, extensible CLI tool written in Crystal 💎 for managing software packages across multiple Linux distributions.

[![License: RPL-v2](https://img.shields.io/badge/RPL-v2?style=flat&label=License&labelColor=ec8f1d&color=ffffff)](/LICENSE)


---

## ✨ Features

- ✅ Install, remove, update and search apps  
- 🔍 Interactive source selection if app is available in multiple sources  
- 🔐 Automatically uses `doas` or `sudo` for privilege elevation  
- 📦 Supports:
  - Apt / Nala (Debian/Ubuntu)
  - Flatpak
  - Pacman (Arch)
  - Paru (AUR helper)
  - DNF (Fedora/RHEL)
  - APK (Alpine)
  - Pacstall
  - Zypper (openSUSE)
  - XBPS (Void)
  - FreeBSD & FreeBSD Ports
  - OpenBSD & OpenBSD Ports
- ⚙️ Configurable via `~/.config/pkgz/config.toml`  
- 🌱 Extensible to support other package managers  

---

## 📦 Requirements

To use **pkgz**, you’ll need the following:

- **Privilege elevation:**  
  Either `sudo` or `doas` must be installed.

- **At least one supported package manager:**  
  Linux: `apt`, `nala`, `flatpak`, `pacman`, `paru`, `yay`, `dnf`, `zypper`, `apk`, `xbps`, or `pacstall`  
  BSD: `FreeBSD pkg`, `FreeBSD Ports`, `OpenBSD pkg`, `OpenBSD Ports`

- **Crystal compiler:**  
  Only needed if you're building from source.  
  (Prebuilt binaries don’t require Crystal.)

---

## ⚙️ Configuration

Create or edit `~/.config/pkgz/config.toml`:

```toml
[sources]
apt = true
nala = false
flatpak = true
paru = false
yay = false
pacman = false
dnf = false
pacstall = true
zypper = false
xbps = false
freebsd = false
freebsd_ports = false
openbsd = false
openbsd_ports = false

[elevator]
command = "sudo"  # or "doas"
```

This controls enabled sources and privilege elevator. If missing, Pkgz will prompt and exit.

---

## 🛠 Installation

### 🧪 Recommended: One-liner Install (Linux x86_64)

You can install the latest prebuilt binary directly with:

```bash
curl -sS https://raw.githubusercontent.com/roguehashrate/pkgz/main/install.sh | bash
```

This installs `pkgz` to `~/.local/bin`.  
Make sure `~/.local/bin` is in your `$PATH`.

---

### 🔐 Verify Download (*Optional*)

To verify the integrity of the binary:

```bash
curl -LO https://github.com/roguehashrate/pkgz/releases/download/v0.1.8/pkgz
curl -LO https://github.com/roguehashrate/pkgz/releases/download/v0.1.8/pkgz.sha256

sha256sum -c pkgz.sha256
```

---

### Build from Source

```bash
git clone https://github.com/roguehashrate/pkgz
cd pkgz
crystal build src/pkgz.cr --release -o pkgz
mv pkgz ~/.local/bin/
```

---

### Prebuilt Binary

Download from [Releases](https://github.com/roguehashrate/pkgz/releases):

```bash
mv pkgz ~/.local/bin
chmod +x ~/.local/bin/pkgz
```

---

### Tarball (for Arch and others)

```bash
wget https://github.com/roguehashrate/pkgz/releases/download/v0.1.8/pkgz-0.1.8.tar.gz
wget https://github.com/roguehashrate/pkgz/releases/download/v0.1.8/pkgz-0.1.8.tar.gz.sha256

sha256sum -c pkgz-0.1.8.tar.gz.sha256
tar -xvf pkgz-0.1.8.tar.gz
cd pkgz
chmod +x install.sh
./install.sh
```

Make sure `~/.local/bin` is in your PATH.

---

## 🚀 Usage

```bash
pkgz <install|remove|update|search|clean|--version> [app-name]
```

Examples:

```bash
pkgz install gimp
pkgz remove neofetch
pkgz clean
pkgz update
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

1. Subclass `Pkgz::Source`  
2. Implement:
   - `name`  
   - `available?(app)`  
   - `install(app)`  
   - `remove(app)`  
   - `update`
   - `search(app)`
3. Add your source to the enabled sources list and config.

---

Created by [roguehashrate](https://github.com/roguehashrate)
