% OCI(1) OCI-User Manuals
% OCI Community
% AUGUST 2016
# NAME

oci-cas \- Content-addressable storage manipulation

# SYNOPSIS

**oci-cas** [command]

# DESCRIPTION

`oci-cas` manipulates content-addressable storage.

# OPTIONS

**--help**
  Print usage statement

# COMMANDS

**get**
  Retrieve a blob from the store.
  See **oci-cas-get**(1) for full documentation on the **get** command.

**put**
  Write a blob to the store.
  See **oci-cas-put**(1) for full documentation on the **put** command.

**delete**
  Remove a blob from the store.
  See **oci-cas-delete**(1) for full documentation on the **delete** command.

# EXAMPLES

```
$ oci-image-init image-layout image
$ echo hello | oci-cas put image
sha256:5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03
$ oci-cas get image sha256:5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03
hello
$ oci-cas delete image sha256:5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03
```

# SEE ALSO

**oci-image-tools**(7), **oci-cas-get**(1), **oci-cas-put**(1), **oci-cas-delete**(1), **oci-image-init**(1)

# HISTORY

August 2016, Originally compiled by W. Trevor King (wking at tremily dot us)
