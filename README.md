<p align="center">
  <img src="/assets/pkgz-logo.png" alt="Pkgz Logo" width="400"/>
</p>

[![tldr-pages](https://img.shields.io/badge/tldr-included-blue?logo=readthedocs&style=flat-square)](https://github.com/tldr-pages/tldr/blob/main/pages/common/pkgz.md)

**Pkgz** is a fast, extensible CLI tool written in Crystal 💎 for managing software packages across multiple Linux distributions.

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
  - Pacstall
  - FreeBSD & FreeBSD Ports
  - OpenBSD & OpenBSD Ports
- ⚙️ Configurable via `~/.config/pkgz/config.toml`  
- 🌱 Extensible to support other package managers  

---

## 📦 Requirements

To use **pkgz**, you’ll need the following:

- **Privilege elevation:**  
  Either `sudo` or `doas` must be installed and properly configured on your system.

- **At least one supported package manager:**  
  One or more of the following package managers must be installed and enabled in your config:
  - `apt` or `nala` (Debian/Ubuntu)
  - `pacman` or `paru` (Arch Linux / AUR)
  - `dnf` (Fedora/RHEL)
  - `flatpak`
  - `pacstall`
  - `pkg` or FreeBSD Ports (FreeBSD)
  - `pkg_add` or OpenBSD Ports (OpenBSD)

- **Crystal compiler:**  
  Only required if you are building **pkgz** from source.  
  Prebuilt binaries are provided and do not require Crystal to be installed.

---

## ⚙️ Configuration

Create or edit `~/.config/pkgz/config.toml`:

```toml
[sources]
apt = true
nala = false
flatpak = true
paru = false
pacman = false
dnf = false
pacstall = true
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
curl -LO https://github.com/roguehashrate/pkgz/releases/download/v0.1.4/pkgz
curl -LO https://github.com/roguehashrate/pkgz/releases/download/v0.1.4/pkgz.sha256

sha256sum -c pkgz.sha256
```

---

### Build from Source

```bash
git clone https://github.com/roguehashrate/pkgz
cd pkgz
crystal build src/pkgz.cr --release -o pkgz
sudo mv pkgz /usr/local/bin/
```

---

### Debian `.deb` Package

```bash
wget https://github.com/roguehashrate/pkgz/releases/download/v0.1.4/pkgz-0.1.4.deb
sudo dpkg -i pkgz-0.1.4.deb
sudo apt-get install -f  # Fix dependencies if needed
```

---

### Prebuilt Binary

Download from [Releases](https://github.com/roguehashrate/pkgz/releases):

```bash
sudo mv pkgz /usr/local/bin/
sudo chmod +x /usr/local/bin/pkgz
```

---

### Tarball (for Arch and others)

```bash
wget https://github.com/roguehashrate/pkgz/releases/download/v0.1.4/pkgz-0.1.4-x86_64.tar.gz
tar -xvf pkgz-0.1.4-x86_64.tar.xz
sudo cp pkgz-0.1.4/usr/bin/pkgz /usr/bin/
```

Or install locally:

```bash
mkdir -p ~/.local/bin
cp pkgz-0.1.4/usr/bin/pkgz ~/.local/bin/
```

Make sure `~/.local/bin` is in your PATH.

---

## 📦 Pacstall Support & Local Installation

`pkgz` supports installing packages via **Pacstall**, a universal package manager for Linux.

1. Ensure **Pacstall** is installed on your system. For installation instructions, visit https://pacstall.dev/

2. Install the local `.pacstall` package file using:

```bash
pacstall -I /path/to/pkgz.pacstall
```

This will install `pkgz` without needing to publish the package remotely.

Afterward, you can use `pkgz` as usual to manage software packages across supported sources.

The pacstall file will likely always be ahead of the pacstall repos.

---

## 🚀 Usage

```bash
pkgz <install|remove|update|search|--version> [app-name]
```

Examples:

```bash
pkgz install gimp
pkgz remove neofetch
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

### 🆘 Need Help Using `pkgz`?

If you're ever unsure how to use `pkgz` or want a quick refresher on commands, the easiest way to get help is through the [`tldr`](https://tldr.sh) tool:

#### 📥 Install `tldr`:

```bash
# Debian/Ubuntu
sudo apt install tldr

# Arch Linux
sudo pacman -S tldr

```

#### 📖 Then run:

```bash
tldr pkgz
```

This will show a concise, community-maintained usage guide for `pkgz` directly in your terminal — no need to scroll through docs or flags.

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
3. Add your source to the enabled sources list and config.

---

## 🪪 License

MIT License

Created by [roguehashrate](https://github.com/roguehashrate)
