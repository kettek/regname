# Regular Expression Batch File Renaming
This program provides a simple popup dialog-driven [regular expression](https://pkg.go.dev/regexp) based file search and match that uses [fmt](https://pkg.go.dev/fmt) to rename the files.

This means that if you provide a regular expression that contain capture groups, you can rearrange them using fmt's argument indexes. For example:

Regular Expression
: `(\d+)\s(\d+)`

Format
: `%[2]s %[1]s`

With the filename input of `0 100`, it will rename it to `100 0`.

