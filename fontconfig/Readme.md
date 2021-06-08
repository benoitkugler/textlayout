# fontconfig for Golang

This package is a port from the C [fontconfig](https://gitlab.freedesktop.org/fontconfig/fontconfig) library.

Its main purpose is to fetch the metadata of the fonts installed in a system, and use that database to select the best font to use given some user-specified query (name, style, weight, etc...).

## Differences from the original C library

While the main fonctionnality of the original library are preserved, some simplifications have been made (mainly to avoid complex logic of file system handling).

### Caching

The package drops support for advanced caching: it is deferred to the users. They can use the provided `Serialize` and `LoadFontset` functions, but its up to them to specified what to cache, when and where.

### Configuration build

The main way to specify complex configurations remains the XML fontconfig format. However, it is not possible to use `<include>` directives. Several config files are simply added one by one.

### Font directories

The XML format does not support specifying font directories. Instead, scans are explicitely triggered by the user, which provide a file (`ScanFontFile`), an in-memory content (`ScanFontRessource`) or a list of directories (`ScanFontDirectories`).

## Dependencies

This is a pure Go implementation, which rely on [fonts](github.com/benoitkugler/fonts) as a substitute of FreeType to handle the scanning of a font file.
