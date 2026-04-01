#!/bin/bash

set -e

VERSION="${VERSION:-dev}"
BUILD_DIR="${BUILD_DIR:-./dist}"
MAIN_PACKAGE="./cmd"

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

clean() {
    log_info "Cleaning build directory..."
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
}

build() {
    local os=$1
    local arch=$2
    local output_name="cloud-swap-${os}-${arch}"
    
    log_info "Building for $os/$arch..."
    
    ext=""
    if [[ "$os" == "windows" ]]; then
        ext=".exe"
        output_name="${output_name}.exe"
    fi
    
    GOOS="$os" GOARCH="$arch" go build \
        -ldflags "-s -w -X main.version=$VERSION" \
        -o "${BUILD_DIR}/${output_name}" \
        "$MAIN_PACKAGE"
    
    if [[ -f "${BUILD_DIR}/${output_name}" ]]; then
        local size=$(du -h "${BUILD_DIR}/${output_name}" | cut -f1)
        log_info "Built: ${output_name} ($size)"
    else
        log_error "Failed to build ${output_name}"
        return 1
    fi
}

build_all() {
    clean
    
    log_info "Building cloud-swap $VERSION for all platforms..."
    echo ""
    
    local platforms=(
        "darwin amd64"
        "darwin arm64"
        "linux amd64"
        "linux arm64"
        "linux armv7"
        "windows amd64"
    )
    
    for platform in "${platforms[@]}"; do
        build $platform
    done
    
    echo ""
    log_info "Build complete! Output in $BUILD_DIR"
}

build_local() {
    log_info "Building for local system..."
    
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $arch in
        x86_64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
    esac
    
    build "$os" "$arch"
}

package() {
    log_info "Creating release packages..."
    
    cd "$BUILD_DIR"
    
    if [[ -f "cloud-swap-linux-amd64" ]]; then
        tar -czvf cloud-swap-linux-amd64.tar.gz cloud-swap-linux-amd64
    fi
    
    if [[ -f "cloud-swap-darwin-amd64" ]]; then
        tar -czvf cloud-swap-darwin-amd64.tar.gz cloud-swap-darwin-amd64
    fi
    
    if [[ -f "cloud-swap-darwin-arm64" ]]; then
        tar -czvf cloud-swap-darwin-arm64.tar.gz cloud-swap-darwin-arm64
    fi
    
    log_info "Packages created:"
    ls -la *.tar.gz 2>/dev/null || true
}

print_help() {
    cat << 'EOF'
cloud-swap Build Script

Usage: 
    ./build.sh [COMMAND]

Commands:
    all         Build for all platforms (default)
    local       Build for local system
    clean       Clean build directory
    package     Create release tarballs
    help        Show this help

Environment:
    VERSION     Set version string (default: dev)
    BUILD_DIR   Set output directory (default: ./dist)

Examples:
    ./build.sh all           # Build for all platforms
    VERSION=v1.0.0 ./build.sh all  # Build with version
    ./build.sh local         # Build for current OS/arch

EOF
}

main() {
    local command="${1:-all}"
    
    case $command in
        all)
            build_all
            ;;
        local)
            build_local
            ;;
        clean)
            clean
            ;;
        package)
            package
            ;;
        help|--help|-h)
            print_help
            ;;
        *)
            log_error "Unknown command: $command"
            print_help
            exit 1
            ;;
    esac
}

main "$@"