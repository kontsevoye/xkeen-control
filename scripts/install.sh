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

echo "Определяю архитектуру"
info_cpu

echo "Читаю конфиг /opt/etc/xkeen-control/config.json"
mkdir -p /opt/etc/xkeen-control
if [ ! -f /opt/etc/xkeen-control/config.json ]; then
  echo "Создаю пустой конфиг /opt/etc/xkeen-control/config.json"
  echo "{}" > /opt/etc/xkeen-control/config.json
fi

echo "Качаю бинарник и init.d конфиг для архитектуры $architecture"

last_tag_name=$(curl -s https://api.github.com/repos/kontsevoye/xkeen-control/releases/latest | jq -r ".tag_name")
echo "Найдена версия $last_tag_name"

installed_tag_name=$(cat /opt/etc/xkeen-control/config.json | jq -r ".installed_tag_name | select (.!=null)")
if [ -z "${installed_tag_name}" ]; then
  tmpfile=$(mktemp)
  cat /opt/etc/xkeen-control/config.json | jq -r ".installed_tag_name |= \"$last_tag_name\"" > $tmpfile
  mv $tmpfile /opt/etc/xkeen-control/config.json
elif [[ "$installed_tag_name" == "$last_tag_name" ]]; then
  echo "Установлена последняя версия"
  exit 0
fi

curl -s -L -o /opt/sbin/xkeen-control "$(curl -s "https://api.github.com/repos/kontsevoye/xkeen-control/releases/tags/$last_tag_name" | jq -r ".assets[] | select(.name | contains (\"$architecture\")) | .browser_download_url")"
if [ -f /opt/etc/init.d/S52xkeencontrol ]; then
  echo "init.d файл сущестует, останавливаю"
  /opt/etc/init.d/S52xkeencontrol stop &>/dev/null
fi
curl -s -L -o /opt/etc/init.d/S52xkeencontrol "https://raw.githubusercontent.com/kontsevoye/xkeen-control/refs/tags/$last_tag_name/init/S52xkeencontrol.sh"

echo "Делаю файлы /opt/sbin/xkeen-control и /opt/etc/init.d/S52xkeencontrol исполняемыми"
chmod +x /opt/sbin/xkeen-control
chmod +x /opt/etc/init.d/S52xkeencontrol

tg_bot_token=$(cat /opt/etc/xkeen-control/config.json | jq -r ".tg_bot_token | select (.!=null)")
if [ -z "${tg_bot_token}" ]; then
  echo "Введите Telegram bot token (можно взять у https://t.me/BotFather):"
  read -r tg_bot_token
  tmpfile=$(mktemp)
  cat /opt/etc/xkeen-control/config.json | jq -r ".tg_bot_token |= \"$tg_bot_token\"" > $tmpfile
  mv $tmpfile /opt/etc/xkeen-control/config.json
fi
tg_admin_id=$(cat /opt/etc/xkeen-control/config.json | jq -r ".tg_admin_id | select (.!=null)")
if [ -z "${tg_admin_id}" ]; then
  echo "Введите Telegram admin ID (можно взять у https://t.me/userinfobot):"
  read -r tg_admin_id
  tmpfile=$(mktemp)
  cat /opt/etc/xkeen-control/config.json | jq -r ".tg_admin_id |= \"$tg_admin_id\"" > $tmpfile
  mv $tmpfile /opt/etc/xkeen-control/config.json
fi

sed -i "s/telegram_bot_token=\".*\"/telegram_bot_token=\"$tg_bot_token\"/" /opt/etc/init.d/S52xkeencontrol
sed -i "s/telegram_admin_id=\".*\"/telegram_admin_id=\"$tg_admin_id\"/" /opt/etc/init.d/S52xkeencontrol

echo "Запускаю /opt/etc/init.d/S52xkeencontrol"
/opt/etc/init.d/S52xkeencontrol start
