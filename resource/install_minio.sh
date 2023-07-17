

mkdir minio
cd minio
wget https://dl.min.io/server/minio/release/linux-amd64/minio

touch minio.service
echo "
[Unit]
Description=minio daemon
After=network.target

[Service]
PIDFile=/tmp/minio.pid
User=root
Group=root
WorkingDirectory=/root/minio/
ExecStart=/root/minio/minio server -config-dir ./config/ --address ':8002' --console-address ':8003' /data
Restart=always

[Install]
WantedBy=multi-user.target" > minio.service

chmod +x minio
chmod +x minio.service
mv minio.service /etc/systemd/system/minio.service
systemctl start minio
systemctl enable minio