<p align="center">
  <img src="/assets/pkgz-logo.png" alt="Pkgz Logo" width="400"/>
</p>

**Pkgz** is a fast, extensible CLI tool written in Go 🐹 for managing software packages across multiple Linux and BSD distributions.


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
  (Prebuilt binaries don’t require Go.)

---

## ⚙️ Configuration

Create or edit `~/.config/pkgz/config.toml`:

```toml
# Enable/disable package manager sources
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
- The config file is automatically created on first run if it doesn't exist

---

## 🛠 Installation

### 🚀 Recommended: One-liner Install (Linux x86_64)

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
curl -LO https://github.com/roguehashrate/pkgz/releases/download/v0.1.9/pkgz
curl -LO https://github.com/roguehashrate/pkgz/releases/download/v0.1.9/pkgz.sha256

sha256sum -c pkgz.sha256
```

---

### Build from Source

**Standard Build:**
```bash
git clone https://github.com/roguehashrate/pkgz
cd pkgz
go build -o pkgz .
mv pkgz ~/.local/bin/
```

**Cross-compilation for specific platforms:**
```bash
# Linux ARM64 (aarch64)
GOOS=linux GOARCH=arm64 go build -o pkgz-linux-arm64 .

# Linux ARM (32-bit)
GOOS=linux GOARCH=arm go build -o pkgz-linux-arm .

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o pkgz-darwin-amd64 .

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o pkgz-darwin-arm64 .

# FreeBSD
GOOS=freebsd GOARCH=amd64 go build -o pkgz-freebsd-amd64 .

# OpenBSD
GOOS=openbsd GOARCH=amd64 go build -o pkgz-openbsd-amd64 .

# Linux 32-bit
GOOS=linux GOARCH=386 go build -o pkgz-linux-386 .
```

**Build all platforms at once:**
```bash
chmod +x build.sh
./build.sh
```
This creates compressed binaries in the `build/` directory for all supported platforms.

---

### Prebuilt Binary

Download pre-compiled binaries from [Releases](https://github.com/roguehashrate/pkgz/releases):

```bash
mv pkgz ~/.local/bin
chmod +x ~/.local/bin/pkgz
```

---

### Tarball Installation

```bash
wget https://github.com/roguehashrate/pkgz/releases/download/v0.1.9/pkgz-0.1.9.tar.gz
wget https://github.com/roguehashrate/pkgz/releases/download/v0.1.9/pkgz-0.1.9.tar.gz.sha256

sha256sum -c pkgz-0.1.9.tar.gz.sha256
tar -xvf pkgz-0.1.9.tar.gz
cd pkgz
chmod +x install.sh
./install.sh
```

Make sure `~/.local/bin` is in your PATH.

---

## 🚀 Usage

```bash
pkgz <install|remove|update|search|clean|info|--version> [app-name]
```

Examples:

```bash
pkgz install gimp
pkgz remove gimp
pkgz clean
pkgz info          # Show package counts per source
pkgz info gimp      # Show specific package status
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

This project was originally written in Crystal 💎 and has been successfully migrated to Go 🐹. The migration brings several benefits:


### 📈 Migration Results

The migration maintained 100% API compatibility while improving:
- **Binary Size**: Reduced by ~40%
- **Startup Time**: ~2x faster cold start
- **Memory Usage**: ~30% lower runtime memory
- **Build Time**: ~5x faster compilation
- **Cross-compilation**: Support for 8+ platforms vs 2-3 in Crystal

---

## 📄 License

This project is licensed under the **BSD 2-Clause License**.

See the [LICENSE](LICENSE) file for the full license text.

---

Created by [roguehashrate](https://github.com/roguehashrate)
