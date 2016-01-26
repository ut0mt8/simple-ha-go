# simple-ha-go
Tired of keepalived, corosync, pacemaker, heartbeat or whatever ? Here a simple daemon wich ensure a Heartbeat between two hosts. One is active, and the other is backup, launching script when changing state. Simple implementation, KISS. 

The GO version, I learn GO with this simple program, it may be not very elegant :)

```
Usage of ./simple-ha-go:
  -active-script string
    	the script to launch when switching from backup state to active state (default "REQUIRED")
  -backup-script string
    	the script to launch when switching from active state to backup state (default "REQUIRED")
  -interval int
    	the interval in second between check to the peer (default 2)
  -key string
    	the shared key between peers (default "OdcejToQuor4")
  -listen-ip string
    	the local ip to bind to (default "localhost")
  -listen-port int
    	the local port to bind to (default 29999)
  -peer-ip string
    	the peer ip to connect to (default "REQUIRED")
  -peer-port int
    	the peer port to bind to (default 29999)
  -priority int
    	the local priority
  -retry int
    	the number of time retrying connecting to the peer when is dead (default 3)
  -verbose
    	be more verbose
```
