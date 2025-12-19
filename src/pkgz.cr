require "toml"

module Pkgz
  VERSION = "0.1.5"
  CONFIG_PATH = "#{ENV["HOME"]}/.config/pkgz/config.toml"
  @@elevator : String? = nil

  def self.get_elevator_command : String
    return @@elevator.not_nil! if @@elevator

    if File.exists?(CONFIG_PATH)
      begin
        config = TOML.parse(File.read(CONFIG_PATH))
        elevator_section = config["elevator"]?.try(&.as_h)
        if elevator_section
          raw_cmd = elevator_section["command"]?.try(&.as_s)
          cmd = raw_cmd ? raw_cmd.strip : nil
          if cmd && !cmd.empty?
            @@elevator = cmd
            return @@elevator.not_nil!
          end
        end
      rescue
        # ignore errors
      end
    end

    # fallback detection
    @@elevator = system("which doas > /dev/null 2>&1") ? "doas" : "sudo"
    @@elevator.not_nil!
  end

  def self.privileged(cmd : String) : Nil
    elevator = get_elevator_command
    system("#{elevator} #{cmd}")
  end

  def self.load_config : Hash(String, Bool)
    unless File.exists?(CONFIG_PATH)
      puts "❌ Config file not found at #{CONFIG_PATH}"
      puts "Please create it manually with the sources you want enabled."
      puts <<-TOML
        [sources]
        apt = true
        nala = false
        flatpak = true
        paru = false
        pacman = false
        dnf = false
        pacstall = true

        [elevator]
        command = "sudo"
      TOML
      exit 1
    end

    config = TOML.parse(File.read(CONFIG_PATH))
    config_sources = config["sources"]?.try(&.as_h) || {} of String => TOML::Any
    config_sources.transform_values(&.as_bool)
  end

  abstract class Source
    abstract def name : String
    abstract def available?(app : String) : Bool
    abstract def install(app : String) : Nil
    abstract def remove(app : String) : Nil
    abstract def update : Nil
    abstract def search(app : String) : Bool
  end

  class AptSource < Source
    def name : String
      "Apt"
    end

    def available?(app : String) : Bool
      `apt-cache search #{app}`.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("apt install -y #{app}")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("apt remove -y #{app}")
    end

    def update : Nil
      Pkgz.privileged("sh -c \"apt update && apt upgrade -y\"")
    end

    def search(app : String) : Bool
      `apt-cache search #{app}`.downcase.includes?(app.downcase)
    end
  end

  class NalaSource < Source
    def name : String
      "Nala"
    end

    def available?(app : String) : Bool
      `nala search #{app}`.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("nala install -y #{app}")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("nala remove -y #{app}")
    end

    def update : Nil
      Pkgz.privileged("nala update && nala upgrade -y")
    end

    def search(app : String) : Bool
      `nala search #{app}`.downcase.includes?(app.downcase)
    end
  end

  class FlatpakSource < Source
    def name : String
      "Flatpak"
    end

    def available?(app : String) : Bool
      `flatpak search #{app}`.includes?(app)
    end

    def install(app : String) : Nil
      system("flatpak install -y #{app}")
    end

    def remove(app : String) : Nil
      system("flatpak uninstall -y #{app}")
    end

    def update : Nil
      system("flatpak update -y")
    end

    def search(app : String) : Bool
      `flatpak search #{app}`.downcase.includes?(app.downcase)
    end
  end

  class PacmanSource < Source
    def name : String
      "Pacman"
    end

    def available?(app : String) : Bool
      `pacman -Ss #{app}`.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("pacman -S --noconfirm #{app}")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("pacman -R --noconfirm #{app}")
    end

    def update : Nil
      Pkgz.privileged("pacman -Syu --noconfirm")
    end

    def search(app : String) : Bool
      `pacman -Ss #{app}`.downcase.includes?(app.downcase)
    end
  end

  class ParuSource < Source
    def name : String
      "Paru (AUR)"
    end

    def available?(app : String) : Bool
      `paru -Ss #{app}`.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("paru -S --noconfirm #{app}")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("paru -R --noconfirm #{app}")
    end

    def update : Nil
      Pkgz.privileged("paru -Syu --noconfirm")
    end

    def search(app : String) : Bool
      `paru -Ss #{app}`.downcase.includes?(app.downcase)
    end
  end

  class YaySource < Source
    def name : String
      "Yay (AUR)"
    end

    def available?(app : String) : Bool
      `yay -Ss #{app}`.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("yay -S --noconfirm #{app}")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("yay -R --noconfirm #{app}")
    end

    def update : Nil
      Pkgz.privileged("yay -Syu --noconfirm")
    end

    def search(app : String) : Bool
      `yay -Ss #{app}`.downcase.includes?(app.downcase)
    end
  end

  class DnfSource < Source
    def name : String
      "DNF"
    end

    def available?(app : String) : Bool
      `dnf search #{app}`.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("dnf install -y #{app}")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("dnf remove -y #{app}")
    end

    def update : Nil
      Pkgz.privileged("dnf upgrade -y")
    end

    def search(app : String) : Bool
      `dnf search #{app}`.downcase.includes?(app.downcase)
    end
  end

  class PacstallSource < Source
    def name : String
      "Pacstall"
    end

    def available?(app : String) : Bool
      `pacstall -S #{app}`.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("pacstall -I #{app}")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("pacstall -R #{app}")
    end

    def update : Nil
      Pkgz.privileged("pacstall -Up")
    end

    def search(app : String) : Bool
      `pacstall -S #{app}`.downcase.includes?(app.downcase)
    end
  end

  class ApkSource < Source
    def name : String
      "Alpine"
    end

    def available?(app : String) : Bool
      `apk search #{app}`.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("apk add #{app}")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("apk del #{app}")
    end

    def update : Nil
      Pkgz.privileged("apk update && apk upgrade")
    end

    def search(app : String) : Bool
      `apk search #{app}`.downcase.includes?(app.downcase)
    end
  end

  class ZypperSource < Source
    def name : String
      "Zypper"
    end

    def available?(app : String) : Bool
      `zypper search #{app}`.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("zypper install -y #{app}")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("zypper remove -y #{app}")
    end

    def update : Nil
      Pkgz.privileged("zypper refresh && zypper update -y")
    end

    def search(app : String) : Bool
      `zypper search #{app}`.downcase.includes?(app.downcase)
    end
  end

  class XbpsSource < Source
    def name : String
      "XBPS"
    end

    def available?(app : String) : Bool
      `xbps-query -Rs #{app}`.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("xbps-install -Sy #{app}")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("xbps-remove -Ry #{app}")
    end

    def update : Nil
      Pkgz.privileged("xbps-install -Syu")
    end

    def search(app : String) : Bool
      `xbps-query -Rs #{app}`.downcase.includes?(app.downcase)
    end
  end

  class FreeBsdSource < Source
    def name : String
      "FreeBSD"
    end

    def available?(app : String) : Bool
      result = %x(pkg search #{app})
      result.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("pkg install -y #{app}")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("pkg delete -y #{app}")
    end

    def update : Nil
      Pkgz.privileged("pkg update")
      Pkgz.privileged("pkg upgrade -y")
    end

    def search(app : String) : Bool
      available?(app)
    end
  end

  class OpenBsdSource < Source
    def name : String
      "OpenBSD"
    end

    def available?(app : String) : Bool
      result = %x(pkg_info | grep #{app})
      result.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("pkg_add #{app}")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("pkg_delete #{app}")
    end

    def update : Nil
      puts "Please use `syspatch` or ports tree for updates on OpenBSD."
    end

    def search(app : String) : Bool
      available?(app)
    end
  end

  class FreeBsdPortsSource < Source
    def name : String
      "FreeBSD Ports"
    end

    def available?(app : String) : Bool
      result = %x(make -C /usr/ports/ search key=#{app} 2>/dev/null)
      result.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("make -C /usr/ports/#{app} install clean BATCH=yes")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("pkg delete -y #{app}")
    end

    def update : Nil
      puts "Update ports tree manually with 'portsnap fetch update' or via git."
    end

    def search(app : String) : Bool
      available?(app)
    end
  end

  class OpenBsdPortsSource < Source
    def name : String
      "OpenBSD Ports"
    end

    def available?(app : String) : Bool
      File.directory?("/usr/ports/#{app}")
    end

    def install(app : String) : Nil
      Pkgz.privileged("cd /usr/ports/#{app} && make install clean")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("pkg_delete #{app}")
    end

    def update : Nil
      puts "Update ports tree manually via 'svn update /usr/ports' or git."
    end

    def search(app : String) : Bool
      available?(app)
    end
  end

  def self.find_and_install(app : String, sources : Array(Source))
    puts "🔍 Searching for '#{app}' in sources..."

    available_sources = sources.select { |s| s.available?(app) }

    if available_sources.empty?
      puts "❌ App '#{app}' not found in any source."
      return
    end

    if available_sources.size == 1
      source = available_sources.first
      puts "✅ Found '#{app}' in #{source.name}. Installing..."
      source.install(app)
      return
    end

    puts "📦 Found '#{app}' in multiple sources:"
    available_sources.each_with_index do |source, i|
      puts "#{i + 1}. #{source.name}"
    end

    print "Which one would you like to use? [1-#{available_sources.size}]: "
    choice = gets.try(&.to_i) || 0
    selected = available_sources[choice - 1]?

    if selected
      puts "🚀 Installing with #{selected.name}..."
      selected.install(app)
    else
      puts "❌ Invalid choice."
    end
  end
end

# CLI Entry Point
if ARGV.size < 1
  puts "Usage: pkgz <install|remove|update|search|--version> [app-name]"
  exit
end

command = ARGV[0]
app_name = ARGV[1]?


if command == "--version"
  puts "pkgz version #{Pkgz::VERSION}"
  exit
end

enabled_sources = Pkgz.load_config
sources = [] of Pkgz::Source

# Linux Sources
sources << Pkgz::AptSource.new       if enabled_sources["apt"]?
sources << Pkgz::NalaSource.new      if enabled_sources["nala"]?
sources << Pkgz::FlatpakSource.new   if enabled_sources["flatpak"]?
sources << Pkgz::PacmanSource.new    if enabled_sources["pacman"]?
sources << Pkgz::ParuSource.new      if enabled_sources["paru"]?
sources << Pkgz::YaySource.new       if enabled_sources["yay"]?
sources << Pkgz::DnfSource.new       if enabled_sources["dnf"]?
sources << Pkgz::ZypperSource.new    if enabled_sources["zypper"]?
sources << Pkgz::XbpsSource.new      if enabled_sources["xbps"]?
sources << Pkgz::ApkSource.new       if enabled_sources["alpine"]?
sources << Pkgz::PacstallSource.new  if enabled_sources["pacstall"]?

# BSD Sources
sources << Pkgz::FreeBsdSource.new      if enabled_sources["freebsd"]?
sources << Pkgz::FreeBsdPortsSource.new if enabled_sources["freebsd_ports"]?
sources << Pkgz::OpenBsdSource.new      if enabled_sources["openbsd"]?
sources << Pkgz::OpenBsdPortsSource.new if enabled_sources["openbsd_ports"]?

case command
when "install"
  if app_name
    Pkgz.find_and_install(app_name, sources)
  else
    puts "Usage: pkgz install <app-name>"
  end
when "remove"
  if app_name
    sources.each do |source|
      puts "❌ Trying to remove '#{app_name}' from #{source.name}..."
      source.remove(app_name)
    end
  else
    puts "Usage: pkgz remove <app-name>"
  end
when "update"
  sources.each do |source|
    puts "⬆️  Updating #{source.name} packages..."
    source.update
  end
when "search"
  if app_name
    puts "🔍 Searching for '#{app_name}' across enabled sources..."
    any_found = false

    sources.each do |source|
      if source.search(app_name)
        puts "✅ Found in #{source.name}"
        any_found = true
      else
        puts "❌ Not found in #{source.name}"
      end
    end

    puts "📦 Package '#{app_name}' not found in any enabled source." unless any_found
  else
    puts "Usage: pkgz search <app-name>"
  end
when "clean"
  sources.each do |source|
    case source
    when Pkgz::AptSource
      puts "🧹 Cleaning Apt cache..."
      Pkgz.privileged("apt clean")
    when Pkgz::NalaSource
      puts "🧹 Cleaning Apt cache with Nala..."
      Pkgz.privileged("nala clean")
    when Pkgz::PacmanSource
      puts "🧹 Cleaning Pacman cache..."
      Pkgz.privileged("pacman -Sc --noconfirm")
    when Pkgz::ParuSource
      puts "🧹 Cleaning Paru cache..."
      Pkgz.privileged("paru -Sc --noconfirm")
    when Pkgz::YaySource
      puts "🧹 Cleaning Paru cache..."
      Pkgz.privileged("yay -Sc --noconfirm")
    when Pkgz::DnfSource
      puts "🧹 Cleaning DNF cache..."
      Pkgz.privileged("dnf clean all")
    when Pkgz::FlatpakSource
      puts "🧹 Cleaning Flatpak cache..."
      system("flatpak uninstall --unused -y")
    when Pkgz::ApkSource
      puts "🧹 Cleaning Alpine cache..."
      Pkgz.privileged("rm -rf /var/cache/apk/*")
    when Pkgz::XbpsSource
      puts "🧹 Cleaning XBPS cache..."
      Pkgz.privileged("xbps-remove -O")
    when Pkgz::ZypperSource
      puts "🧹 Cleaning Zypper cache..."
      Pkgz.privileged("zypper clean --all")
    when Pkgz::PacstallSource
      puts "🧹 Cleaning Pacstall cache..."
      Pkgz.privileged("pacstall -C")
    when Pkgz::FreeBsdSource, Pkgz::FreeBsdPortsSource,
         Pkgz::OpenBsdSource, Pkgz::OpenBsdPortsSource
      puts "⚠️  No automatic clean command available for #{source.name}."
    else
      puts "⚠️  Unknown source #{source.name}, skipping clean."
    end
  end
else
  puts "Unknown command: #{command}"
  puts "Usage: pkgz <install|remove|update|search|clean|--version> [app-name]"
end
