% OCI-IMAGE-TOOL-VALIDATE(1) OCI Image Tool User Manuals
% OCI Community
% JULY 2016
# NAME
oci-image-tool validate \- Validate one or more image files

# SYNOPSIS
**oci-image-tool validate** FILE... [OPTIONS]

# DESCRIPTION
`oci-image-tool validate` validates the given file(s) against the OCI image specification.


# OPTIONS
**--help**
  Print usage statement

**--ref**=[]
  The reference to validate (should point to a manifest).
  Can be specified multiple times to validate multiple references.
  `NAME` must be present in the `refs` subdirectory of the image.
  Only applicable if type is image or imageLayout.

**--type**=""
  Type of the file to validate. If unset, oci-image-tool will try to auto-detect the type. One of "imageLayout,image,imageZip,manifest,imageIndex,config"

# EXAMPLES
```
$ skopeo copy docker://busybox oci:busybox-oci:latest
$ oci-image-tool validate --type imageLayout --ref latest busybox-oci
busybox-oci: OK
```

# SEE ALSO
**skopeo**(1)

# HISTORY
Sept 2016, Originally compiled by Antonio Murdaca (runcom at redhat dot com)
