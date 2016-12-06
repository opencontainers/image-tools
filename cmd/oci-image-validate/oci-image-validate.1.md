% OCI(1) OCI-IMAGE-VALIDATE User Manuals
% OCI Community
% JULY 2016
# NAME
oci-image-validate \- Validate one or more image files

# SYNOPSIS
**oci-image-validate** FILE... [flags]
**oci-image-validate** [--help|-v|--version]

# DESCRIPTION
`oci-image-validate` validates the given file(s) against the OCI image specification.


# OPTIONS
**--help**
  Print usage statement

**--ref**=[]
  The reference to validate (should point to a manifest).
  Can be specified multiple times to validate multiple references.
  `NAME` must be present in the `refs` subdirectory of the image.
  Defaults to `v1.0`.
  Only applicable if type is image or imageLayout.

**--type**=""
  Type of the file to validate. If unset, oci-image-validate will try to auto-detect the type. One of "imageLayout,image,manifest,manifestList,config"

**-v**, **--version**
  Print version information and exit.

# EXAMPLES
```
$ skopeo copy docker://busybox oci:busybox-oci
$ oci-image-validate --type imageLayout --ref latest busybox-oci
busybox-oci: OK
```

# SEE ALSO
**skopeo**(1)

# HISTORY
Sept 2016, Originally compiled by Antonio Murdaca (runcom at redhat dot com)
