% OCI(1) OCI-UNPACK User Manuals
% OCI Community
% JULY 2016
# NAME
oci-unpack \- Unpack an image or image source layout

# SYNOPSIS
**oci-unpack** [src] [dest] [flags]

# DESCRIPTION
`oci-unpack` validates an application/vnd.oci.image.manifest.v1+json and unpacks its layered filesystem to `dest`.

# OPTIONS
**--help**
  Print usage statement

**--ref**
  The ref pointing to the manifest to be unpacked. This must be present in the "refs" subdirectory of the image. (default "v1.0")

**--same-owner**
  Preserve the owner and group of the layer entries when unpacking the image (default for superuser, but not for ordinary users).

**--type**
  Type of the file to unpack. If unset, oci-unpack will try to auto-detect the type. One of "imageLayout,image"

# EXAMPLES
```
$ skopeo copy docker://busybox oci:busybox-oci
$ mkdir busybox-bundle
$ oci-unpack --ref latest busybox-oci busybox-bundle
tree busybox-bundle
busybox-bundle
├── bin
│   ├── [
│   ├── [[
│   ├── acpid
│   ├── addgroup
│   ├── add-shell
│   ├── adduser
│   ├── adjtimex
│   ├── ar
│   ├── arp
│   ├── arping
│   ├── ash
[...]
```

# SEE ALSO
**skopeo**(1)

# HISTORY
Sept 2016, Originally compiled by Antonio Murdaca (runcom at redhat dot com)
