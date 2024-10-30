#!/bin/sh
# /opt/etc/init.d/S52xkeencontrol
### Начало информации о службе
# Краткое описание: Запуск / Остановка xkeen-control
# version="0.2"  # Версия скрипта
### Конец информации о службе

# coreutils-nohup required (opkg install coreutils-nohup)

# EDIT ME
telegram_bot_token=""
telegram_admin_id=""
enable_backups="true"
config_location="/opt/etc/xray/configs/05_routing.json"
log_location="/opt/tmp/xkeen-control.log"

green="\033[32m"
red="\033[31m"
yellow="\033[33m"
reset="\033[0m"

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
        nohup $xkeen_control_initd -config="$config_location" -token="$telegram_bot_token" -admin="$telegram_admin_id" -enableBackups="$enable_backups" >>$log_location 2>&1 &
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
