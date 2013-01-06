# BGP API

`bgpapi` runs as a "parsed-route-backend" to [ExaBGP](http://code.google.com/p/exabgp/)
and provides an HTTP API to some BGP data.

## Compilation

After installing Go (golang) and configuring the development environment, you can run

    go get -v github.com/abh/bgpapi
    go build
    go install

## Configuration

bgpapi itself doesn't currently take any configuration. Configure a
`neighbor` in ExaBGP and a `parsed-route-backend` process, like:

        process parsed-route-backend {
                parse-routes;
                peer-updates;
                run /Users/ask/go/bin/bgpapi;
        }
