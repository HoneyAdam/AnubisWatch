# AnubisWatch Homebrew Formula
#
# Install: brew install anubiswatch
# Tap: brew tap AnubisWatch/anubiswatch

class Anubiswatch < Formula
  desc "The Judgment Never Sleeps — Zero-dependency uptime monitoring"
  homepage "https://anubis.watch"
  url "https://github.com/AnubisWatch/anubiswatch/archive/v0.0.1.tar.gz"
  sha256 "CHANGE_ME"
  license "Apache-2.0"

  head "https://github.com/AnubisWatch/anubiswatch.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args, "-ldflags", flags, "./cmd/anubis"
    bin.install "anubis"
  end

  def flags
    %W[
      -s -w
      -X main.Version=#{version}
      -X main.Commit=#{Utils.popen_read("git rev-parse HEAD").chomp}
      -X main.BuildDate=#{Time.now.strftime("%Y-%m-%dT%H:%M:%SZ")}
    ]
  end

  def post_install
    (var/"lib/anubis/data").mkpath
    (etc/"anubis").mkpath
  end

  service do
    run [opt_bin/"anubis", "serve", "--config", etc/"anubis/anubis.yaml"]
    keep_alive true
    working_dir var/"lib/anubis/data"
    log_path var/"log/anubis.log"
    error_log_path var/"log/anubis.error.log"
  end

  test do
    # Version command
    output = shell_output("#{bin}/anubis version")
    assert_match "AnubisWatch", output

    # Init command
    system "#{bin}/anubis", "init", "--config", testpath/"anubis.yaml"
    assert_predicate testpath/"anubis.yaml", :exist?

    # Health check (will fail as no server running, but validates binary)
    assert_raises do
      shell_output("#{bin}/anubis health")
    end
  end
end

__END__

## AnubisWatch — The Judgment Never Sleeps

AnubisWatch is a single-binary, zero-dependency uptime monitoring solution
written in Go 1.26 with an embedded React 19 dashboard.

### Features

- **8 Protocols:** HTTP, TCP, UDP, DNS, SMTP, ICMP, gRPC, WebSocket, TLS
- **Embedded Dashboard:** Beautiful React 19 UI compiled into the binary
- **Raft Consensus:** Distributed clustering with automatic leader election
- **CobaltDB Storage:** Embedded B+Tree database with encryption at rest
- **Alert System:** Slack, Discord, PagerDuty, Email, Webhooks, SMS
- **Status Pages:** Public status pages with custom domains
- **Multi-Tenant:** SaaS-ready workspace isolation

### Quick Start

```bash
# Install
brew install anubiswatch

# Generate config
anubis init

# Start
anubis serve

# Open dashboard
open https://localhost:8443
```

### Configuration

Edit `/opt/homebrew/etc/anubis/anubis.yaml`

### Service Management

```bash
# Start service
brew services start anubiswatch

# Stop service
brew services stop anubiswatch

# Status
brew services info anubiswatch
```

### Links

- Homepage: https://anubis.watch
- Documentation: https://github.com/AnubisWatch/anubiswatch
- Issues: https://github.com/AnubisWatch/anubiswatch/issues
