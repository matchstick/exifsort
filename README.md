# exifsort

![Under Construction](data/construction.jpg) 

Program are both under construction but for now here are some notes.

exifsort
========
Libraries and CLI to sort to sort media by date using EXIF information.

This is for folks who have a closet full of hard drives and network drives full
of photos and want to centralize them in one folder structure.

The functionality and API live in the lib directory. Check out
[godocs](https://godoc.org/github.com/matchstick/exifsort/lib) for details.

Overview
========
The program is written to employ several stages to let the user verify the
step results as they organize their photos. It cannot hurt to be careful.

Huge thanks to [dsoprea](https://github.com/dsoprea) for his [exif library](https://github.com/dsoprea/go-exif) and fast responses.

Commands
========

`exifsort -> scan <srcDir> --summarize --quiet`
 * walk a directory report of exif state and number of files. Useful to test that exif library will be fine.

`exifsort -> sort <srcDir> <dstDir> <year | month | day>  <copy | move>`
 * walk a directory and then transfer files to a target directory of nested 

`exifsort -> merge is coming`
* Take a dstDir then merge it with a pre-existing one.

`exifsort -> eval <files>`
 * Prints the date information of files specified. 

TODOs include:
* Add movie formats to suffix
* Sort file formats?
* concurrency
* marshal to json after scan, load sort from json
* Clean up tests now that APIs are tighter
* Clean up doc.go in lib now that APIS are changed.
* Write more complete tests. coverage is not high enough and for sort it is 0%.
* Transfer invalid photos to an unsorted directory
* Write a merge step to merge two sorted directories.
* Set up CI on github.


