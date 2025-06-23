# Pkgz

A simple Crystal CLI tool to install, remove, and update applications from APT and Flatpak on Debian-based GNU/Linux systems.

Supports interactive selection when apps are available from multiple sources and manages privilege escalation via system commands.

---

## Features

- Install apps from APT or Flatpak  
- Remove apps from all supported sources  
- Update system and Flatpak packages  
- Interactive choice when multiple sources provide the same app  
- Extensible design for adding more package sources  

---

## Requirements

- Crystal programming language installed (for building from source)  
- Debian-based Linux (Ubuntu, Debian, etc.)  
- APT and Flatpak installed and configured  
- Optional: sudo or doas for privilege escalation  

---

## Installation

### Option 1: Use the Precompiled Binary

Download the `pkgz` binary from the [Releases](https://github.com/yourusername/pkgz/releases) section.

Move it to a directory in your PATH (e.g., `/usr/local/bin`):

```bash
sudo mv pkgz /usr/local/bin/
sudo chmod +x /usr/local/bin/pkgz
```

### Option 2: Build from Source

Clone this repository or download the `pkgz.cr` file.

Compile the program:

```bash
crystal build src/pkgz.cr --release -o pkgz
```

(Optional) Move the compiled binary to your PATH:

```bash
sudo mv pkgz /usr/local/bin/
```

---

## Usage

```bash
pkgz <command> [app-name]
```

You can run commands like:

- `install <app-name>` — install an app  
- `remove <app-name>` — remove an app from all sources  
- `update` — update all package sources  

If multiple sources provide the app, you’ll be prompted to choose.

---

## Example

```bash
$ pkgz install gimp
🔍 Searching for 'gimp' in sources...
📦 Found 'gimp' in multiple sources:
1. APT
2. Flatpak
Which one would you like to use? [1-2]: 2
🚀 Installing with Flatpak...
```

```bash
$ pkgz remove tmux
❌ Trying to remove 'tmux' from APT...
❌ Trying to remove 'tmux' from Flatpak...
```

```bash
$ pkgz update
⬆️  Updating APT packages...
⬆️  Updating Flatpak packages...
```

---

## License

MIT License

Created with Crystal 💎 by roguehashrate