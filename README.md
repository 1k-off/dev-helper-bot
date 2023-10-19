# Dev Helper Slack Bot

Slack bot written in go with mongodb store. What it can do:
- create nginx or caddy configurations from template and reload nginx (personal domain for any developer mapped to his workstation through VPN connection)
- delete created nginx configurations after a time
- update nginx configurations (basic auth, proxy port, full-ssl, target IP)
- create and delete VPN configurations (pritunl) (admin only)
- send welcome message to new VPN users
- send user's personal VPN config (URL to download from pritunl)

## How to run
### systemd service
```
[Unit]
Description=Ooops bot service
ConditionPathExists=/opt/ooops
After=network.target

[Service]
Type=simple

Restart=always
RestartSec=5s
StartLimitIntervalSec=60

WorkingDirectory=/opt/ooops
ExecStart=/opt/ooops/ooops

# make sure log directory exists and owned by syslog
PermissionsStartOnly=true
ExecStartPre=/bin/mkdir -p /var/log/ooops
ExecStartPre=/bin/chown syslog:adm /var/log/ooops
ExecStartPre=/bin/chmod 755 /var/log/ooops
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=ooops

[Install]
WantedBy=multi-user.target
```