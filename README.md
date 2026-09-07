<p align="center">
  <img src="/assets/pkgz-logo.png" alt="Pkgz Logo" width="400"/>
</p>

**Pkgz** is a fast, extensible CLI tool written in Go 🐹 for managing multiple package types on Linux distributions.


---

## ✨ Features

- ✅ Install, remove, update, refresh and search apps  
- 🖥️ Custom TUI (bubbletea) that drives every operation (`install`, `remove`, `update`, `refresh`, `search`, `info`, `clean`) in a single unified interface — per-source status, live spinner, and a scrollable output pane — instead of each package manager taking over your terminal.
- 🔍 Interactive source selection if app is available in multiple sources  
- 🔐 Automatically uses `doas` or `sudo` for privilege elevation  
- 📦 Supports:
  - Apt (Debian/Ubuntu)
  - Flatpak
  - Pacman (Arch)
  - Paru (AUR helper)
  - Yay (AUR helper)
  - DNF (Fedora/RHEL)
  - Zypper (openSUSE)
- ⚙️ Configurable via `~/.config/pkgz/config.toml`  
- 🌱 Extensible to support other package managers  
- 💾 Falls back to plain console output when stdout is not a terminal (pipes, scripts, CI)

---

## 📦 Requirements

To use **pkgz**, you’ll need the following:

- **Privilege elevation:**  
  Either `sudo` or `doas` must be installed.

- **At least one supported package manager:**  
  `apt`, `flatpak`, `pacman`, `paru`, `yay`, `dnf`, or `zypper`

- **Go compiler:**  
  Only needed if you're building from source.  
  Required for building from source.

---

## ⚙️ Configuration

On the **first run**, pkgz automatically detects the package managers installed on your system (apt, flatpak, pacman, paru, yay, dnf, zypper), picks a usable elevator (`doas` or `sudo`), and writes `~/.config/pkgz/config.toml` for you. You can edit it afterwards to enable/disable sources:

```toml
# Enable/disable package manager sources
[sources]
apt = true       # enabled because apt was detected
flatpak = true
paru = false
yay = false
pacman = false
dnf = false
zypper = false

# Privilege escalation method
[elevator]
command = "sudo"  # or "doas"
```

**Configuration Notes:**
- Only sources whose binaries actually exist are used. If you enable a source whose package manager is not installed, pkgz prints a warning and skips it (it never reports phantom results for it).
- The elevator command is honored as configured. If it can't be found, pkgz stops with a clear error instead of failing mid-operation.
- No config file is required — the tool is usable as-is on first run.

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
1. Operating system (linux)
2. Architecture (amd64, 386, arm64, arm)

**Output**: Creates a compressed binary in `build/{OS}/{ARCH}/`
- Binary: `build/{OS}/{ARCH}/pkgz.gz` (extracts to 'pkgz')

For a quick non-interactive host build that also produces `.tar.gz` and `.deb` packages:

```bash
./build.sh --dev
```

**Output**:
- Binary: `build/{OS}/{ARCH}/pkgz`
- Tarball: `build/{OS}/{ARCH}/pkgz-v{VERSION}-{OS}-{ARCH}.tar.gz`
- Debian package: `build/{OS}/{ARCH}/pkgz_{VERSION}_{arch}.deb` (Linux only)

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
pkgz install emacs --source flatpak   # pick the source explicitly
pkgz install vim curl                  # install several packages at once
pkgz remove gimp
pkgz clean
pkgz info          # Show package counts per source
pkgz info gimp      # Show specific package status
pkgz refresh        # Check for available updates without installing them
pkgz update         # Apply all available updates
pkgz search vim --source apt
pkgz --version
```

`--source NAME` (also accepted as `--from NAME`) skips the source picker, which makes
pkgz deterministic for scripts and CI. Without it, pkgz shows the picker only when
more than one source has the package; a single candidate runs directly.

When you run any command, pkgz opens its TUI showing each enabled source with its status and live output. For example `pkgz update`:

```
╭─────────────╮
│ pkgz update │
╰─────────────╯

