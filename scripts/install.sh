#!/bin/sh

echo "Устанавливаю curl"
opkg update &>/dev/null
opkg install curl &>/dev/null

echo "Качаю актуальный init.d"
curl -s -L -o /opt/etc/init.d/S52xkeencontrol "https://raw.githubusercontent.com/kontsevoye/xkeen-control/refs/tags/$(curl -s https://api.github.com/repos/kontsevoye/xkeen-control/releases/latest | jq -r ".tag_name")/init/S52xkeencontrol.sh"
chmod +x /opt/etc/init.d/S52xkeencontrol

echo "Запускаю установку"
/opt/etc/init.d/S52xkeencontrol update

if [[ $? -eq 1 ]]; then
  echo "Чето не пошло. Удаляют init.d. Перезапустите установку."
  rm -rf /opt/etc/init.d/S52xkeencontrol
fi
