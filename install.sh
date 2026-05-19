#!/bin/sh
set -eu

REPO="Nam-Cheol/namba-ai"
VERSION="${NAMBA_VERSION:-latest}"
INSTALL_DIR="${NAMBA_INSTALL_DIR:-$HOME/.local/bin}"

# Test-only overrides used by installer regression tests. They are local-file
# inputs only and do not provide a checksum bypass for normal installs.
TEST_ASSET_PATH="${NAMBA_INSTALL_TEST_ASSET_PATH:-}"
TEST_CHECKSUMS_PATH="${NAMBA_INSTALL_TEST_CHECKSUMS_PATH:-}"
TEST_INSTALL_DIR="${NAMBA_INSTALL_TEST_DIR:-}"
TEST_VERSION="${NAMBA_INSTALL_TEST_VERSION:-}"
TEST_OS="${NAMBA_INSTALL_TEST_OS:-}"
TEST_ARCH="${NAMBA_INSTALL_TEST_ARCH:-}"

if [ -n "$TEST_VERSION" ]; then
    VERSION="$TEST_VERSION"
fi
if [ -n "$TEST_INSTALL_DIR" ]; then
    INSTALL_DIR="$TEST_INSTALL_DIR"
fi

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

copy_or_download() {
    source="$1"
    output="$2"
    if [ -f "$source" ]; then
        cp "$source" "$output"
        return
    fi
    download "$source" "$output"
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

find_expected_checksum() {
    checksums_path="$1"
    asset_name="$2"
    awk -v asset="$asset_name" '
        /^[[:space:]]*$/ { next }
        {
            digest=$1
            name=$2
            sub(/^\*/, "", name)
            sub(/^\.\//, "", name)
            base=name
            sub(/^.*\//, "", base)
            if (base == asset && name !~ /\.\./ && name !~ /^\//) {
                print digest
                found=1
                exit
            }
        }
        END { if (!found) exit 1 }
    ' "$checksums_path"
}

sha256_file() {
    path="$1"
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$path" | awk '{print $1}'
        return
    fi
    if command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$path" | awk '{print $1}'
        return
    fi
    if command -v openssl >/dev/null 2>&1; then
        openssl dgst -sha256 "$path" | awk '{print $NF}'
        return
    fi
    printf '%s\n' "No SHA-256 checksum tool found. Install sha256sum, shasum, or openssl and retry." >&2
    exit 1
}

verify_checksum() {
    archive_path="$1"
    checksums_path="$2"
    asset_name="$3"

    expected="$(find_expected_checksum "$checksums_path" "$asset_name" || true)"
    if [ -z "$expected" ]; then
        printf '%s\n' "checksums.txt does not contain an entry for $asset_name." >&2
        exit 1
    fi

    actual="$(sha256_file "$archive_path")"
    expected_lc="$(printf '%s' "$expected" | tr '[:upper:]' '[:lower:]')"
    actual_lc="$(printf '%s' "$actual" | tr '[:upper:]' '[:lower:]')"
    if [ "$expected_lc" != "$actual_lc" ]; then
        printf '%s\n' "Checksum verification failed for $asset_name." >&2
        printf '%s\n' "Expected: $expected" >&2
        printf '%s\n' "Actual:   $actual" >&2
        exit 1
    fi
}

verify_tar_entries() {
    archive_path="$1"
    if tar -tzf "$archive_path" | awk '
        $0 ~ /^\/|(^|\/)\.\.($|\/)/ { bad=1 }
        END { exit bad ? 1 : 0 }
    '; then
        return
    fi
    printf '%s\n' "Archive contains unsafe paths and will not be extracted." >&2
    exit 1
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

OS_NAME="${TEST_OS:-$(detect_os)}"
ARCH_NAME="${TEST_ARCH:-$(detect_arch)}"

if [ "$OS_NAME" = "unsupported" ] || [ "$ARCH_NAME" = "unsupported" ]; then
    printf '%s\n' "Unsupported platform: $(uname -s)/$(uname -m)" >&2
    exit 1
fi

ASSET_NAME="namba_${OS_NAME}_${ARCH_NAME}.tar.gz"
if [ "$VERSION" = "latest" ]; then
    DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/$ASSET_NAME"
    CHECKSUMS_URL="https://github.com/$REPO/releases/latest/download/checksums.txt"
else
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$ASSET_NAME"
    CHECKSUMS_URL="https://github.com/$REPO/releases/download/$VERSION/checksums.txt"
fi
if [ -n "$TEST_ASSET_PATH" ]; then
    DOWNLOAD_URL="$TEST_ASSET_PATH"
fi
if [ -n "$TEST_CHECKSUMS_PATH" ]; then
    CHECKSUMS_URL="$TEST_CHECKSUMS_PATH"
fi

printf '%s\n' "Installing NambaAI from $DOWNLOAD_URL"
mkdir -p "$INSTALL_DIR"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM

ARCHIVE_PATH="$TMP_DIR/$ASSET_NAME"
CHECKSUMS_PATH="$TMP_DIR/checksums.txt"
EXTRACT_DIR="$TMP_DIR/extract"
mkdir -p "$EXTRACT_DIR"
copy_or_download "$DOWNLOAD_URL" "$ARCHIVE_PATH"
copy_or_download "$CHECKSUMS_URL" "$CHECKSUMS_PATH"
verify_checksum "$ARCHIVE_PATH" "$CHECKSUMS_PATH" "$ASSET_NAME"
verify_tar_entries "$ARCHIVE_PATH"
tar -xzf "$ARCHIVE_PATH" -C "$EXTRACT_DIR"
if [ ! -f "$EXTRACT_DIR/namba" ]; then
    printf '%s\n' "namba was not found in the downloaded archive." >&2
    exit 1
fi
install -m 0755 "$EXTRACT_DIR/namba" "$INSTALL_DIR/namba"

if [ -z "$TEST_INSTALL_DIR" ]; then
    case "${SHELL:-}" in
        */zsh)
            append_path "$HOME/.zshrc"
            ;;
        *)
            append_path "$HOME/.profile"
            ;;
    esac
fi

case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *) export PATH="$INSTALL_DIR:$PATH" ;;
esac

printf '\n%s\n' "NambaAI installed."
printf '%s\n' "Binary: $INSTALL_DIR/namba"
printf '%s\n' "Command: namba"
printf '%s\n' "If the command is not available in your current shell, run 'exec ${SHELL:-sh} -l' or open a new terminal."
