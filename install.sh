#!/bin/sh

set -eu

OWNER=sdcio
REPO=kubectl-sdc
BINARY=kubectl-sdc
COMPLETION_BINARY=kubectl_complete-sdc

usage() {
    cat <<EOF
Install ${BINARY} from GitHub releases.

The installer also installs ${COMPLETION_BINARY}, which is required for shell completion.

Environment variables:
  VERSION      Release tag to install (default: latest)
  INSTALL_DIR  Destination directory (default: ~/.local/bin)

Examples:
  curl -fsSL https://raw.githubusercontent.com/${OWNER}/${REPO}/main/install.sh | sh
  curl -fsSL https://raw.githubusercontent.com/${OWNER}/${REPO}/main/install.sh | VERSION=v0.1.3 INSTALL_DIR=/usr/local/bin sh
EOF
}
log() {
    printf "%s\n" "$*" >&2
}

require_cmd() {
    if ! command -v "$1" >/dev/null 2>&1; then
        log "missing required command: $1"
        exit 1
    fi
}

detect_os() {
    case "$(uname -s)" in
        Linux)
            printf 'Linux'
            ;;
        Darwin)
            printf 'Darwin'
            ;;
        *)
            log "unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)
            printf 'amd64'
            ;;
        aarch64|arm64)
            printf 'aarch64'
            ;;
        *)
            log "unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
}

resolve_version() {
    requested_version=${VERSION:-latest}
    if [ "$requested_version" != "latest" ]; then
        printf '%s' "$requested_version"
        return
    fi

    require_cmd curl
    curl -fsSL "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" |
        sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' |
        head -n 1
}

download() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$1" -o "$2"
        return
    fi
    if command -v wget >/dev/null 2>&1; then
        wget -qO "$2" "$1"
        return
    fi

    log "missing required command: curl or wget"
    exit 1
}

if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
    usage
    exit 0
fi

require_cmd tar
require_cmd mktemp
require_cmd uname

os=$(detect_os)
arch=$(detect_arch)
tag=$(resolve_version)

if [ -z "$tag" ]; then
    log "failed to resolve release version"
    exit 1
fi

version=${tag#v}
install_dir=${INSTALL_DIR:-${HOME}/.local/bin}
archive_name="${REPO}_${version}_${os}_${arch}.tar.gz"
archive_url="https://github.com/${OWNER}/${REPO}/releases/download/${tag}/${archive_name}"

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT INT TERM

archive_path="${tmpdir}/${archive_name}"

log "downloading ${archive_url}"
download "$archive_url" "$archive_path"

mkdir -p "$install_dir"
tar -xzf "$archive_path" -C "$tmpdir"
cp "${tmpdir}/${BINARY}" "${install_dir}/${BINARY}"
cp "${tmpdir}/${COMPLETION_BINARY}" "${install_dir}/${COMPLETION_BINARY}"
chmod 0755 "${install_dir}/${BINARY}"
chmod 0755 "${install_dir}/${COMPLETION_BINARY}"

log "installed ${BINARY} ${tag} to ${install_dir}/${BINARY}"
log "installed ${COMPLETION_BINARY} ${tag} to ${install_dir}/${COMPLETION_BINARY}"
log "ensure ${install_dir} is in PATH"