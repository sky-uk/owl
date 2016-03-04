# OWL - Systemd journal watcher

Reads the systemd journal and matches entries based on a set of include patterns, e.g ERROR, and then filters out the matching entries based on an exclude list, e.g `KNOWN SILLY ERROR`

Example configuration:

```
[default]
time=6
errorsToReport=5
fatalNumberOfErrors=1

[service "kube-apiserver"]
include=: E

[service "kube-controller-manager"]
include=: E
exclude=is forbidden: unable to admit pod without exceeding quota for resource memory
exclude=cat
```

This will check errors for the `kube-apiserver` and `kube-controller-manager` services in the last 6 seconds, 
report (print out) the last, and consider any more than 1 errors to be fatal and exit with a non 0 exit code.

The printing out and non 0 exist code is useful if you are executing owl from a service like monit.



