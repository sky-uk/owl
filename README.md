# OWL - Systemd journal watcher

Reads the systemd journal and matches entries based on a set of include patterns, e.g ERROR, and then filters out the matching entries based on an exclude list, e.g `KNOWN SILLY ERROR`.

This is intended to be used by a daemon such as [monit](https://mmonit.com/monit/). `owl` will return a non-zero exit code if the threshold of errors is met.

Patterns are treated as regular expressions - see https://golang.org/pkg/regexp/ for details.

Example configuration:

    [global]
    time=6
    errorsToReport=5
    alertThreshold=1

    [service "kube-apiserver"]
    include=: E

    [service "kube-controller-manager"]
    include=: E
    include=: W
    exclude=is forbidden:.*exceeding quota for resource memory
    exclude=cat

This will check errors for the `kube-apiserver` and `kube-controller-manager` services in the last 6 seconds, 
report (print out) the last, and consider any more than 1 errors to be fatal and exit with a non 0 exit code.

The config file should be placed in `/etc/owl/config`

# Development

Needs `libsystemd` to build. On fedora:

    dnf install -y systemd-devel

Dependencies are managed with `godep`. Set them up before building:

    godep restore

## Releasing

For contributors, release only from master branch:

    go build
    strip owl

Then tag it and upload it to releases.

    git tag 0.3.0 && git push --tags
