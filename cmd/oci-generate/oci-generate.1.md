% OCI(1) OCI-UNPACK User Manuals
% OCI Community
% JULY 2016
# NAME
oci-generate \- Generate an OCI image or an OCI imageLayout

# SYNOPSIS
**oci-unpack** [dest] [flags]
**oci-unpack** [--help|--version]

# DESCRIPTION
`oci-generate` generate an application/vnd.oci.image.manifest.v1+json or application/vnd.oci.image.manifest.list+json to `dest`.

# OPTIONS
**--help**
  Print usage statement

**--type**
  Type of the file to generate.One of "imageLayout,image".(detault "imageLayout")

**--version**
  Print version information and exit.

# EXAMPLES
```
$ oci-generate example
$ tree example
example
├── blobs
│   └── sha256
│       ├── 5d82e6cbf19cd18f40301df14ffbd62ba1acae8a097e75011ec603ff38a7450a
│       ├── 89b46f9224fbcedd165dd6b317329b69182471c6706993bedc2c2d2fd2d76fba
│       └── a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4
├── oci-layout
└── refs
    └── latest
[...]
```
# HISTORY
Sept 2016, Originally compiled by Antonio Murdaca (runcom at redhat dot com)
