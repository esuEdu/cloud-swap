#!/bin/bash

set -e

VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.cloud-swap/bin}"
CONFIG_DIR="${CONFIG_DIR:-$HOME/.cloud-swap}"
BINARY_NAME="cloud-swap"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "darwin"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    else
        echo "$OSTYPE"
    fi
}

detect_arch() {
    architecture=$(uname -m)
    case $architecture in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        armv7)
            echo "armv7"
            ;;
        *)
            echo "$architecture"
            ;;
    esac
}

download_binary() {
    local version=$1
    local os=$2
    local arch=$3
    local temp_dir=$(mktemp -d)
    
    log_info "Downloading cloud-swap ${version} for ${os}/${arch}..."
    
    local download_url
    if [[ "$version" == "latest" ]]; then
        download_url="https://github.com/esuEdu/cloud-swap/releases/latest/download/cloud-swap-${os}-${arch}"
    else
        download_url="https://github.com/esuEdu/cloud-swap/releases/download/${version}/cloud-swap-${os}-${arch}"
    fi
    
    if command -v curl &> /dev/null; then
        curl -fsSL "$download_url" -o "${temp_dir}/${BINARY_NAME}" || {
            log_error "Download failed. Trying to build from source..."
            rm -rf "$temp_dir"
            return 1
        }
    elif command -v wget &> /dev/null; then
        wget -q "$download_url" -O "${temp_dir}/${BINARY_NAME}" || {
            log_error "Download failed. Trying to build from source..."
            rm -rf "$temp_dir"
            return 1
        }
    else
        log_error "Neither curl nor wget found. Please install one of them."
        rm -rf "$temp_dir"
        return 1
    fi
    
    chmod +x "${temp_dir}/${BINARY_NAME}"
    echo "$temp_dir"
}

build_from_source() {
    log_info "Building cloud-swap from source..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.24+ first."
        exit 1
    fi
    
    local temp_dir=$(mktemp -d)
    cd "$temp_dir"
    
    if [[ ! -d "$HOME/go/src/github.com/esuEdu/cloud-swap" ]]; then
        mkdir -p "$HOME/go/src/github.com/esuEdu"
        cd "$HOME/go/src/github.com/esuEdu"
        git clone https://github.com/esuEdu/cloud-swap.git || {
            log_error "Failed to clone repository"
            exit 1
        }
        cd cloud-swap
    else
        cd "$HOME/go/src/github.com/esuEdu/cloud-swap"
        git pull origin develop 2>/dev/null || true
    fi
    
    go build -ldflags "-s -w" -o "${temp_dir}/${BINARY_NAME}" ./cmd
    
    if [[ ! -f "${temp_dir}/${BINARY_NAME}" ]]; then
        log_error "Build failed"
        exit 1
    fi
    
    echo "$temp_dir"
}

install_binary() {
    local binary_path=$1
    
    mkdir -p "$INSTALL_DIR"
    cp "$binary_path/$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    log_info "Installed to $INSTALL_DIR/$BINARY_NAME"
}

setup_completions() {
    local shell="$1"
    local completion_file="$CONFIG_DIR/completions/${shell}.sh"
    
    mkdir -p "$CONFIG_DIR/completions"
    
    case $shell in
        bash)
            "$INSTALL_DIR/$BINARY_NAME" completion bash > "$completion_file"
            ;;
        zsh)
            "$INSTALL_DIR/$BINARY_NAME" completion zsh > "$completion_file"
            ;;
        fish)
            "$INSTALL_DIR/$BINARY_NAME" completion fish > "$completion_file"
            ;;
    esac
    
    log_info "Generated ${shell} completions to $completion_file"
}

