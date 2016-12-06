% OCI(1) OCI-IMAGE-TOOL User Manuals
% OCI Community
% JULY 2016
# NAME
oci-image-tool \- OCI (Open Container Initiative) image tools

# SYNOPSIS
**oci-image-tool** [OPTIONS] COMMAND [arg...]

**oci-image-tool** [--help|-v|--version]

# DESCRIPTION
oci-image-tool is a collection of tools for working with the [OCI image specification](https://github.com/opencontainers/image-spec).


# OPTIONS
**--help**
  Print usage statement.

**--debug**
  Enable debug output

**-v**, **--version**
  Print version information.

# COMMANDS
**validate**
  Validate the given file(s) against the OCI image specification
  See **oci-image-tool-validate**(1) for full documentation on the **validate** command.

**unpack**
  Unpack the file which against the OCI image specification into a bundle directory.
  See **oci-image-tool-unpack**(1) for full documentation on the **unpack** command.

**create**
  Create an OCI runtime bundle
  See **oci-image-tool-create**(1) for full documentation on the **create** command.

# SEE ALSO
**oci-image-tool-validate**(1), **oci-image-tool-unpack**(1), **oci-image-tool-create**(1)

# HISTORY
Sept 2016, Originally compiled by Antonio Murdaca (runcom at redhat dot com)
