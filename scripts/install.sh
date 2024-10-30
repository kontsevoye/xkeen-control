#!/bin/sh

opkg update &>/dev/null
opkg install curl &>/dev/null

if [[ -e "/opt/etc/init.d/S52xkeencontrol" && $(grep -q "update() {" "/opt/etc/init.d/S52xkeencontrol") ]]; then
  echo "init.d файл сущестует, умеет обновляться"
  echo "Останавливаю запущенный экземпляр"
  /opt/etc/init.d/S52xkeencontrol stop &>/dev/null
else
  echo "Качаю актуальный установщик"
  curl -s -L -o /opt/etc/init.d/S52xkeencontrol "https://raw.githubusercontent.com/kontsevoye/xkeen-control/refs/tags/$(curl -s https://api.github.com/repos/kontsevoye/xkeen-control/releases/latest | jq -r ".tag_name")/init/S52xkeencontrol.sh"
  chmod +x /opt/etc/init.d/S52xkeencontrol
fi

echo "Запускаю установку"
/opt/etc/init.d/S52xkeencontrol update

if [[ $? -eq 1 ]]; then
  echo "Чето не пошло. Удаляют init.d. Перезапустите установку."
  rm -rf /opt/etc/init.d/S52xkeencontrol
fi
