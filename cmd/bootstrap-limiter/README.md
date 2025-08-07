# Bootstrap Limiter

## Description

This utility is aimed to provide an easy way to setup the system so that a non-privileged user can set resource limits for Bless Functions.
This tool is designed to be run a single time, before running the Bless Worker Node.
If the user has no intention of using cgroups to set resource limits, there is no need in running this tool.

## Workaround

This tool aims to provide an easy way to manage the _typical_ use-case.
It is not mandatory to use the tool to setup your system.

If you would like to setup your system manually, the steps are listed below.
The readme assumes your `cgroup` mountpoint is the default - `/sys/fs/cgroup`.

    1. Create a directory `/sys/fs/cgroup/bless`
    2. Change owner of the directory and its subdirectories to the user that will be running the node. Example: `sudo chown -R <user> /sys/fs/cgroup/bless`.
    3. Set files `/sys/fs/cgroup/cgroup.procs` and `/sys/fs/cgroup/cgroup.subtree_control` to be group writable. Example: `sudo chmod 0664 /sys/fs/cgroup/cgroup.procs` (same for `cgroup.subtree_control`)
    4. Add user to the group that owns the files listed in step 3. By default this would be `root`, so for example `sudo usermod -a -G root <user>`.

## Removing Cgroup

You can remove a cgroup, effectively reverting the changes done by the tool by running `sudo rmdir /sys/fs/cgroup/bless`.

## Further Reading

To read more about resource limits see [here](https://docs.kernel.org/admin-guide/cgroup-v2.html).
