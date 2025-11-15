#!/bin/sh
# POSIX sh build script — runs all targets, records failures, prints summary.

BIN_DIR=/bin
mkdir -p .${BIN_DIR}

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

TITLE=convert-cbz

FAILED_BUILDS=""

build() {
    label="$1"
    shift
    logfile=".${BIN_DIR}/${label}.err"

    # Run the command using env so VAR=val tokens are treated as environment assignments
    # and the final token(s) form the command to execute.
    if env "$@" 2> "$logfile"; then
        printf '%b\n' "${GREEN}Build succeeded: ${label}${NC}"
        rm -f "$logfile"
    else
        printf '%b\n' "${RED}Build failed: ${label}${NC}"
        printf '   (see %s for error output)\n' "$logfile"
        FAILED_BUILDS="${FAILED_BUILDS} ${label}"
    fi
}

build_cmd() {
  printf 'CGO_ENABLED=0 GOOS=%s GOARCH=%s go build -trimpath -buildvcs=false -o .%s/%s-%s-%s ./main.go' \
    "$1" "$2" "$BIN_DIR" "$TITLE" "$1" "$2"
}

build_all() {
    # Linux
    build "linux_amd64"    $(build_cmd linux amd64)
    build "linux_arm"      $(build_cmd linux arm)
    build "linux_arm64"    $(build_cmd linux arm64)
    build "linux_ppc64"    $(build_cmd linux ppc64)
    build "linux_ppc64le"  $(build_cmd linux ppc64le)
    build "linux_mips"     $(build_cmd linux mips)
    build "linux_mipsle"   $(build_cmd linux mipsle)
    build "linux_mips64"   $(build_cmd linux mips64)
    build "linux_mips64le" $(build_cmd linux mips64le)
    build "linux_s390x"    $(build_cmd linux s390x)

    # Darwin
    build "darwin_amd64"  $(build_cmd darwin amd64)
    build "darwin_arm64"  $(build_cmd darwin arm64)

    # FreeBSD
    build "freebsd_amd64" $(build_cmd freebsd amd64)
    build "freebsd_386"   $(build_cmd freebsd 386)

    # OpenBSD
    build "openbsd_amd64" $(build_cmd openbsd amd64)
    build "openbsd_386"   $(build_cmd openbsd 386)
    build "openbsd_arm64" $(build_cmd openbsd arm64)

    # NetBSD
    build "netbsd_amd64"  $(build_cmd netbsd amd64)
    build "netbsd_386"    $(build_cmd netbsd 386)
    build "netbsd_arm"    $(build_cmd netbsd arm)

    # DragonFlyBSD
    build "dragonfly_amd64" $(build_cmd dragonfly amd64)

    # Solaris
    build "solaris_amd64" $(build_cmd solaris amd64)

    # Plan 9
    build "plan9_386"     $(build_cmd plan9 386)
    build "plan9_amd64"   $(build_cmd plan9 amd64)
}

usage() {
    echo "Usage: $0 [all]"
    echo
    echo "Build ${TITLE} in different modes:"
    echo "  ./build.sh       Build a single binary for the current system (default)."
    echo "  ./build.sh all   Cross-compile for all supported OS/architectures."
    echo
    echo "Examples:"
    echo "  ./build.sh       -> builds .${BIN_DIR}/${TITLE}"
    echo "  ./build.sh all   -> builds multiple binaries into ./bin/"
    exit 1
}

if [ $# -eq 0 ]; then
    build "host_default" CGO_ENABLED=0 go build -trimpath -buildvcs=false -o .${BIN_DIR}/${TITLE} ./main.go            
elif [ $# -eq 1 ] && [ "$1" = "all" ]; then
    build_all
else
    usage
fi

# Final summary
if [ -n "$FAILED_BUILDS" ]; then
    echo
    printf '%b\n' "${RED}Some builds failed:${NC}"
    for target in $FAILED_BUILDS; do
        printf '   - %s\n' "$target"
    done
    exit 1
else
    echo
    printf '%b\n' "${GREEN}All builds completed successfully.${NC}"
    exit 0
fi
