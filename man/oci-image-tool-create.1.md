% OCI-IMAGE-TOOL-CREATE(1) OCI Image Tool User Manuals
% OCI Community
% JULY 2016
# NAME
oci-image-tool create \- Create an OCI runtime bundle

# SYNOPSIS
**oci-image-tool create** [src] [dest] [OPTIONS]

# DESCRIPTION
`oci-image-tool create` validates an application/vnd.oci.image.manifest.v1+json and unpacks its layered filesystem to `dest/rootfs`, although the target directory is configurable with `--rootfs`. See **oci-image-tool unpack**(1) for more details on this process.

Also translates the referenced config from application/vnd.oci.image.config.v1+json to a
runtime-spec-compatible `dest/config.json`.

# OPTIONS
**--help**
  Print usage statement

**--ref**=[]
  Specify the search criteria for the validated reference, format is A=B.
  Reference should point to a manifest or index.
  e.g. --ref name=v1.0 --ref platform.os=latest
  Only support `name`, `platform.os` and `digest` three cases.

**--rootfs**=""
  A directory representing the root filesystem of the container in the OCI runtime bundle. It is strongly recommended to keep the default value. (default "rootfs")

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
$ oci-image-tool create --ref name=latest busybox-oci busybox-bundle
$ cd busybox-bundle && sudo runc run busybox
[...]
```

# SEE ALSO
**runc**(1), **skopeo**(1)

# HISTORY
Sept 2016, Originally compiled by Antonio Murdaca (runcom at redhat dot com)
