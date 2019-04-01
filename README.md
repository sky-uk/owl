![travis](https://travis-ci.org/sky-uk/owl.svg?branch=master)

# OWL - Systemd journal watcher

Reads the systemd journal and matches entries based on a set of include patterns, e.g ERROR, and then filters out the matching entries based on an exclude list, e.g `KNOWN SILLY ERROR`.

This is intended to be used by a daemon such as [monit](https://mmonit.com/monit/). `owl` will return a non-zero exit code if the threshold of errors is met.

## Configuration 

Example configuration:

    [global]
    time=6
    errorsToReport=5
    alertThreshold=1

    [service "kube-apiserver"]
    include=: E.*

    [service "kube-controller-manager"]
    include=: E.*
    include=: W.*
    exclude=.*is forbidden:.*exceeding quota for resource memory
    exclude=cat

    [service "*"]
    include=.*FATAL.*

This will check errors for the `kube-apiserver` and `kube-controller-manager` services in the last 6 seconds, 
report (print out) the last, and consider any more than 1 errors to be fatal and exit with a non 0 exit code.
It will also match any log lines with the word FATAL in it.

The config file should be placed in `/etc/owl/config`

## Pattern matching

Patterns are treated as regular expressions - see https://golang.org/pkg/regexp/ for details.

 - Include and exclude filters can match entries spanning multiple lines.
   For instance an `Include: "E.*starts\n.*ends here"` will find 2 matches in the following journal logs entries:
      
        : E error starts
        continues and ends here
        : E error starts
        and ends here
        something else comes up

 - Only the matching portion of the journal entry is returned.
   For instance, in the following journal entry, using `Include: ": E.*"` will return the whole journal entry line, 
   while `Include ": E"` will only return `": E"`.
   
        : E error description

 - No journal entries will be matched when no includes are specified


# Download

You can find the latest x86_64 binary at https://github.com/sky-uk/owl/releases.

To install in `/usr/local/bin`:

    sudo curl -sSL -o /usr/local/bin/owl https://github.com/sky-uk/owl/releases/download/0.8.0/owl && \
    sudo chmod 755 /usr/local/bin/owl

# Development

## Build with Docker

    docker build . -t owl-build
    docker run owl-build

To include your local changes and save the resulting binary locally:

    docker run -v $PWD:/go/src/github.com/sky-uk/owl owl-build

## Build without Docker

    make test

Prerequisites:

- [Go Dep](https://github.com/golang/dep)
- `libsystemd` (Fedora: `dnf install -y systemd-devel`, Ubuntu: `apt install -y libsystemd-dev`)

## Releasing

For contributors, release only from master branch:

    docker build . -t owl-build
    docker run -v $PWD:/go/src/github.com/sky-uk/owl owl-build release

or

    make release

Then tag it and upload it to releases.

    git tag 0.9.0 && git push --tags
