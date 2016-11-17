% OCI(1) OCI-CREATE-LAYER User Manuals
% OCI Community
% October 2016
# NAME
oci-create-layer \- Create filesystem changeset

# SYNOPSIS
**oci-create-layer** [child] [parent] [flags]

# DESCRIPTION
`oci-create-layer` creates a filesystem changeset from two layers. It compares child with parent and generates a filsystem diff, pack the diff into a uncompressed tar archive. The default output is stdout, use `--dest` to specify a custom one.

# OPTIONS
**--help**
  Print usage statement

**--dest**
The dest specify a particular filename where the layer write to

# EXAMPLES
```
$ oci-create-layer --dest rootfs-1-s.tar rootfs-1-s rootfs-1
$ ls
rootfs-1  rootfs-1-s  rootfs-1-s.tar

```

# HISTORY
Oct 2016, Originally compiled by Lei Jitang (coolljt0725 at huawei dot com)
