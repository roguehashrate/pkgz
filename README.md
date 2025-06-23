# Pkgz

**Pkgz** is a fast, extensible CLI tool written in Crystal 💎 for managing software packages across multiple Linux distributions. It supports system and Flatpak packages, with automatic privilege elevation (`sudo`/`doas`) and interactive source selection.

---

## ✨ Features

- ✅ Install, remove, and update apps  
- 🔍 Interactive source selection if app is available in multiple places  
- 🔐 Uses `doas` or `sudo` automatically  
- 📦 Supports:
  - APT/Nala (Debian/Ubuntu)
  - Flatpak
  - Pacman (Arch)
  - Paru (AUR helper)
  - DNF (Fedora/RHEL)
- ⚙️ Configurable via `~/.config/pkgz/config.toml`  
- 🌱 Extensible for other package managers  

---

## 📦 Requirements

- [Crystal](https://crystal-lang.org) (for building from source)
- One or more of the following package managers installed:
  - `apt`, `nala`, `flatpak`, `pacman`, `paru`, `dnf`
- `sudo` or `doas`
- (Optional) `flatpak` for Flatpak support

---

## ⚙️ Configuration

Create the file:  
`~/.config/pkgz/config.toml`

Example:

```toml
[sources]
apt = true
flatpak = true
paru = true
pacman = false
dnf = false
```

This file controls which sources Pkgz uses. If the file is missing, Pkgz will notify the user and exit.

---

## 🛠 Installation

### Option 1: Build from Source

```bash
git clone https://github.com/roguehashrate/pkgz
cd pkgz
crystal build src/pkgz.cr --release -o pkgz
sudo mv pkgz /usr/local/bin/
```

---

### Option 2: Install via `.deb`

If you're on a Debian-based system, you can install Pkgz with the prebuilt `.deb` package:

```bash
wget https://github.com/roguehashrate/pkgz/releases/download/v0.1.2/pkgz_0.1.2.deb
sudo dpkg -i pkgz_0.1.2.deb
sudo apt-get install -f  # Installs any missing dependencies
```

This will install `pkgz` into `/usr/bin`.

> 💡 **Note:** If you previously copied a binary to `/usr/local/bin`, remove it first:
>
> ```bash
> sudo rm /usr/local/bin/pkgz
> ```

---

### Option 3: Prebuilt Binary (Manual)

Download from [Releases](https://github.com/yourusername/pkgz/releases), then:

```bash
sudo mv pkgz /usr/local/bin/
sudo chmod +x /usr/local/bin/pkgz
```

---

### Option 4: Tarball (Recommended for Arch and others)

Download the tarball from [Releases](https://github.com/yourusername/pkgz/releases):

```bash
wget https://github.com/roguehashrate/pkgz/releases/download/v0.1.2/pkgz-0.1.2-x86_64.tar.xz
```

Extract and install:

```bash
tar -xvf pkgz-0.1.2-x86_64.tar.xz
sudo cp pkgz-0.1.2/usr/bin/pkgz /usr/bin/
```

Or for a user-local install:

```bash
tar -xvf pkgz-0.1.2-x86_64.tar.xz
mkdir -p ~/.local/bin
cp pkgz-0.1.2/usr/bin/pkgz ~/.local/bin/
```

Make sure `~/.local/bin` is in your PATH.

---

## 🚀 Usage

```bash
pkgz <install|remove|update|--version> [app-name]
```

### Examples:

```bash
pkgz install gimp
pkgz remove neofetch
pkgz update
pkgz --version
```

### Sample Output

```bash
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

- Automatically uses `doas` if available, otherwise falls back to `sudo`.
- Commands requiring root privileges (like installing or updating system packages) are handled automatically.

---

## 🧩 Extending Support

To add a new package manager, subclass `Pkgz::Source` and implement:

- `name`
- `available?(app)`
- `install(app)`
- `remove(app)`
- `update`

Then load it in `Pkgz.available_sources` based on binary presence and config.

---

## 🪪 License

MIT License

Created with ❤️ by [roguehashrate](https://github.com/roguehashrate)
