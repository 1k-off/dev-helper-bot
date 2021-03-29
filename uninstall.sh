#!/usr/bin/env bash
INSTALL_DIR="/opt/ooops-bot"

systemctl disable ooops-bot
systemctl stop ooops-bot

rm -rf $INSTALL_DIR
rm /lib/systemd/system/ooops-bot.service
rm /etc/rsyslog.d/ooops-bot.conf
rm -rf /var/log/ooops-bot

systemctl daemon-reload
systemctl restart rsyslog

echo '
Ooops bot uninstalled. Plak plak ;(
Do not forget to remove include from your nginx.
'
