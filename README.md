# exifsort

This README and program is Under construction but for now here is some notes.

TODOs include:
* Clean up tests now that APIs are tighter
* Clean up doc.go in lib now that APIS are changed.
* Redo the cobra pieces so that lint does not complain.
* Write more complete tests. coverage is not high enough and for sort it is 0%.
* Update this readme.
* Transfer invalid phnotos to an unsorted directoryin src
* Write a merge step to merge two sorted directories.
* Set up CI on github.

Main motivation for this was to get my hands dirty writing golang again.
Program to sort and scan photos and movies for their Creation date using their
Exif info.  The main motivation is to have a program that can handle all the
100K of photos nestled in the terabyes of hard drives I have been finding in my
closet. 

The program is written to use several stages to let the user verify the steps.
as these are photos it cannot hurt to be careful.

Huge thanks to dsoprea for his exif library and fast responses.

Inputs:

exifsort -> scan <in dir> --summarize --quiet
 * walk a directory report of exif state and number of files

exifsort -> sort <in dir> <out dir> <year | month | day> 
 * walk a directory and then process files to a target directory of nested 

exifsort -> eval <file>
 * Prints the date information of one file specified. TODO needs glob support.
 TODO (check for duplciate files, first checktimestamp, then check contents)

exifsort -> merge is coming
