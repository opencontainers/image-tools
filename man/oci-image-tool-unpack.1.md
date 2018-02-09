% OCI-IMAGE-TOOL-UNPACK(1) OCI Image Tool User Manuals
% OCI Community
% JULY 2016
# NAME
oci-image-tool unpack \- Unpack an image or image source layout

# SYNOPSIS
**oci-image-tool unpack** [src] [dest] [OPTIONS]

# DESCRIPTION
`oci-image-tool unpack` validates an application/vnd.oci.image.manifest.v1+json and unpacks its layered filesystem to `dest`.

# OPTIONS
**--help**
  Print usage statement

**--select**=[]
  Select the search criteria for the validated reference, format is A=B.
  Reference should point to a manifest or index.
  e.g. --select org.opencontainers.ref.name=v1.0 --select platform.os=latest
  Only support `org.opencontainers.ref.name`, `platform.os` and `digest` three cases.

**--type**=""
  Type of the file to unpack. If unset, oci-image-tool will try to auto-detect the type. One of "imageLayout,image,imageZip"

**--platform**=""
  Specify the os and architecture of the manifest, format is OS:Architecture.
  e.g. --platform linux:amd64
  Only applicable if reftype is index.

# EXAMPLES
```
$ skopeo copy docker://busybox oci:busybox-oci:latest
$ mkdir busybox-bundle
$ oci-image-tool unpack --select org.opencontainers.ref.name=latest busybox-oci busybox-bundle
$ tree busybox-bundle
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
