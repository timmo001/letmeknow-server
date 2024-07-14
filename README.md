# LetMeKnow - Server

### Installation

Download the `letmeknow-server-*` binary from the releases page.

### Linux

#### Install

```bash
sudo mv letmeknow-server-* /usr/local/bin/letmeknow-server
sudo chmod +x /usr/local/bin/letmeknow-server
```

#### Setup as a service

```bash
sudo systemctl edit --force --full letmeknow-server.service
```

```ini
[Unit]
Description=LetMeKnow Server
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/letmeknow-server
Restart=always
RestartSec=1

[Install]
WantedBy=default.target
```

Now enable and start the service:

```bash
sudo systemctl enable --now letmeknow-server.service
```

Check the status:

```bash
sudo systemctl status letmeknow-server.service
```

#### Logs

```bash
sudo journalctl -u letmeknow-server.service -f
```
