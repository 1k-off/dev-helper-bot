server {
#    listen 443 ssl http2;
    listen 80;
    server_name {{ .domain }};
    access_log off;
    error_log  /dev/null;
#    include snippets/letsencrypt.conf;
#    include snippets/ssl_dev.conf;
#    include snippets/protected_server.conf;

    auth_basic {{ .basicauth }};
    auth_basic_user_file /etc/nginx/passwd/default;

    location / {
        proxy_pass {{ .scheme }}://{{ .ip }};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_connect_timeout 120;
        proxy_send_timeout 120;
        proxy_read_timeout 180;
    }
}