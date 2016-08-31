% OCI(1) OCI-IMAGE-TOOL User Manuals
% OCI Community
% AUGUST 2016
# NAME

oci-refs \- Name-based reference manipulation

# SYNOPSIS

**oci-refs** [command]

# DESCRIPTION

`oci-refs` manipulates name-based references.

# OPTIONS

**--help**
  Print usage statement

# COMMANDS

**get**
  Retrieve a reference from the store.
  See **oci-refs-get**(1) for full documentation on the **get** command.

**list**
  Return available names from the store.
  See **oci-refs-list**(1) for full documentation on the **list** command.

**put**
  Write a reference to the store.
  See **oci-refs-put**(1) for full documentation on the **put** command.

# EXAMPLES

```
$ oci-image-init image-layout image.tar
$ DIGEST=$(echo hello | oci-cas put image.tar)
$ SIZE=$(echo hello | wc -c)
$ printf '{"mediaType": "text/plain", "digest": "%s", "size": %d}' "${DIGEST}" "${SIZE}" |
>   oci-refs put image.tar greeting
$ oci-refs list image.tar
greeting
$ oci-refs get image.tar greeting
{"mediaType":"text/plain","digest":"sha256:5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03","size":6}
```

# SEE ALSO

**oci-image-tools**(7), **oci-cas-put**(1), **oci-refs-get**(1), **oci-refs-list**(1), **oci-refs-put**(1)

# HISTORY

August 2016, Originally compiled by W. Trevor King (wking at tremily dot us)
