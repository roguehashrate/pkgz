
<p align="center">
  <img src="/assets/pkgz-logo.png" alt="Pkgz Logo" width="400"/>
</p>



**Pkgz** is a fast, extensible CLI tool written in Crystal 💎 for managing software packages across multiple Linux distributions. It supports system and Flatpak packages, with automatic privilege elevation (`sudo`/`doas`) and interactive source selection.

---

## ✨ Features

- ✅ Install, remove, and update apps  
- 🔍 Interactive source selection if app is available in multiple sources  
- 🔐 Automatically uses `doas` or `sudo` for privilege elevation  
- 📦 Supports:
  - APT / Nala (Debian/Ubuntu)
  - Flatpak
  - Pacman (Arch)
  - Paru (AUR helper)
  - DNF (Fedora/RHEL)
  - Pacstall
- ⚙️ Configurable via `~/.config/pkgz/config.toml`  
- 🌱 Extensible to support other package managers  

---

## 📦 Requirements

- [Crystal](https://crystal-lang.org) (if building from source)  
- One or more package managers installed:
  - `apt`, `nala`, `flatpak`, `pacman`, `paru`, `dnf`, `pacstall`  
- `sudo` or `doas` installed  
- (Optional) `flatpak` for Flatpak support

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

[elevator]
command = "sudo"  # or "doas"
```

This controls enabled sources and privilege elevator. If missing, Pkgz will prompt and exit.

---

## 🛠 Installation

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
wget https://github.com/roguehashrate/pkgz/releases/download/v0.1.2/pkgz_0.1.2.deb
sudo dpkg -i pkgz_0.1.2.deb
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
wget https://github.com/roguehashrate/pkgz/releases/download/v0.1.2/pkgz-0.1.2-x86_64.tar.xz
tar -xvf pkgz-0.1.2-x86_64.tar.xz
sudo cp pkgz-0.1.2/usr/bin/pkgz /usr/bin/
```

Or install locally:

```bash
mkdir -p ~/.local/bin
cp pkgz-0.1.2/usr/bin/pkgz ~/.local/bin/
```

Make sure `~/.local/bin` is in your PATH.

---

## 📦 Pacstall Support & Local Installation

`pkgz` supports installing packages via **Pacstall**, a universal package manager for Linux.

1. Ensure **Pacstall** is installed on your system. For installation instructions, visit https://pacstall.dev/

2. Install the local `.pacstall` package file using:

```bash
pacstall -I /path/to/pkgz-0.1.3.pacstall
```

This will install `pkgz` without needing to publish the package remotely.

Afterward, you can use `pkgz` as usual to manage software packages across supported sources.

The pacstall file will likely always be ahead of the pacstall repos.

---

## 🚀 Usage

```bash
pkgz <install|remove|update|--version> [app-name]
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