add_to_path() {
    local shell="$1"
    local path_entry="export PATH=\"$INSTALL_DIR:\$PATH\""
    
    case $shell in
        bash)
            local bashrc="$HOME/.bashrc"
            if [[ -f "$HOME/.bash_profile" ]]; then
                bashrc="$HOME/.bash_profile"
            fi
            
            if ! grep -q "$INSTALL_DIR" "$bashrc" 2>/dev/null; then
                echo "" >> "$bashrc"
                echo "# cloud-swap" >> "$bashrc"
                echo "$path_entry" >> "$bashrc"
                log_info "Added to $bashrc"
            fi
            ;;
        zsh)
            local zshrc="$HOME/.zshrc"
            if [[ ! -f "$zshrc" ]]; then
                touch "$zshrc"
            fi
            
            if ! grep -q "$INSTALL_DIR" "$zshrc" 2>/dev/null; then
                echo "" >> "$zshrc"
                echo "# cloud-swap" >> "$zshrc"
                echo "$path_entry" >> "$zshrc"
                log_info "Added to $zshrc"
            fi
            ;;
    esac
}

init_config() {
    mkdir -p "$CONFIG_DIR"
    
    if [[ ! -f "$CONFIG_DIR/config.json" ]]; then
        cat > "$CONFIG_DIR/config.json" << 'EOF'
{
  "default_region": "us-east-1",
  "default_output": "json",
  "fzf_enabled": true
}
EOF
        log_info "Created default config at $CONFIG_DIR/config.json"
    fi
}

setup_shell_integration() {
    local shell="${1:-bash}"
    
    setup_completions "$shell"
    add_to_path "$shell"
    
    log_info "Shell integration complete for $shell"
}

uninstall() {
    log_info "Uninstalling cloud-swap..."
    
    rm -f "$INSTALL_DIR/$BINARY_NAME"
    rm -rf "$CONFIG_DIR"
    
    log_info "Uninstalled successfully"
}

print_help() {
    cat << 'EOF'
cloud-swap Installer

Usage: 
    ./install.sh [OPTIONS]

Options:
    --version VERSION     Install specific version (default: latest)
    --install-dir DIR     Install to custom directory (default: ~/.cloud-swap/bin)
    --config-dir DIR     Config directory (default: ~/.cloud-swap)
    --shell SHELL        Setup shell integration: bash, zsh, or fish
    --source             Build from source instead of downloading
    --uninstall          Uninstall cloud-swap
    --help               Show this help message

Examples:
    ./install.sh                          # Install latest version
    ./install.sh --version v1.0.0         # Install specific version
    ./install.sh --shell bash             # Install with bash integration
    ./install.sh --source                 # Build from source
    ./install.sh --uninstall              # Uninstall

EOF
}

main() {
    local install_source=false
    local shell=""
    
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
            --config-dir)
                CONFIG_DIR="$2"
                shift 2
                ;;
            --shell)
                shell="$2"
                shift 2
                ;;
            --source)
                install_source=true
                shift
                ;;
            --uninstall)
                uninstall
                exit 0
                ;;
            --help)
                print_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                print_help
                exit 1
                ;;
        esac
    done
    
    echo ""
    echo "========================================"
    echo "  cloud-swap Installer"
    echo "========================================"
    echo ""
    
    local os=$(detect_os)
    local arch=$(detect_arch)
    local temp_dir=""
    
    if [[ "$install_source" == true ]]; then
        temp_dir=$(build_from_source)
    else
        temp_dir=$(download_binary "$VERSION" "$os" "$arch") || {
            log_warn "Download failed. Building from source..."
            temp_dir=$(build_from_source)
        }
    fi
    
    install_binary "$temp_dir"
    init_config
    
    if [[ -n "$shell" ]]; then
        setup_shell_integration "$shell"
    else
        echo ""
        log_info "Setup shell integration? (bash/zsh/fish/n)"
        read -r response
        case $response in
            bash|zsh|fish)
                setup_shell_integration "$response"
                ;;
            *)
                log_info "Skipping shell integration"
                ;;
        esac
    fi
    
    rm -rf "$temp_dir"
    
    echo ""
    echo "========================================"
    log_info "Installation complete!"
    echo "========================================"
    echo ""
    echo "Add to PATH:"
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    echo ""
    echo "Then run:"
    echo "  cloud-swap --help"
    echo ""
}

main "$@"