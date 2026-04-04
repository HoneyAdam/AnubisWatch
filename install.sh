#!/bin/sh
# AnubisWatch Install Script
# ═══════════════════════════════════════════════════════════
#
# Quick install:
#   curl -fsSL https://anubis.watch/install.sh | sh
#
# Or with wget:
#   wget -qO- https://anubis.watch/install.sh | sh
#
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="AnubisWatch/anubiswatch"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
DATA_DIR="${DATA_DIR:-/var/lib/anubis}"
CONFIG_DIR="${CONFIG_DIR:-/etc/anubis}"
SERVICE_NAME="anubis"

# Parse arguments
VERSION=""
SKIP_SERVICE=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --data-dir)
            DATA_DIR="$2"
            shift 2
            ;;
        --config-dir)
            CONFIG_DIR="$2"
            shift 2
            ;;
        --skip-service)
            SKIP_SERVICE=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --help|-h)
            echo "AnubisWatch Install Script"
            echo ""
            echo "Usage:"
            echo "  curl -fsSL https://anubis.watch/install.sh | sh"
            echo "  curl -fsSL https://anubis.watch/install.sh | sh -s -- --version v1.0.0"
            echo ""
            echo "Options:"
            echo "  --version <ver>     Install specific version"
            echo "  --install-dir <dir> Installation directory (default: /usr/local/bin)"
            echo "  --data-dir <dir>    Data directory (default: /var/lib/anubis)"
            echo "  --config-dir <dir>  Config directory (default: /etc/anubis)"
            echo "  --skip-service      Skip systemd service creation"
            echo "  --verbose, -v       Verbose output"
            echo "  --help, -h          Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Helper functions
log_info() {
    printf "${BLUE}ℹ️  %s${NC}\n" "$1"
}

log_success() {
    printf "${GREEN}✓ %s${NC}\n" "$1"
}

log_warning() {
    printf "${YELLOW}⚠️  %s${NC}\n" "$1"
}

log_error() {
    printf "${RED}✗ %s${NC}\n" "$1"
}

log_verbose() {
    if [ "$VERBOSE" = true ]; then
        printf "  %s\n" "$1"
    fi
}

# Check requirements
check_requirements() {
    log_info "Checking requirements..."

    # Check for curl or wget
    if command -v curl >/dev/null 2>&1; then
        DOWNLOAD_CMD="curl -fsSL"
        log_verbose "Using curl for downloads"
    elif command -v wget >/dev/null 2>&1; then
        DOWNLOAD_CMD="wget -qO-"
        log_verbose "Using wget for downloads"
    else
        log_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi

    # Check for sudo
    if [ "$(id -u)" -ne 0 ]; then
        log_error "This script must be run as root (use sudo)"
        exit 1
    fi

    log_success "Requirements check passed"
}

# Detect system architecture
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            GOARCH="amd64"
            ;;
        aarch64|arm64)
            GOARCH="arm64"
            ;;
        armv7l)
            GOARCH="arm"
            ;;
        *)
            log_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    log_verbose "Detected architecture: $ARCH ($GOARCH)"
}

# Detect OS
detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case $OS in
        linux)
            GOOS="linux"
            ;;
        darwin)
            GOOS="darwin"
            ;;
        *)
            log_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac
    log_verbose "Detected OS: $OS ($GOOS)"
}

