#!/bin/sh
set -eu

REPO="Nam-Cheol/namba-ai"
VERSION="${NAMBA_VERSION:-latest}"
INSTALL_DIR="${NAMBA_INSTALL_DIR:-$HOME/.local/bin}"

detect_os() {
    case "$(uname -s)" in
        Linux) printf '%s' "Linux" ;;
        Darwin) printf '%s' "macOS" ;;
        *) printf '%s' "unsupported" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) printf '%s' "x86_64" ;;
        arm64|aarch64) printf '%s' "arm64" ;;
        *) printf '%s' "unsupported" ;;
    esac
}

download() {
    url="$1"
    output="$2"
    if command -v curl >/dev/null 2>&1; then
        if curl -fsSL "$url" -o "$output"; then
            return
        fi
        print_download_error "$url"
        exit 1
    fi
    if command -v wget >/dev/null 2>&1; then
        if wget -qO "$output" "$url"; then
            return
        fi
        print_download_error "$url"
        exit 1
    fi
    printf '%s\n' "curl or wget is required to install NambaAI." >&2
    exit 1
}

print_download_error() {
    url="$1"
    printf '%s\n' "Failed to download $url" >&2
    if [ "$VERSION" = "latest" ]; then
        printf '%s\n' "No GitHub Release has been published yet, or the latest release does not contain $ASSET_NAME." >&2
    else
        printf '%s\n' "Release '$VERSION' was not found, or it does not contain $ASSET_NAME." >&2
    fi
    printf '%s\n' "Common causes: no published release, missing asset, repository access restrictions, or a network error." >&2
    printf '%s\n' "Fallback: go install github.com/Nam-Cheol/namba-ai/cmd/namba@main" >&2
}

append_path() {
    line="export PATH=\"$INSTALL_DIR:\$PATH\""
    profile="$1"
    if [ -f "$profile" ] && grep -F "$line" "$profile" >/dev/null 2>&1; then
        return
    fi
    if [ ! -f "$profile" ]; then
        : > "$profile"
    fi
    printf '\n%s\n' "$line" >> "$profile"
}

OS_NAME="$(detect_os)"
ARCH_NAME="$(detect_arch)"

if [ "$OS_NAME" = "unsupported" ] || [ "$ARCH_NAME" = "unsupported" ]; then
    printf '%s\n' "Unsupported platform: $(uname -s)/$(uname -m)" >&2
    exit 1
fi

ASSET_NAME="namba_${OS_NAME}_${ARCH_NAME}.tar.gz"
if [ "$VERSION" = "latest" ]; then
    DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/$ASSET_NAME"
else
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$ASSET_NAME"
fi

printf '%s\n' "Installing NambaAI from $DOWNLOAD_URL"
mkdir -p "$INSTALL_DIR"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM

ARCHIVE_PATH="$TMP_DIR/$ASSET_NAME"
download "$DOWNLOAD_URL" "$ARCHIVE_PATH"
tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"
install -m 0755 "$TMP_DIR/namba" "$INSTALL_DIR/namba"

case "${SHELL:-}" in
    */zsh)
        append_path "$HOME/.zshrc"
        ;;
    *)
        append_path "$HOME/.profile"
        ;;
esac

case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *) export PATH="$INSTALL_DIR:$PATH" ;;
esac

printf '\n%s\n' "NambaAI installed."
printf '%s\n' "Binary: $INSTALL_DIR/namba"
printf '%s\n' "Command: namba"
printf '%s\n' "If the command is not available in your current shell, run 'exec $SHELL -l' or open a new terminal."

