# xkeen-control

## Установка на роутер

- ставим вспомогательный пакет из opkg
  - `opkg install coreutils-nohup`
- берем бинарник нужной архитектуры из последнего релиза [xkeen-control/releases](https://github.com/kontsevoye/xkeen-control/releases) и кладем его в `/opt/sbin/xkeen-control`
  - например `scp -P222 -O ~/Downloads/xkeen-control/xkeen-control-linux-arm64 root@192.168.69.1:/opt/sbin/xkeen-control`
- берем заготовку init.d скрипта из последнего релиза, например [xkeen-control/v1.0.0/initd_example.sh](https://github.com/kontsevoye/xkeen-control/blob/v1.0.0/initd_example.sh) и кладем в `/opt/etc/init.d/S52xkeencontrol`
  - например `scp -P222 -O ~/Downloads/xkeen-control/initd_example.sh root@192.168.69.1:/opt/etc/init.d/S52xkeencontrol`
- делаем наши файлы исполняемыми
  - `chmod +x /opt/sbin/xkeen-control`
  - `chmod +x /opt/etc/init.d/S52xkeencontrol`
- заполняем параметры запуска в `/opt/etc/init.d/S52xkeencontrol` под строкой EDIT ME. `-token t -admin 1` меняем на `-token ТОКЕН_ТЕЛЕГРАМ_БОТА_СЮДА -admin СВОЙ_ТЕЛЕГРАМ_АЙДИ_СЮДА`
  - `ТОКЕН_ТЕЛЕГРАМ_БОТА_СЮДА` берем у [@BotFather](https://t.me/BotFather)
  - `СВОЙ_ТЕЛЕГРАМ_АЙДИ_СЮДА` берем у [@userinfobot](https://t.me/@userinfobot)
- запускаем софтину путем выполнения `/opt/etc/init.d/S52xkeencontrol start`
- теперь если отправить своему боту `/list`, он должен ответить текущим списком доменов

опционально можно у BotFather задать список доступных команд, чтобы они появлялись при наборе слэша
```
list - Список проксируемых доменов
add - Добавить домен в список проксируемых
delete - Убрать домен из списка проксируемых
restart - Перезапустить xkeen для применения конфига
backups - Список бэкапов конфига
restore - Восстановить конфига из бэкапа
help - Помощь по префиксам xray
```
