#!/bin/sh
# /opt/etc/init.d/S52xkeencontrol
### Начало информации о службе
# Краткое описание: Запуск / Остановка xkeen-control
# version="0.1"  # Версия скрипта
### Конец информации о службе

green=""
red=""
yellow=""
reset=""

xkeen_control_initd="/opt/sbin/xkeen-control"

# Функция для проверки статуса xkeen-control
xkeen_control_status() {
    if ps | grep -v grep | grep -q "$xkeen_control_initd"; then
        return 0 # Процесс существует и работает
    else
        return 1 # Процесс не существует
    fi
}

# Функция для запуска xkeen-control
start() {
    if xkeen_control_status; then
        echo -e "  xkeen-control ${yellow}уже запущен${reset}"
    else
	# EDIT ME
	# coreutils-nohup required (opkg install coreutils-nohup)
        nohup $xkeen_control_initd -config /opt/etc/xray/configs/05_routing.json -token t -admin 1 >/opt/tmp/xkeen-control.log 2>&1 &
        echo -e "  xkeen-control ${green}запущен${reset}"
    fi
}

# Функция для остановки xkeen-control
stop() {
    if xkeen_control_status; then
        killall -9 "xkeen-control"
        echo -e "  xkeen-control ${yellow}остановлен${reset}"
    else
        echo -e "  xkeen-control ${red}не запущен${reset}"
    fi
}

# Функция для перезапуска xkeen-control
restart() {
    stop > /dev/null 2>&1
    start > /dev/null 2>&1
    echo -e "  xkeen-control ${green}перезапущен${reset}"
}

# Обработка аргументов командной строки
case "$1" in
    start)
        start;;
    stop)
        stop;;
    restart)
        restart;;
    status)
        if xkeen_control_status; then
            echo -e "  xkeen-control ${green}запущен${reset}"
        else
            echo -e "  xkeen-control ${red}не запущен${reset}"
        fi;;
    *)
        echo -e "  Команды: ${green}start${reset} | ${red}stop${reset} | ${yellow}restart${reset} | status";;
esac

exit 0

