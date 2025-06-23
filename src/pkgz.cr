# src/pkgz.cr
module Pkgz
  VERSION = "0.1.1"

  # Auto-detect whether to use doas or sudo
  @@elevator = nil
  @@apt_cmd = nil

  def self.privileged(cmd : String) : Nil
    if @@elevator.nil?
      @@elevator = system("which doas > /dev/null 2>&1") ? "doas" : "sudo"
    end
    system("#{@@elevator} #{cmd}")
  end

  def self.apt_command : String
    if @@apt_cmd.nil?
      @@apt_cmd = system("which nala > /dev/null 2>&1") ? "nala" : "apt"
    end
    @@apt_cmd.not_nil!
  end

  # Abstract blueprint for a package source
  abstract class Source
    abstract def name : String
    abstract def available?(app : String) : Bool
    abstract def install(app : String) : Nil
    abstract def remove(app : String) : Nil
    abstract def update : Nil
  end

  # APT/Nala Source implementation
  class AptSource < Source
    def name : String
      Pkgz.apt_command.upcase
    end

    def available?(app : String) : Bool
      output = `apt-cache search #{app}`
      output.includes?(app)
    end

    def install(app : String) : Nil
      Pkgz.privileged("#{Pkgz.apt_command} install -y #{app}")
    end

    def remove(app : String) : Nil
      Pkgz.privileged("#{Pkgz.apt_command} remove -y #{app}")
    end

    def update : Nil
      Pkgz.privileged("#{Pkgz.apt_command} update && #{Pkgz.apt_command} upgrade -y")
    end
  end

  # Flatpak Source implementation
  class FlatpakSource < Source
    def name : String
      "Flatpak"
    end

    def available?(app : String) : Bool
      output = `flatpak search #{app}`
      output.includes?(app)
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
  end

  # App install logic
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
    choice = gets.try &.to_i || 0
    selected = available_sources[choice - 1]?

    if selected
      puts "🚀 Installing with #{selected.name}..."
      selected.install(app)
    else
      puts "❌ Invalid choice."
    end
  end
end

# ─────────────────────────────────────────────
# CLI Entry Point (outside the module)
# ─────────────────────────────────────────────

if ARGV.size < 1
  puts "Usage: pkgz <install|remove|update> [app-name]"
  exit
end

command = ARGV[0]
app_name = ARGV[1]? # may be nil

sources = [Pkgz::AptSource.new, Pkgz::FlatpakSource.new]

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
else
  puts "❓ Unknown command: #{command}"
  puts "Available commands: install, remove, update"
end
