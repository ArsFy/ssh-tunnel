# SSH Tunnel

![](https://img.shields.io/badge/license-MIT-blue)
![](https://img.shields.io/badge/Golang-1.24-blue)

Easy to use SSH Tunnel, support automatic reconnection.

### Config

```json
{
    "server": [
        {
            "host": "your_server_host",
            "port": 22,
            "user": "your_server_user",
            "password": "your_server_password",
            "services": [
                {
                    "type": "remote",
                    "local": "localhost:8080",
                    "remote": "localhost:3000"
                },
                {
                    "type": "local",
                    "local": "localhost:5173",
                    "remote": "localhost:5173"
                }
            ]
        }
    ]
}
```

### Build

```bash
go build -ldflags="-s -w" .
```

### Run

```bash
./ssh-tunnel
```