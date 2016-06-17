% OCI(1) OCI-IMAGE-VALIDATE User Manuals
% OCI Community
% JULY 2016
# NAME
oci-image-validate \- Validate one or more image files

# SYNOPSIS
**oci-image-validate** FILE... [flags]

# DESCRIPTION
`oci-image-validate` validates the given file(s) against the OCI image specification.


# OPTIONS
**--help**
  Print usage statement

**--ref** NAME
  The reference to validate (should point to a manifest).
  Can be specified multiple times to validate multiple references.
  `NAME` must be present in the `refs` subdirectory of the image.
  Defaults to `v1.0`.
  Only applicable if type is image.

**--type**
  Type of the file to validate. If unset, oci-image-validate will try to auto-detect the type. One of "image,manifest,manifestList,config"

# EXAMPLES
```
$ skopeo copy docker://busybox oci:busybox-oci
$ oci-image-validate --type image --ref latest busybox-oci
busybox-oci: OK
```

# SEE ALSO
**skopeo**(1)

# HISTORY
Sept 2016, Originally compiled by Antonio Murdaca (runcom at redhat dot com)
