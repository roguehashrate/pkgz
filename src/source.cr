require "http/client"
require "uri"
require "file_utils"
require "./pkgz"

module Pkgz
  class PkgzFileSource < Source
    BASE_URL = "https://raw.githubusercontent.com/roguehashrate/pkgz_repo/main/common/"

    def name : String
      "Pkgz File"
    end

    def available?(app : String) : Bool
      local_path = "./#{app}.pkgz"
      remote_url = BASE_URL + "#{app}.pkgz"
      File.exists?(local_path) || remote_file_exists?(remote_url)
    end

    def install(app : String) : Nil
      puts "📦 Installing from .pkgz file..."

      pkgz_data = if File.exists?("#{app}.pkgz")
                    File.read("#{app}.pkgz")
                  else
                    fetch_remote_pkgz_file(app)
                  end

      begin
        parsed = TOML.parse(pkgz_data)
        name = parsed["name"].as_s
        version = parsed["version"]?.try(&.as_s) || "unknown"
        maintainer = parsed["maintainer"]?.try(&.as_s) || "unknown"

        puts "🛠️  Installing #{name} v#{version} by #{maintainer}"

        binaries_section = parsed["binaries"]?.try(&.as_h)
        binary_entry = binaries_section.try { |b| b["linux_x86_64"]? }
        binary_url = binary_entry.try(&.as_s)

        if binary_url && binary_url.ends_with?(".AppImage")
          home = ENV["HOME"]?
          unless home
            puts "❌ Cannot determine HOME directory."
            return
          end

          apps_dir = File.join(home, ".local", "bin", "pkgz_apps")
          bin_dir = File.join(home, ".local", "bin")

          FileUtils.mkdir_p(apps_dir)
          FileUtils.mkdir_p(bin_dir)

          filename = File.basename(URI.parse(binary_url).path)
          destination = File.join(apps_dir, filename)
          symlink_path = File.join(bin_dir, name.downcase)

          puts "⬇️  Downloading AppImage #{filename} to #{apps_dir}..."
          system("curl -L -o #{destination} #{binary_url}")

          puts "🔧 Making AppImage executable..."
          system("chmod +x #{destination}")

          if File.exists?(symlink_path)
            safe_delete(symlink_path)
          end

          puts "🔗 Creating symlink #{symlink_path} -> #{destination}"
          FileUtils.ln_s(destination, symlink_path)

          puts "✅ Installed #{name} successfully."
        else
          puts "❌ No suitable AppImage binary found for this platform."
        end
      rescue ex
        puts "❌ Failed to parse .pkgz file: #{ex.message}"
      end
    end

    def remove(app : String) : Nil
      home = ENV["HOME"]?
      unless home
        puts "❌ Cannot determine HOME directory."
        return
      end

      apps_dir = File.join(home, ".local", "bin", "pkgz_apps")
      bin_dir = File.join(home, ".local", "bin")

      pkgz_data = if File.exists?("#{app}.pkgz")
                    File.read("#{app}.pkgz")
                  else
                    begin
                      fetch_remote_pkgz_file(app)
                    rescue ex
                      puts "❌ Could not fetch .pkgz for removal: #{ex.message}"
                      return
                    end
                  end

      begin
        parsed = TOML.parse(pkgz_data)
        name = parsed["name"].as_s

        binaries_section = parsed["binaries"]?.try(&.as_h)
        binary_entry = binaries_section.try { |b| b["linux_x86_64"]? }
        binary_url = binary_entry.try(&.as_s)

        if binary_url && binary_url.ends_with?(".AppImage")
          filename = File.basename(URI.parse(binary_url).path)
          destination = File.join(apps_dir, filename)
          symlink_path = File.join(bin_dir, name.downcase)

          puts "🗑️  Removing AppImage #{destination} and symlink #{symlink_path}"

          safe_delete(destination)
          safe_delete(symlink_path)

          puts "✅ Removed #{name}."
        else
          puts "❌ No AppImage binary found for removal."
        end
      rescue ex
        puts "❌ Failed to parse .pkgz file for removal: #{ex.message}"
      end
    end

    def update : Nil
      puts "🔄 Update not supported for .pkgz sources yet."
    end

    def search(app : String) : Bool
      available?(app)
    end

    private def fetch_remote_pkgz_file(app : String) : String
      url = BASE_URL + "#{app}.pkgz"
      response = HTTP::Client.get(url)
      if response.status_code == 200
        response.body
      else
        raise "Could not fetch remote .pkgz file from #{url}"
      end
    rescue ex
      raise "HTTP error: #{ex.message}"
    end

    private def remote_file_exists?(url : String) : Bool
      response = HTTP::Client.head(url)
      response.status_code == 200
    rescue
      false
    end

    private def safe_delete(path : String)
      File.delete(path) if File.exists?(path)
    rescue
      # Ignore errors
    end
  end
end
