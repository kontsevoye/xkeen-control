# xkeen-control

Telegram UI для управления доменным роутингом xkeen. Можно добавлять/удалять маршруты через бота.

[![Demo](docs/demo.png)](docs/demo.mp4)

## Автоматическая установка на роутер

Запихиваем [файл установки](https://github.com/kontsevoye/xkeen-control/blob/master/scripts/install.sh) на роутер, делаем исполняемым, запускаем.

Например запустив такую команду на роутере:
```shell
curl https://github.com/kontsevoye/xkeen-control/blob/master/scripts/install.sh -o /opt/tmp/install.sh \
  && chmod +x /opt/tmp/install.sh \
  && /opt/tmp/install.sh
```

Если нет curl, то предварительно нужно его установить:
```shell
opkg update && opkg install curl
```

## Ручная установка на роутер

- ставим вспомогательный пакет из opkg
  - `opkg install coreutils-nohup`
- берем бинарник нужной архитектуры из последнего релиза [xkeen-control/releases](https://github.com/kontsevoye/xkeen-control/releases) и кладем его в `/opt/sbin/xkeen-control`
  - например `scp -P222 -O ~/Downloads/xkeen-control/xkeen-control-linux-arm64 root@192.168.69.1:/opt/sbin/xkeen-control`
- берем заготовку init.d скрипта из последнего релиза, например [xkeen-control/v1.2.0/init/S52xkeencontrol.sh](https://github.com/kontsevoye/xkeen-control/blob/v1.2.0/init/S52xkeencontrol.sh) и кладем в `/opt/etc/init.d/S52xkeencontrol`
  - например `scp -P222 -O ~/Downloads/xkeen-control/S52xkeencontrol.sh root@192.168.69.1:/opt/etc/init.d/S52xkeencontrol`
- делаем наши файлы исполняемыми
  - `chmod +x /opt/sbin/xkeen-control`
  - `chmod +x /opt/etc/init.d/S52xkeencontrol`
- заполняем параметры запуска в `/opt/etc/init.d/S52xkeencontrol` под строкой EDIT ME. `telegram_bot_token=""\ntelegram_admin_id=""` меняем на `telegram_bot_token="ТОКЕН_ТЕЛЕГРАМ_БОТА_СЮДА"\ntelegram_admin_id="СВОЙ_ТЕЛЕГРАМ_АЙДИ_СЮДА"`
  - `ТОКЕН_ТЕЛЕГРАМ_БОТА_СЮДА` берем у [@BotFather](https://t.me/BotFather)
  - `СВОЙ_ТЕЛЕГРАМ_АЙДИ_СЮДА` берем у [@userinfobot](https://t.me/@userinfobot)
- запускаем софтину путем выполнения `/opt/etc/init.d/S52xkeencontrol start`
- теперь если отправить своему боту `/list`, он должен ответить текущим списком доменов

## Конфигурация списка команд в боте

У BotFather можно задать список доступных команд, чтобы они появлялись при наборе слэша.
Делается это через `/mybots` -> _bot name_ -> `Edit Bot` -> `Edit Commands`
```
list - Список проксируемых доменов
add - Добавить домен в список проксируемых
delete - Убрать домен из списка проксируемых
restart - Перезапустить xkeen для применения конфига
backups - Список бэкапов конфига
restore - Восстановить конфига из бэкапа
help - Помощь по префиксам xray
```
