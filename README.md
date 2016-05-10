![travis](https://travis-ci.org/sky-uk/owl.svg?branch=master)

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

# Download

You can find the latest x86_64 binary at https://github.com/sky-uk/owl/releases.

To install in `/usr/local/bin`:

    sudo -i
    curl -O https://github.com/sky-uk/owl/releases/download/0.4.0/owl > /usr/local/bin/owl
    chmod u+x /usr/local/bin/owl

# Development

Build and test with make:

    make test

Requires `libsystemd` to build. On fedora:

    dnf install -y systemd-devel

Golang dependencies are managed with `godep`. Set them up before building:

    godep restore

## Releasing

For contributors, release only from master branch:

    make release

Then tag it and upload it to releases.

    git tag 0.3.0 && git push --tags