# Get latest version or use specified
get_version() {
    if [ -n "$VERSION" ]; then
        log_info "Using specified version: $VERSION"
        return
    fi

    log_info "Fetching latest version..."
    VERSION=$($DOWNLOAD_CMD https://api.github.com/repos/$REPO/releases/latest | grep '"tag_name"' | cut -d'"' -f4)

    if [ -z "$VERSION" ]; then
        log_error "Failed to fetch latest version"
        exit 1
    fi

    log_success "Latest version: $VERSION"
}

# Download binary
download_binary() {
    BINARY_NAME="anubis-${GOOS}-${GOARCH}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"

    log_info "Downloading AnubisWatch ${VERSION}..."
    log_verbose "Download URL: $DOWNLOAD_URL"

    TMPFILE=$(mktemp)
    if ! $DOWNLOAD_CMD "$DOWNLOAD_URL" -o "$TMPFILE" 2>/dev/null; then
        log_error "Failed to download binary"
        log_verbose "URL: $DOWNLOAD_URL"
        exit 1
    fi

    log_success "Binary downloaded"
    BINARY_PATH="$TMPFILE"
}

# Install binary
install_binary() {
    log_info "Installing binary to ${INSTALL_DIR}..."

    # Create install directory if needed
    if [ ! -d "$INSTALL_DIR" ]; then
        mkdir -p "$INSTALL_DIR"
        log_verbose "Created directory: $INSTALL_DIR"
    fi

    # Install binary
    install -m 755 "$BINARY_PATH" "${INSTALL_DIR}/anubis"
    log_verbose "Binary installed with executable permissions"

    # Cleanup temp file
    rm -f "$BINARY_PATH"

    log_success "Binary installed to ${INSTALL_DIR}/anubis"
}

# Create user and directories
setup_directories() {
    log_info "Setting up directories..."

    # Create anubis user if it doesn't exist
    if ! id "$SERVICE_NAME" >/dev/null 2>&1; then
        useradd --system --no-create-home --shell /usr/sbin/nologin "$SERVICE_NAME"
        log_verbose "Created system user: $SERVICE_NAME"
    else
        log_verbose "User $SERVICE_NAME already exists"
    fi

    # Create data directory
    if [ ! -d "$DATA_DIR" ]; then
        mkdir -p "$DATA_DIR"
        log_verbose "Created data directory: $DATA_DIR"
    fi
    chown -R "$SERVICE_NAME:$SERVICE_NAME" "$DATA_DIR"
    chmod 750 "$DATA_DIR"

    # Create config directory
    if [ ! -d "$CONFIG_DIR" ]; then
        mkdir -p "$CONFIG_DIR"
        log_verbose "Created config directory: $CONFIG_DIR"
    fi

    log_success "Directories configured"
}

# Generate default config
generate_config() {
    CONFIG_FILE="${CONFIG_DIR}/anubis.yaml"

    if [ -f "$CONFIG_FILE" ]; then
        log_warning "Configuration already exists: $CONFIG_FILE"
        log_warning "Skipping config generation"
        return
    fi

    log_info "Generating default configuration..."

    cat > "$CONFIG_FILE" <<'EOF'
# ⚖️ AnubisWatch Configuration
# The Judgment Never Sleeps

server:
  host: "0.0.0.0"
  port: 8443
  tls:
    enabled: false

storage:
  path: "/var/lib/anubis/data"
  retention_days: 90

# Example soul (monitor)
souls:
  - name: "Example API"
    type: http
    target: "https://httpbin.org/get"
    weight: 60s
    timeout: 10s
    enabled: true
    http:
      method: GET
      valid_status: [200]

# Alert channels (configure as needed)
channels: []

# Alert rules
verdicts:
  rules: []
EOF

    chown "$SERVICE_NAME:$SERVICE_NAME" "$CONFIG_FILE"
    chmod 640 "$CONFIG_FILE"

    log_success "Configuration created: $CONFIG_FILE"
}

# Install systemd service
install_service() {
    if [ "$SKIP_SERVICE" = true ]; then
        log_info "Skipping systemd service installation (--skip-service)"
        return
    fi

    log_info "Installing systemd service..."

    # Get script directory for service file
    SCRIPT_DIR=$(dirname "$0")
    SERVICE_FILE="${SCRIPT_DIR}/anubis.service"

    # If service file doesn't exist in script dir, create it
    if [ ! -f "$SERVICE_FILE" ]; then
        cat > "$SERVICE_FILE" <<'EOF'
[Unit]
Description=AnubisWatch Uptime Monitoring
Documentation=https://github.com/AnubisWatch/anubiswatch
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=anubis
Group=anubis
Environment="ANUBIS_CONFIG=/etc/anubis/anubis.yaml"
Environment="ANUBIS_DATA_DIR=/var/lib/anubis/data"
Environment="ANUBIS_LOG_LEVEL=info"
ExecStart=/usr/local/bin/anubis serve
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=10
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
ReadWritePaths=/var/lib/anubis
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF
    fi

    # Install service file
    cp "$SERVICE_FILE" "/etc/systemd/system/${SERVICE_NAME}.service"
    chmod 644 "/etc/systemd/system/${SERVICE_NAME}.service"

    # Reload systemd
    systemctl daemon-reload

    log_success "Systemd service installed"
}

# Enable and start service
start_service() {
    if [ "$SKIP_SERVICE" = true ]; then
        return
    fi

    log_info "Enabling and starting AnubisWatch service..."

    systemctl enable "$SERVICE_NAME"
    systemctl start "$SERVICE_NAME"

    # Wait for service to start
    sleep 2

    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log_success "AnubisWatch is running"
    else
        log_warning "Service may not have started correctly"
        log_warning "Check logs with: journalctl -u anubis -f"
    fi
}

# Print next steps
print_next_steps() {
    echo ""
    echo "═══════════════════════════════════════════════════════"
    echo ""
    log_success "AnubisWatch installed successfully!"
    echo ""
    echo "Binary location:  ${INSTALL_DIR}/anubis"
    echo "Config file:      ${CONFIG_DIR}/anubis.yaml"
    echo "Data directory:   ${DATA_DIR}"
    echo ""
    echo "Next steps:"
    echo ""
    echo "  1. Edit configuration:"
    echo "     sudo nano ${CONFIG_DIR}/anubis.yaml"
    echo ""
    echo "  2. View service status:"
    echo "     sudo systemctl status anubis"
    echo ""
    echo "  3. View logs:"
    echo "     sudo journalctl -u anubis -f"
    echo ""
    echo "  4. Access dashboard:"
    echo "     http://localhost:8443"
    echo ""
    echo "Commands:"
    echo "  anubis version    - Show version"
    echo "  anubis serve      - Start server"
    echo "  anubis init       - Generate config"
    echo "  anubis watch URL  - Quick-add monitor"
    echo ""
    echo "═══════════════════════════════════════════════════════"
    echo ""
    echo "⚖️  The Judgment Never Sleeps"
    echo ""
}

# Main installation
main() {
    echo ""
    echo "⚖️  AnubisWatch Installation"
    echo "═══════════════════════════════════════════════════════"
    echo ""

    check_requirements
    detect_arch
    detect_os
    get_version
    download_binary
    install_binary
    setup_directories
    generate_config

    # Only install systemd service on Linux
    if [ "$GOOS" = "linux" ] && [ "$SKIP_SERVICE" = false ]; then
        install_service
        start_service
    fi

    print_next_steps
}

# Run installation
main
