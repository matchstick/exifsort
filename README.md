# exifsort

![Under Construction](data/construction.jpg) 

Program are both under construction but for now here are some notes.

# exifsort

Libraries and CLI to sort to sort media by date using EXIF information.

This is for folks who have a closet full of hard drives and network drives full
of photos and want to centralize them in one folder structure.

The functionality and API live in the lib directory. Check out
[godocs](https://godoc.org/github.com/matchstick/exifsort/lib) for details.

# Overview

The program is written to employ several stages to let the user verify the
step results as they organize their photos. It cannot hurt to be careful.

Huge thanks to [dsoprea](https://github.com/dsoprea) for his [exif
library](https://github.com/dsoprea/go-exif) and fast responses.

# Commands

## scan

Scanning is when exifsort will read the data from the directory of files,
filter for media and retrieve time. Useful to test that exif library will be fine.

`exifsort scan -input <dir> --summarize --quiet`

## sort

Walk an input directory, index the data and then transfer files to an output  directory.

`exifsort sort -input <dir> -output <dir> -method <year | month | day> -action <copy | move>`

### Methods

| Method | Structure |
| ------ | --------- |
| Year   | dst -> year -> media |
| Month  | dst -> year-> month -> media |
| Day    | dst -> year-> month -> day -> media |

### Actions

An action specifies whether to move or copy the files from input to output 

## eval

Just useful for debugging and looking at files. Prints the date information of files specified. 

`exifsort eval <files>`

## merge

Soon we will be able to merge output from sort to a pre-existing directory.

TODOs
=====
* Sort out file formats?
* concurrency
* sort tests are not started.
* Transfer invalid photos to an unsorted directory
* Write a merge step to merge two sorted directories.
* Set up CI on github.
