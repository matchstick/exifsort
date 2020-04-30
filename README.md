# exifSort
Program to sort photos and movies via Exif

Make sure to discuss concurrency and file server is not fun.

Inputs:
exifSort 
 provide help and list

exifSort -> scan <in dir> --summary
 * walk a directory report of exif state and number of files

exifSort -> sort <in dir> <out dir>  <log file>
 * walk a directory and then process files to a target directory of nested 

exifSort -> eval <file>
