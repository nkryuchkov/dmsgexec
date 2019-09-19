# IP Tables

**dmsg.Discovery (http://139.162.29.5:9090)**
```
$ sudo iptables -A INPUT -p tcp --dport 9090 -j ACCEPT

$ messaging-discovery
```

**dmsg.Server (139.162.58.186:8080)**
```
$ sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT

$ messaging-server
```
