# exifSort

This README is Under construction but for now here is some notes.

TODOs include:
* Clean up tests now that APIs are tighter
* Clean up doc.go in lib now that APIS are changed.
* Figure out git hooks
* Write more complete sort tests.
* Update this readme.

Program to sort and scan photos and movies for their Creation date using their Exif info.

Huge thanks to dsoprea for his exif library and fast responses.

Inputs:

exifSort -> scan <in dir> --summarize --quiet
 * walk a directory report of exif state and number of files

exifSort -> sort <in dir> <out dir> <year | month | day> 
 * walk a directory and then process files to a target directory of nested 

exifSort -> eval <file>
 * Prints the date information of one file specified. TODO needs glob support.
 TODO (check for duplciate files, first checktimestamp, then check contents)
