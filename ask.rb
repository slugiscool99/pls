class Ask < Formula
  desc "Ask me anything"
  homepage "https://github.com/slugiscool99/ask"
  url "https://github.com/slugiscool99/ask"
  sha256 "dab75b5a49687e891e22d3355464993d0b4b9bf946bc310dea37fc3a9d74fb67"

  depends_on "bash" # If you need bash for the wrapper script

  def install
    # Install the Go binary
    bin.install "ask"

    # Install the wrapper script
    (bin/"ask_wrapper").write <<~EOS
      #!/bin/bash

      LOG_FILE="$HOME/terminal_output.log"
      MAX_LOG_FILE_SIZE=$((10 * 1024 * 1024)) # 10 MB
      MAX_LOG_LINES=10000

      rotate_log_file() {
        if [ -f "$LOG_FILE" ]; then
          log_file_size=$(stat -c%s "$LOG_FILE")
          if [ "$log_file_size" -ge "$MAX_LOG_FILE_SIZE" ]; then
            archive_path="${LOG_FILE}.$(date +%s)"
            mv "$LOG_FILE" "$archive_path"
            touch "$LOG_FILE"
          fi
        fi
      }

      trim_log_file() {
        if [ -f "$LOG_FILE" ]; then
          line_count=$(wc -l < "$LOG_FILE")
          if [ "$line_count" -gt "$MAX_LOG_LINES" ]; then
            tail -n "$MAX_LOG_LINES" "$LOG_FILE" > "${LOG_FILE}.tmp"
            mv "${LOG_FILE}.tmp" "$LOG_FILE"
          fi
        fi
      }

      rotate_log_file
      trim_log_file
      "#{bin}/ask" "$@" &> >(tee -a "$LOG_FILE")
    EOS
    bin.install "ask_wrapper"
    bin.install_symlink "ask_wrapper" => "ask"

    # Modify shell configuration files
    bashrc = File.expand_path("~/.bashrc")
    zshrc = File.expand_path("~/.zshrc")

    if File.exist?(bashrc)
      unless File.readlines(bashrc).grep(/log_terminal_output/).any?
        File.open(bashrc, "a") do |file|
          file.puts "\nlog_terminal_output() {"
          file.puts "  LOG_FILE=\"$HOME/terminal_output.log\""
          file.puts "  script -q -a -c \"$BASH_COMMAND\" \"$LOG_FILE\""
          file.puts "}"
          file.puts "trap 'log_terminal_output' DEBUG\n"
        end
      end
    end

    if File.exist?(zshrc)
      unless File.readlines(zshrc).grep(/log_terminal_output/).any?
        File.open(zshrc, "a") do |file|
          file.puts "\nlog_terminal_output() {"
          file.puts "  LOG_FILE=\"$HOME/terminal_output.log\""
          file.puts "  script -q -a -c \"$ZSH_COMMAND\" \"$LOG_FILE\""
          file.puts "}"
          file.puts "trap 'log_terminal_output' DEBUG\n"
        end
      end
    end
  end

  def caveats
    <<~EOS
      To finish the install, start a new terminal session or run:
        source ~/.bashrc
        source ~/.zshrc
    EOS
  end

  def post_uninstall
    # Define the uninstallation script
    uninstall_script = <<~EOS
      #!/bin/bash

      # Remove logging from .bashrc
      sed -i '' '/log_terminal_output/d' ~/.bashrc
      sed -i '' '/trap '\''log_terminal_output'\'' DEBUG/d' ~/.bashrc

      # Remove logging from .zshrc
      sed -i '' '/log_terminal_output/d' ~/.zshrc
      sed -i '' '/trap '\''log_terminal_output'\'' DEBUG/d' ~/.zshrc

      echo "Logging mechanism removed. Please restart your terminal."
    EOS

    # Write and execute the uninstallation script
    File.write("/tmp/uninstall_logging.sh", uninstall_script)
    system "bash", "/tmp/uninstall_logging.sh"
  end
end
