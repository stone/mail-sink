mail-sink
=========
[![Build Status](https://drone.io/github.com/stone/mail-sink/status.png)](https://drone.io/github.com/stone/mail-sink/latest)

mail-sink is a utility program that implements a "black hole" function. It listens on the named host (or address) and port. It accepts Simple Mail Transfer Protocol (SMTP) messages from the network and discards them. 

Written in Go: http://www.golang.org/

Install with:

    go get github.com/stone/mail-sink
    
Usage:

    $ ./mail-sink -h
    Usage of mail-sink:
    -H="localhost": hostname to greet with
    -i="localhost": listen on interface
    -p=25: listen port
    -v=false: log the mail body