● ⣻ ▸ Updating Apt
✓   Updating Flatpak
○   Updating DNF
○   Updating Pacman

Running...

↑/↓ select task · o toggle log · q quit
```

Status markers:
- `●` running (active task) · `✓` done · `▲` updates available (shown with a count) · `✗` failed · `○` not yet started

For a status check, `pkgz refresh` marks each source with `✓`/`▲`/`✗` and puts the count right in the label:

```
╭───────────────╮
│ pkgz refresh  │
╰───────────────╯

✓   Checking Apt — up to date
▲   Checking Pacman — 3 update(s)
✗   Checking DNF

2 done · 1 failed — finished with errors.

o toggle log · any other key to exit
```

When all tasks finish, pkgz stays on a summary screen until you dismiss it. For `pkgz update`:

```
╭─────────────╮
│ pkgz update │
╰─────────────╯

✓   Updating Apt
✓   Updating Flatpak
✗   Updating Pacman
✓   Updating DNF

3 done · 1 failed — finished with errors.

o toggle log · any other key to exit
```

When an app is available from multiple sources (e.g. emacs in both Apt and Flatpak), `pkgz install` shows a picker first so you can choose the source, then runs the install in the same window:

```
╭─────────────────────────╮
│ pkgz install emacs      │
╰─────────────────────────╯

'emacs' is available via multiple sources. Choose one:

▸ Apt
  Flatpak

↑/↓ move · enter select · q quit
```

**TUI controls:**

*During the run:*
- `↑`/`↓` (or `←`/`→`) — select a source / move between entries
- `enter` (or `space`) — confirm a selection (e.g. which source to install from)
- `o` — toggle the captured-output log pane
- `q`, `esc`, or `Ctrl+C` — quit

*On the done/summary screen:*
- `o` — keep toggling the captured-output log pane (does **not** quit)
- `q`, `esc`, `Ctrl+C`, `enter` — or any other key — return to the shell

**Exit codes:** pkgz exits `0` when every operation succeeded and `1` when anything
failed (a rejected source choice, a package not found, a failed update/install,
etc.), so scripts can rely on it.

**Non-TTY fallback:** when stdout is not a terminal (e.g. `pkgz update | tee log`,
pipes, scripts, CI), pkgz falls back to plain console output automatically. If the
console is still interactive, a multi-source `install`/`remove` shows a plain
numbered prompt. If input is not interactive either, pkgz lists the candidate
sources and tells you to pick one with `--source`, so scripts never hang on a
silent prompt:

```
$ pkgz install gimp
⚠️ 'gimp' is available via multiple sources:
1. APT
2. Flatpak
Which one would you like to use? [1-2]: 2
```

---

## 🔐 Privilege Elevation

- Automatically detects and uses `doas` or `sudo`.
- Privileged commands are run with the configured elevator command.

---

## 🧩 Extending Pkgz

To add support for a new package manager:

1. Implement the `Source` interface (see `pkg/sources/interface.go`):
   - `Name() string`
   - `Available(app string) (bool, error)`
   - `Installed(app string) (bool, error)`
   - `Install(app string) error`
   - `Remove(app string) error`
   - `Update() error`
   - `ListUpdates() ([]string, error)`
   - `Search(app string) (bool, error)`
   - `InstalledCount() (int, error)`
2. Optionally implement `SetTask(utils.Task)` on your source so run-time output streams into the TUI (pkgz type-asserts for it via `withTask`; it is not part of the `Source` interface).
3. Add your source to `main.go`'s enabled-sources list and to the config.

---

## 🔄 Migration from Crystal

This project was originally written in Crystal and has moved to Go mostly for compatablity reasons and also because a user is more likely to have Go installed on their system than Crystal so it reduces friction a little.

---

## 📄 License

This project is licensed under the **BSD 2-Clause License**.

See the [LICENSE](LICENSE) file for the full license text.

---

Created by [roguehashrate](https://github.com/roguehashrate)