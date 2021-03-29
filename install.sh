#!/usr/bin/env bash
INSTALL_DIR="/opt/ooops-bot"
mkdir -p $INSTALL_DIR

cp -R pkg/. $INSTALL_DIR
cd $INSTALL_DIR

echo "Installing as system service..."
mv data/ooops-bot.service /lib/systemd/system/ooops-bot.service

echo "if \$programname == 'ooops-bot' then /var/log/ooops-bot/ooops-bot.log
& stop
" > /etc/rsyslog.d/ooops-bot.conf

echo "Restarting rsyslog..."
systemctl restart rsyslog
echo "Reloading daemons..."
systemctl daemon-reload
echo "Adding to autostart..."
systemctl enable ooops-bot
echo "Starting..."
systemctl start ooops-bot

echo '
Ooops bot installed and started :)
Do not forget to add `include /opt/ooops-bot/nginx/*;` to your nginx.
'
