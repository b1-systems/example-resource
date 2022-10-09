# example-resource

## Installation

```bash
git clone https://tk-sls.de/gitlab/golang-oidc/example-resource.git
cd example-resource
go mod tidy
go build
sudo mkdir /usr/local/example-resource
sudo cp example-resource /usr/local/example-resource
sudo cp example-resource.service /etc/systemd/system
sudo systemctl daemon-reload
```

## Configuration

```bash
sudo cp example-resource.ini.sample /usr/local/example-resource/example-resource.ini
sudo vi /usr/local/example-resource/example-resource.ini
```

Example `example-resource.ini`:

```
[example-resource]
# This URL will be used for endpoint discovery of your IdP
providerUrl = https://your_idp_server/realms/golang-oidc

# Plain HTTP service address of this "example-frontend" server:
listenAddress = 0.0.0.0:8080
```

# Start

```bash
systemctl start example-resource.service
journalctl -xefu  example-resource.service
```

