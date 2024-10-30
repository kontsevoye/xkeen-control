#!/bin/sh

opkg update &>/dev/null
opkg install coreutils-nohup curl jq lscpu &>/dev/null

architecture=""
info_cpu() {
    local uname_arch=$(uname -m | tr '[:upper:]' '[:lower:]')
    case "$uname_arch" in
        *'armv5tel'*)
            architecture='arm'
            ;;
        *'armv6l'*)
            architecture='arm'
            if grep Features /proc/cpuinfo | grep -qw 'vfp'; then
                architecture='arm'
            fi
            ;;
        *'armv7'*)
            architecture='arm'
            if grep Features /proc/cpuinfo | grep -qw 'vfp'; then
                architecture='arm'
            fi
            ;;
        *'armv8'* | *'aarch64'*)
            architecture='arm64'
            ;;
        *'mips64le'* | *'mips64'* )
            architecture='mips64'
            ;;
        *'mipsle'* | *'mips 1004'* | *'mips 34'* | *'mips 24'*)
            architecture='mipsle'
            ;;
        *'mips'*)
            architecture='mips'
            ;;
        *)
            local cpuinfo=$(grep -i 'model name' /proc/cpuinfo | sed -e 's/.*: //i' | tr '[:upper:]' '[:lower:]')
            if echo "$cpuinfo" | grep -q -e *'armv8'* -e *'aarch64'* -e *'cortex-a'*; then
                architecture='arm64'
            elif echo "$cpuinfo" | grep -q -e *'mips64le'*; then
                architecture='mips64le'
            elif echo "$cpuinfo" | grep -q -e *'mips64'*; then
                architecture='mips64'
            elif echo "$cpuinfo" | grep -q -e *'mips32le'* -e *'mips 1004'* -e *'mips 34'* -e *'mips 24'*; then
                architecture='mipsle'
            elif echo "$cpuinfo" | grep -q -e *'mips'*; then
                architecture='mips'
            else
                echo "unsupported arch $uname_arch"
                exit 1
            fi
            ;;
    esac

    if [ "$architecture" = 'mips64' ] || [ "$architecture" = 'mips' ]; then
        local lscpu_output="$(lscpu 2>/dev/null | tr '[:upper:]' '[:lower:]')"
        if echo "$lscpu_output" | grep -q "little endian"; then
            architecture="${architecture}le"
        fi
    fi
}

info_cpu

curl -L -o /opt/sbin/xkeen-control "$(curl -s https://api.github.com/repos/kontsevoye/xkeen-control/releases/latest | jq -r ".assets[] | select(.name | contains (\"$architecture\")) | .browser_download_url")"
chmod +x /opt/sbin/xkeen-control
curl -L -o /opt/etc/init.d/S52xkeencontrol "https://raw.githubusercontent.com/kontsevoye/xkeen-control/refs/tags/$(curl -s https://api.github.com/repos/kontsevoye/xkeen-control/releases/latest | jq -r ".tag_name")/init/S52xkeencontrol.sh"
chmod +x /opt/etc/init.d/S52xkeencontrol
